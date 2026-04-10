import{setup as setupApi,auth,dash,pts,apts,cal,rem,usr,prof}from'./api.js';

// ── Helpers ──
function $(s,p){return(p||document).querySelector(s)}
function $$(s,p){return[...(p||document).querySelectorAll(s)]}
function el(tag,attrs,kids){const n=document.createElement(tag);if(attrs)for(const[k,v]of Object.entries(attrs)){if(k==='text')n.textContent=v;else if(k==='class')n.className=v;else if(k==='style')n.style.cssText=v;else if(k.startsWith('on'))n[k]=v;else n.setAttribute(k,v)}if(kids)(Array.isArray(kids)?kids:kids instanceof NodeList?[...kids]:[kids]).forEach(c=>{if(typeof c==='string')n.appendChild(document.createTextNode(c));else if(c)n.appendChild(c)});return n}
function svg(name,size){const s=document.createElementNS('http://www.w3.org/2000/svg','svg');s.setAttribute('width',size||18);s.setAttribute('height',size||18);s.setAttribute('viewBox','0 0 24 24');s.setAttribute('fill','none');s.setAttribute('stroke','currentColor');s.setAttribute('stroke-width','1.5');s.setAttribute('stroke-linecap','round');s.setAttribute('stroke-linejoin','round');const u=document.createElementNS('http://www.w3.org/2000/svg','use');u.setAttributeNS('http://www.w3.org/1999/xlink','href','#i-'+name);s.appendChild(u);return s}
function toast(msg,type){const c=$('#ms-toasts');const t=el('div',{class:'ms-toast ms-toast-'+(type||'info'),text:msg});c.appendChild(t);requestAnimationFrame(()=>t.classList.add('show'));setTimeout(()=>{t.classList.remove('show');setTimeout(()=>t.remove(),300)},4000)}
function pill(status){const m={active:['pill-success','Actif'],a_surveiller:['pill-warning','A surveiller'],perdu_de_vue:['pill-danger','Perdu de vue'],sorti:['pill-neutral','Sorti'],confirme:['pill-success','Confirme'],en_attente:['pill-warning','En attente'],termine:['pill-info','Termine'],manque:['pill-danger','Manque'],annule:['pill-neutral','Annule'],reporte:['pill-warning','Reporte'],planifie:['pill-warning','Planifie'],envoye:['pill-success','Envoye'],recu:['pill-info','Recu'],echec:['pill-danger','Echec'],conge:['pill-neutral','Conge'],desactive:['pill-danger','Desactive']};const[cls,lbl]=m[status]||['pill-neutral',status||'—'];return el('span',{class:'pill '+cls,text:lbl})}
function risk(score){const s=score||5;let cls='low',lbl='Faible';if(s>6){cls='high';lbl='Eleve'}else if(s>3){cls='med';lbl='Moyen'}const sp=el('span',{class:'risk '+cls});sp.appendChild(el('span',{class:'risk-dot'}));sp.appendChild(document.createTextNode(' '+lbl));return sp}
function fmtDate(d){if(!d)return'—';const dt=new Date(d);return isNaN(dt)?d:dt.toLocaleDateString('fr-FR',{day:'2-digit',month:'short',year:'numeric'})}
function fmtType(t){return{consultation:'Consultation',retrait_medicaments:'Retrait medicaments',bilan_sanguin:'Bilan sanguin',club_adherence:"Club d'adherence"}[t]||t||''}
function initials(name){return(name||'??').split(' ').map(w=>w[0]).join('').substring(0,2).toUpperCase()}
function loading(){return el('div',{class:'ms-loading',text:'Chargement...'})}
function empty(msg){return el('div',{class:'ms-empty',text:msg||'Aucun element'})}

// ── State ──
let user=null;
let currentPage='';

// ── Router ──
const pages={
  setup:{title:'Configuration',render:pageSetup},
  login:{title:'Connexion',render:pageLogin},
  dashboard:{title:'Tableau de bord',render:pageDashboard},
  calendar:{title:'Calendrier',render:pageCalendar},
  'new-apt':{title:'Prise de RDV',render:pageNewApt},
  'new-patient':{title:'Nouveau patient',render:pageNewPatient},
  patients:{title:'Liste patients',render:pagePatients},
  'patient-file':{title:'Fiche patient',render:pagePatientFile},
  reminders:{title:'Rappels',render:pageReminders},
  users:{title:'Utilisateurs',render:pageUsers},
  profile:{title:'Mon profil',render:pageProfile},
  settings:{title:'Parametres',render:pageSettings},
  help:{title:"Centre d'aide",render:pageHelp},
};

async function route(){
  const hash=location.hash.slice(1)||'dashboard';
  if(hash!=='login'&&hash!=='setup'&&!user){location.hash='#login';return}
  const page=pages[hash];if(!page)return;
  currentPage=hash;
  const content=$('#page-content');if(!content)return;
  content.textContent='';
  // Update sidebar.
  $$('.sb-item').forEach(i=>i.classList.toggle('active',i.dataset.page===hash));
  const title=$('#topbar-title');if(title)title.textContent=page.title;
  const search=$('#topbar-search');
  if(search)search.style.display=['setup','login','settings','help','profile','new-patient','new-apt','users'].includes(hash)?'none':'';
  try{await page.render(content)}catch(e){content.textContent='Erreur';console.error(e)}
}
window.addEventListener('hashchange',route);

// ── Boot ──
async function boot(){
  const app=$('#app');app.textContent='';
  const res=await setupApi.status();
  if(res.ok&&!res.data.setup_complete){renderFullPage(app);location.hash='#setup';route();return}
  const me=await auth.me();
  if(me.ok){user=me.data;renderShell(app);if(!location.hash||location.hash==='#login'||location.hash==='#setup')location.hash='#dashboard';route();return}
  renderFullPage(app);location.hash='#login';route();
}

function renderFullPage(app){app.textContent='';app.appendChild(el('div',{id:'page-content'}))}

function renderShell(app){
  app.textContent='';
  const layout=el('div',{class:'app visible'});
  layout.appendChild(buildSidebar());
  const main=el('div',{class:'main'});
  main.appendChild(buildTopbar());
  main.appendChild(el('div',{class:'content',id:'page-content'}));
  layout.appendChild(main);
  app.appendChild(layout);
}

function buildSidebar(){
  const sb=el('aside',{class:'sidebar'});
  // Brand.
  const brand=el('div',{class:'sb-brand'});
  const logo=el('div',{class:'sb-logo'});logo.appendChild(svg('leaf',18));brand.appendChild(logo);
  const bt=el('div',{class:'sb-brand-text'});bt.appendChild(el('h3',{text:'MaSante'}));bt.appendChild(el('span',{text:'Plateforme de sante'}));brand.appendChild(bt);
  sb.appendChild(brand);
  // Nav.
  const nav=el('nav',{class:'sb-nav'});
  const sections=[
    ['Principal',[['chart','Tableau de bord','dashboard'],['calendar','Calendrier RDV','calendar'],['plus','Prise de RDV','new-apt']]],
    ['Patients',[['user-plus','Nouveau patient','new-patient'],['users','Liste patients','patients']]],
    ['Terrain',[['bell','Rappels','reminders']]],
    ['Systeme',[['shield','Utilisateurs','users'],['settings','Parametres','settings'],['help',"Centre d'aide",'help']]],
  ];
  sections.forEach(([label,items])=>{
    const sec=el('div',{class:'sb-section'});
    sec.appendChild(el('div',{class:'sb-section-label',text:label}));
    items.forEach(([ic,text,page])=>{
      const item=el('div',{class:'sb-item','data-page':page});
      item.appendChild(svg(ic,18));item.appendChild(document.createTextNode(' '+text));
      item.onclick=()=>{location.hash='#'+page};
      sec.appendChild(item);
    });
    nav.appendChild(sec);
  });
  sb.appendChild(nav);
  // User.
  const uSec=el('div',{class:'sb-user'});
  const av=el('div',{class:'sb-avatar',text:user?initials(user.full_name):'?',style:'cursor:pointer'});
  av.onclick=()=>{location.hash='#profile'};uSec.appendChild(av);
  const ui=el('div',{class:'sb-user-info',style:'cursor:pointer',onclick:()=>{location.hash='#profile'}});
  ui.appendChild(el('div',{class:'name',text:user?user.full_name:''}));
  ui.appendChild(el('div',{class:'role',text:user?user.role:''}));
  uSec.appendChild(ui);
  const logoutBtn=el('button',{class:'icon-btn',title:'Deconnecter',style:'color:rgba(255,255,255,.35);flex-shrink:0'});
  logoutBtn.appendChild(svg('logout',18));
  logoutBtn.onclick=async()=>{await auth.logout();user=null;location.hash='#login';location.reload()};
  uSec.appendChild(logoutBtn);
  sb.appendChild(uSec);
  return sb;
}

