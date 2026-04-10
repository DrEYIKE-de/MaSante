import { profile as pApi } from '../api.js';
import { el, iconEl, notify } from '../ui.js';
import { state } from '../app.js';
export const title = 'Mon profil';
export async function render(container) {
  const grid = el('div', { class: 'grid-2-wide', style: 'align-items:start' });

  // Left — info.
  const card1 = el('div', { class: 'card' });
  const head1 = el('div', { class: 'card-head' });
  const t1 = el('h3');
  t1.appendChild(iconEl('user', 18));
  t1.appendChild(document.createTextNode(' Informations personnelles'));
  head1.appendChild(t1);
  card1.appendChild(head1);
  const body1 = el('div', { class: 'card-body' });
  const u = state.user || {};
  const nameInput = el('input', { class: 'form-input', type: 'text', value: u.full_name || '' });
  const emailInput = el('input', { class: 'form-input', type: 'email', value: u.email || '' });
  const phoneInput = el('input', { class: 'form-input', type: 'tel', value: '' });

  [['Nom complet', nameInput], ['Email', emailInput], ['Telephone', phoneInput]].forEach(([label, inp]) => {
    const g = el('div', { class: 'form-group' });
    g.appendChild(el('label', { text: label }));
    g.appendChild(inp);
    body1.appendChild(g);
  });

  const saveBtn = el('button', { class: 'btn btn-primary', style: 'width:auto;margin-top:8px' });
  saveBtn.appendChild(iconEl('check', 16));
  saveBtn.appendChild(document.createTextNode(' Enregistrer'));
  saveBtn.onclick = async () => {
    const res = await pApi.update({ full_name: nameInput.value, email: emailInput.value, phone: phoneInput.value });
    if (res.ok) { notify.success('Profil mis a jour'); state.user.full_name = nameInput.value; }
    else notify.error(res.error);
  };
  body1.appendChild(saveBtn);
  card1.appendChild(body1);
  grid.appendChild(card1);

  // Right — password.
  const card2 = el('div', { class: 'card' });
  const head2 = el('div', { class: 'card-head' });
  const t2 = el('h3');
  t2.appendChild(iconEl('shield', 18));
  t2.appendChild(document.createTextNode(' Securite'));
  head2.appendChild(t2);
  card2.appendChild(head2);
  const body2 = el('div', { class: 'card-body' });

  const currentPwd = el('input', { class: 'form-input', type: 'password', placeholder: 'Mot de passe actuel' });
  const newPwd = el('input', { class: 'form-input', type: 'password', placeholder: 'Nouveau mot de passe (min 8 + 1 chiffre)' });
  const confirmPwd = el('input', { class: 'form-input', type: 'password', placeholder: 'Confirmer' });
  const pwdError = el('div', { style: 'display:none;padding:8px 12px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.82rem;margin-top:8px' });

  [['Mot de passe actuel', currentPwd], ['Nouveau mot de passe', newPwd], ['Confirmer', confirmPwd]].forEach(([label, inp]) => {
    const g = el('div', { class: 'form-group' });
    g.appendChild(el('label', { text: label }));
    g.appendChild(inp);
    body2.appendChild(g);
  });

  const changePwdBtn = el('button', { class: 'btn btn-secondary', style: 'width:auto;margin-top:8px' });
  changePwdBtn.appendChild(iconEl('shield', 16));
  changePwdBtn.appendChild(document.createTextNode(' Changer le mot de passe'));
  changePwdBtn.onclick = async () => {
    pwdError.style.display = 'none';
    if (newPwd.value !== confirmPwd.value) { pwdError.textContent = 'Les mots de passe ne correspondent pas'; pwdError.style.display = 'block'; return; }
    const res = await pApi.changePassword(currentPwd.value, newPwd.value);
    if (res.ok) { notify.success('Mot de passe change — reconnectez-vous'); }
    else { pwdError.textContent = res.error; pwdError.style.display = 'block'; }
  };
  body2.appendChild(changePwdBtn);
  body2.appendChild(pwdError);
  card2.appendChild(body2);
  grid.appendChild(card2);

  container.appendChild(grid);
}
