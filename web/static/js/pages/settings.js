import { el, iconEl } from '../ui.js';
import { exports as exp } from '../api.js';
export const title = 'Parametres';
export async function render(container) {
  const grid = el('div', { class: 'grid-2-wide', style: 'align-items:start' });

  // Exports.
  const card1 = el('div', { class: 'card' });
  const head1 = el('div', { class: 'card-head' });
  const t1 = el('h3');
  t1.appendChild(iconEl('chart', 18));
  t1.appendChild(document.createTextNode(' Rapports et exports'));
  head1.appendChild(t1);
  card1.appendChild(head1);
  const body1 = el('div', { class: 'card-body' });

  const lastMonth = new Date();
  lastMonth.setMonth(lastMonth.getMonth() - 1);
  const monthStr = lastMonth.toISOString().slice(0, 7);

  [['Rapport mensuel — ' + monthStr, exp.monthlyExcel(monthStr), exp.monthlyPdf(monthStr)],
   ['Liste patients actifs', exp.patientsExcel('active'), exp.patientsPdf('active')],
   ['Patients perdus de vue', exp.patientsExcel('perdu_de_vue'), exp.patientsPdf('perdu_de_vue')]].forEach(([label, excelUrl, pdfUrl]) => {
    const item = el('div', { class: 'list-item' });
    item.appendChild(el('div', { style: 'flex:1', text: label }));
    const acts = el('div', { class: 'pt-acts' });

    const excelBtn = el('a', { class: 'btn btn-sm btn-secondary', href: excelUrl, style: 'width:auto;text-decoration:none' });
    excelBtn.appendChild(iconEl('download', 14));
    excelBtn.appendChild(document.createTextNode(' Excel'));
    acts.appendChild(excelBtn);

    const pdfBtn = el('a', { class: 'btn btn-sm btn-secondary', href: pdfUrl, style: 'width:auto;text-decoration:none' });
    pdfBtn.appendChild(iconEl('download', 14));
    pdfBtn.appendChild(document.createTextNode(' PDF'));
    acts.appendChild(pdfBtn);

    item.appendChild(acts);
    body1.appendChild(item);
  });
  card1.appendChild(body1);
  grid.appendChild(card1);

  // Placeholder — team link.
  const card2 = el('div', { class: 'card' });
  const head2 = el('div', { class: 'card-head' });
  const t2 = el('h3');
  t2.appendChild(iconEl('users', 18));
  t2.appendChild(document.createTextNode(' Equipe soignante'));
  head2.appendChild(t2);
  const link = el('a', { class: 'card-link', text: 'Gerer les utilisateurs' });
  link.onclick = () => { location.hash = '#users'; };
  head2.appendChild(link);
  card2.appendChild(head2);
  const body2 = el('div', { class: 'card-body' });
  body2.appendChild(el('div', { style: 'text-align:center;padding:20px;color:var(--gray-300)', text: 'Accedez a la gestion des utilisateurs pour ajouter ou modifier des comptes.' }));
  card2.appendChild(body2);
  grid.appendChild(card2);

  container.appendChild(grid);
}