function buildTopbar(){
  const tb=el('header',{class:'topbar'});
  tb.appendChild(el('div',{class:'topbar-title',id:'topbar-title',text:'Tableau de bord'}));
  const search=el('div',{class:'topbar-search',id:'topbar-search'});
  search.appendChild(svg('search',16));
  const si=el('input',{type:'text',placeholder:'Rechercher un patient...'});
  si.onkeydown=e=>{if(e.key==='Enter'&&si.value.trim()){window._searchQuery=si.value.trim();location.hash='#patients'}};
  search.appendChild(si);tb.appendChild(search);
  const right=el('div',{class:'topbar-right'});
  const badge=el('div',{class:'topbar-badge online'});badge.appendChild(el('span',{class:'bdot'}));badge.appendChild(el('span',{text:'En ligne'}));
  right.appendChild(badge);
  const helpBtn=el('button',{class:'icon-btn'});helpBtn.appendChild(svg('help',18));helpBtn.onclick=()=>{location.hash='#help'};right.appendChild(helpBtn);
  tb.appendChild(right);return tb;
}

// ══════════════════════════════
// PAGES
// ══════════════════════════════

// ── SETUP ──
async function pageSetup(c){
  let step=1;const total=5;const data={center:{},admin:{},schedule:{},sms:{}};
  const res=await setupApi.status();
  if(res.ok&&res.data.setup_complete){location.hash='#login';location.reload();return}
  if(res.ok&&res.data.current_step>0)step=res.data.current_step+1;
  if(step>total)step=total;

  const wrap=el('div',{style:'height:100vh;display:flex;flex-direction:column;background:var(--white);margin:-24px'});
  // Header.
  const header=el('div',{class:'setup-header'});
  const slogo=el('div',{class:'sh-logo'});slogo.appendChild(svg('leaf',16));header.appendChild(slogo);
  header.appendChild(el('h2',{text:'Configuration de MaSante'}));
  const stepLbl=el('span',{class:'sh-step',text:'Etape '+step+' sur '+total});header.appendChild(stepLbl);
  wrap.appendChild(header);
  const progBg=el('div',{class:'setup-progress'});
  const progFill=el('div',{class:'setup-progress-fill',style:'width:'+(step/total*100)+'%'});
  progBg.appendChild(progFill);wrap.appendChild(progBg);
  // Body.
  const body=el('div',{class:'setup-body'});
  const content=el('div',{class:'setup-content'});body.appendChild(content);wrap.appendChild(body);
  const errEl=el('div',{style:'display:none;padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin:0 40px 10px'});
  wrap.appendChild(errEl);
  // Footer.
  const footer=el('div',{class:'setup-footer'});
  const prevBtn=el('button',{class:'btn btn-secondary',style:'width:auto;visibility:'+(step===1?'hidden':'visible')});
  prevBtn.appendChild(svg('left',14));prevBtn.appendChild(document.createTextNode(' Precedent'));
  const nextBtn=el('button',{class:'btn btn-primary',style:'width:auto'});
  function updateNext(){nextBtn.textContent='';if(step===total){nextBtn.appendChild(document.createTextNode('Lancer MaSante '));nextBtn.appendChild(svg('check',14))}else{nextBtn.appendChild(document.createTextNode('Suivant '));nextBtn.appendChild(svg('right',14))}}
  updateNext();

  prevBtn.onclick=()=>{if(step<=1)return;step--;renderStep();stepLbl.textContent='Etape '+step+' sur '+total;progFill.style.width=(step/total*100)+'%';prevBtn.style.visibility=step===1?'hidden':'visible';updateNext();errEl.style.display='none'};

  nextBtn.onclick=async()=>{
    errEl.style.display='none';nextBtn.disabled=true;
    const err=await saveStep();
    nextBtn.disabled=false;
    if(err){errEl.textContent=err;errEl.style.display='block';return}
    step++;
    if(step>total){location.hash='#login';location.reload();return}
    renderStep();stepLbl.textContent='Etape '+step+' sur '+total;progFill.style.width=(step/total*100)+'%';prevBtn.style.visibility='visible';updateNext()
  };

  footer.appendChild(prevBtn);footer.appendChild(nextBtn);wrap.appendChild(footer);
  c.appendChild(wrap);renderStep();

  function renderStep(){
    content.textContent='';
    if(step===1)renderStep1();else if(step===2)renderStep2();else if(step===3)renderStep3();else if(step===4)renderStep4();else renderStep5();
  }

  function input(lbl,type,id,ph,full){const g=el('div',{class:'form-group'+(full?' setup-full':'')});g.appendChild(el('label',{text:lbl}));g.appendChild(el('input',{class:'form-input',type:type,id:id,placeholder:ph||''}));return g}
  function sel(lbl,id,opts,def){const g=el('div',{class:'form-group'});g.appendChild(el('label',{text:lbl}));const s=el('select',{class:'form-input',id:id});opts.forEach(o=>{const op=el('option',{value:o,text:o});if(o===def)op.selected=true;s.appendChild(op)});g.appendChild(s);return g}

  function renderStep1(){
    content.appendChild(el('h3',{text:'Votre etablissement'}));
    content.appendChild(el('p',{class:'setup-desc',text:'Ces informations identifient votre centre de sante.'}));
    const g=el('div',{class:'setup-grid'});
    g.appendChild(input('Nom','text','s-name','Ex: Hopital Laquintinie',true));
    const types=el('div',{class:'type-options',id:'s-types',style:'margin:6px 0 14px'});
    ['Hopital public','Centre de sante','Clinique privee'].forEach((t,i)=>{
      const b=el('div',{class:'type-opt'+(i===1?' on':''),text:t});
      b.onclick=()=>{$$('.type-opt',types).forEach(x=>x.classList.remove('on'));b.classList.add('on')};
      types.appendChild(b)});
    g.appendChild(types);
    g.appendChild(sel('Pays','s-country',['Cameroun',"Cote d'Ivoire",'RD Congo','Senegal','Tchad','Gabon','Congo','Burkina Faso','Mali','Niger','Guinee','Togo','Benin','Rwanda','Kenya','Madagascar'],'Cameroun'));
    g.appendChild(input('Ville','text','s-city','Ex: Douala'));
    g.appendChild(input('Quartier','text','s-district','Ex: Bonanjo'));
    g.appendChild(input('Latitude','text','s-lat','Optionnel'));
    g.appendChild(input('Longitude','text','s-lng','Optionnel'));
    content.appendChild(g);
  }
  function renderStep2(){
    content.appendChild(el('h3',{text:'Compte administrateur'}));
    content.appendChild(el('p',{class:'setup-desc',text:'Ce sera le premier utilisateur avec tous les droits.'}));
    const g=el('div',{class:'setup-grid'});
    g.appendChild(input('Nom complet','text','s-name2','Ex: Dr. Adele Mbarga'));
    g.appendChild(input('Email','email','s-email','Ex: adele@hopital.cm'));
    g.appendChild(input('Identifiant','text','s-user','Choisir un identifiant'));
    g.appendChild(sel('Fonction','s-title',['Medecin referent','Chef de service','Directeur','Coordinateur programme']));
    g.appendChild(input('Mot de passe','password','s-pwd','Min 8 caracteres + 1 chiffre',true));
    g.appendChild(input('Confirmer','password','s-pwd2','Retapez le mot de passe',true));
    content.appendChild(g);
  }
  function renderStep3(){
    content.appendChild(el('h3',{text:'Horaires de consultation'}));
    content.appendChild(el('p',{class:'setup-desc',text:'Definissez les creneaux. Modifiable a tout moment.'}));
    const dg=el('div',{class:'form-group'});dg.appendChild(el('label',{text:'Jours'}));
    const checks=el('div',{class:'day-checks'});
    ['L','M','M','J','V','S','D'].forEach((d,i)=>{const b=el('div',{class:'day-check'+(i<5?' on':''),text:d});b.onclick=()=>b.classList.toggle('on');checks.appendChild(b)});
    dg.appendChild(checks);content.appendChild(dg);
    const g=el('div',{class:'setup-grid',style:'margin-top:14px'});
    const tg1=el('div',{class:'form-group'});tg1.appendChild(el('label',{text:'Debut'}));tg1.appendChild(el('input',{class:'form-input',type:'time',id:'s-start',value:'08:00'}));g.appendChild(tg1);
    const tg2=el('div',{class:'form-group'});tg2.appendChild(el('label',{text:'Fin'}));tg2.appendChild(el('input',{class:'form-input',type:'time',id:'s-end',value:'16:00'}));g.appendChild(tg2);
    g.appendChild(sel('Duree creneau','s-slot',['15','30','45','60'],'30'));
    const ng=el('div',{class:'form-group'});ng.appendChild(el('label',{text:'Max patients/jour'}));ng.appendChild(el('input',{class:'form-input',type:'number',id:'s-max',value:'40',min:'1'}));g.appendChild(ng);
    content.appendChild(g);
  }
  function renderStep4(){
    content.appendChild(el('h3',{text:'Configuration SMS (optionnel)'}));
    content.appendChild(el('p',{class:'setup-desc',text:'Les rappels SMS reduisent les RDV manques de 25 a 50%.'}));
    const opts=el('div',{class:'type-options',id:'s-sms-enable',style:'margin-bottom:16px'});
    ['Oui, activer','Plus tard'].forEach((t,i)=>{
      const b=el('div',{class:'type-opt'+(i===1?' on':''),text:t});
      b.onclick=()=>{$$('.type-opt',opts).forEach(x=>x.classList.remove('on'));b.classList.add('on')};
      opts.appendChild(b)});
    content.appendChild(opts);
    const g=el('div',{class:'setup-grid'});
    g.appendChild(sel('Fournisseur','s-sms-prov',["Africa's Talking",'MTN','Orange','Twilio','Infobip']));
    g.appendChild(input('Cle API','password','s-sms-key','Votre cle API'));
    g.appendChild(input('Secret','password','s-sms-sec','Votre secret'));
    g.appendChild(input('Expediteur','text','s-sms-sender','Ex: MaSante'));
    content.appendChild(g);
  }
  function renderStep5(){
    const s=el('div',{style:'text-align:center;padding:40px 0'});
    const icw=el('div',{style:'width:72px;height:72px;border-radius:50%;background:var(--success-bg);color:var(--success);display:flex;align-items:center;justify-content:center;margin:0 auto 20px'});
    icw.appendChild(svg('check',32));s.appendChild(icw);
    s.appendChild(el('h3',{text:'Votre plateforme est prete',style:'margin-bottom:10px'}));
    s.appendChild(el('p',{text:'Cliquez sur "Lancer MaSante" pour commencer.',style:'color:var(--gray-400)'}));
    const recap=el('div',{class:'setup-recap',style:'text-align:left;margin-top:24px'});
    [['Etablissement',data.center.name||'—'],['Localisation',[data.center.city,data.center.district,data.center.country].filter(Boolean).join(', ')||'—'],['Administrateur',data.admin.full_name||'—'],['Identifiant',data.admin.username||'—'],['SMS',data.sms.enabled?'Active':'Non configure']].forEach(([l,v])=>{
      const r=el('div',{class:'sr-row'});r.appendChild(el('span',{class:'sr-label',text:l}));r.appendChild(el('span',{class:'sr-val',text:v}));recap.appendChild(r)});
    s.appendChild(recap);content.appendChild(s);
  }

  function v(id){const e=$('#'+id);return e?e.value.trim():''}
  async function saveStep(){
    let res;
    if(step===1){const typeMap={"Hopital public":'hopital_public',"Centre de sante":'centre_sante',"Clinique privee":'clinique_privee'};const sel=$('#s-types .type-opt.on');data.center={name:v('s-name'),type:sel?typeMap[sel.textContent]||'centre_sante':'centre_sante',country:v('s-country'),city:v('s-city'),district:v('s-district')};if(v('s-lat'))data.center.lat=parseFloat(v('s-lat'));if(v('s-lng'))data.center.lng=parseFloat(v('s-lng'));if(!data.center.name||!data.center.country||!data.center.city)return'Nom, pays et ville requis';res=await setupApi.center(data.center)}
    else if(step===2){data.admin={full_name:v('s-name2'),email:v('s-email'),username:v('s-user'),password:v('s-pwd'),title:v('s-title')};if(!data.admin.full_name||!data.admin.username||!data.admin.password)return'Nom, identifiant et mot de passe requis';if(v('s-pwd')!==v('s-pwd2'))return'Les mots de passe ne correspondent pas';res=await setupApi.admin(data.admin)}
    else if(step===3){const days=[];$$('.day-check.on').forEach((_,i)=>days.push(i+1));data.schedule={consultation_days:days.join(','),start_time:v('s-start'),end_time:v('s-end'),slot_duration:parseInt(v('s-slot'))||30,max_patients_day:parseInt(v('s-max'))||40};res=await setupApi.schedule(data.schedule)}
    else if(step===4){const en=$('#s-sms-enable .type-opt.on');const enabled=en&&en.textContent.indexOf('Oui')>=0;const provMap={"Africa's Talking":'africastalking','MTN':'mtn','Orange':'orange','Twilio':'twilio','Infobip':'infobip'};data.sms={enabled:enabled,provider:enabled?(provMap[v('s-sms-prov')]||''):'',api_key:v('s-sms-key'),api_secret:v('s-sms-sec'),sender_id:v('s-sms-sender')};res=await setupApi.sms(data.sms)}
    else if(step===5){res=await setupApi.complete();if(res.ok){toast('Configuration terminee !','success');return null}}
    if(res&&!res.ok)return res.error;return null;
  }
}

