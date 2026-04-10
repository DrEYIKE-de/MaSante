// setup.js — 5-step setup wizard, connected to the API.
import { setup as setupApi } from '../api.js';
import { el, iconEl, notify } from '../ui.js';
import { setupDone } from '../app.js';

export const title = 'Configuration';

let step = 1;
const totalSteps = 5;
const data = { center: {}, admin: {}, schedule: {}, sms: {} };

export async function render(container) {
  // Check current step from server.
  const res = await setupApi.status();
  if (res.ok && res.data.setup_complete) {
    setupDone();
    return;
  }
  if (res.ok && res.data.current_step > 0) {
    step = res.data.current_step + 1;
    if (step > totalSteps) step = totalSteps;
  }

  const wrapper = el('div', { class: 'setup-screen', style: 'position:relative;height:100vh;margin:-24px;display:flex;flex-direction:column;background:var(--white)' });

  // Header.
  const header = el('div', { class: 'setup-header' });
  const logo = el('div', { class: 'sh-logo' });
  logo.appendChild(iconEl('leaf', 16));
  header.appendChild(logo);
  header.appendChild(el('h2', { text: 'Configuration de MaSante' }));
  const stepLabel = el('span', { class: 'sh-step', id: 'wiz-step-label', text: 'Etape ' + step + ' sur ' + totalSteps });
  header.appendChild(stepLabel);
  wrapper.appendChild(header);

  // Progress bar.
  const progressBg = el('div', { class: 'setup-progress' });
  const progressFill = el('div', { class: 'setup-progress-fill', id: 'wiz-progress', style: 'width:' + (step / totalSteps * 100) + '%' });
  progressBg.appendChild(progressFill);
  wrapper.appendChild(progressBg);

  // Body.
  const body = el('div', { class: 'setup-body' });
  const content = el('div', { class: 'setup-content', id: 'wiz-content' });
  body.appendChild(content);
  wrapper.appendChild(body);

  // Error.
  const errorEl = el('div', { id: 'wiz-error', style: 'display:none;padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin:0 40px 10px' });
  wrapper.appendChild(errorEl);

  // Footer.
  const footer = el('div', { class: 'setup-footer' });
  const prevBtn = el('button', { class: 'btn btn-secondary', id: 'wiz-prev', style: 'width:auto;visibility:' + (step === 1 ? 'hidden' : 'visible') });
  prevBtn.appendChild(iconEl('left', 14));
  prevBtn.appendChild(document.createTextNode(' Precedent'));
  prevBtn.onclick = () => goStep(-1, content, stepLabel, progressFill, prevBtn, nextBtn, errorEl);

  const nextBtn = el('button', { class: 'btn btn-primary', id: 'wiz-next', style: 'width:auto' });
  updateNextBtn(nextBtn);
  nextBtn.onclick = () => goStep(1, content, stepLabel, progressFill, prevBtn, nextBtn, errorEl);

  footer.appendChild(prevBtn);
  footer.appendChild(nextBtn);
  wrapper.appendChild(footer);

  container.appendChild(wrapper);
  renderStep(content);
}

function updateNextBtn(btn) {
  btn.textContent = '';
  if (step === totalSteps) {
    btn.appendChild(document.createTextNode('Lancer MaSante '));
    btn.appendChild(iconEl('check', 14));
  } else {
    btn.appendChild(document.createTextNode('Suivant '));
    btn.appendChild(iconEl('right', 14));
  }
}

async function goStep(dir, content, stepLabel, progressFill, prevBtn, nextBtn, errorEl) {
  errorEl.style.display = 'none';

  if (dir === 1) {
    // Save current step.
    const err = await saveStep();
    if (err) {
      errorEl.textContent = err;
      errorEl.style.display = 'block';
      return;
    }
  }

  step += dir;
  if (step < 1) { step = 1; return; }
  if (step > totalSteps + 1) return;

  if (step > totalSteps) {
    // Done.
    setupDone();
    return;
  }

  stepLabel.textContent = 'Etape ' + step + ' sur ' + totalSteps;
  progressFill.style.width = (step / totalSteps * 100) + '%';
  prevBtn.style.visibility = step === 1 ? 'hidden' : 'visible';
  updateNextBtn(nextBtn);
  renderStep(content);
}

async function saveStep() {
  let res;
  switch (step) {
    case 1:
      collectStep1();
      res = await setupApi.center(data.center);
      break;
    case 2:
      collectStep2();
      res = await setupApi.admin(data.admin);
      break;
    case 3:
      collectStep3();
      res = await setupApi.schedule(data.schedule);
      break;
    case 4:
      collectStep4();
      res = await setupApi.sms(data.sms);
      break;
    case 5:
      res = await setupApi.complete();
      break;
    default:
      return null;
  }
  if (!res.ok) return res.error;
  return null;
}

function collectStep1() {
  data.center = {
    name: val('s-name'),
    type: getSelectedType('s1-types') || 'centre_sante',
    country: val('s-country'),
    city: val('s-city'),
    district: val('s-district'),
  };
  const lat = val('s-lat');
  const lng = val('s-lng');
  if (lat) data.center.lat = parseFloat(lat);
  if (lng) data.center.lng = parseFloat(lng);
}

