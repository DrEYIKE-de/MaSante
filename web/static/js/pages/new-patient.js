// new-patient.js — Patient registration form, connected to API.
import { patients } from '../api.js';
import { el, iconEl, notify } from '../ui.js';

export const title = 'Nouveau patient';

export async function render(container) {
  const grid = el('div', { class: 'apt-grid' });

  // Left column — identity.
  const leftCard = el('div', { class: 'card' });
  const leftHead = el('div', { class: 'card-head' });
  const leftTitle = el('h3');
  leftTitle.appendChild(iconEl('user-plus', 18));
  leftTitle.appendChild(document.createTextNode(' Inscription'));
  leftHead.appendChild(leftTitle);
  leftCard.appendChild(leftHead);

  const leftBody = el('div', { class: 'card-body' });
  const formGrid = el('div', { style: 'display:grid;grid-template-columns:1fr 1fr;gap:14px' });

  const fields = {
    last_name: input('Nom', 'text', 'Nom de famille'),
    first_name: input('Prenom', 'text', 'Prenom'),
    date_of_birth: input('Date de naissance', 'date', ''),
    sex: select('Sexe', ['', 'M', 'F'], ['Selectionner', 'Masculin', 'Feminin']),
    phone: input('Telephone', 'tel', '+237 6XX XXX XXX'),
    phone_secondary: input('Telephone secondaire', 'tel', 'Optionnel'),
    district: input('Quartier / Zone', 'text', 'Ex: Akwa'),
    address: input('Adresse / Repere', 'text', 'Description du lieu'),
  };

  Object.values(fields).forEach(f => formGrid.appendChild(f.group));
  leftBody.appendChild(formGrid);

  // Language.
  const langGroup = el('div', { class: 'form-group', style: 'margin-top:14px' });
  langGroup.appendChild(el('label', { text: 'Langue preferee' }));
  const langOpts = el('div', { class: 'remind-opts' });
  let selectedLang = 'fr';
  ['Francais', 'Anglais', 'Duala', 'Ewondo', 'Bamileke'].forEach((l, i) => {
    const codes = ['fr', 'en', 'duala', 'ewondo', 'bamileke'];
    const opt = el('div', { class: 'r-opt' + (i === 0 ? ' on' : ''), text: l });
    opt.onclick = () => {
      langOpts.querySelectorAll('.r-opt').forEach(o => o.classList.remove('on'));
      opt.classList.add('on');
      selectedLang = codes[i];
    };
    langOpts.appendChild(opt);
  });
  langGroup.appendChild(langOpts);
  leftBody.appendChild(langGroup);

  leftCard.appendChild(leftBody);
  grid.appendChild(el('div', {}, [leftCard]));

  // Right column — contact + medical.
  const rightCol = el('div');

  // Contact.
  const contactCard = el('div', { class: 'card', style: 'margin-bottom:16px' });
  const contactHead = el('div', { class: 'card-head' });
  const contactTitle = el('h3');
  contactTitle.appendChild(iconEl('msg', 18));
  contactTitle.appendChild(document.createTextNode(' Contact & rappels'));
  contactHead.appendChild(contactTitle);
  contactCard.appendChild(contactHead);

  const contactBody = el('div', { class: 'card-body' });
  const channelGroup = el('div', { class: 'form-group' });
  channelGroup.appendChild(el('label', { text: 'Canal de rappel' }));
  const channelOpts = el('div', { class: 'remind-opts' });
  let selectedChannel = 'sms';
  [['SMS', 'sms', 'msg'], ['WhatsApp', 'whatsapp', 'smartphone'], ['Appel vocal', 'voice', 'phone'], ['Aucun', 'none', 'ban']].forEach(([label, code, ic], i) => {
    const opt = el('div', { class: 'r-opt' + (i === 0 ? ' on' : '') });
    opt.appendChild(iconEl(ic, 16));
    opt.appendChild(document.createTextNode(' ' + label));
    opt.onclick = () => {
      channelOpts.querySelectorAll('.r-opt').forEach(o => o.classList.remove('on'));
      opt.classList.add('on');
      selectedChannel = code;
    };
    channelOpts.appendChild(opt);
  });
  channelGroup.appendChild(channelOpts);
  contactBody.appendChild(channelGroup);

  const contactFields = {
    contact_name: input('Personne de confiance', 'text', 'Nom du contact'),
    contact_phone: input('Telephone contact', 'tel', '+237 6XX XXX XXX'),
    contact_relation: select('Lien', ['', 'Conjoint(e)', 'Parent', 'Frere/Soeur', 'Ami(e)', 'Autre']),
  };
  Object.values(contactFields).forEach(f => contactBody.appendChild(f.group));
  contactCard.appendChild(contactBody);
  rightCol.appendChild(contactCard);

  // Medical.
  const medCard = el('div', { class: 'card', style: 'margin-bottom:16px' });
  const medHead = el('div', { class: 'card-head' });
  const medTitle = el('h3');
  medTitle.appendChild(iconEl('clipboard', 18));
  medTitle.appendChild(document.createTextNode(' Informations medicales'));
  medHead.appendChild(medTitle);
  medCard.appendChild(medHead);

  const medBody = el('div', { class: 'card-body' });
  const medFields = {
    referred_by: select('Refere par', ['', 'Centre de depistage', 'Transfert', 'Auto-presentation', 'Agent communautaire', 'Autre']),
  };
  Object.values(medFields).forEach(f => medBody.appendChild(f.group));
  medCard.appendChild(medBody);
  rightCol.appendChild(medCard);

  // Error.
  const errorEl = el('div', { style: 'display:none;padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin-bottom:12px' });
  rightCol.appendChild(errorEl);

  // Submit.
  const submitBtn = el('button', { class: 'btn btn-primary' });
  submitBtn.appendChild(iconEl('check', 16));
  submitBtn.appendChild(document.createTextNode(' Inscrire le patient'));
  submitBtn.onclick = async () => {
    errorEl.style.display = 'none';

    const body = {
      last_name: fields.last_name.value(),
      first_name: fields.first_name.value(),
      date_of_birth: fields.date_of_birth.value(),
      sex: fields.sex.value(),
      phone: fields.phone.value(),
      phone_secondary: fields.phone_secondary.value(),
      district: fields.district.value(),
      address: fields.address.value(),
      language: selectedLang,
      reminder_channel: selectedChannel,
      contact_name: contactFields.contact_name.value(),
      contact_phone: contactFields.contact_phone.value(),
      contact_relation: contactFields.contact_relation.value(),
      referred_by: medFields.referred_by.value(),
    };

    if (!body.last_name || !body.first_name || !body.sex) {
      errorEl.textContent = 'Nom, prenom et sexe sont requis';
      errorEl.style.display = 'block';
      return;
    }

    submitBtn.disabled = true;
    submitBtn.textContent = 'Inscription...';
    const res = await patients.create(body);
    if (!res.ok) {
      errorEl.textContent = res.error;
      errorEl.style.display = 'block';
      submitBtn.disabled = false;
      submitBtn.textContent = '';
      submitBtn.appendChild(iconEl('check', 16));
      submitBtn.appendChild(document.createTextNode(' Inscrire le patient'));
      return;
    }

    notify.success('Patient inscrit — Code: ' + (res.data.Code || ''));
    location.hash = '#patients';
  };
  rightCol.appendChild(submitBtn);

  grid.appendChild(rightCol);
  container.appendChild(grid);
}

function input(label, type, placeholder) {
  const group = el('div', { class: 'form-group' });
  group.appendChild(el('label', { text: label }));
  const inp = el('input', { class: 'form-input', type, placeholder: placeholder || '' });
  group.appendChild(inp);
  return { group, value: () => inp.value.trim() };
}

function select(label, values, labels) {
  const group = el('div', { class: 'form-group' });
  group.appendChild(el('label', { text: label }));
  const sel = el('select', { class: 'form-input' });
  values.forEach((v, i) => {
    sel.appendChild(el('option', { value: v, text: (labels && labels[i]) || v }));
  });
  group.appendChild(sel);
  return { group, value: () => sel.value };
}