// ── LOGIN ──
async function pageLogin(c){
  const wrap=el('div',{style:'display:flex;height:100vh;margin:-24px'});
  const left=el('div',{style:'flex:1;background:var(--primary);display:flex;flex-direction:column;justify-content:center;align-items:center;color:#fff'});
  const brand=el('div',{style:'text-align:center'});
  const logo=el('div',{style:'width:80px;height:80px;border:2px solid rgba(255,255,255,.15);border-radius:20px;display:flex;align-items:center;justify-content:center;margin:0 auto 28px;background:rgba(255,255,255,.06)'});
  logo.appendChild(svg('leaf',36));brand.appendChild(logo);
  brand.appendChild(el('h1',{text:'MaSante',style:'font-size:2.8rem;font-weight:700;letter-spacing:-1.5px;margin-bottom:8px'}));
  brand.appendChild(el('p',{text:'Plateforme de suivi sante communautaire',style:'font-size:1rem;color:rgba(255,255,255,.5)'}));
  left.appendChild(brand);wrap.appendChild(left);

  const right=el('div',{style:'width:440px;background:var(--white);display:flex;flex-direction:column;justify-content:center;padding:56px'});
  right.appendChild(el('h2',{text:'Connexion',style:'font-size:1.7rem;margin-bottom:6px'}));
  right.appendChild(el('p',{text:'Entrez vos identifiants',style:'color:var(--gray-400);margin-bottom:32px;font-size:.9rem'}));
  const errEl=el('div',{style:'display:none;padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin-bottom:16px'});right.appendChild(errEl);
  const ug=el('div',{class:'form-group'});ug.appendChild(el('label',{text:'Identifiant'}));const ui=el('input',{class:'form-input',type:'text',placeholder:'Votre identifiant'});ug.appendChild(ui);right.appendChild(ug);
  const pg=el('div',{class:'form-group'});pg.appendChild(el('label',{text:'Mot de passe'}));const pi=el('input',{class:'form-input',type:'password',placeholder:'Votre mot de passe'});pg.appendChild(pi);right.appendChild(pg);
  const btn=el('button',{class:'btn btn-primary',style:'margin-top:8px'});btn.appendChild(svg('check',16));btn.appendChild(document.createTextNode(' Se connecter'));

  async function doLogin(){
    errEl.style.display='none';
    if(!ui.value.trim()||!pi.value){errEl.textContent='Identifiant et mot de passe requis';errEl.style.display='block';return}
    btn.disabled=true;btn.textContent='Connexion...';
    const res=await auth.login(ui.value.trim(),pi.value);
    if(!res.ok){errEl.textContent=res.error;errEl.style.display='block';btn.disabled=false;btn.textContent='';btn.appendChild(svg('check',16));btn.appendChild(document.createTextNode(' Se connecter'));return}
    user=res.data;toast('Connexion reussie','success');
    const app=$('#app');renderShell(app);location.hash='#dashboard';route();
  }
  btn.onclick=doLogin;pi.onkeydown=e=>{if(e.key==='Enter')doLogin()};
  right.appendChild(btn);wrap.appendChild(right);c.appendChild(wrap);ui.focus();
}

