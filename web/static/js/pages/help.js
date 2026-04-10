import { el, iconEl } from '../ui.js';
export const title = "Centre d'aide";
export async function render(container) {
  // Header.
  const header = el('div', { style: 'text-align:center;margin-bottom:28px' });
  header.appendChild(el('h3', { text: 'Comment pouvons-nous vous aider ?', style: 'font-size:1.4rem;margin-bottom:16px' }));
  container.appendChild(header);

  // Guides.
  const guidesTitle = el('h4', { style: 'font-size:.95rem;margin-bottom:12px' });
  guidesTitle.appendChild(iconEl('book', 18));
  guidesTitle.appendChild(document.createTextNode(' Guides de demarrage'));
  container.appendChild(guidesTitle);

  const guides = el('div', { class: 'grid-2', style: 'align-items:start;margin-bottom:24px' });

  const guideData = [
    ['user-plus', 'green', 'Inscrire un nouveau patient', [
      '1. Cliquez sur Nouveau patient dans la barre laterale',
      '2. Remplissez les informations: nom, prenom, telephone, quartier',
      '3. Selectionnez la langue preferee du patient',
      '4. Ajoutez un contact de confiance',
      '5. Cliquez sur Inscrire le patient — un code unique sera genere',
    ]],
    ['calendar', 'blue', 'Programmer un rendez-vous', [
      '1. Allez dans Prise de RDV',
      '2. Recherchez le patient par nom ou code',
      '3. Selectionnez le type de visite',
      '4. Choisissez la date et un creneau disponible',
      '5. Cliquez sur Confirmer le rendez-vous',
    ]],
    ['clipboard', 'amber', 'Gerer un rendez-vous', [
      'Depuis le dashboard: cliquez sur le RDV dans la liste du jour',
      'Termine: ajoutez vos notes et programmez le prochain RDV',
      'Manque: choisissez une action (rappel SMS, reprogrammer)',
      'Reporte: selectionnez une nouvelle date',
    ]],
    ['bell', 'red', 'Configurer les rappels SMS', [
      '1. Allez dans Parametres ou Rappels',
      '2. Configurez votre fournisseur SMS (cle API)',
      '3. Activez les rappels J-7, J-2, jour J',
      '4. Les rappels sont envoyes automatiquement',
    ]],
  ];

  guideData.forEach(([icon, color, title, steps]) => {
    const card = el('div', { class: 'card', style: 'cursor:pointer' });
    const body = el('div', { class: 'card-body', style: 'padding:18px 20px' });

    const header = el('div', { style: 'display:flex;align-items:center;gap:12px' });
    const iconWrap = el('div', { class: 'stat-icon ' + color, style: 'width:36px;height:36px;margin-bottom:0' });
    iconWrap.appendChild(iconEl(icon, 18));
    header.appendChild(iconWrap);
    header.appendChild(el('div', { text: title, style: 'font-weight:600;font-size:.9rem' }));
    const chevron = iconEl('down', 16);
    chevron.style.cssText = 'margin-left:auto;color:var(--gray-300);transition:.2s';
    header.appendChild(chevron);
    body.appendChild(header);

    const content = el('div', { style: 'display:none;padding-top:10px;border-top:1px solid var(--gray-50);margin-top:10px;font-size:.85rem;color:var(--gray-500);line-height:1.7' });
    steps.forEach(s => content.appendChild(el('p', { text: s })));
    body.appendChild(content);

    card.onclick = () => {
      const open = content.style.display === 'block';
      content.style.display = open ? 'none' : 'block';
      chevron.style.transform = open ? '' : 'rotate(180deg)';
    };

    card.appendChild(body);
    guides.appendChild(card);
  });

  container.appendChild(guides);

  // FAQ.
  const faqTitle = el('h4', { style: 'font-size:.95rem;margin-bottom:12px' });
  faqTitle.appendChild(iconEl('help', 18));
  faqTitle.appendChild(document.createTextNode(' Questions frequentes'));
  container.appendChild(faqTitle);

  const faqGrid = el('div', { class: 'grid-2-wide', style: 'align-items:start' });
  const faqs = [
    ['Les messages mentionnent-ils la maladie ?', 'Jamais. Tous les messages sont generiques: "rappel de votre rendez-vous sante". Aucune mention de pathologie.'],
    ['MaSante fonctionne-t-il sans internet ?', 'Oui. Toutes les fonctions fonctionnent hors ligne. Seuls les rappels SMS necessitent une connexion.'],
    ['Comment ajouter un nouvel utilisateur ?', 'Allez dans Utilisateurs (menu Systeme), cliquez sur Ajouter, remplissez les infos et attribuez un role.'],
    ['Comment declarer un deces ?', 'Ouvrez la fiche du patient, cliquez sur Sortie du programme, selectionnez le motif.'],
  ];

  const faqCard = el('div', { class: 'card' });
  const faqBody = el('div', { class: 'card-body' });
  faqs.forEach(([q, a]) => {
    const item = el('div', { style: 'padding:12px 0;border-bottom:1px solid var(--gray-50);cursor:pointer' });
    const question = el('div', { style: 'display:flex;align-items:center;justify-content:space-between;font-size:.88rem;font-weight:600;color:var(--gray-700)' });
    question.appendChild(el('span', { text: q }));
    const ch = iconEl('down', 14);
    ch.style.cssText = 'flex-shrink:0;color:var(--gray-300);transition:.2s';
    question.appendChild(ch);
    item.appendChild(question);

    const answer = el('div', { style: 'display:none;padding-top:8px;font-size:.82rem;color:var(--gray-500);line-height:1.6', text: a });
    item.appendChild(answer);

    item.onclick = () => {
      const open = answer.style.display === 'block';
      answer.style.display = open ? 'none' : 'block';
      ch.style.transform = open ? '' : 'rotate(180deg)';
    };
    faqBody.appendChild(item);
  });
  faqCard.appendChild(faqBody);
  faqGrid.appendChild(faqCard);
  container.appendChild(faqGrid);

  // Version.
  container.appendChild(el('div', { style: 'text-align:center;margin-top:24px;padding:16px;color:var(--gray-300);font-size:.78rem', text: 'MaSante v1.0.0 — Logiciel libre et gratuit — masante.africa' }));
}
