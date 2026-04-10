import { reminders as rApi } from '../api.js';
import { el, iconEl, statusPill, loading, notify } from '../ui.js';
export const title = 'Gestion des rappels';
export async function render(container) {
  loading(container);
  const [statsRes, queueRes, tplRes] = await Promise.all([rApi.stats(), rApi.list(), rApi.templates()]);
  container.textContent = '';

  // Stats.
  const stats = statsRes.ok ? statsRes.data : {};
  const statsRow = el('div', { class: 'rem-stats' });
  [['Taux de livraison', (stats.DeliveryRate || 0).toFixed(1) + '%', 'var(--success)'],
   ['En attente', String(stats.PendingCount || 0), 'var(--warning)'],
   ['Echecs', String(stats.FailedCount || 0), 'var(--danger)']].forEach(([label, val, color]) => {
    const s = el('div', { class: 'rem-stat' });
    s.appendChild(el('div', { class: 'rem-stat-val', text: val, style: 'color:' + color }));
    s.appendChild(el('div', { class: 'rem-stat-label', text: label }));
    statsRow.appendChild(s);
  });
  container.appendChild(statsRow);

  const grid = el('div', { class: 'grid-2-wide', style: 'align-items:start' });

  // Queue.
  const qCard = el('div', { class: 'card' });
  const qHead = el('div', { class: 'card-head' });
  const qTitle = el('h3');
  qTitle.appendChild(iconEl('send', 18));
  qTitle.appendChild(document.createTextNode(" File d'attente"));
  qHead.appendChild(qTitle);
  const sendAllBtn = el('a', { class: 'card-link', text: 'Tout envoyer' });
  sendAllBtn.onclick = async () => { const r = await rApi.sendAll(); if (r.ok) notify.success('Rappels envoyes'); else notify.error(r.error); };
  qHead.appendChild(sendAllBtn);
  qCard.appendChild(qHead);

  const qBody = el('div', { class: 'card-body' });
  const queue = queueRes.ok ? (queueRes.data || []) : [];
  if (queue.length === 0) {
    qBody.appendChild(el('div', { style: 'text-align:center;padding:20px;color:var(--gray-300)', text: 'Aucun rappel en attente' }));
  } else {
    queue.forEach(r => {
      const item = el('div', { class: 'rq-item' });
      const chMap = { sms: 'ch-sms', whatsapp: 'ch-wa', voice: 'ch-voice' };
      const chIcon = el('div', { class: 'rq-ch ' + (chMap[r.Channel] || 'ch-sms') });
      chIcon.appendChild(iconEl(r.Channel === 'voice' ? 'phone' : 'msg', 18));
      item.appendChild(chIcon);
      const info = el('div', { class: 'rq-info' });
      info.appendChild(el('div', { class: 'rq-name', text: r.PatientName || 'Patient' }));
      info.appendChild(el('div', { class: 'rq-sched', text: r.Type + ' — ' + (r.Status || '') }));
      item.appendChild(info);
      item.appendChild(statusPill(r.Status));
      qBody.appendChild(item);
    });
  }
  qCard.appendChild(qBody);
  grid.appendChild(qCard);

  // Templates.
  const tCard = el('div', { class: 'card' });
  const tHead = el('div', { class: 'card-head' });
  const tTitle = el('h3');
  tTitle.appendChild(iconEl('clipboard', 18));
  tTitle.appendChild(document.createTextNode(' Modeles de messages'));
  tHead.appendChild(tTitle);
  tCard.appendChild(tHead);

  const tBody = el('div', { class: 'card-body' });
  const templates = tplRes.ok ? (tplRes.data || []) : [];
  templates.forEach(t => {
    const group = el('div', { class: 'form-group' });
    group.appendChild(el('label', { text: t.Name }));
    const textarea = el('textarea', { class: 'form-input', rows: '3' });
    textarea.value = t.Body;
    group.appendChild(textarea);
    const saveBtn = el('button', { class: 'btn btn-sm btn-secondary', text: 'Enregistrer', style: 'width:auto;margin-top:6px' });
    saveBtn.onclick = async () => {
      const r = await rApi.updateTemplate(t.ID, { body: textarea.value, is_active: t.IsActive });
      if (r.ok) notify.success('Modele enregistre'); else notify.error(r.error);
    };
    group.appendChild(saveBtn);
    tBody.appendChild(group);
  });
  tCard.appendChild(tBody);
  grid.appendChild(tCard);

  container.appendChild(grid);
}
