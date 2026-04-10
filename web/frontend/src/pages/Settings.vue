<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()
const apiBase = '/api/v1'

const exports = [
  { label: 'Rapport mensuel (Excel)', url: apiBase + '/export/monthly?format=xlsx', icon: '\uD83D\uDCCA', desc: 'Statistiques du mois en cours' },
  { label: 'Rapport mensuel (PDF)', url: apiBase + '/export/monthly?format=pdf', icon: '\uD83D\uDCC4', desc: 'Version imprimable du rapport' },
  { label: 'Patients actifs (Excel)', url: apiBase + '/export/patients?status=active&format=xlsx', icon: '\uD83D\uDC65', desc: 'Liste de tous les patients actifs' },
  { label: 'Patients perdus de vue (Excel)', url: apiBase + '/export/patients?status=perdu_de_vue&format=xlsx', icon: '\u26A0', desc: 'Patients sans visite recente' },
]
</script>

<template>
  <div class="grid-2" style="align-items:start">
    <div class="card">
      <div class="card-head"><h3>Exports et rapports</h3></div>
      <div class="card-body">
        <a v-for="exp in exports" :key="exp.url" :href="exp.url" target="_blank" rel="noopener" class="list-item" style="cursor:pointer;text-decoration:none;color:inherit">
          <div style="font-size:1.4rem;flex-shrink:0;width:36px;text-align:center">{{ exp.icon }}</div>
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
          <div style="font-size:1.4rem;flex-shrink:0;width:36px;text-align:center">&#128100;</div>
          <div class="item-info">
            <div class="item-name">Gestion des utilisateurs</div>
            <div class="item-sub">Ajouter, modifier ou desactiver des comptes</div>
          </div>
        </div>
        <div class="list-item" style="cursor:pointer" @click="router.push('/profile')">
          <div style="font-size:1.4rem;flex-shrink:0;width:36px;text-align:center">&#9881;</div>
          <div class="item-info">
            <div class="item-name">Mon profil</div>
            <div class="item-sub">Modifier vos informations personnelles</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
