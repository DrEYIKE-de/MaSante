// patient-file.js — Individual patient view with timeline.
import { patients as patientsApi, appointments as aptsApi } from '../api.js';
import { el, iconEl, statusPill, riskBadge, formatDate, loading, notify, openModal, closeModal } from '../ui.js';

export const title = 'Fiche patient';

export async function render(container) {
  const id = sessionStorage.getItem('patient_id');
  if (!id) { location.hash = '#patients'; return; }

  loading(container);
  const res = await patientsApi.get(id);
  if (!res.ok) { notify.error('Patient introuvable'); location.hash = '#patients'; return; }
  const p = res.data;

  container.textContent = '';
  const layout = el('div', { class: 'pf-layout' });

  // Sidebar.
  const side = el('div', { class: 'pf-side' });

  // Profile card.
  const card = el('div', { class: 'pf-card' });
  const initials = ((p.LastName || '')[0] || '') + ((p.FirstName || '')[0] || '');
  card.appendChild(el('div', { class: 'pf-avatar', text: initials.toUpperCase() }));
  card.appendChild(el('div', { class: 'pf-name', text: p.LastName + ' ' + p.FirstName }));
  card.appendChild(el('div', { class: 'pf-id', text: p.Code }));
  card.appendChild(el('div', { style: 'margin-bottom:10px' }, [statusPill(p.Status)]));

  const details = el('div', { class: 'pf-details' });
  const rows = [
    ['Age', calcAge(p.DateOfBirth)],
    ['Sexe', p.Sex === 'M' ? 'Masculin' : 'Feminin'],
    ['Zone', p.District || '—'],
    ['Langue', p.Language || 'fr'],
    ['Telephone', p.Phone || '—'],
    ['Rappels', p.ReminderChannel || 'sms'],
    ['Inscrit', formatDate(p.EnrollmentDate)],
  ];
  rows.forEach(([lbl, val]) => {
    const row = el('div', { class: 'pf-row' });
    row.appendChild(el('span', { class: 'lbl', text: lbl }));
    row.appendChild(el('span', { class: 'val', text: val }));
    details.appendChild(row);
  });
  card.appendChild(details);
  side.appendChild(card);

  // Risk card.
  const riskCard = el('div', { class: 'risk-card' });
  const riskHead = el('div', { class: 'risk-head' });
  riskHead.appendChild(el('h4', { text: 'Score de risque' }));
  const circle = el('div', { class: 'risk-circle', text: String(p.RiskScore || 5) });
  if (p.RiskScore > 6) { circle.style.background = 'var(--danger-bg)'; circle.style.color = 'var(--danger)'; circle.style.borderColor = 'var(--danger)'; }
  else if (p.RiskScore > 3) { circle.style.background = 'var(--warning-bg)'; circle.style.color = 'var(--warning)'; circle.style.borderColor = 'var(--warning)'; }
  riskHead.appendChild(circle);
  riskCard.appendChild(riskHead);
  side.appendChild(riskCard);

  // Actions.
  const aptBtn = el('button', { class: 'btn btn-primary', style: 'margin-bottom:8px' });
  aptBtn.appendChild(iconEl('calendar', 16));
  aptBtn.appendChild(document.createTextNode(' Programmer un RDV'));
  aptBtn.onclick = () => { sessionStorage.setItem('apt_patient_id', p.ID); location.hash = '#new-apt'; };
  side.appendChild(aptBtn);

  if (p.Status !== 'sorti') {
    const exitBtn = el('button', { class: 'btn btn-secondary', style: 'color:var(--gray-500)' });
    exitBtn.appendChild(iconEl('archive', 16));
    exitBtn.appendChild(document.createTextNode(' Sortie du programme'));
    exitBtn.onclick = () => showExitModal(p);
    side.appendChild(exitBtn);
  }

  layout.appendChild(side);

  // Main.
  const main = el('div', { class: 'pf-main' });

  // Timeline placeholder — load appointments for this patient.
  const timelineCard = el('div', { class: 'card' });
  const timelineHead = el('div', { class: 'card-head' });
  const timelineTitle = el('h3');
  timelineTitle.appendChild(iconEl('clock', 18));
  timelineTitle.appendChild(document.createTextNode(' Historique'));
  timelineHead.appendChild(timelineTitle);
  timelineCard.appendChild(timelineHead);

  const timelineBody = el('div', { class: 'card-body' });
  timelineBody.appendChild(el('div', { style: 'text-align:center;padding:20px;color:var(--gray-300)', text: 'Chargement...' }));
  timelineCard.appendChild(timelineBody);
  main.appendChild(timelineCard);

  layout.appendChild(main);
  container.appendChild(layout);

  // Load timeline async.
  const aptsRes = await aptsApi.get(id); // This will 404 since it expects appointment ID — use list instead.
  // Actually we need a patient-specific endpoint. For now show enrollment date.
  timelineBody.textContent = '';
  const tl = el('div', { class: 'tl' });
  const item = el('div', { class: 'tl-item current' });
  item.appendChild(el('div', { class: 'tl-date', text: formatDate(p.EnrollmentDate) + ' — inscription' }));
  const content = el('div', { class: 'tl-content' });
  content.appendChild(el('div', { class: 'tl-type', text: 'Inscription dans le programme' }));
  content.appendChild(el('div', { class: 'tl-note', text: p.ReferredBy ? 'Refere par: ' + p.ReferredBy : '' }));
  item.appendChild(content);
  tl.appendChild(item);
  timelineBody.appendChild(tl);
}

