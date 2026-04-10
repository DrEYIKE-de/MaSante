<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

const exports = [
  { label: 'Rapport mensuel (Excel)', url: '/api/v1/export/monthly/excel', desc: 'Statistiques du mois en cours' },
  { label: 'Rapport mensuel (PDF)', url: '/api/v1/export/monthly/pdf', desc: 'Version imprimable du rapport' },
  { label: 'Patients actifs (Excel)', url: '/api/v1/export/patients/excel?status=active', desc: 'Liste de tous les patients actifs' },
  { label: 'Patients actifs (PDF)', url: '/api/v1/export/patients/pdf?status=active', desc: 'Version imprimable' },
  { label: 'Patients perdus de vue (Excel)', url: '/api/v1/export/patients/excel?status=perdu_de_vue', desc: 'Patients sans visite recente' },
  { label: 'Patients perdus de vue (PDF)', url: '/api/v1/export/patients/pdf?status=perdu_de_vue', desc: 'Version imprimable' },
]
</script>

<template>
  <div class="grid-2" style="align-items:start">
    <div class="card">
      <div class="card-head"><h3>Exports et rapports</h3></div>
      <div class="card-body">
        <a v-for="exp in exports" :key="exp.url" :href="exp.url" target="_blank" rel="noopener" class="list-item" style="cursor:pointer;text-decoration:none;color:inherit">
          <div style="width:36px;height:36px;border-radius:8px;background:var(--gray-50);display:flex;align-items:center;justify-content:center;flex-shrink:0">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--gray-400)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
          </div>
          <div class="item-info">
            <div class="item-name">{{ exp.label }}</div>
            <div class="item-sub">{{ exp.desc }}</div>
          </div>
        </a>
      </div>
    </div>

    <div class="card">
      <div class="card-head"><h3>Administration</h3></div>
      <div class="card-body">
        <div class="list-item" style="cursor:pointer" @click="router.push('/users')">
          <div style="width:36px;height:36px;border-radius:8px;background:var(--gray-50);display:flex;align-items:center;justify-content:center;flex-shrink:0">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--gray-400)" stroke-width="1.5"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/></svg>
          </div>
          <div class="item-info">
            <div class="item-name">Gestion des utilisateurs</div>
            <div class="item-sub">Ajouter, modifier ou desactiver des comptes</div>
          </div>
        </div>
        <div class="list-item" style="cursor:pointer" @click="router.push('/profile')">
          <div style="width:36px;height:36px;border-radius:8px;background:var(--gray-50);display:flex;align-items:center;justify-content:center;flex-shrink:0">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--gray-400)" stroke-width="1.5"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-2.82 1.18V21a2 2 0 0 1-4 0v-.09a1.65 1.65 0 0 0-1.08-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1.08H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68V3a2 2 0 0 1 4 0v.09c.003.9.543 1.67 1.08 1.51a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.32 9c.26.6.85 1 1.51 1.08H21a2 2 0 0 1 0 4h-.09c-.66.003-1.25.4-1.51 1.08z"/></svg>
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