function collectStep2() {
  data.admin = {
    full_name: val('s-admin-name'),
    email: val('s-admin-email'),
    username: val('s-admin-user'),
    password: val('s-admin-pwd'),
    title: val('s-admin-title'),
  };
}

function collectStep3() {
  const days = [];
  document.querySelectorAll('.day-check.on').forEach((el, i) => days.push(i + 1));
  data.schedule = {
    consultation_days: days.join(','),
    start_time: val('s-start'),
    end_time: val('s-end'),
    slot_duration: parseInt(val('s-slot')) || 30,
    max_patients_day: parseInt(val('s-max')) || 40,
  };
}

function collectStep4() {
  const enabled = getSelectedType('s4-enable') === 'Oui, activer';
  data.sms = {
    enabled: enabled,
    provider: enabled ? getSelectedType('s4-provider') || '' : '',
    api_key: val('s-sms-key'),
    api_secret: val('s-sms-secret'),
    sender_id: val('s-sms-sender'),
  };
}

function val(id) {
  const input = document.getElementById(id);
  return input ? input.value.trim() : '';
}

function getSelectedType(groupId) {
  const group = document.getElementById(groupId);
  if (!group) return '';
  const selected = group.querySelector('.type-opt.on');
  return selected ? selected.textContent : '';
}

function renderStep(content) {
  content.textContent = '';
  const fns = [null, renderStep1, renderStep2, renderStep3, renderStep4, renderStep5];
  if (fns[step]) fns[step](content);
}

function renderStep1(c) {
  c.appendChild(el('h3', { text: 'Votre etablissement' }));
  c.appendChild(el('p', { class: 'setup-desc', text: 'Ces informations identifient votre centre de sante dans le systeme.' }));
  const grid = el('div', { class: 'setup-grid' });
  grid.appendChild(formGroup('Nom de l\'etablissement', 'text', 's-name', 'Ex: Hopital Laquintinie', true));
  grid.appendChild(typeSelector('s1-types', ['Hopital public', 'Centre de sante', 'Clinique privee'], 'Centre de sante'));
  grid.appendChild(selectGroup('Pays', 's-country', ['Cameroun', 'Cote d\'Ivoire', 'RD Congo', 'Senegal', 'Tchad', 'Gabon', 'Congo', 'Burkina Faso', 'Mali', 'Niger', 'Guinee', 'Togo', 'Benin', 'Rwanda', 'Kenya', 'Madagascar']));
  grid.appendChild(formGroup('Ville', 'text', 's-city', 'Ex: Douala'));
  grid.appendChild(formGroup('Quartier / Zone', 'text', 's-district', 'Ex: Bonanjo'));
  grid.appendChild(formGroup('Latitude (optionnel)', 'text', 's-lat', 'Ex: 4.0435'));
  grid.appendChild(formGroup('Longitude (optionnel)', 'text', 's-lng', 'Ex: 9.6948'));
  c.appendChild(grid);
}

function renderStep2(c) {
  c.appendChild(el('h3', { text: 'Compte administrateur' }));
  c.appendChild(el('p', { class: 'setup-desc', text: 'Ce sera le premier utilisateur avec tous les droits.' }));
  const grid = el('div', { class: 'setup-grid' });
  grid.appendChild(formGroup('Nom complet', 'text', 's-admin-name', 'Ex: Dr. Adele Mbarga'));
  grid.appendChild(formGroup('Email', 'email', 's-admin-email', 'Ex: adele@hopital.cm'));
  grid.appendChild(formGroup('Identifiant de connexion', 'text', 's-admin-user', 'Choisir un identifiant'));
  grid.appendChild(selectGroup('Fonction', 's-admin-title', ['Medecin referent', 'Chef de service', 'Directeur', 'Coordinateur programme', 'Infirmier(e) chef']));
  grid.appendChild(formGroup('Mot de passe', 'password', 's-admin-pwd', 'Min 8 caracteres dont 1 chiffre', true));
  grid.appendChild(formGroup('Confirmer', 'password', 's-admin-pwd2', 'Retapez le mot de passe', true));
  c.appendChild(grid);
}

function renderStep3(c) {
  c.appendChild(el('h3', { text: 'Jours et horaires de consultation' }));
  c.appendChild(el('p', { class: 'setup-desc', text: 'Definissez les creneaux disponibles. Modifiable a tout moment.' }));

  // Day checks.
  const dayGroup = el('div', { class: 'form-group' });
  dayGroup.appendChild(el('label', { text: 'Jours de consultation' }));
  const checks = el('div', { class: 'day-checks' });
  ['L', 'M', 'M', 'J', 'V', 'S', 'D'].forEach((d, i) => {
    const btn = el('div', { class: 'day-check' + (i < 5 ? ' on' : ''), text: d });
    btn.onclick = () => btn.classList.toggle('on');
    checks.appendChild(btn);
  });
  dayGroup.appendChild(checks);
  c.appendChild(dayGroup);

  const grid = el('div', { class: 'setup-grid', style: 'margin-top:16px' });
  grid.appendChild(timeGroup('Heure de debut', 's-start', '08:00'));
  grid.appendChild(timeGroup('Heure de fin', 's-end', '16:00'));
  grid.appendChild(selectGroup('Duree creneau', 's-slot', ['15', '30', '45', '60'], '30'));
  grid.appendChild(numberGroup('Max patients / jour', 's-max', 40));
  c.appendChild(grid);
}

