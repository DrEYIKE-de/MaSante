<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

const reports = [
  {
    label: 'Rapport mensuel',
    desc: 'Resume des patients, rendez-vous et indicateurs du mois',
    excel: '/api/v1/export/monthly/excel',
    pdf: '/api/v1/export/monthly/pdf',
  },
  {
    label: 'Liste de tous les patients',
    desc: 'Tous les patients enregistres avec statut, risque et coordonnees',
    excel: '/api/v1/export/patients/excel',
    pdf: '/api/v1/export/patients/pdf',
  },
  {
    label: 'Patients actifs',
    desc: 'Patients actuellement suivis dans le programme',
    excel: '/api/v1/export/patients/excel?status=active',
    pdf: '/api/v1/export/patients/pdf?status=active',
  },
  {
    label: 'Patients a surveiller',
    desc: 'Patients avec un risque eleve de rendez-vous manque',
    excel: '/api/v1/export/patients/excel?status=a_surveiller',
    pdf: '/api/v1/export/patients/pdf?status=a_surveiller',
  },
  {
    label: 'Patients perdus de vue',
    desc: 'Patients sans visite depuis plus de 90 jours',
    excel: '/api/v1/export/patients/excel?status=perdu_de_vue',
    pdf: '/api/v1/export/patients/pdf?status=perdu_de_vue',
  },
  {
    label: 'Patients sortis du programme',
    desc: 'Deces, transferts, abandons, guerisons',
    excel: '/api/v1/export/patients/excel?status=sorti',
    pdf: '/api/v1/export/patients/pdf?status=sorti',
  },
]
</script>

<template>
  <div class="grid-2" style="align-items:start">
    <!-- Exports -->
    <div class="card">
      <div class="card-head"><h3>Rapports et exports</h3></div>
      <div class="card-body">
        <div v-for="r in reports" :key="r.label" class="list-item" style="gap:12px">
          <div style="width:36px;height:36px;border-radius:8px;background:var(--gray-50);display:flex;align-items:center;justify-content:center;flex-shrink:0">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--gray-400)" stroke-width="1.5" stroke-linecap="round">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>
            </svg>
          </div>
          <div class="item-info" style="flex:1">
            <div class="item-name">{{ r.label }}</div>
            <div class="item-sub">{{ r.desc }}</div>
          </div>
          <div style="display:flex;gap:6px;flex-shrink:0">
            <a :href="r.excel" target="_blank" rel="noopener" class="btn btn-sm btn-secondary" style="width:auto;text-decoration:none;display:flex;align-items:center;gap:4px">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
              Excel
            </a>
            <a :href="r.pdf" target="_blank" rel="noopener" class="btn btn-sm btn-secondary" style="width:auto;text-decoration:none;display:flex;align-items:center;gap:4px">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
              PDF
            </a>
          </div>
        </div>
      </div>
    </div>

    <!-- Administration -->
    <div class="card">
      <div class="card-head"><h3>Administration</h3></div>
      <div class="card-body">
        <div class="list-item" style="cursor:pointer" @click="router.push('/users')">
          <div style="width:36px;height:36px;border-radius:8px;background:var(--gray-50);display:flex;align-items:center;justify-content:center;flex-shrink:0">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--gray-400)" stroke-width="1.5" stroke-linecap="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/></svg>
          </div>
          <div class="item-info">
            <div class="item-name">Gestion des utilisateurs</div>
            <div class="item-sub">Ajouter, modifier ou desactiver des comptes</div>
          </div>
        </div>
        <div class="list-item" style="cursor:pointer" @click="router.push('/profile')">
          <div style="width:36px;height:36px;border-radius:8px;background:var(--gray-50);display:flex;align-items:center;justify-content:center;flex-shrink:0">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--gray-400)" stroke-width="1.5" stroke-linecap="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
          </div>
          <div class="item-info">
            <div class="item-name">Mon profil</div>
            <div class="item-sub">Modifier vos informations personnelles</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
