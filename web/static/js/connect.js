// connect.js — Wires the prototype UI to the MaSante REST API.
// This script overrides the dummy handlers in the prototype to make real API calls.
import { setup, auth, dashboard, patients, appointments, calendar, reminders, users, profile } from './api.js';

// ── Toast notification ──

function toast(msg, type) {
  let c = document.getElementById('ms-toast');
  if (!c) {
    c = document.createElement('div');
    c.id = 'ms-toast';
    c.style.cssText = 'position:fixed;top:20px;right:20px;z-index:99999;display:flex;flex-direction:column;gap:8px';
    document.body.appendChild(c);
  }
  const t = document.createElement('div');
  t.textContent = msg;
  const colors = { success: '#2d7a4f', error: '#a63d3d', warning: '#b8860b', info: '#3d6b8a' };
  t.style.cssText = 'padding:12px 20px;border-radius:8px;font-size:.85rem;font-weight:500;color:#fff;box-shadow:0 8px 30px rgba(0,0,0,.12);transform:translateX(120%);transition:transform .3s;max-width:400px;background:' + (colors[type] || colors.info);
  c.appendChild(t);
  requestAnimationFrame(() => t.style.transform = 'translateX(0)');
  setTimeout(() => { t.style.transform = 'translateX(120%)'; setTimeout(() => t.remove(), 300); }, 4000);
}

// ── Boot: check setup status ──

async function boot() {
  const res = await setup.status();
  if (res.ok && res.data.setup_complete) {
    // Setup done — hide wizard, check auth.
    document.getElementById('setupScreen').classList.add('hidden');
    const me = await auth.me();
    if (me.ok) {
      document.getElementById('loginScreen').classList.add('hidden');
      document.getElementById('app').classList.add('visible');
      loadDashboard();
    } else {
      document.getElementById('loginScreen').classList.remove('hidden');
    }
  }
  // If setup not done, wizard is already visible (default state).
}

// ── Setup Wizard ──

const wizData = {};
const origWizStep = window.wizStep;

window.wizStep = async function(dir) {
  const wizCur = window.wizCur || 1;

  if (dir === 1) {
    const err = await saveWizStep(wizCur);
    if (err) {
      toast(err, 'error');
      return;
    }
  }

  // Call original navigation logic.
  if (typeof origWizStep === 'function') {
    origWizStep(dir);
  }
};

async function saveWizStep(step) {
  let res;
  switch (step) {
    case 1: {
      const nameEl = document.getElementById('setupName');
      const country = document.querySelector('#wiz-1 select');
      const textInputs = document.querySelectorAll('#wiz-1 input[type="text"]');
      const typeOpt = document.querySelector('#wiz-1 .type-opt.on');
      const typeMap = { 'Hopital public': 'hopital_public', 'Centre de sante': 'centre_sante', 'Clinique privee': 'clinique_privee' };
      const data = {
        name: nameEl ? nameEl.value : '',
        type: typeOpt ? (typeMap[typeOpt.textContent] || 'centre_sante') : 'centre_sante',
        country: country ? country.value : '',
        city: textInputs[1] ? textInputs[1].value : '',
        district: textInputs[2] ? textInputs[2].value : '',
      };
      if (textInputs[3] && textInputs[3].value) data.lat = parseFloat(textInputs[3].value);
      if (textInputs[4] && textInputs[4].value) data.lng = parseFloat(textInputs[4].value);
      if (!data.name || !data.country || !data.city) return 'Nom, pays et ville requis';
      wizData.center = data;
      res = await setup.center(data);
      break;
    }
    case 2: {
      const inputs = document.querySelectorAll('#wiz-2 .form-input');
      const data = {
        full_name: inputs[0] ? inputs[0].value : '',
        email: inputs[1] ? inputs[1].value : '',
        username: inputs[2] ? inputs[2].value : '',
        title: inputs[3] ? inputs[3].value : '',
        password: inputs[4] ? inputs[4].value : '',
      };
      const confirm = inputs[5] ? inputs[5].value : '';
      if (!data.full_name || !data.username || !data.password) return 'Nom, identifiant et mot de passe requis';
      if (data.password !== confirm) return 'Les mots de passe ne correspondent pas';
      wizData.admin = data;
      res = await setup.admin(data);
      break;
    }
    case 3: {
      const days = [];
      document.querySelectorAll('#wiz-3 .day-check.on').forEach((el, i) => days.push(i + 1));
      const timeInputs = document.querySelectorAll('#wiz-3 input[type="time"]');
      const selects = document.querySelectorAll('#wiz-3 select');
      const numInputs = document.querySelectorAll('#wiz-3 input[type="number"]');
      const data = {
        consultation_days: days.join(','),
        start_time: timeInputs[0] ? timeInputs[0].value : '08:00',
        end_time: timeInputs[1] ? timeInputs[1].value : '16:00',
        slot_duration: selects[0] ? parseInt(selects[0].value) : 30,
        max_patients_day: numInputs[0] ? parseInt(numInputs[0].value) : 40,
      };
      res = await setup.schedule(data);
      break;
    }
    case 4: {
      const enableOpt = document.querySelector('#wiz-4 .type-opt.on');
      const enabled = enableOpt && enableOpt.textContent.indexOf('Oui') >= 0;
      const providerOpt = document.querySelector('#wiz-4 .sms-provider.selected .sp-name');
      const provMap = { "Africa's Talking": 'africastalking', 'MTN SMS API': 'mtn', 'Orange SMS API': 'orange', 'Twilio': 'twilio', 'Infobip': 'infobip' };
      const inputs = document.querySelectorAll('#smsConfigBlock .form-input');
      const data = {
        enabled: enabled,
        provider: enabled && providerOpt ? (provMap[providerOpt.textContent] || '') : '',
        api_key: inputs[0] ? inputs[0].value : '',
        api_secret: inputs[1] ? inputs[1].value : '',
        sender_id: inputs[2] ? inputs[2].value : '',
      };
      res = await setup.sms(data);
      break;
    }
    case 5: {
      res = await setup.complete();
      if (res.ok) {
        toast('Configuration terminee ! Connectez-vous.', 'success');
        setTimeout(() => {
          document.getElementById('setupScreen').classList.add('hidden');
          document.getElementById('loginScreen').classList.remove('hidden');
        }, 500);
        return null;
      }
      break;
    }
    default:
      return null;
  }
  if (res && !res.ok) return res.error;
  return null;
}

