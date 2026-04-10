// ui.js — Reusable UI components using safe DOM manipulation only.
// No innerHTML with dynamic content — all user data goes through textContent.

// ── Icons (safe SVG creation) ──

const NS = 'http://www.w3.org/2000/svg';

const ICON_PATHS = {
  leaf: ['M17 8C8 10 5.9 16.17 3.82 21.34l1.89.66.95-2.3c.48.17.98.3 1.34.3C19 20 22 3 22 3c-1 2-8 2.25-13 3.25S2 11.5 2 13.5s1.75 3.75 1.75 3.75'],
  calendar: ['M3 10h18', 'M16 2v4', 'M8 2v4'],
  users: ['M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2'],
  chart: ['M18 20V10', 'M12 20V4', 'M6 20v-6'],
  plus: ['M12 5v14', 'M5 12h14'],
  clipboard: ['M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2'],
  bell: ['M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9', 'M13.73 21a2 2 0 0 1-3.46 0'],
  search: ['M21 21l-4.35-4.35'],
  check: ['M20 6L9 17l-5-5'],
  x: ['M18 6L6 18', 'M6 6l12 12'],
  alert: ['M12 8v4', 'M12 16h.01'],
  clock: ['M12 6v6l4 2'],
  send: ['M22 2L11 13'],
  download: ['M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4', 'M7 10l5 5 5-5', 'M12 15V3'],
  logout: ['M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4', 'M16 17l5-5-5-5', 'M21 12H9'],
  settings: [],
  help: ['M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3', 'M12 17h.01'],
  shield: ['M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z'],
  user: ['M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2'],
  'user-plus': ['M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2', 'M20 8v6', 'M23 11h-6'],
  right: ['M9 18l6-6-6-6'],
  left: ['M15 18l-6-6 6-6'],
  up: ['M18 15l-6-6-6 6'],
  down: ['M6 9l6 6 6-6'],
  wifi: ['M5 12.55a11 11 0 0 1 14.08 0', 'M1.42 9a16 16 0 0 1 21.16 0', 'M8.53 16.11a6 6 0 0 1 6.95 0', 'M12 20h.01'],
  msg: ['M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z'],
  archive: ['M21 8v13H3V8', 'M1 3h22v5H1z', 'M10 12h4'],
  edit: ['M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7', 'M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z'],
  trash: ['M3 6h18', 'M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2'],
  ban: ['M4.93 4.93l14.14 14.14'],
  smartphone: ['M12 18h.01'],
  phone: ['M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6A19.79 19.79 0 0 1 2.12 4.11 2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72c.127.96.361 1.903.7 2.81a2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45c.907.339 1.85.573 2.81.7A2 2 0 0 1 22 16.92z'],
  map: ['M1 6l7-4 8 4 7-4v16l-7 4-8-4-7 4z', 'M8 2v16', 'M16 6v16'],
  book: ['M4 19.5A2.5 2.5 0 0 1 6.5 17H20', 'M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z'],
};

// Circles for icons that need them.
const ICON_CIRCLES = {
  search: [{ cx: 11, cy: 11, r: 8 }],
  alert: [{ cx: 12, cy: 12, r: 10 }],
  clock: [{ cx: 12, cy: 12, r: 10 }],
  help: [{ cx: 12, cy: 12, r: 10 }],
  user: [{ cx: 12, cy: 7, r: 4 }],
  'user-plus': [{ cx: 8.5, cy: 7, r: 4 }],
  users: [{ cx: 9, cy: 7, r: 4 }],
  ban: [{ cx: 12, cy: 12, r: 10 }],
};

// Rects.
const ICON_RECTS = {
  calendar: [{ x: 3, y: 4, w: 18, h: 18, rx: 2 }],
  clipboard: [{ x: 8, y: 2, w: 8, h: 4, rx: 1 }],
  smartphone: [{ x: 5, y: 2, w: 14, h: 20, rx: 2 }],
};

export function iconEl(name, size = 18) {
  const svg = document.createElementNS(NS, 'svg');
  svg.setAttribute('width', size);
  svg.setAttribute('height', size);
  svg.setAttribute('viewBox', '0 0 24 24');
  svg.setAttribute('fill', 'none');
  svg.setAttribute('stroke', 'currentColor');
  svg.setAttribute('stroke-width', '1.5');
  svg.setAttribute('stroke-linecap', 'round');
  svg.setAttribute('stroke-linejoin', 'round');

  const paths = ICON_PATHS[name] || [];
  paths.forEach(d => {
    const p = document.createElementNS(NS, 'path');
    p.setAttribute('d', d);
    svg.appendChild(p);
  });

  const circles = ICON_CIRCLES[name] || [];
  circles.forEach(c => {
    const el = document.createElementNS(NS, 'circle');
    el.setAttribute('cx', c.cx);
    el.setAttribute('cy', c.cy);
    el.setAttribute('r', c.r);
    svg.appendChild(el);
  });

  const rects = ICON_RECTS[name] || [];
  rects.forEach(r => {
    const el = document.createElementNS(NS, 'rect');
    el.setAttribute('x', r.x);
    el.setAttribute('y', r.y);
    el.setAttribute('width', r.w);
    el.setAttribute('height', r.h);
    if (r.rx) el.setAttribute('rx', r.rx);
    svg.appendChild(el);
  });

  return svg;
}

