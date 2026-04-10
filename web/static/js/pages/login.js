// login.js — Login page.
import { el, iconEl, notify } from '../ui.js';
import { doLogin } from '../app.js';

export const title = 'Connexion';

export async function render(container) {
  const wrapper = el('div', { style: 'display:flex;height:100vh;margin:-24px' });

  // Left panel.
  const left = el('div', { class: 'login-left', style: 'flex:1;display:flex;flex-direction:column;justify-content:center;align-items:center;position:relative;overflow:hidden;background:var(--primary)' });
  const brand = el('div', { style: 'text-align:center;color:#fff;position:relative;z-index:1' });
  const logo = el('div', { class: 'login-logo', style: 'width:80px;height:80px;border:2px solid rgba(255,255,255,.15);border-radius:20px;display:flex;align-items:center;justify-content:center;margin:0 auto 28px;background:rgba(255,255,255,.06)' });
  logo.appendChild(iconEl('leaf', 36));
  brand.appendChild(logo);
  brand.appendChild(el('h1', { text: 'MaSante', style: 'font-size:2.8rem;font-weight:700;letter-spacing:-1.5px;margin-bottom:8px' }));
  brand.appendChild(el('p', { text: 'Plateforme de suivi sante communautaire', style: 'font-size:1rem;color:rgba(255,255,255,.5)' }));
  left.appendChild(brand);

  // Right panel.
  const right = el('div', { style: 'width:440px;background:var(--white);display:flex;flex-direction:column;justify-content:center;padding:56px' });
  right.appendChild(el('h2', { text: 'Connexion', style: 'font-size:1.7rem;margin-bottom:6px;color:var(--gray-900)' }));
  right.appendChild(el('p', { text: 'Entrez vos identifiants pour acceder a votre espace', style: 'color:var(--gray-400);margin-bottom:32px;font-size:.9rem' }));

  // Error message.
  const errorEl = el('div', { style: 'display:none;padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin-bottom:16px' });
  right.appendChild(errorEl);

  // Username.
  const usernameGroup = el('div', { class: 'form-group' });
  usernameGroup.appendChild(el('label', { text: 'Identifiant' }));
  const usernameInput = el('input', { class: 'form-input', type: 'text', placeholder: 'Votre identifiant' });
  usernameGroup.appendChild(usernameInput);
  right.appendChild(usernameGroup);

  // Password.
  const passwordGroup = el('div', { class: 'form-group' });
  passwordGroup.appendChild(el('label', { text: 'Mot de passe' }));
  const passwordInput = el('input', { class: 'form-input', type: 'password', placeholder: 'Votre mot de passe' });
  passwordGroup.appendChild(passwordInput);
  right.appendChild(passwordGroup);

  // Submit.
  const submitBtn = el('button', { class: 'btn btn-primary', style: 'margin-top:8px' });
  submitBtn.appendChild(iconEl('check', 16));
  submitBtn.appendChild(document.createTextNode(' Se connecter'));

  async function handleLogin() {
    const username = usernameInput.value.trim();
    const password = passwordInput.value;
    if (!username || !password) {
      errorEl.textContent = 'Identifiant et mot de passe requis';
      errorEl.style.display = 'block';
      return;
    }
    submitBtn.disabled = true;
    submitBtn.textContent = 'Connexion...';
    const err = await doLogin(username, password);
    if (err) {
      errorEl.textContent = err;
      errorEl.style.display = 'block';
      submitBtn.disabled = false;
      submitBtn.textContent = '';
      submitBtn.appendChild(iconEl('check', 16));
      submitBtn.appendChild(document.createTextNode(' Se connecter'));
    }
  }

  submitBtn.onclick = handleLogin;
  passwordInput.onkeydown = (e) => { if (e.key === 'Enter') handleLogin(); };
  right.appendChild(submitBtn);

  wrapper.appendChild(left);
  wrapper.appendChild(right);
  container.appendChild(wrapper);

  usernameInput.focus();
}