// ── Login ──

const origDoLogin = window.doLogin;

window.doLogin = async function() {
  const inputs = document.querySelectorAll('.login-right .form-input');
  const username = inputs[0] ? inputs[0].value.trim() : '';
  const password = inputs[1] ? inputs[1].value : '';

  if (!username || !password) {
    toast('Identifiant et mot de passe requis', 'error');
    return;
  }

  const res = await auth.login(username, password);
  if (!res.ok) {
    toast(res.error, 'error');
    return;
  }

  toast('Connexion reussie', 'success');
  document.getElementById('loginScreen').classList.add('hidden');
  setTimeout(() => {
    document.getElementById('app').classList.add('visible');
    loadDashboard();
  }, 200);
};

// ── Logout ──

const origDoLogout = window.doLogout;

window.doLogout = async function() {
  await auth.logout();
  toast('Deconnecte', 'info');
  document.getElementById('app').classList.remove('visible');
  setTimeout(() => {
    document.getElementById('loginScreen').classList.remove('hidden');
  }, 300);
};

// ── Dashboard: load real data ──

async function loadDashboard() {
  const [statsRes, todayRes, overdueRes] = await Promise.all([
    dashboard.stats(),
    dashboard.today(),
    dashboard.overdue(),
  ]);

  if (statsRes.ok) {
    const p = statsRes.data.patients || {};
    const a = statsRes.data.appointments || {};
    const total = Object.values(p).reduce((s, v) => s + v, 0);
    const todayCount = Object.values(a).reduce((s, v) => s + v, 0);
    const lost = p['perdu_de_vue'] || 0;

    // Update stat cards.
    const statVals = document.querySelectorAll('.stat-val');
    if (statVals[0]) statVals[0].textContent = total;
    if (statVals[1]) statVals[1].textContent = todayCount;
    // Retention rate — approximate.
    const retention = total > 0 ? Math.round((1 - lost / total) * 100) : 0;
    if (statVals[2]) statVals[2].textContent = retention + '%';
    if (statVals[3]) statVals[3].textContent = lost;
  }

  // Load today's appointments into the list.
  if (todayRes.ok && todayRes.data) {
    updateTodayList(todayRes.data);
  }

  // Load overdue into the overdue list.
  if (overdueRes.ok && overdueRes.data) {
    updateOverdueList(overdueRes.data);
  }
}