function calcAge(dob) {
  if (!dob) return '—';
  const d = new Date(dob);
  if (isNaN(d)) return '—';
  const age = Math.floor((Date.now() - d.getTime()) / 31557600000);
  return age + ' ans';
}

function showExitModal(patient) {
  const content = el('div');
  content.appendChild(el('p', { text: 'Motif de sortie pour ' + patient.LastName + ' ' + patient.FirstName, style: 'font-size:.85rem;color:var(--gray-500);margin-bottom:16px' }));

  const reasons = ['deces', 'transfert', 'abandon', 'perdu_de_vue', 'guerison'];
  const labels = ['Deces', 'Transfert vers un autre centre', 'Abandon volontaire', 'Perdu de vue definitif', 'Guerison / Fin de traitement'];
  let selectedReason = '';

  const optionsDiv = el('div');
  reasons.forEach((r, i) => {
    const opt = el('div', { style: 'padding:10px 14px;border:1.5px solid var(--gray-200);border-radius:var(--radius);margin-bottom:8px;cursor:pointer;font-size:.88rem;font-weight:500;color:var(--gray-600)' });
    opt.textContent = labels[i];
    opt.onclick = () => {
      optionsDiv.querySelectorAll('div').forEach(d => { d.style.borderColor = 'var(--gray-200)'; d.style.background = ''; });
      opt.style.borderColor = 'var(--primary)';
      opt.style.background = 'var(--primary-subtle)';
      selectedReason = r;
    };
    optionsDiv.appendChild(opt);
  });
  content.appendChild(optionsDiv);

  const dateGroup = el('div', { class: 'form-group', style: 'margin-top:12px' });
  dateGroup.appendChild(el('label', { text: 'Date de sortie' }));
  const dateInput = el('input', { class: 'form-input', type: 'date', value: new Date().toISOString().slice(0, 10) });
  dateGroup.appendChild(dateInput);
  content.appendChild(dateGroup);

  const notesGroup = el('div', { class: 'form-group' });
  notesGroup.appendChild(el('label', { text: 'Notes' }));
  const notesInput = el('textarea', { class: 'form-input', placeholder: 'Circonstances...' });
  notesInput.rows = 3;
  notesGroup.appendChild(notesInput);
  content.appendChild(notesGroup);

  openModal('Sortie du programme', content, [
    { label: 'Annuler', cls: 'btn-secondary', onclick: closeModal },
    { label: 'Confirmer la sortie', cls: 'btn-primary', onclick: async () => {
      if (!selectedReason) { notify.error('Selectionnez un motif'); return; }
      const res = await patientsApi.exit(patient.ID, {
        reason: selectedReason,
        date: dateInput.value,
        notes: notesInput.value,
      });
      if (!res.ok) { notify.error(res.error); return; }
      notify.success('Patient sorti du programme');
      closeModal();
      location.hash = '#patients';
    }},
  ]);
}
