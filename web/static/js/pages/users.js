import { users as uApi } from '../api.js';
import { el, iconEl, statusPill, notify, openModal, closeModal } from '../ui.js';
export const title = 'Gestion des utilisateurs';
export async function render(container) {
  const headerDiv = el('div', { style: 'display:flex;align-items:center;justify-content:space-between;margin-bottom:20px' });
  headerDiv.appendChild(el('h3', { text: 'Equipe soignante', style: 'font-size:1.1rem' }));
  const addBtn = el('button', { class: 'btn btn-primary', style: 'width:auto' });
  addBtn.appendChild(iconEl('user-plus', 16));
  addBtn.appendChild(document.createTextNode(' Ajouter'));
  addBtn.onclick = () => showAddModal(loadUsers);
  headerDiv.appendChild(addBtn);
  container.appendChild(headerDiv);

  const tableEl = el('div');
  container.appendChild(tableEl);

  async function loadUsers() {
    tableEl.textContent = '';
    const res = await uApi.list();
    if (!res.ok) { notify.error(res.error); return; }
    const list = res.data || [];
    const table = el('div', { class: 'ptable' });
    const thead = el('div', { class: 'pt-head', style: 'grid-template-columns:44px 2fr 1fr 1fr 1fr 80px' });
    ['', 'Utilisateur', 'Role', 'Derniere connexion', 'Statut', ''].forEach(h => thead.appendChild(el('div', { text: h })));
    table.appendChild(thead);
    list.forEach(u => {
      const row = el('div', { class: 'pt-row', style: 'grid-template-columns:44px 2fr 1fr 1fr 1fr 80px' });
      const initials = (u.full_name || '??').split(' ').map(w => w[0]).join('').substring(0, 2).toUpperCase();
      row.appendChild(el('div', {}, [el('div', { class: 'avatar a1', text: initials })]));
      const nameDiv = el('div');
      nameDiv.appendChild(el('div', { class: 'pt-name', text: u.full_name }));
      nameDiv.appendChild(el('div', { class: 'pt-code', text: u.username + (u.email ? ' — ' + u.email : '') }));
      row.appendChild(nameDiv);
      const roleMap = { admin: 'pill-danger', medecin: 'pill-info', infirmier: 'pill-warning', asc: 'pill-success' };
      row.appendChild(el('div', {}, [el('span', { class: 'pill ' + (roleMap[u.role] || 'pill-neutral'), text: u.role })]));
      row.appendChild(el('div', { class: 'pt-cell', text: u.last_login_at ? new Date(u.last_login_at).toLocaleDateString('fr-FR') : '—' }));
      row.appendChild(el('div', {}, [statusPill(u.status)]));
      const acts = el('div', { class: 'pt-acts' });
      const delBtn = el('button', { class: 'icon-btn' });
      delBtn.appendChild(iconEl('trash', 16));
      delBtn.onclick = async () => { if (confirm('Desactiver ' + u.full_name + ' ?')) { await uApi.disable(u.id); loadUsers(); } };
      acts.appendChild(delBtn);
      row.appendChild(acts);
      table.appendChild(row);
    });
    tableEl.appendChild(table);
  }
  loadUsers();
}

function showAddModal(onDone) {
  const content = el('div');
  const fields = {};
  [['Nom complet', 'text', 'full_name', 'Ex: Ngassa Marie'],
   ['Email', 'email', 'email', 'Ex: ngassa@mail.cm'],
   ['Identifiant', 'text', 'username', 'Choisir un identifiant'],
   ['Mot de passe', 'password', 'password', 'Min 8 caracteres + 1 chiffre']].forEach(([label, type, key, ph]) => {
    const g = el('div', { class: 'form-group' });
    g.appendChild(el('label', { text: label }));
    const inp = el('input', { class: 'form-input', type, placeholder: ph });
    g.appendChild(inp);
    fields[key] = inp;
    content.appendChild(g);
  });
  const roleGroup = el('div', { class: 'form-group' });
  roleGroup.appendChild(el('label', { text: 'Role' }));
  const roleSel = el('select', { class: 'form-input' });
  [['admin', 'Admin'], ['medecin', 'Medecin'], ['infirmier', 'Infirmier'], ['asc', 'ASC']].forEach(([v, l]) => roleSel.appendChild(el('option', { value: v, text: l })));
  roleSel.value = 'asc';
  roleGroup.appendChild(roleSel);
  content.appendChild(roleGroup);

  const errorEl = el('div', { style: 'display:none;padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin-top:10px' });
  content.appendChild(errorEl);

  openModal('Ajouter un utilisateur', content, [
    { label: 'Annuler', cls: 'btn-secondary', onclick: closeModal },
    { label: 'Creer', cls: 'btn-primary', onclick: async () => {
      errorEl.style.display = 'none';
      const data = { full_name: fields.full_name.value, email: fields.email.value, username: fields.username.value, password: fields.password.value, role: roleSel.value };
      if (!data.full_name || !data.username || !data.password) { errorEl.textContent = 'Champs requis manquants'; errorEl.style.display = 'block'; return; }
      const res = await uApi.create(data);
      if (!res.ok) { errorEl.textContent = res.error; errorEl.style.display = 'block'; return; }
      notify.success('Utilisateur cree');
      closeModal();
      onDone();
    }},
  ]);
}