function updateTodayList(apts) {
  // Find the RDV du jour card body.
  const cards = document.querySelectorAll('#s-dashboard .card-body');
  if (!cards[0]) return;
  const body = cards[0];
  body.textContent = '';

  if (apts.length === 0) {
    const empty = document.createElement('div');
    empty.style.cssText = 'text-align:center;padding:20px;color:var(--gray-300)';
    empty.textContent = "Aucun rendez-vous aujourd'hui";
    body.appendChild(empty);
    return;
  }

  apts.forEach(apt => {
    const item = document.createElement('div');
    item.className = 'list-item';

    const time = document.createElement('span');
    time.className = 'time-label';
    time.textContent = apt.Time || '—';
    item.appendChild(time);

    const initials = (apt.PatientName || '??').split(' ').map(w => w[0]).join('').substring(0, 2);
    const av = document.createElement('div');
    av.className = 'avatar a1';
    av.textContent = initials;
    item.appendChild(av);

    const info = document.createElement('div');
    info.className = 'item-info';
    const name = document.createElement('div');
    name.className = 'item-name';
    name.textContent = apt.PatientName || 'Patient';
    info.appendChild(name);
    const typeLbl = document.createElement('div');
    typeLbl.className = 'item-sub';
    const typeMap = { consultation: 'Consultation', retrait_medicaments: 'Retrait medicaments', bilan_sanguin: 'Bilan sanguin', club_adherence: "Club d'adherence" };
    typeLbl.textContent = typeMap[apt.Type] || apt.Type || '';
    info.appendChild(typeLbl);
    item.appendChild(info);

    const statusMap = { confirme: 'pill-success', en_attente: 'pill-warning', manque: 'pill-danger', termine: 'pill-info' };
    const pill = document.createElement('span');
    pill.className = 'pill ' + (statusMap[apt.Status] || 'pill-neutral');
    pill.textContent = apt.Status || '';
    item.appendChild(pill);

    body.appendChild(item);
  });
}

function updateOverdueList(apts) {
  const cards = document.querySelectorAll('#s-dashboard .card-body');
  if (!cards[1]) return;
  const body = cards[1];
  body.textContent = '';

  if (apts.length === 0) {
    const empty = document.createElement('div');
    empty.style.cssText = 'text-align:center;padding:20px;color:var(--gray-300)';
    empty.textContent = 'Aucun patient en retard';
    body.appendChild(empty);
    return;
  }

  apts.forEach(apt => {
    const item = document.createElement('div');
    item.className = 'list-item';

    const initials = (apt.PatientName || '??').split(' ').map(w => w[0]).join('').substring(0, 2);
    const av = document.createElement('div');
    av.className = 'overdue-avatar';
    av.textContent = initials;
    item.appendChild(av);

    const info = document.createElement('div');
    info.className = 'item-info';
    const name = document.createElement('div');
    name.className = 'item-name';
    name.textContent = apt.PatientName || 'Patient';
    info.appendChild(name);
    const days = document.createElement('div');
    days.className = 'overdue-days';
    const daysLate = Math.max(1, Math.floor((Date.now() - new Date(apt.Date).getTime()) / 86400000));
    days.textContent = daysLate + ' jours de retard';
    info.appendChild(days);
    item.appendChild(info);

    const btn = document.createElement('button');
    btn.className = 'btn btn-sm btn-secondary';
    btn.textContent = 'Reprogrammer';
    btn.style.cssText = 'width:auto;flex-shrink:0';
    btn.onclick = () => go('new-apt', document.querySelectorAll('.sb-item')[3]);
    item.appendChild(btn);

    body.appendChild(item);
  });
}

// ── Navigate: reload data when switching screens ──

const origGo = window.go;
window.go = function(id, navEl) {
  if (typeof origGo === 'function') origGo(id, navEl);

  // Load data for the target screen.
  switch (id) {
    case 'dashboard': loadDashboard(); break;
    case 'patients': loadPatientList(); break;
    case 'users': loadUserList(); break;
    case 'reminders': loadReminders(); break;
  }
};

// ── Patients list ──