// ── DASHBOARD ──
async function pageDashboard(c){
  c.appendChild(loading());
  const[sR,tR,oR]=await Promise.all([dash.stats(),dash.today(),dash.overdue()]);
  c.textContent='';
  const p=sR.ok?sR.data.patients||{}:{};const a=sR.ok?sR.data.appointments||{}:{};
  const total=Object.values(p).reduce((s,v)=>s+v,0);const todayN=Object.values(a).reduce((s,v)=>s+v,0);const lost=p.perdu_de_vue||0;
  const retention=total>0?Math.round((1-lost/total)*100):0;
  // Stats.
  const sr=el('div',{class:'stats-row'});
  [[svg('users'),total,'Patients actifs','green'],[svg('calendar'),todayN,"RDV aujourd'hui",'blue'],[svg('chart'),retention+'%','Retention','amber'],[svg('alert'),lost,'Perdus de vue','red']].forEach(([ic,val,lbl,col])=>{
    const card=el('div',{class:'stat'});const iw=el('div',{class:'stat-icon '+col});iw.appendChild(ic);card.appendChild(iw);card.appendChild(el('div',{class:'stat-val',text:String(val)}));card.appendChild(el('div',{class:'stat-label',text:lbl}));sr.appendChild(card)});
  c.appendChild(sr);

  const grid=el('div',{class:'grid-2',style:'align-items:start'});
  // Today.
  const tc=el('div',{class:'card'});
  const th=el('div',{class:'card-head'});const tt=el('h3');tt.appendChild(svg('calendar',18));tt.appendChild(document.createTextNode(' RDV du jour'));th.appendChild(tt);
  const calLink=el('a',{class:'card-link',text:'Calendrier'});calLink.onclick=()=>{location.hash='#calendar'};th.appendChild(calLink);tc.appendChild(th);
  const tb=el('div',{class:'card-body'});
  const today=tR.ok?(tR.data||[]):[];
  if(!today.length)tb.appendChild(empty("Aucun RDV aujourd'hui"));
  else today.forEach(apt=>{
    const item=el('div',{class:'list-item'});item.appendChild(el('span',{class:'time-label',text:apt.Time||'—'}));
    item.appendChild(el('div',{class:'avatar a1',text:initials(apt.PatientName)}));
    const info=el('div',{class:'item-info'});info.appendChild(el('div',{class:'item-name',text:apt.PatientName||'Patient'}));info.appendChild(el('div',{class:'item-sub',text:fmtType(apt.Type)}));item.appendChild(info);
    item.appendChild(pill(apt.Status));tb.appendChild(item)});
  tc.appendChild(tb);grid.appendChild(tc);

  // Overdue.
  const oc=el('div',{class:'card'});
  const oh=el('div',{class:'card-head'});const ot=el('h3');ot.appendChild(svg('clock',18));ot.appendChild(document.createTextNode(' Patients en retard'));oh.appendChild(ot);oc.appendChild(oh);
  const ob=el('div',{class:'card-body'});
  const overdue=oR.ok?(oR.data||[]):[];
  if(!overdue.length)ob.appendChild(empty('Aucun patient en retard'));
  else overdue.forEach(apt=>{
    const item=el('div',{class:'list-item'});item.appendChild(el('div',{class:'overdue-avatar',text:initials(apt.PatientName)}));
    const info=el('div',{class:'item-info'});info.appendChild(el('div',{class:'item-name',text:apt.PatientName||'Patient'}));
    const days=Math.max(1,Math.floor((Date.now()-new Date(apt.Date).getTime())/864e5));
    info.appendChild(el('div',{class:'overdue-days',text:days+' jours de retard'}));item.appendChild(info);
    const btn=el('button',{class:'btn btn-sm btn-secondary',text:'Reprogrammer',style:'width:auto;flex-shrink:0'});btn.onclick=()=>{location.hash='#new-apt'};item.appendChild(btn);ob.appendChild(item)});
  oc.appendChild(ob);grid.appendChild(oc);c.appendChild(grid);
}

// ── CALENDAR ──
async function pageCalendar(c){
  c.appendChild(loading());
  const today=new Date();const mon=new Date(today);mon.setDate(today.getDate()-((today.getDay()+6)%7));
  const ds=mon.toISOString().slice(0,10);const res=await cal.week(ds);c.textContent='';
  c.appendChild(el('div',{class:'cal-date',text:'Semaine du '+mon.toLocaleDateString('fr-FR',{day:'numeric',month:'long',year:'numeric'}),style:'margin-bottom:16px'}));
  const grid=el('div',{class:'cal-grid'});
  const header=el('div',{class:'cal-header'});header.appendChild(el('div',{class:'cal-hcell'}));
  ['Lun','Mar','Mer','Jeu','Ven','Sam','Dim'].forEach((d,i)=>{const dd=new Date(mon);dd.setDate(mon.getDate()+i);header.appendChild(el('div',{class:'cal-hcell'+(dd.toDateString()===today.toDateString()?' today':''),text:d+' '+dd.getDate()}))});
  grid.appendChild(header);
  const body=el('div',{class:'cal-body'});
  const aptsList=res.ok?(res.data||[]):[];const byDT={};
  aptsList.forEach(a=>{const k=(a.Date||'').slice(0,10)+'|'+(a.Time||'');if(!byDT[k])byDT[k]=[];byDT[k].push(a)});
  for(let h=8;h<=16;h++){const time=String(h).padStart(2,'0')+':00';const row=el('div',{class:'cal-row'});row.appendChild(el('div',{class:'cal-time',text:time}));
    for(let d=0;d<7;d++){const dd=new Date(mon);dd.setDate(mon.getDate()+d);const ds2=dd.toISOString().slice(0,10);const slot=el('div',{class:'cal-slot'});
      (byDT[ds2+'|'+time]||[]).forEach(a=>{const sm={confirme:'c-ok',en_attente:'c-wait',manque:'c-miss',termine:'c-done'};slot.appendChild(el('div',{class:'cal-evt '+(sm[a.Status]||'c-ok'),text:(a.PatientName||'').split(' ')[0]}))});
      row.appendChild(slot)}body.appendChild(row)}
  grid.appendChild(body);c.appendChild(grid);
}

