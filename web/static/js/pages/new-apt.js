import { patients as pApi, appointments as aApi } from '../api.js';
import { el, iconEl, notify } from '../ui.js';
export const title = 'Prise de rendez-vous';
export async function render(container) {
  const grid = el('div', { class: 'apt-grid' });
  const left = el('div');
  const right = el('div');

  // Patient search.
  const card1 = el('div', { class: 'card' });
  const head1 = el('div', { class: 'card-head' });
  const t1 = el('h3');
  t1.appendChild(iconEl('users', 18));
  t1.appendChild(document.createTextNode(' Patient'));
  head1.appendChild(t1);
  card1.appendChild(head1);
  const body1 = el('div', { class: 'card-body' });
  const searchInput = el('input', { class: 'form-input', type: 'text', placeholder: 'Rechercher par nom ou code...' });
  const resultsList = el('div', { style: 'margin-top:10px' });
  let selectedPatientId = sessionStorage.getItem('apt_patient_id') || null;
  sessionStorage.removeItem('apt_patient_id');
  const selectedInfo = el('div', { style: 'margin-top:10px' });

  searchInput.oninput = async () => {
    const q = searchInput.value.trim();
    if (q.length < 2) { resultsList.textContent = ''; return; }
    const res = await pApi.search(q);
    resultsList.textContent = '';
    if (res.ok && res.data) {
      res.data.slice(0, 5).forEach(p => {
        const item = el('div', { style: 'padding:8px 12px;cursor:pointer;border-bottom:1px solid var(--gray-50);font-size:.88rem', text: p.LastName + ' ' + p.FirstName + ' — ' + p.Code });
        item.onmouseenter = () => item.style.background = 'var(--gray-25)';
        item.onmouseleave = () => item.style.background = '';
        item.onclick = () => { selectedPatientId = p.ID; resultsList.textContent = ''; searchInput.value = p.LastName + ' ' + p.FirstName; selectedInfo.textContent = p.Code + ' — ' + (p.District || ''); };
        resultsList.appendChild(item);
      });
    }
  };
  body1.appendChild(searchInput);
  body1.appendChild(resultsList);
  body1.appendChild(selectedInfo);

  // Type.
  const typeSelect = el('select', { class: 'form-input', style: 'margin-top:12px' });
  ['consultation', 'retrait_medicaments', 'bilan_sanguin', 'club_adherence'].forEach(t => {
    const labels = { consultation: 'Consultation de suivi', retrait_medicaments: 'Retrait medicaments', bilan_sanguin: 'Bilan sanguin', club_adherence: "Club d'adherence" };
    typeSelect.appendChild(el('option', { value: t, text: labels[t] }));
  });
  const typeGroup = el('div', { class: 'form-group', style: 'margin-top:12px' });
  typeGroup.appendChild(el('label', { text: 'Type de visite' }));
  typeGroup.appendChild(typeSelect);
  body1.appendChild(typeGroup);

  // Notes.
  const notesInput = el('textarea', { class: 'form-input', placeholder: 'Notes...', rows: '3' });
  const notesGroup = el('div', { class: 'form-group' });
  notesGroup.appendChild(el('label', { text: 'Notes' }));
  notesGroup.appendChild(notesInput);
  body1.appendChild(notesGroup);

  card1.appendChild(body1);
  left.appendChild(card1);

  // Date + slots.
  const card2 = el('div', { class: 'card' });
  const head2 = el('div', { class: 'card-head' });
  const t2 = el('h3');
  t2.appendChild(iconEl('calendar', 18));
  t2.appendChild(document.createTextNode(' Date et creneau'));
  head2.appendChild(t2);
  card2.appendChild(head2);
  const body2 = el('div', { class: 'card-body' });
  const dateInput = el('input', { class: 'form-input', type: 'date' });
  const dateGroup = el('div', { class: 'form-group' });
  dateGroup.appendChild(el('label', { text: 'Date du rendez-vous' }));
  dateGroup.appendChild(dateInput);
  body2.appendChild(dateGroup);

  const slotsContainer = el('div');
  let selectedTime = '';
  dateInput.onchange = async () => {
    slotsContainer.textContent = '';
    selectedTime = '';
    const res = await aApi.slots(dateInput.value);
    if (!res.ok) return;
    const slotsGrid = el('div', { class: 'slots-grid' });
    (res.data || []).forEach(s => {
      const cls = 'slot' + (!s.Available ? ' off' : '');
      const slot = el('div', { class: cls, text: s.Time });
      if (s.Available) {
        slot.onclick = () => {
          slotsGrid.querySelectorAll('.slot').forEach(x => x.classList.remove('picked'));
          slot.classList.add('picked');
          selectedTime = s.Time;
        };
      }
      slotsGrid.appendChild(slot);
    });
    slotsContainer.appendChild(slotsGrid);
  };
  body2.appendChild(slotsContainer);
  card2.appendChild(body2);
  right.appendChild(card2);

  // Error + submit.
  const errorEl = el('div', { style: 'display:none;padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin:12px 0' });
  right.appendChild(errorEl);

  const submitBtn = el('button', { class: 'btn btn-primary', style: 'margin-top:12px' });
  submitBtn.appendChild(iconEl('check', 16));
  submitBtn.appendChild(document.createTextNode(' Confirmer le rendez-vous'));
  submitBtn.onclick = async () => {
    errorEl.style.display = 'none';
    if (!selectedPatientId) { errorEl.textContent = 'Selectionnez un patient'; errorEl.style.display = 'block'; return; }
    if (!dateInput.value) { errorEl.textContent = 'Selectionnez une date'; errorEl.style.display = 'block'; return; }
    if (!selectedTime) { errorEl.textContent = 'Selectionnez un creneau'; errorEl.style.display = 'block'; return; }
    submitBtn.disabled = true;
    const res = await aApi.create({ patient_id: parseInt(selectedPatientId), date: dateInput.value, time: selectedTime, type: typeSelect.value, notes: notesInput.value });
    if (!res.ok) { errorEl.textContent = res.error; errorEl.style.display = 'block'; submitBtn.disabled = false; return; }
    notify.success('Rendez-vous programme');
    location.hash = '#calendar';
  };
  right.appendChild(submitBtn);

  grid.appendChild(left);
  grid.appendChild(right);
  container.appendChild(grid);
}
