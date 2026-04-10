// dashboard.js — Main dashboard with live stats from API.
import { dashboard } from '../api.js';
import { el, iconEl, statusPill, riskBadge, formatDate, loading, notify } from '../ui.js';

export const title = 'Tableau de bord';

export async function render(container) {
  loading(container);

  const [statsRes, todayRes, overdueRes] = await Promise.all([
    dashboard.stats(),
    dashboard.today(),
    dashboard.overdue(),
  ]);

  container.textContent = '';

  if (!statsRes.ok) {
    notify.error('Erreur de chargement du tableau de bord');
    return;
  }

  const pCounts = statsRes.data.patients || {};
  const aCounts = statsRes.data.appointments || {};

  // Stat cards.
  const statsRow = el('div', { class: 'stats-row' });

  const totalPatients = Object.values(pCounts).reduce((a, b) => a + b, 0);
  const todayCount = Object.values(aCounts).reduce((a, b) => a + b, 0);
  const confirmedCount = aCounts['confirme'] || 0;
  const lostCount = pCounts['perdu_de_vue'] || 0;

  statsRow.appendChild(statCard('users', 'green', totalPatients, 'Patients actifs'));
  statsRow.appendChild(statCard('calendar', 'blue', todayCount, "RDV aujourd'hui"));
  statsRow.appendChild(statCard('check', 'amber', confirmedCount, 'Confirmes'));
  statsRow.appendChild(statCard('alert', 'red', lostCount, 'Perdus de vue'));
  container.appendChild(statsRow);

  const grid = el('div', { class: 'grid-2', style: 'align-items:start' });

  // Today's appointments.
  const todayCard = el('div', { class: 'card' });
  const todayHead = el('div', { class: 'card-head' });
  const todayTitle = el('h3');
  todayTitle.appendChild(iconEl('calendar', 18));
  todayTitle.appendChild(document.createTextNode(" RDV du jour"));
  todayHead.appendChild(todayTitle);
  const calLink = el('a', { class: 'card-link', text: 'Voir tout' });
  calLink.onclick = () => { location.hash = '#calendar'; };
  todayHead.appendChild(calLink);
  todayCard.appendChild(todayHead);

  const todayBody = el('div', { class: 'card-body' });
  const todayApts = todayRes.ok ? (todayRes.data || []) : [];
  if (todayApts.length === 0) {
    todayBody.appendChild(el('div', { style: 'text-align:center;padding:20px;color:var(--gray-300)', text: 'Aucun rendez-vous aujourd\'hui' }));
  } else {
    todayApts.forEach(apt => {
      const item = el('div', { class: 'list-item' });
      item.appendChild(el('span', { class: 'time-label', text: apt.Time || '—' }));
      const initials = (apt.PatientName || '??').split(' ').map(w => w[0]).join('').substring(0, 2);
      item.appendChild(el('div', { class: 'avatar a1', text: initials }));
      const info = el('div', { class: 'item-info' });
      info.appendChild(el('div', { class: 'item-name', text: apt.PatientName || 'Patient' }));
      info.appendChild(el('div', { class: 'item-sub', text: formatType(apt.Type) }));
      item.appendChild(info);
      item.appendChild(statusPill(apt.Status));
      todayBody.appendChild(item);
    });
  }
  todayCard.appendChild(todayBody);
  grid.appendChild(todayCard);

  // Overdue patients.
  const overdueCard = el('div', { class: 'card' });
  const overdueHead = el('div', { class: 'card-head' });
  const overdueTitle = el('h3');
  overdueTitle.appendChild(iconEl('clock', 18));
  overdueTitle.appendChild(document.createTextNode(' Patients en retard'));
  overdueHead.appendChild(overdueTitle);
  overdueCard.appendChild(overdueHead);

  const overdueBody = el('div', { class: 'card-body' });
  const overdueApts = overdueRes.ok ? (overdueRes.data || []) : [];
  if (overdueApts.length === 0) {
    overdueBody.appendChild(el('div', { style: 'text-align:center;padding:20px;color:var(--gray-300)', text: 'Aucun patient en retard' }));
  } else {
    overdueApts.forEach(apt => {
      const item = el('div', { class: 'list-item' });
      const initials = (apt.PatientName || '??').split(' ').map(w => w[0]).join('').substring(0, 2);
      item.appendChild(el('div', { class: 'overdue-avatar', text: initials }));
      const info = el('div', { class: 'item-info' });
      info.appendChild(el('div', { class: 'item-name', text: apt.PatientName || 'Patient' }));
      const daysLate = Math.floor((Date.now() - new Date(apt.Date).getTime()) / 86400000);
      info.appendChild(el('div', { class: 'overdue-days', text: daysLate + ' jours de retard' }));
      item.appendChild(info);
      const reprogBtn = el('button', { class: 'btn btn-sm btn-secondary', text: 'Reprogrammer' });
      reprogBtn.style.width = 'auto';
      reprogBtn.style.flexShrink = '0';
      reprogBtn.onclick = () => { location.hash = '#new-apt'; };
      item.appendChild(reprogBtn);
      overdueBody.appendChild(item);
    });
  }
  overdueCard.appendChild(overdueBody);
  grid.appendChild(overdueCard);

  container.appendChild(grid);
}

function statCard(iconName, color, value, label) {
  const card = el('div', { class: 'stat' });
  const iconWrap = el('div', { class: 'stat-icon ' + color });
  iconWrap.appendChild(iconEl(iconName, 20));
  card.appendChild(iconWrap);
  card.appendChild(el('div', { class: 'stat-val', text: String(value) }));
  card.appendChild(el('div', { class: 'stat-label', text: label }));
  return card;
}

function formatType(type) {
  const map = {
    consultation: 'Consultation de suivi',
    retrait_medicaments: 'Retrait medicaments',
    bilan_sanguin: 'Bilan sanguin',
    club_adherence: "Club d'adherence",
  };
  return map[type] || type || '';
}