// ── NEW APPOINTMENT ──
async function pageNewApt(c){
  const grid=el('div',{class:'apt-grid'});let selPatientId=null,selTime='';
  // Left.
  const left=el('div');const c1=el('div',{class:'card'});
  const h1=el('div',{class:'card-head'});const t1=el('h3');t1.appendChild(svg('users',18));t1.appendChild(document.createTextNode(' Patient'));h1.appendChild(t1);c1.appendChild(h1);
  const b1=el('div',{class:'card-body'});
  const si=el('input',{class:'form-input',type:'text',placeholder:'Rechercher par nom ou code...'});
  const results=el('div');const selInfo=el('div',{style:'margin-top:10px'});
  si.oninput=async()=>{const q=si.value.trim();if(q.length<2){results.textContent='';return}const r=await pts.search(q);results.textContent='';if(r.ok&&r.data)(r.data).slice(0,5).forEach(p=>{const it=el('div',{style:'padding:8px 12px;cursor:pointer;border-bottom:1px solid var(--gray-50);font-size:.88rem',text:p.LastName+' '+p.FirstName+' — '+p.Code});it.onclick=()=>{selPatientId=p.ID;results.textContent='';si.value=p.LastName+' '+p.FirstName;selInfo.textContent=p.Code+' — '+(p.District||'')};results.appendChild(it)})};
  b1.appendChild(si);b1.appendChild(results);b1.appendChild(selInfo);
  const typeG=el('div',{class:'form-group',style:'margin-top:12px'});typeG.appendChild(el('label',{text:'Type'}));
  const typeSel=el('select',{class:'form-input'});
  [['consultation','Consultation'],['retrait_medicaments','Retrait medicaments'],['bilan_sanguin','Bilan sanguin'],['club_adherence',"Club d'adherence"]].forEach(([v,l])=>typeSel.appendChild(el('option',{value:v,text:l})));
  typeG.appendChild(typeSel);b1.appendChild(typeG);
  const notesI=el('textarea',{class:'form-input',placeholder:'Notes...',rows:'3'});
  const ng=el('div',{class:'form-group'});ng.appendChild(el('label',{text:'Notes'}));ng.appendChild(notesI);b1.appendChild(ng);
  c1.appendChild(b1);left.appendChild(c1);grid.appendChild(left);

  // Right.
  const right=el('div');const c2=el('div',{class:'card'});
  const h2=el('div',{class:'card-head'});const t2=el('h3');t2.appendChild(svg('calendar',18));t2.appendChild(document.createTextNode(' Date et creneau'));h2.appendChild(t2);c2.appendChild(h2);
  const b2=el('div',{class:'card-body'});
  const dg=el('div',{class:'form-group'});dg.appendChild(el('label',{text:'Date'}));const di=el('input',{class:'form-input',type:'date'});dg.appendChild(di);b2.appendChild(dg);
  const slotsC=el('div');
  di.onchange=async()=>{slotsC.textContent='';selTime='';const r=await apts.slots(di.value);if(!r.ok)return;const sg=el('div',{class:'slots-grid'});(r.data||[]).forEach(s=>{const sl=el('div',{class:'slot'+(s.Available?'':' off'),text:s.Time});if(s.Available)sl.onclick=()=>{$$('.slot',sg).forEach(x=>x.classList.remove('picked'));sl.classList.add('picked');selTime=s.Time};sg.appendChild(sl)});slotsC.appendChild(sg)};
  b2.appendChild(slotsC);c2.appendChild(b2);right.appendChild(c2);
  const errEl=el('div',{style:'display:none;padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin:12px 0'});right.appendChild(errEl);
  const subBtn=el('button',{class:'btn btn-primary',style:'margin-top:12px'});subBtn.appendChild(svg('check',16));subBtn.appendChild(document.createTextNode(' Confirmer'));
  subBtn.onclick=async()=>{errEl.style.display='none';if(!selPatientId){errEl.textContent='Selectionnez un patient';errEl.style.display='block';return}if(!di.value){errEl.textContent='Selectionnez une date';errEl.style.display='block';return}if(!selTime){errEl.textContent='Selectionnez un creneau';errEl.style.display='block';return}subBtn.disabled=true;const r=await apts.create({patient_id:selPatientId,date:di.value,time:selTime,type:typeSel.value,notes:notesI.value});if(!r.ok){errEl.textContent=r.error;errEl.style.display='block';subBtn.disabled=false;return}toast('RDV programme','success');location.hash='#calendar'};
  right.appendChild(subBtn);grid.appendChild(right);c.appendChild(grid);
}

// ── NEW PATIENT ──
async function pageNewPatient(c){
  const grid=el('div',{class:'apt-grid'});let lang='fr',channel='sms';
  function inp(lbl,type,ph){const g=el('div',{class:'form-group'});g.appendChild(el('label',{text:lbl}));const i=el('input',{class:'form-input',type:type,placeholder:ph||''});g.appendChild(i);return{g,v:()=>i.value.trim()}}
  function selG(lbl,opts,vals){const g=el('div',{class:'form-group'});g.appendChild(el('label',{text:lbl}));const s=el('select',{class:'form-input'});(vals||opts).forEach((v,i)=>s.appendChild(el('option',{value:v,text:opts[i]})));g.appendChild(s);return{g,v:()=>s.value}}

  const left=el('div');const c1=el('div',{class:'card'});
  const h1=el('div',{class:'card-head'});const t1=el('h3');t1.appendChild(svg('user-plus',18));t1.appendChild(document.createTextNode(' Inscription'));h1.appendChild(t1);c1.appendChild(h1);
  const b1=el('div',{class:'card-body'});const fg=el('div',{style:'display:grid;grid-template-columns:1fr 1fr;gap:14px'});
  const f={last_name:inp('Nom','text','Nom de famille'),first_name:inp('Prenom','text','Prenom'),dob:inp('Date de naissance','date'),sex:selG('Sexe',['Selectionner','Masculin','Feminin'],['','M','F']),phone:inp('Telephone','tel','+237 6XX XXX XXX'),phone2:inp('Tel secondaire','tel','Optionnel'),district:inp('Quartier','text','Ex: Akwa'),address:inp('Adresse','text','Repere')};
  Object.values(f).forEach(x=>fg.appendChild(x.g));b1.appendChild(fg);
  // Language.
  const lg=el('div',{class:'form-group',style:'margin-top:14px'});lg.appendChild(el('label',{text:'Langue'}));
  const lo=el('div',{class:'remind-opts'});
  [['Francais','fr'],['Anglais','en'],['Duala','duala'],['Ewondo','ewondo'],['Bamileke','bamileke']].forEach(([l,code],i)=>{const o=el('div',{class:'r-opt'+(i===0?' on':''),text:l});o.onclick=()=>{$$('.r-opt',lo).forEach(x=>x.classList.remove('on'));o.classList.add('on');lang=code};lo.appendChild(o)});
  lg.appendChild(lo);b1.appendChild(lg);c1.appendChild(b1);left.appendChild(c1);grid.appendChild(left);

  // Right.
  const right=el('div');
  // Channel.
  const c2=el('div',{class:'card',style:'margin-bottom:16px'});const h2=el('div',{class:'card-head'});const t2=el('h3');t2.appendChild(svg('msg',18));t2.appendChild(document.createTextNode(' Contact'));h2.appendChild(t2);c2.appendChild(h2);
  const b2=el('div',{class:'card-body'});
  const cg=el('div',{class:'form-group'});cg.appendChild(el('label',{text:'Canal de rappel'}));
  const co=el('div',{class:'remind-opts'});
  [['SMS','sms','msg'],['WhatsApp','whatsapp','smartphone'],['Appel','voice','phone'],['Aucun','none','ban']].forEach(([l,code,ic],i)=>{const o=el('div',{class:'r-opt'+(i===0?' on':'')});o.appendChild(svg(ic,16));o.appendChild(document.createTextNode(' '+l));o.onclick=()=>{$$('.r-opt',co).forEach(x=>x.classList.remove('on'));o.classList.add('on');channel=code};co.appendChild(o)});
  cg.appendChild(co);b2.appendChild(cg);
  const cf={contact_name:inp('Personne de confiance','text','Nom'),contact_phone:inp('Tel contact','tel','+237'),contact_rel:selG('Lien',['Selectionner','Conjoint(e)','Parent','Frere/Soeur','Ami(e)','Autre'],['','Conjoint(e)','Parent','Frere/Soeur','Ami(e)','Autre']),referred:selG('Refere par',['Selectionner','Centre de depistage','Transfert','Auto-presentation','Agent communautaire','Autre'],['','Centre de depistage','Transfert','Auto-presentation','Agent communautaire','Autre'])};
  Object.values(cf).forEach(x=>b2.appendChild(x.g));c2.appendChild(b2);right.appendChild(c2);

  const errEl=el('div',{style:'display:none;padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin-bottom:12px'});right.appendChild(errEl);
  const subBtn=el('button',{class:'btn btn-primary'});subBtn.appendChild(svg('check',16));subBtn.appendChild(document.createTextNode(' Inscrire le patient'));
  subBtn.onclick=async()=>{errEl.style.display='none';const body={last_name:f.last_name.v(),first_name:f.first_name.v(),date_of_birth:f.dob.v(),sex:f.sex.v(),phone:f.phone.v(),phone_secondary:f.phone2.v(),district:f.district.v(),address:f.address.v(),language:lang,reminder_channel:channel,contact_name:cf.contact_name.v(),contact_phone:cf.contact_phone.v(),contact_relation:cf.contact_rel.v(),referred_by:cf.referred.v()};
    if(!body.last_name||!body.first_name||!body.sex){errEl.textContent='Nom, prenom et sexe requis';errEl.style.display='block';return}subBtn.disabled=true;const r=await pts.create(body);if(!r.ok){errEl.textContent=r.error;errEl.style.display='block';subBtn.disabled=false;return}toast('Patient inscrit — Code: '+(r.data.Code||''),'success');location.hash='#patients'};
  right.appendChild(subBtn);grid.appendChild(right);c.appendChild(grid);
}