async function loadPatientList() {
  const res = await patients.list('page=1&per_page=50');
  if (!res.ok) { toast(res.error, 'error'); return; }

  const table = document.querySelector('#s-patients .ptable');
  if (!table) return;

  // Remove existing rows (keep header).
  const rows = table.querySelectorAll('.pt-row');
  rows.forEach(r => r.remove());

  const list = res.data.patients || [];
  if (list.length === 0) {
    const empty = document.createElement('div');
    empty.style.cssText = 'text-align:center;padding:20px;color:var(--gray-300)';
    empty.textContent = 'Aucun patient enregistre';
    table.appendChild(empty);
    return;
  }

  const avatarColors = ['a1', 'a2', 'a3', 'a4', 'a5', 'a6'];
  list.forEach((p, i) => {
    const row = document.createElement('div');
    row.className = 'pt-row';
    row.style.cssText = 'grid-template-columns:44px 2fr 1fr 1fr 1fr 1fr 100px;cursor:pointer';

    // Avatar.
    const avCell = document.createElement('div');
    const av = document.createElement('div');
    av.className = 'avatar ' + avatarColors[i % avatarColors.length];
    av.textContent = ((p.LastName || '')[0] || '') + ((p.FirstName || '')[0] || '');
    avCell.appendChild(av);
    row.appendChild(avCell);

    // Name.
    const nameCell = document.createElement('div');
    const nameDiv = document.createElement('div');
    nameDiv.className = 'pt-name';
    nameDiv.textContent = p.LastName + ' ' + p.FirstName;
    nameCell.appendChild(nameDiv);
    const codeDiv = document.createElement('div');
    codeDiv.className = 'pt-code';
    codeDiv.textContent = p.Code;
    nameCell.appendChild(codeDiv);
    row.appendChild(nameCell);

    // Risk.
    const riskCell = document.createElement('div');
    const risk = document.createElement('span');
    const score = p.RiskScore || 5;
    risk.className = 'risk ' + (score <= 3 ? 'low' : score <= 6 ? 'med' : 'high');
    const dot = document.createElement('span');
    dot.className = 'risk-dot';
    risk.appendChild(dot);
    risk.appendChild(document.createTextNode(' ' + (score <= 3 ? 'Faible' : score <= 6 ? 'Moyen' : 'Eleve')));
    riskCell.appendChild(risk);
    row.appendChild(riskCell);

    // Last visit.
    const lastCell = document.createElement('div');
    lastCell.className = 'pt-cell';
    lastCell.textContent = p.UpdatedAt ? new Date(p.UpdatedAt).toLocaleDateString('fr-FR') : '—';
    row.appendChild(lastCell);

    // Next apt — placeholder.
    const nextCell = document.createElement('div');
    nextCell.className = 'pt-cell';
    nextCell.textContent = '—';
    row.appendChild(nextCell);

    // Status.
    const statusCell = document.createElement('div');
    const statusMap = { active: ['pill-success', 'Actif'], a_surveiller: ['pill-warning', 'A surveiller'], perdu_de_vue: ['pill-danger', 'Perdu de vue'], sorti: ['pill-neutral', 'Sorti'] };
    const [pillCls, pillText] = statusMap[p.Status] || ['pill-neutral', p.Status];
    const pill = document.createElement('span');
    pill.className = 'pill ' + pillCls;
    pill.textContent = pillText;
    statusCell.appendChild(pill);
    row.appendChild(statusCell);

    // Actions.
    const actCell = document.createElement('div');
    actCell.className = 'pt-acts';
    row.appendChild(actCell);

    row.onclick = () => {
      // Store patient ID and navigate to file.
      window._currentPatientId = p.ID;
      go('patient-file', document.querySelectorAll('.sb-item')[5]);
    };

    table.appendChild(row);
  });
}

// ── Users list ──