function renderStep4(c) {
  c.appendChild(el('h3', { text: 'Configuration des rappels SMS' }));
  c.appendChild(el('p', { class: 'setup-desc', text: 'Les rappels SMS reduisent les rendez-vous manques de 25 a 50%. Cette etape est optionnelle.' }));
  c.appendChild(typeSelector('s4-enable', ['Oui, activer', 'Plus tard'], 'Plus tard'));

  const configBlock = el('div', { id: 's4-config', style: 'margin-top:16px' });
  configBlock.appendChild(el('label', { text: 'Fournisseur SMS', style: 'display:block;font-size:.8rem;font-weight:600;color:var(--gray-600);margin-bottom:8px;text-transform:uppercase;letter-spacing:.5px' }));
  configBlock.appendChild(typeSelector('s4-provider', ["Africa's Talking", 'MTN', 'Orange', 'Twilio', 'Infobip'], "Africa's Talking"));

  const grid = el('div', { class: 'setup-grid', style: 'margin-top:14px' });
  grid.appendChild(formGroup('Cle API', 'password', 's-sms-key', 'Collez votre cle API'));
  grid.appendChild(formGroup('Secret API', 'password', 's-sms-secret', 'Collez votre secret'));
  grid.appendChild(formGroup('Nom expediteur', 'text', 's-sms-sender', 'Ex: MaSante', true));
  configBlock.appendChild(grid);
  c.appendChild(configBlock);
}

function renderStep5(c) {
  const success = el('div', { style: 'text-align:center;padding:40px 0' });
  const iconWrap = el('div', { style: 'width:72px;height:72px;border-radius:50%;background:var(--success-bg);color:var(--success);display:flex;align-items:center;justify-content:center;margin:0 auto 20px' });
  iconWrap.appendChild(iconEl('check', 32));
  success.appendChild(iconWrap);
  success.appendChild(el('h3', { text: 'Votre plateforme est prete', style: 'margin-bottom:10px' }));
  success.appendChild(el('p', { text: 'Cliquez sur "Lancer MaSante" pour commencer.', style: 'color:var(--gray-400);font-size:.9rem' }));

  // Recap.
  const recap = el('div', { class: 'setup-recap', style: 'text-align:left;margin-top:24px' });
  const rows = [
    ['Etablissement', data.center.name || '—'],
    ['Localisation', [data.center.city, data.center.district, data.center.country].filter(Boolean).join(', ') || '—'],
    ['Administrateur', data.admin.full_name || '—'],
    ['Identifiant', data.admin.username || '—'],
    ['Rappels SMS', data.sms.enabled ? 'Active' : 'Non configure'],
  ];
  rows.forEach(([label, value]) => {
    const row = el('div', { class: 'sr-row' });
    row.appendChild(el('span', { class: 'sr-label', text: label }));
    row.appendChild(el('span', { class: 'sr-val', text: value }));
    recap.appendChild(row);
  });
  success.appendChild(recap);
  c.appendChild(success);
}

// ── Form helpers ──

function formGroup(label, type, id, placeholder, full = false) {
  const g = el('div', { class: 'form-group' + (full ? ' setup-full' : '') });
  g.appendChild(el('label', { text: label }));
  g.appendChild(el('input', { class: 'form-input', type, id, placeholder: placeholder || '' }));
  return g;
}

function selectGroup(label, id, options, selected) {
  const g = el('div', { class: 'form-group' });
  g.appendChild(el('label', { text: label }));
  const select = el('select', { class: 'form-input', id });
  options.forEach(opt => {
    const o = el('option', { text: opt, value: opt });
    if (opt === selected) o.selected = true;
    select.appendChild(o);
  });
  g.appendChild(select);
  return g;
}

function timeGroup(label, id, value) {
  const g = el('div', { class: 'form-group' });
  g.appendChild(el('label', { text: label }));
  g.appendChild(el('input', { class: 'form-input', type: 'time', id, value }));
  return g;
}

function numberGroup(label, id, value) {
  const g = el('div', { class: 'form-group' });
  g.appendChild(el('label', { text: label }));
  g.appendChild(el('input', { class: 'form-input', type: 'number', id, value: String(value), min: '1' }));
  return g;
}

function typeSelector(groupId, options, defaultOpt) {
  const wrapper = el('div', { class: 'type-options', id: groupId, style: 'margin-top:6px' });
  options.forEach(opt => {
    const btn = el('div', { class: 'type-opt' + (opt === defaultOpt ? ' on' : ''), text: opt });
    btn.onclick = () => {
      wrapper.querySelectorAll('.type-opt').forEach(b => b.classList.remove('on'));
      btn.classList.add('on');
    };
    wrapper.appendChild(btn);
  });
  return wrapper;
}