// ── PATIENTS ──
async function pagePatients(c){
  const filters=el('div',{class:'filters'});let activeF='';
  [['Tous',''],['Actifs','active'],['A surveiller','a_surveiller'],['Perdus de vue','perdu_de_vue'],['Sortis','sorti']].forEach(([l,v])=>{const b=el('button',{class:'fbtn'+(v===''?' on':''),text:l});b.onclick=()=>{activeF=v;$$('.fbtn',filters).forEach(x=>x.classList.remove('on'));b.classList.add('on');load()};filters.appendChild(b)});
  c.appendChild(filters);const tC=el('div');c.appendChild(tC);

  async function load(){tC.textContent='';tC.appendChild(loading());let q='page=1&per_page=50';if(activeF)q+='&status='+activeF;
    if(window._searchQuery){q+='&q='+encodeURIComponent(window._searchQuery);delete window._searchQuery}
    const r=await pts.list(q);tC.textContent='';if(!r.ok){toast(r.error,'error');return}
    const list=r.data.patients||[];const total=r.data.total||0;
    if(!list.length){tC.appendChild(empty('Aucun patient'));return}
    const t=el('div',{class:'ptable'});
    const th=el('div',{class:'pt-head',style:'grid-template-columns:44px 2fr 1fr 1fr 1fr 100px'});['','Patient','Risque','Derniere visite','Statut',''].forEach(h=>th.appendChild(el('div',{text:h})));t.appendChild(th);
    const cols=['a1','a2','a3','a4','a5','a6'];
    list.forEach((p,i)=>{const row=el('div',{class:'pt-row',style:'grid-template-columns:44px 2fr 1fr 1fr 1fr 100px;cursor:pointer'});
      row.appendChild(el('div',{},[el('div',{class:'avatar '+cols[i%6],text:initials(p.LastName+' '+p.FirstName)})]));
      const nc=el('div');nc.appendChild(el('div',{class:'pt-name',text:p.LastName+' '+p.FirstName}));nc.appendChild(el('div',{class:'pt-code',text:p.Code}));row.appendChild(nc);
      row.appendChild(el('div',{},[risk(p.RiskScore)]));row.appendChild(el('div',{class:'pt-cell',text:fmtDate(p.UpdatedAt)}));row.appendChild(el('div',{},[pill(p.Status)]));
      const ac=el('div',{class:'pt-acts'});const ab=el('button',{class:'icon-btn'});ab.appendChild(svg('calendar',16));ab.onclick=e=>{e.stopPropagation();location.hash='#new-apt'};ac.appendChild(ab);row.appendChild(ac);
      row.onclick=()=>{window._patientId=p.ID;location.hash='#patient-file'};
      t.appendChild(row)});
    tC.appendChild(t);tC.appendChild(el('div',{style:'text-align:center;padding:12px;color:var(--gray-400);font-size:.82rem',text:total+' patients'}))}
  load();
}

// ── PATIENT FILE ──
async function pagePatientFile(c){
  const id=window._patientId;if(!id){location.hash='#patients';return}
  c.appendChild(loading());const r=await pts.get(id);if(!r.ok){toast('Patient introuvable','error');location.hash='#patients';return}
  const p=r.data;c.textContent='';
  const layout=el('div',{class:'pf-layout'});
  // Sidebar.
  const side=el('div',{class:'pf-side'});
  const card=el('div',{class:'pf-card'});card.appendChild(el('div',{class:'pf-avatar',text:initials(p.LastName+' '+p.FirstName)}));
  card.appendChild(el('div',{class:'pf-name',text:p.LastName+' '+p.FirstName}));card.appendChild(el('div',{class:'pf-id',text:p.Code}));
  card.appendChild(el('div',{style:'margin-bottom:10px'},[pill(p.Status)]));
  const det=el('div',{class:'pf-details'});
  [['Sexe',p.Sex==='M'?'Masculin':'Feminin'],['Zone',p.District||'—'],['Telephone',p.Phone||'—'],['Langue',p.Language||'fr'],['Rappels',p.ReminderChannel||'—'],['Inscrit',fmtDate(p.EnrollmentDate)]].forEach(([l,v])=>{const r=el('div',{class:'pf-row'});r.appendChild(el('span',{class:'lbl',text:l}));r.appendChild(el('span',{class:'val',text:v}));det.appendChild(r)});
  card.appendChild(det);side.appendChild(card);
  // Risk.
  const rc=el('div',{class:'risk-card'});const rh=el('div',{class:'risk-head'});rh.appendChild(el('h4',{text:'Score de risque'}));
  const circ=el('div',{class:'risk-circle',text:String(p.RiskScore||5)});rh.appendChild(circ);rc.appendChild(rh);side.appendChild(rc);
  // Actions.
  const aptBtn=el('button',{class:'btn btn-primary',style:'margin-bottom:8px'});aptBtn.appendChild(svg('calendar',16));aptBtn.appendChild(document.createTextNode(' Programmer un RDV'));aptBtn.onclick=()=>{location.hash='#new-apt'};side.appendChild(aptBtn);
  if(p.Status!=='sorti'){const exitBtn=el('button',{class:'btn btn-secondary',style:'color:var(--gray-500)'});exitBtn.appendChild(svg('archive',16));exitBtn.appendChild(document.createTextNode(' Sortie du programme'));exitBtn.onclick=()=>showExit(p);side.appendChild(exitBtn)}
  layout.appendChild(side);
  // Main.
  const main=el('div',{class:'pf-main'});
  const tc=el('div',{class:'card'});const tch=el('div',{class:'card-head'});const tct=el('h3');tct.appendChild(svg('clock',18));tct.appendChild(document.createTextNode(' Historique'));tch.appendChild(tct);tc.appendChild(tch);
  const tcb=el('div',{class:'card-body'});
  const tl=el('div',{class:'tl'});const item=el('div',{class:'tl-item current'});item.appendChild(el('div',{class:'tl-date',text:fmtDate(p.EnrollmentDate)+' — inscription'}));
  const cont=el('div',{class:'tl-content'});cont.appendChild(el('div',{class:'tl-type',text:'Inscription dans le programme'}));
  if(p.ReferredBy)cont.appendChild(el('div',{class:'tl-note',text:'Refere par: '+p.ReferredBy}));
  item.appendChild(cont);tl.appendChild(item);tcb.appendChild(tl);tc.appendChild(tcb);main.appendChild(tc);
  layout.appendChild(main);c.appendChild(layout);
}

function showExit(p){
  const content=el('div');content.appendChild(el('p',{text:'Motif de sortie pour '+p.LastName+' '+p.FirstName,style:'font-size:.85rem;color:var(--gray-500);margin-bottom:16px'}));
  const reasons=['deces','transfert','abandon','perdu_de_vue','guerison'];const labels=['Deces','Transfert','Abandon volontaire','Perdu de vue definitif','Guerison'];let selReason='';
  const opts=el('div');reasons.forEach((r,i)=>{const o=el('div',{style:'padding:10px 14px;border:1.5px solid var(--gray-200);border-radius:var(--radius);margin-bottom:8px;cursor:pointer;font-size:.88rem',text:labels[i]});o.onclick=()=>{$$('div',opts).forEach(d=>{d.style.borderColor='var(--gray-200)';d.style.background=''});o.style.borderColor='var(--primary)';o.style.background='var(--primary-subtle)';selReason=r};opts.appendChild(o)});
  content.appendChild(opts);
  const di=el('input',{class:'form-input',type:'date',value:new Date().toISOString().slice(0,10)});const dg=el('div',{class:'form-group',style:'margin-top:12px'});dg.appendChild(el('label',{text:'Date'}));dg.appendChild(di);content.appendChild(dg);
  const ni=el('textarea',{class:'form-input',placeholder:'Notes...',rows:'3'});const ng2=el('div',{class:'form-group'});ng2.appendChild(el('label',{text:'Notes'}));ng2.appendChild(ni);content.appendChild(ng2);

  // Simple modal.
  const overlay=el('div',{class:'modal-overlay open'});const modal=el('div',{class:'modal open'});modal.onclick=e=>e.stopPropagation();
  const mh=el('div',{class:'modal-head'});mh.appendChild(el('h3',{text:'Sortie du programme'}));const xBtn=el('button',{class:'icon-btn'});xBtn.appendChild(svg('x'));xBtn.onclick=()=>{overlay.remove();modal.remove()};mh.appendChild(xBtn);
  modal.appendChild(mh);const mb=el('div',{class:'modal-body'});mb.appendChild(content);modal.appendChild(mb);
  const mf=el('div',{class:'modal-footer'});
  const cancelBtn=el('button',{class:'btn btn-secondary',text:'Annuler',style:'width:auto'});cancelBtn.onclick=()=>{overlay.remove();modal.remove()};mf.appendChild(cancelBtn);
  const confirmBtn=el('button',{class:'btn btn-primary',text:'Confirmer',style:'width:auto'});
  confirmBtn.onclick=async()=>{if(!selReason){toast('Selectionnez un motif','error');return}const r=await pts.exit(p.ID,{reason:selReason,date:di.value,notes:ni.value});if(!r.ok){toast(r.error,'error');return}toast('Patient sorti du programme','success');overlay.remove();modal.remove();location.hash='#patients'};
  mf.appendChild(confirmBtn);modal.appendChild(mf);
  overlay.onclick=()=>{overlay.remove();modal.remove()};
  document.body.appendChild(overlay);document.body.appendChild(modal);
}