// ── Toasts ──

let toastContainer = null;

function ensureToastContainer() {
  if (toastContainer) return;
  toastContainer = document.createElement('div');
  toastContainer.id = 'toast-container';
  document.body.appendChild(toastContainer);
}

export function toast(message, type = 'info', duration = 4000) {
  ensureToastContainer();
  const el = document.createElement('div');
  el.className = 'toast toast-' + type;
  el.textContent = message;
  toastContainer.appendChild(el);
  requestAnimationFrame(() => el.classList.add('show'));
  setTimeout(() => {
    el.classList.remove('show');
    setTimeout(() => el.remove(), 300);
  }, duration);
}

export const notify = {
  success: (msg) => toast(msg, 'success'),
  error: (msg) => toast(msg, 'error'),
  warning: (msg) => toast(msg, 'warning'),
  info: (msg) => toast(msg, 'info'),
};

// ── Modal ──

export function openModal(title, contentEl, actions = []) {
  closeModal();
  const overlay = document.createElement('div');
  overlay.className = 'modal-overlay open';
  overlay.onclick = closeModal;

  const modal = document.createElement('div');
  modal.className = 'modal open';
  modal.onclick = (e) => e.stopPropagation();

  const head = document.createElement('div');
  head.className = 'modal-head';
  const h3 = document.createElement('h3');
  h3.textContent = title;
  head.appendChild(h3);
  const closeBtn = document.createElement('button');
  closeBtn.className = 'icon-btn';
  closeBtn.appendChild(iconEl('x'));
  closeBtn.onclick = closeModal;
  head.appendChild(closeBtn);

  const body = document.createElement('div');
  body.className = 'modal-body';
  body.appendChild(contentEl);

  modal.appendChild(head);
  modal.appendChild(body);

  if (actions.length > 0) {
    const footer = document.createElement('div');
    footer.className = 'modal-footer';
    actions.forEach(({ label, cls, onclick }) => {
      const btn = document.createElement('button');
      btn.className = 'btn ' + (cls || 'btn-secondary');
      btn.textContent = label;
      btn.style.width = 'auto';
      btn.onclick = onclick;
      footer.appendChild(btn);
    });
    modal.appendChild(footer);
  }

  document.body.appendChild(overlay);
  document.body.appendChild(modal);
}

export function closeModal() {
  document.querySelectorAll('.modal-overlay, .modal').forEach(el => el.remove());
}

// ── Safe DOM helpers ──

export function el(tag, attrs = {}, children = []) {
  const node = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs)) {
    if (k === 'class') node.className = v;
    else if (k === 'text') node.textContent = v;
    else if (k === 'onclick') node.onclick = v;
    else if (k === 'style') node.style.cssText = v;
    else node.setAttribute(k, v);
  }
  children.forEach(child => {
    if (typeof child === 'string') {
      node.appendChild(document.createTextNode(child));
    } else if (child) {
      node.appendChild(child);
    }
  });
  return node;
}

// ── Formatting ──

export function formatDate(dateStr) {
  if (!dateStr) return '\u2014';
  const d = new Date(dateStr);
  if (isNaN(d)) return dateStr;
  return d.toLocaleDateString('fr-FR', { day: '2-digit', month: 'short', year: 'numeric' });
}

export function statusPill(status) {
  const map = {
    active: ['pill-success', 'Actif'],
    a_surveiller: ['pill-warning', 'A surveiller'],
    perdu_de_vue: ['pill-danger', 'Perdu de vue'],
    sorti: ['pill-neutral', 'Sorti'],
    confirme: ['pill-success', 'Confirme'],
    en_attente: ['pill-warning', 'En attente'],
    termine: ['pill-info', 'Termine'],
    manque: ['pill-danger', 'Manque'],
    annule: ['pill-neutral', 'Annule'],
    reporte: ['pill-warning', 'Reporte'],
    planifie: ['pill-warning', 'Planifie'],
    envoye: ['pill-success', 'Envoye'],
    recu: ['pill-info', 'Recu'],
    echec: ['pill-danger', 'Echec'],
    conge: ['pill-neutral', 'Conge'],
    desactive: ['pill-danger', 'Desactive'],
  };
  const [cls, label] = map[status] || ['pill-neutral', status || '—'];
  return el('span', { class: 'pill ' + cls, text: label });
}

export function riskBadge(score) {
  let cls, label;
  if (score <= 3) { cls = 'low'; label = 'Faible'; }
  else if (score <= 6) { cls = 'med'; label = 'Moyen'; }
  else { cls = 'high'; label = 'Eleve'; }
  return el('span', { class: 'risk ' + cls }, [
    el('span', { class: 'risk-dot' }),
    document.createTextNode(' ' + label),
  ]);
}

export function avatar(initials, variant = 'a1') {
  return el('div', { class: 'avatar ' + variant, text: initials });
}

// ── Loading / Empty ──

export function loading(container) {
  container.textContent = '';
  container.appendChild(el('div', {
    style: 'text-align:center;padding:40px;color:var(--gray-300)',
    text: 'Chargement...',
  }));
}

export function empty(container, message = 'Aucun element') {
  container.textContent = '';
  container.appendChild(el('div', {
    style: 'text-align:center;padding:40px;color:var(--gray-300)',
    text: message,
  }));
}
