// app.js — SPA router and application bootstrap.
import { auth, setup as setupApi } from './api.js';
import { iconEl, el, notify } from './ui.js';

// ── State ──

export const state = {
  user: null, // current logged-in user
};

// ── Page registry ──

const pages = {};

export function registerPage(name, mod) {
  pages[name] = mod;
}

// ── Router ──

let currentPage = null;

async function route() {
  const hash = location.hash.slice(1) || 'dashboard';
  const content = document.getElementById('page-content');
  if (!content) return;

  // Check auth.
  if (hash !== 'login' && hash !== 'setup' && !state.user) {
    location.hash = '#login';
    return;
  }

  // Load page module.
  const page = pages[hash];
  if (!page) {
    content.textContent = 'Page introuvable';
    return;
  }

  currentPage = hash;
  content.textContent = '';

  // Update sidebar active state.
  document.querySelectorAll('.sb-item').forEach(item => {
    item.classList.toggle('active', item.dataset.page === hash);
  });

  // Update topbar title.
  const titleEl = document.getElementById('topbar-title');
  if (titleEl && page.title) titleEl.textContent = page.title;

  // Show/hide search bar.
  const searchEl = document.getElementById('topbar-search');
  if (searchEl) {
    const hideSearch = ['setup', 'login', 'settings', 'help', 'profile', 'new-patient', 'new-apt', 'users'];
    searchEl.style.display = hideSearch.includes(hash) ? 'none' : '';
  }

  // Render page.
  try {
    await page.render(content);
  } catch (e) {
    content.textContent = 'Erreur de chargement';
    console.error(e);
  }
}

// ── Boot ──

async function boot() {
  const app = document.getElementById('app');
  app.textContent = '';

  // Check if setup is done.
  const setupRes = await setupApi.status();
  if (setupRes.ok && !setupRes.data.setup_complete) {
    location.hash = '#setup';
    renderShell(app, false);
    route();
    return;
  }

  // Check if already logged in.
  const meRes = await auth.me();
  if (meRes.ok) {
    state.user = meRes.data;
    renderShell(app, true);
    if (!location.hash || location.hash === '#login' || location.hash === '#setup') {
      location.hash = '#dashboard';
    }
    route();
    return;
  }

  // Not logged in.
  location.hash = '#login';
  renderShell(app, false);
  route();
}

// ── Shell ──

function renderShell(app, withSidebar) {
  app.textContent = '';

  if (!withSidebar) {
    const content = el('div', { id: 'page-content' });
    app.appendChild(content);
    return;
  }

  const layout = el('div', { class: 'app visible' });

  // Sidebar.
  const sidebar = buildSidebar();
  layout.appendChild(sidebar);

  // Main area.
  const main = el('div', { class: 'main' });

  // Topbar.
  const topbar = buildTopbar();
  main.appendChild(topbar);

  // Content.
  const content = el('div', { class: 'content', id: 'page-content' });
  main.appendChild(content);

  layout.appendChild(main);
  app.appendChild(layout);
}