// ── REMINDERS ──
async function pageReminders(c){
  c.appendChild(loading());const[sR,qR,tR]=await Promise.all([rem.stats(),rem.list(),rem.templates()]);c.textContent='';
  const s=sR.ok?sR.data:{};
  const sr=el('div',{class:'rem-stats'});
  [[(s.DeliveryRate||0).toFixed(1)+'%','Taux livraison','var(--success)'],[String(s.PendingCount||0),'En attente','var(--warning)'],[String(s.FailedCount||0),'Echecs','var(--danger)']].forEach(([v,l,col])=>{const st=el('div',{class:'rem-stat'});st.appendChild(el('div',{class:'rem-stat-val',text:v,style:'color:'+col}));st.appendChild(el('div',{class:'rem-stat-label',text:l}));sr.appendChild(st)});
  c.appendChild(sr);
  const grid=el('div',{class:'grid-2-wide',style:'align-items:start'});
  // Queue.
  const qc=el('div',{class:'card'});const qh=el('div',{class:'card-head'});const qt=el('h3');qt.appendChild(svg('send',18));qt.appendChild(document.createTextNode(" File d'attente"));qh.appendChild(qt);qc.appendChild(qh);
  const qb=el('div',{class:'card-body'});const queue=qR.ok?(qR.data||[]):[];
  if(!queue.length)qb.appendChild(empty('Aucun rappel en attente'));
  else queue.forEach(r=>{const item=el('div',{class:'rq-item'});const ch=el('div',{class:'rq-ch ch-sms'});ch.appendChild(svg('msg',18));item.appendChild(ch);
    const info=el('div',{class:'rq-info'});info.appendChild(el('div',{class:'rq-name',text:r.PatientName||'Patient'}));info.appendChild(el('div',{class:'rq-sched',text:r.Type+' — '+(r.Status||'')}));item.appendChild(info);item.appendChild(pill(r.Status));qb.appendChild(item)});
  qc.appendChild(qb);grid.appendChild(qc);
  // Templates.
  const tc=el('div',{class:'card'});const tch=el('div',{class:'card-head'});const tct=el('h3');tct.appendChild(svg('clipboard',18));tct.appendChild(document.createTextNode(' Modeles'));tch.appendChild(tct);tc.appendChild(tch);
  const tcb=el('div',{class:'card-body'});const tpls=tR.ok?(tR.data||[]):[];
  tpls.forEach(t=>{const g=el('div',{class:'form-group'});g.appendChild(el('label',{text:t.Name}));const ta=el('textarea',{class:'form-input',rows:'3'});ta.value=t.Body;g.appendChild(ta);
    const sb=el('button',{class:'btn btn-sm btn-secondary',text:'Enregistrer',style:'width:auto;margin-top:6px'});sb.onclick=async()=>{const r=await rem.updateTpl(t.ID,{body:ta.value,is_active:t.IsActive});if(r.ok)toast('Modele enregistre','success');else toast(r.error,'error')};g.appendChild(sb);tcb.appendChild(g)});
  tc.appendChild(tcb);grid.appendChild(tc);c.appendChild(grid);
}

// ── USERS ──
async function pageUsers(c){
  const hdr=el('div',{style:'display:flex;align-items:center;justify-content:space-between;margin-bottom:20px'});
  hdr.appendChild(el('h3',{text:'Equipe soignante',style:'font-size:1.1rem'}));
  const addBtn=el('button',{class:'btn btn-primary',style:'width:auto'});addBtn.appendChild(svg('user-plus',16));addBtn.appendChild(document.createTextNode(' Ajouter'));
  addBtn.onclick=()=>showAddUser(load);hdr.appendChild(addBtn);c.appendChild(hdr);
  const tC=el('div');c.appendChild(tC);

  async function load(){tC.textContent='';tC.appendChild(loading());const r=await usr.list();tC.textContent='';if(!r.ok){toast(r.error,'error');return}
    const list=r.data||[];const t=el('div',{class:'ptable'});
    const th=el('div',{class:'pt-head',style:'grid-template-columns:44px 2fr 1fr 1fr 1fr 80px'});['','Utilisateur','Role','Derniere connexion','Statut',''].forEach(h=>th.appendChild(el('div',{text:h})));t.appendChild(th);
    list.forEach(u=>{const row=el('div',{class:'pt-row',style:'grid-template-columns:44px 2fr 1fr 1fr 1fr 80px'});
      row.appendChild(el('div',{},[el('div',{class:'avatar a1',text:initials(u.full_name)})]));
      const nc=el('div');nc.appendChild(el('div',{class:'pt-name',text:u.full_name}));nc.appendChild(el('div',{class:'pt-code',text:u.username+(u.email?' — '+u.email:'')}));row.appendChild(nc);
      const rm={admin:'pill-danger',medecin:'pill-info',infirmier:'pill-warning',asc:'pill-success'};row.appendChild(el('div',{},[el('span',{class:'pill '+(rm[u.role]||'pill-neutral'),text:u.role})]));
      row.appendChild(el('div',{class:'pt-cell',text:'—'}));row.appendChild(el('div',{},[pill(u.status)]));
      const ac=el('div',{class:'pt-acts'});const db=el('button',{class:'icon-btn'});db.appendChild(svg('trash',16));db.onclick=async()=>{if(confirm('Desactiver '+u.full_name+' ?')){await usr.disable(u.id);load()}};ac.appendChild(db);row.appendChild(ac);
      t.appendChild(row)});tC.appendChild(t)}
  load();
}

function showAddUser(onDone){
  const content=el('div');const fields={};
  [['Nom complet','text','au-name'],['Email','email','au-email'],['Identifiant','text','au-user'],['Mot de passe','password','au-pwd']].forEach(([l,t,id])=>{const g=el('div',{class:'form-group'});g.appendChild(el('label',{text:l}));const i=el('input',{class:'form-input',type:t,id:id,placeholder:l});g.appendChild(i);fields[id]=i;content.appendChild(g)});
  const rg=el('div',{class:'form-group'});rg.appendChild(el('label',{text:'Role'}));const rs=el('select',{class:'form-input',id:'au-role'});
  [['admin','Admin'],['medecin','Medecin'],['infirmier','Infirmier'],['asc','ASC']].forEach(([v,l])=>rs.appendChild(el('option',{value:v,text:l})));rs.value='asc';rg.appendChild(rs);content.appendChild(rg);
  const errEl=el('div',{style:'display:none;padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin-top:10px'});content.appendChild(errEl);

  const overlay=el('div',{class:'modal-overlay open'});const modal=el('div',{class:'modal open'});modal.onclick=e=>e.stopPropagation();
  const mh=el('div',{class:'modal-head'});mh.appendChild(el('h3',{text:'Ajouter un utilisateur'}));const xBtn=el('button',{class:'icon-btn'});xBtn.appendChild(svg('x'));xBtn.onclick=()=>{overlay.remove();modal.remove()};mh.appendChild(xBtn);
  modal.appendChild(mh);const mb=el('div',{class:'modal-body'});mb.appendChild(content);modal.appendChild(mb);
  const mf=el('div',{class:'modal-footer'});
  const cb=el('button',{class:'btn btn-secondary',text:'Annuler',style:'width:auto'});cb.onclick=()=>{overlay.remove();modal.remove()};mf.appendChild(cb);
  const sb=el('button',{class:'btn btn-primary',text:'Creer',style:'width:auto'});
  sb.onclick=async()=>{errEl.style.display='none';const d={full_name:fields['au-name'].value,email:fields['au-email'].value,username:fields['au-user'].value,password:fields['au-pwd'].value,role:rs.value};if(!d.full_name||!d.username||!d.password){errEl.textContent='Champs requis';errEl.style.display='block';return}const r=await usr.create(d);if(!r.ok){errEl.textContent=r.error;errEl.style.display='block';return}toast('Utilisateur cree','success');overlay.remove();modal.remove();onDone()};
  mf.appendChild(sb);modal.appendChild(mf);overlay.onclick=()=>{overlay.remove();modal.remove()};
  document.body.appendChild(overlay);document.body.appendChild(modal);
}

