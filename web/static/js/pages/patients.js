// patients.js — Patient list with search, filters, and pagination.
import { patients as patientsApi } from '../api.js';
import { el, iconEl, statusPill, riskBadge, formatDate, loading, notify } from '../ui.js';

export const title = 'Liste des patients';

export async function render(container) {
  loading(container);

  // Check for search query from topbar.
  const query = sessionStorage.getItem('search_query') || '';
  sessionStorage.removeItem('search_query');

  // Filters.
  const filters = el('div', { class: 'filters' });
  const statuses = [
    { label: 'Tous', value: '' },
    { label: 'Actifs', value: 'active' },
    { label: 'A surveiller', value: 'a_surveiller' },
    { label: 'Perdus de vue', value: 'perdu_de_vue' },
    { label: 'Sortis', value: 'sorti' },
  ];
  let activeFilter = '';
  statuses.forEach(s => {
    const btn = el('button', { class: 'fbtn' + (s.value === '' ? ' on' : ''), text: s.label });
    btn.onclick = () => {
      activeFilter = s.value;
      filters.querySelectorAll('.fbtn').forEach(b => b.classList.remove('on'));
      btn.classList.add('on');
      loadList();
    };
    filters.appendChild(btn);
  });

  container.textContent = '';
  container.appendChild(filters);

  const tableContainer = el('div');
  container.appendChild(tableContainer);

  async function loadList() {
    loading(tableContainer);
    let params = 'page=1&per_page=50';
    if (activeFilter) params += '&status=' + activeFilter;
    if (query) params += '&q=' + encodeURIComponent(query);
    const res = await patientsApi.list(params);
    tableContainer.textContent = '';

    if (!res.ok) {
      notify.error(res.error);
      return;
    }

    const list = res.data.patients || [];
    const total = res.data.total || 0;

    if (list.length === 0) {
      tableContainer.appendChild(el('div', { style: 'text-align:center;padding:40px;color:var(--gray-300)', text: 'Aucun patient trouve' }));
      return;
    }

    const table = el('div', { class: 'ptable' });

    // Header.
    const thead = el('div', { class: 'pt-head', style: 'grid-template-columns:44px 2fr 1fr 1fr 1fr 100px' });
    ['', 'Patient', 'Risque', 'Derniere visite', 'Statut', 'Actions'].forEach(h => {
      thead.appendChild(el('div', { text: h }));
    });
    table.appendChild(thead);

    // Rows.
    list.forEach(p => {
      const row = el('div', { class: 'pt-row', style: 'grid-template-columns:44px 2fr 1fr 1fr 1fr 100px;cursor:pointer' });
      row.onclick = () => {
        sessionStorage.setItem('patient_id', p.ID);
        location.hash = '#patient-file';
      };

      const initials = ((p.LastName || '')[0] || '') + ((p.FirstName || '')[0] || '');
      row.appendChild(el('div', {}, [el('div', { class: 'avatar a1', text: initials.toUpperCase() })]));

      const nameCell = el('div');
      nameCell.appendChild(el('div', { class: 'pt-name', text: p.LastName + ' ' + p.FirstName }));
      nameCell.appendChild(el('div', { class: 'pt-code', text: p.Code }));
      row.appendChild(nameCell);

      row.appendChild(el('div', {}, [riskBadge(p.RiskScore)]));
      row.appendChild(el('div', { class: 'pt-cell', text: formatDate(p.UpdatedAt) }));
      row.appendChild(el('div', {}, [statusPill(p.Status)]));

      const acts = el('div', { class: 'pt-acts' });
      const aptBtn = el('button', { class: 'icon-btn' });
      aptBtn.appendChild(iconEl('calendar', 16));
      aptBtn.onclick = (e) => { e.stopPropagation(); location.hash = '#new-apt'; };
      acts.appendChild(aptBtn);
      row.appendChild(acts);

      table.appendChild(row);
    });

    tableContainer.appendChild(table);
    tableContainer.appendChild(el('div', { style: 'text-align:center;padding:12px;color:var(--gray-400);font-size:.82rem', text: total + ' patients au total' }));
  }

  loadList();
}
