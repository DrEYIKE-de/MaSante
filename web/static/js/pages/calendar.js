import { calendar as calApi } from '../api.js';
import { el, iconEl, statusPill, loading, notify } from '../ui.js';
export const title = 'Calendrier des rendez-vous';
export async function render(container) {
  const today = new Date();
  const monday = new Date(today);
  monday.setDate(today.getDate() - ((today.getDay() + 6) % 7));
  const dateStr = monday.toISOString().slice(0, 10);

  loading(container);
  const res = await calApi.week(dateStr);
  container.textContent = '';

  const controls = el('div', { class: 'cal-controls' });
  controls.appendChild(el('span', { class: 'cal-date', text: 'Semaine du ' + monday.toLocaleDateString('fr-FR', { day: 'numeric', month: 'long', year: 'numeric' }) }));
  container.appendChild(controls);

  const grid = el('div', { class: 'cal-grid' });
  const header = el('div', { class: 'cal-header' });
  header.appendChild(el('div', { class: 'cal-hcell' }));
  const dayNames = ['Lun', 'Mar', 'Mer', 'Jeu', 'Ven', 'Sam', 'Dim'];
  for (let i = 0; i < 7; i++) {
    const d = new Date(monday);
    d.setDate(monday.getDate() + i);
    const isToday = d.toDateString() === today.toDateString();
    header.appendChild(el('div', { class: 'cal-hcell' + (isToday ? ' today' : ''), text: dayNames[i] + ' ' + d.getDate() }));
  }
  grid.appendChild(header);

  const body = el('div', { class: 'cal-body' });
  const apts = res.ok ? (res.data || []) : [];
  const slotsByDayTime = {};
  apts.forEach(a => {
    const dayKey = a.Date ? a.Date.slice(0, 10) : '';
    const key = dayKey + '|' + (a.Time || '');
    if (!slotsByDayTime[key]) slotsByDayTime[key] = [];
    slotsByDayTime[key].push(a);
  });

  for (let h = 8; h <= 16; h++) {
    const time = String(h).padStart(2, '0') + ':00';
    const row = el('div', { class: 'cal-row' });
    row.appendChild(el('div', { class: 'cal-time', text: time }));
    for (let d = 0; d < 7; d++) {
      const dayDate = new Date(monday);
      dayDate.setDate(monday.getDate() + d);
      const dayStr = dayDate.toISOString().slice(0, 10);
      const slot = el('div', { class: 'cal-slot' });
      const key = dayStr + '|' + time;
      (slotsByDayTime[key] || []).forEach(a => {
        const statusMap = { confirme: 'c-ok', en_attente: 'c-wait', manque: 'c-miss', termine: 'c-done' };
        const evt = el('div', { class: 'cal-evt ' + (statusMap[a.Status] || 'c-ok'), text: (a.PatientName || '').split(' ')[0] });
        slot.appendChild(evt);
      });
      row.appendChild(slot);
    }
    body.appendChild(row);
  }
  grid.appendChild(body);
  container.appendChild(grid);
}