async function loadUserList() {
  const res = await users.list();
  if (!res.ok) return;

  const table = document.querySelector('#s-users .ptable');
  if (!table) return;

  const rows = table.querySelectorAll('.pt-row');
  rows.forEach(r => r.remove());

  (res.data || []).forEach(u => {
    const row = document.createElement('div');
    row.className = 'pt-row';
    row.style.cssText = 'grid-template-columns:44px 2fr 1fr 1fr 1fr 120px';

    const avCell = document.createElement('div');
    const av = document.createElement('div');
    av.className = 'avatar a1';
    av.textContent = (u.full_name || '??').split(' ').map(w => w[0]).join('').substring(0, 2);
    avCell.appendChild(av);
    row.appendChild(avCell);

    const nameCell = document.createElement('div');
    const nameDiv = document.createElement('div');
    nameDiv.className = 'pt-name';
    nameDiv.textContent = u.full_name;
    nameCell.appendChild(nameDiv);
    const codeDiv = document.createElement('div');
    codeDiv.className = 'pt-code';
    codeDiv.textContent = u.username + (u.email ? ' — ' + u.email : '');
    nameCell.appendChild(codeDiv);
    row.appendChild(nameCell);

    const roleCell = document.createElement('div');
    const roleMap = { admin: 'pill-danger', medecin: 'pill-info', infirmier: 'pill-warning', asc: 'pill-success' };
    const rp = document.createElement('span');
    rp.className = 'pill ' + (roleMap[u.role] || 'pill-neutral');
    rp.textContent = u.role;
    roleCell.appendChild(rp);
    row.appendChild(roleCell);

    const loginCell = document.createElement('div');
    loginCell.className = 'pt-cell';
    loginCell.textContent = '—';
    row.appendChild(loginCell);

    const statusCell = document.createElement('div');
    const sMap = { active: ['pill-success', 'Actif'], conge: ['pill-neutral', 'Conge'], desactive: ['pill-danger', 'Desactive'] };
    const [sc, sl] = sMap[u.status] || ['pill-neutral', u.status];
    const sp = document.createElement('span');
    sp.className = 'pill ' + sc;
    sp.textContent = sl;
    statusCell.appendChild(sp);
    row.appendChild(statusCell);

    const actCell = document.createElement('div');
    actCell.className = 'pt-acts';
    row.appendChild(actCell);

    table.appendChild(row);
  });
}

// ── Reminders ──

async function loadReminders() {
  const [statsRes, queueRes] = await Promise.all([reminders.stats(), reminders.list()]);

  if (statsRes.ok) {
    const s = statsRes.data;
    const vals = document.querySelectorAll('#s-reminders .rem-stat-val');
    if (vals[0]) vals[0].textContent = (s.DeliveryRate || 0).toFixed(1) + '%';
    if (vals[1]) vals[1].textContent = '—';
    if (vals[2]) vals[2].textContent = s.PendingCount || 0;
    if (vals[3]) vals[3].textContent = s.FailedCount || 0;
  }
}

// ── New patient form ──

// Find the submit button in new-patient screen and wire it.
setTimeout(() => {
  const btn = document.querySelector('#s-new-patient .btn-primary');
  if (btn) {
    btn.onclick = async (e) => {
      e.preventDefault();
      const inputs = document.querySelectorAll('#s-new-patient .form-input');
      const langOpt = document.querySelector('#s-new-patient .r-opt.on');
      const langMap = { 'Francais': 'fr', 'Anglais': 'en', 'Duala': 'duala', 'Ewondo': 'ewondo', 'Bamileke': 'bamileke' };
      const channelOpts = document.querySelectorAll('#s-new-patient .remind-opts');
      let channel = 'sms';
      if (channelOpts[1]) {
        const chOpt = channelOpts[1].querySelector('.r-opt.on');
        if (chOpt) {
          const chMap = { 'SMS': 'sms', 'WhatsApp': 'whatsapp', 'Appel vocal': 'voice', 'Aucun': 'none' };
          channel = chMap[chOpt.textContent.trim()] || 'sms';
        }
      }

      const body = {
        last_name: inputs[0] ? inputs[0].value : '',
        first_name: inputs[1] ? inputs[1].value : '',
        date_of_birth: inputs[2] ? inputs[2].value : '',
        sex: inputs[3] ? inputs[3].value : '',
        phone: inputs[4] ? inputs[4].value : '',
        phone_secondary: inputs[5] ? inputs[5].value : '',
        district: inputs[6] ? inputs[6].value : '',
        address: inputs[7] ? inputs[7].value : '',
        language: langOpt ? (langMap[langOpt.textContent] || 'fr') : 'fr',
        reminder_channel: channel,
        contact_name: inputs[8] ? inputs[8].value : '',
        contact_phone: inputs[9] ? inputs[9].value : '',
        contact_relation: inputs[10] ? inputs[10].value : '',
        referred_by: inputs[11] ? inputs[11].value : '',
      };

      if (!body.last_name || !body.first_name || !body.sex) {
        toast('Nom, prenom et sexe requis', 'error');
        return;
      }

      const res = await patients.create(body);
      if (!res.ok) {
        toast(res.error, 'error');
        return;
      }

      toast('Patient inscrit — Code: ' + (res.data.Code || ''), 'success');
      go('patients', document.querySelectorAll('.sb-item')[4]);
    };
  }
}, 500);

// ── Boot ──

boot();