// ── PROFILE ──
async function pageProfile(c){
  if(!user)return;
  const grid=el('div',{class:'grid-2-wide',style:'align-items:start'});
  // Info.
  const c1=el('div',{class:'card'});const h1=el('div',{class:'card-head'});const t1=el('h3');t1.appendChild(svg('user',18));t1.appendChild(document.createTextNode(' Informations'));h1.appendChild(t1);c1.appendChild(h1);
  const b1=el('div',{class:'card-body'});
  const ni=el('input',{class:'form-input',type:'text',value:user.full_name||''});const ei=el('input',{class:'form-input',type:'email',value:user.email||''});const pi=el('input',{class:'form-input',type:'tel',value:''});
  [['Nom',ni],['Email',ei],['Telephone',pi]].forEach(([l,i])=>{const g=el('div',{class:'form-group'});g.appendChild(el('label',{text:l}));g.appendChild(i);b1.appendChild(g)});
  const saveBtn=el('button',{class:'btn btn-primary',style:'width:auto;margin-top:8px'});saveBtn.appendChild(svg('check',16));saveBtn.appendChild(document.createTextNode(' Enregistrer'));
  saveBtn.onclick=async()=>{const r=await prof.update({full_name:ni.value,email:ei.value,phone:pi.value});if(r.ok){toast('Profil mis a jour','success');user.full_name=ni.value}else toast(r.error,'error')};
  b1.appendChild(saveBtn);c1.appendChild(b1);grid.appendChild(c1);
  // Password.
  const c2=el('div',{class:'card'});const h2=el('div',{class:'card-head'});const t2=el('h3');t2.appendChild(svg('shield',18));t2.appendChild(document.createTextNode(' Securite'));h2.appendChild(t2);c2.appendChild(h2);
  const b2=el('div',{class:'card-body'});
  const cp=el('input',{class:'form-input',type:'password',placeholder:'Mot de passe actuel'});const np=el('input',{class:'form-input',type:'password',placeholder:'Nouveau (min 8 + 1 chiffre)'});const cfp=el('input',{class:'form-input',type:'password',placeholder:'Confirmer'});
  const pe=el('div',{style:'display:none;padding:8px 12px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.82rem;margin-top:8px'});
  [['Actuel',cp],['Nouveau',np],['Confirmer',cfp]].forEach(([l,i])=>{const g=el('div',{class:'form-group'});g.appendChild(el('label',{text:l}));g.appendChild(i);b2.appendChild(g)});
  const cpBtn=el('button',{class:'btn btn-secondary',style:'width:auto;margin-top:8px'});cpBtn.appendChild(svg('shield',16));cpBtn.appendChild(document.createTextNode(' Changer'));
  cpBtn.onclick=async()=>{pe.style.display='none';if(np.value!==cfp.value){pe.textContent='Mots de passe differents';pe.style.display='block';return}const r=await prof.changePwd(cp.value,np.value);if(r.ok)toast('Mot de passe change','success');else{pe.textContent=r.error;pe.style.display='block'}};
  b2.appendChild(cpBtn);b2.appendChild(pe);c2.appendChild(b2);grid.appendChild(c2);c.appendChild(grid);
}

// ── SETTINGS ──
async function pageSettings(c){
  const grid=el('div',{class:'grid-2-wide',style:'align-items:start'});
  const c1=el('div',{class:'card'});const h1=el('div',{class:'card-head'});const t1=el('h3');t1.appendChild(svg('chart',18));t1.appendChild(document.createTextNode(' Rapports'));h1.appendChild(t1);c1.appendChild(h1);
  const b1=el('div',{class:'card-body'});
  const now=new Date();const lm=new Date(now);lm.setMonth(lm.getMonth()-1);const ms=lm.toISOString().slice(0,7);
  [['Rapport mensuel — '+ms,'/api/v1/export/monthly/excel?month='+ms,'/api/v1/export/monthly/pdf?month='+ms],['Patients actifs','/api/v1/export/patients/excel?status=active','/api/v1/export/patients/pdf?status=active'],['Perdus de vue','/api/v1/export/patients/excel?status=perdu_de_vue','/api/v1/export/patients/pdf?status=perdu_de_vue']].forEach(([label,exUrl,pdfUrl])=>{
    const item=el('div',{class:'list-item'});item.appendChild(el('div',{style:'flex:1',text:label}));
    const acts=el('div',{class:'pt-acts'});
    const eb=el('a',{class:'btn btn-sm btn-secondary',href:exUrl,style:'width:auto;text-decoration:none'});eb.appendChild(svg('download',14));eb.appendChild(document.createTextNode(' Excel'));acts.appendChild(eb);
    const pb=el('a',{class:'btn btn-sm btn-secondary',href:pdfUrl,style:'width:auto;text-decoration:none'});pb.appendChild(svg('download',14));pb.appendChild(document.createTextNode(' PDF'));acts.appendChild(pb);
    item.appendChild(acts);b1.appendChild(item)});
  c1.appendChild(b1);grid.appendChild(c1);
  const c2=el('div',{class:'card'});const h2=el('div',{class:'card-head'});const t2=el('h3');t2.appendChild(svg('users',18));t2.appendChild(document.createTextNode(' Equipe'));h2.appendChild(t2);
  const link=el('a',{class:'card-link',text:'Gerer'});link.onclick=()=>{location.hash='#users'};h2.appendChild(link);c2.appendChild(h2);
  const b2=el('div',{class:'card-body'});b2.appendChild(empty('Accedez a la gestion des utilisateurs'));c2.appendChild(b2);grid.appendChild(c2);c.appendChild(grid);
}

// ── HELP ──
async function pageHelp(c){
  c.appendChild(el('div',{style:'text-align:center;margin-bottom:28px'},[el('h3',{text:'Comment pouvons-nous vous aider ?',style:'font-size:1.4rem;margin-bottom:16px'})]));
  const guides=el('div',{class:'grid-2',style:'align-items:start;margin-bottom:24px'});
  [['user-plus','green','Inscrire un patient',['Cliquez sur Nouveau patient','Remplissez nom, prenom, telephone','Selectionnez la langue','Cliquez Inscrire']],
   ['calendar','blue','Programmer un RDV',['Allez dans Prise de RDV','Recherchez le patient','Choisissez date et creneau','Confirmez']],
   ['clipboard','amber','Gerer un RDV',['Depuis le calendrier, cliquez sur un evenement','Marquez termine, manque ou reporte','Ajoutez des notes']],
   ['bell','red','Configurer les SMS',['Allez dans Parametres','Configurez le fournisseur SMS','Activez les rappels J-7, J-2']]
  ].forEach(([ic,col,title,steps])=>{
    const card=el('div',{class:'card',style:'cursor:pointer'});const body=el('div',{class:'card-body',style:'padding:18px 20px'});
    const hdr=el('div',{style:'display:flex;align-items:center;gap:12px'});const iw=el('div',{class:'stat-icon '+col,style:'width:36px;height:36px;margin-bottom:0'});iw.appendChild(svg(ic,18));hdr.appendChild(iw);hdr.appendChild(el('div',{text:title,style:'font-weight:600;font-size:.9rem'}));
    const ch=svg('down',16);ch.style.cssText='margin-left:auto;color:var(--gray-300);transition:.2s';hdr.appendChild(ch);body.appendChild(hdr);
    const cnt=el('div',{style:'display:none;padding-top:10px;border-top:1px solid var(--gray-50);margin-top:10px;font-size:.85rem;color:var(--gray-500);line-height:1.7'});
    steps.forEach(s=>cnt.appendChild(el('p',{text:s})));body.appendChild(cnt);
    card.onclick=()=>{const open=cnt.style.display==='block';cnt.style.display=open?'none':'block';ch.style.transform=open?'':'rotate(180deg)'};
    card.appendChild(body);guides.appendChild(card)});
  c.appendChild(guides);
  c.appendChild(el('div',{style:'text-align:center;margin-top:24px;padding:16px;color:var(--gray-300);font-size:.78rem',text:'MaSante v1.0.0 — masante.africa'}));
}

// ── Start ──
boot();