function buildSidebar() {
  const sidebar = el('aside', { class: 'sidebar' });

  // Brand.
  const brand = el('div', { class: 'sb-brand' });
  const logo = el('div', { class: 'sb-logo' });
  logo.appendChild(iconEl('leaf', 18));
  brand.appendChild(logo);
  const brandText = el('div', { class: 'sb-brand-text' });
  brandText.appendChild(el('h3', { text: 'MaSante' }));
  brandText.appendChild(el('span', { text: 'Plateforme de sante' }));
  brand.appendChild(brandText);
  sidebar.appendChild(brand);

  // Nav.
  const nav = el('nav', { class: 'sb-nav' });

  const sections = [
    { label: 'Principal', items: [
      { icon: 'chart', text: 'Tableau de bord', page: 'dashboard' },
      { icon: 'calendar', text: 'Calendrier RDV', page: 'calendar' },
      { icon: 'plus', text: 'Prise de RDV', page: 'new-apt' },
    ]},
    { label: 'Patients', items: [
      { icon: 'user-plus', text: 'Nouveau patient', page: 'new-patient' },
      { icon: 'users', text: 'Liste patients', page: 'patients' },
    ]},
    { label: 'Terrain', items: [
      { icon: 'bell', text: 'Rappels', page: 'reminders' },
    ]},
    { label: 'Systeme', items: [
      { icon: 'shield', text: 'Utilisateurs', page: 'users' },
      { icon: 'settings', text: 'Parametres', page: 'settings' },
      { icon: 'help', text: "Centre d'aide", page: 'help' },
    ]},
  ];

  sections.forEach(sec => {
    const section = el('div', { class: 'sb-section' });
    section.appendChild(el('div', { class: 'sb-section-label', text: sec.label }));
    sec.items.forEach(item => {
      const navItem = el('div', { class: 'sb-item', 'data-page': item.page });
      navItem.appendChild(iconEl(item.icon, 18));
      navItem.appendChild(document.createTextNode(' ' + item.text));
      navItem.onclick = () => { location.hash = '#' + item.page; };
      section.appendChild(navItem);
    });
    nav.appendChild(section);
  });

  sidebar.appendChild(nav);

  // User.
  const userSection = el('div', { class: 'sb-user' });
  const initials = state.user ? state.user.full_name.split(' ').map(w => w[0]).join('').substring(0, 2).toUpperCase() : '?';
  const avatarEl = el('div', { class: 'sb-avatar', text: initials, style: 'cursor:pointer' });
  avatarEl.onclick = () => { location.hash = '#profile'; };
  userSection.appendChild(avatarEl);

  const userInfo = el('div', { class: 'sb-user-info', style: 'cursor:pointer' });
  userInfo.onclick = () => { location.hash = '#profile'; };
  userInfo.appendChild(el('div', { class: 'name', text: state.user ? state.user.full_name : '' }));
  userInfo.appendChild(el('div', { class: 'role', text: state.user ? state.user.role : '' }));
  userSection.appendChild(userInfo);

  const logoutBtn = el('button', { class: 'icon-btn', title: 'Se deconnecter', style: 'color:rgba(255,255,255,.35);flex-shrink:0' });
  logoutBtn.appendChild(iconEl('logout', 18));
  logoutBtn.onclick = async () => {
    await auth.logout();
    state.user = null;
    location.hash = '#login';
    location.reload();
  };
  userSection.appendChild(logoutBtn);

  sidebar.appendChild(userSection);
  return sidebar;
}

function buildTopbar() {
  const topbar = el('header', { class: 'topbar' });
  topbar.appendChild(el('div', { class: 'topbar-title', id: 'topbar-title', text: 'Tableau de bord' }));

  // Search.
  const search = el('div', { class: 'topbar-search', id: 'topbar-search' });
  search.appendChild(iconEl('search', 16));
  const searchInput = el('input', { type: 'text', placeholder: 'Rechercher un patient...' });
  searchInput.onkeydown = (e) => {
    if (e.key === 'Enter' && searchInput.value.trim()) {
      location.hash = '#patients';
      // Will be picked up by patients page.
      sessionStorage.setItem('search_query', searchInput.value.trim());
    }
  };
  search.appendChild(searchInput);
  topbar.appendChild(search);

  const right = el('div', { class: 'topbar-right' });

  // Online badge.
  const badge = el('div', { class: 'topbar-badge online', id: 'connect-badge' });
  badge.appendChild(el('span', { class: 'bdot' }));
  badge.appendChild(el('span', { text: 'En ligne', id: 'connect-text' }));
  right.appendChild(badge);

  // Help button.
  const helpBtn = el('button', { class: 'icon-btn' });
  helpBtn.appendChild(iconEl('help', 18));
  helpBtn.onclick = () => { location.hash = '#help'; };
  right.appendChild(helpBtn);

  topbar.appendChild(right);
  return topbar;
}

// ── Login handling (called from login page) ──

export async function doLogin(username, password) {
  const res = await auth.login(username, password);
  if (!res.ok) {
    return res.error;
  }
  state.user = res.data;
  const app = document.getElementById('app');
  renderShell(app, true);
  location.hash = '#dashboard';
  route();
  return null;
}

// ── Setup complete (called from setup page) ──

export async function setupDone() {
  location.hash = '#login';
  location.reload();
}

// ── Listen for hash changes ──

window.addEventListener('hashchange', route);

// ── Load all pages then boot ──

async function loadPages() {
  const modules = [
    ['setup', () => import('./pages/setup.js')],
    ['login', () => import('./pages/login.js')],
    ['dashboard', () => import('./pages/dashboard.js')],
    ['patients', () => import('./pages/patients.js')],
    ['new-patient', () => import('./pages/new-patient.js')],
    ['patient-file', () => import('./pages/patient-file.js')],
    ['calendar', () => import('./pages/calendar.js')],
    ['new-apt', () => import('./pages/new-apt.js')],
    ['reminders', () => import('./pages/reminders.js')],
    ['users', () => import('./pages/users.js')],
    ['profile', () => import('./pages/profile.js')],
    ['settings', () => import('./pages/settings.js')],
    ['help', () => import('./pages/help.js')],
  ];
  for (const [name, loader] of modules) {
    const mod = await loader();
    registerPage(name, mod);
  }
}

loadPages().then(boot);
