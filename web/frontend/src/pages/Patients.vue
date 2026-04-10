<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { patients as api } from '../api'

const router = useRouter()
const route = useRoute()
const list = ref([])
const total = ref(0)
const loading = ref(true)
const activeFilter = ref('')

const filters = [
  { label: 'Tous', value: '' },
  { label: 'Actifs', value: 'active' },
  { label: 'A surveiller', value: 'a_surveiller' },
  { label: 'Perdus de vue', value: 'perdu_de_vue' },
  { label: 'Sortis', value: 'sorti' },
]

async function load() {
  loading.value = true
  let params = 'page=1&per_page=50'
  if (activeFilter.value) params += '&status=' + activeFilter.value
  if (route.query.q) params += '&q=' + encodeURIComponent(route.query.q)
  const res = await api.list(params)
  loading.value = false
  if (res.ok) { list.value = res.data.patients || []; total.value = res.data.total || 0 }
}

function initials(p) { return ((p.LastName || '')[0] + (p.FirstName || '')[0]).toUpperCase() }
function riskLabel(s) { return s <= 3 ? 'Faible' : s <= 6 ? 'Moyen' : 'Eleve' }
function riskClass(s) { return s <= 3 ? 'low' : s <= 6 ? 'med' : 'high' }
function statusClass(s) { return { active: 'pill-success', a_surveiller: 'pill-warning', perdu_de_vue: 'pill-danger', sorti: 'pill-neutral' }[s] || 'pill-neutral' }
function statusLabel(s) { return { active: 'Actif', a_surveiller: 'A surveiller', perdu_de_vue: 'Perdu de vue', sorti: 'Sorti' }[s] || s }

onMounted(load)
watch(() => route.query.q, load)

const avatarColors = ['a1', 'a2', 'a3', 'a4', 'a5', 'a6']
</script>

<template>
  <div class="filters">
    <button v-for="f in filters" :key="f.value" class="fbtn" :class="{ on: activeFilter === f.value }" @click="activeFilter = f.value; load()">{{ f.label }}</button>
  </div>

  <div v-if="loading" class="ms-loading">Chargement...</div>
  <div v-else-if="!list.length" class="ms-empty">Aucun patient</div>
  <template v-else>
    <div class="ptable">
      <div class="pt-head" style="grid-template-columns:44px 2fr 1fr 1fr 1fr 100px">
        <div></div><div>Patient</div><div>Risque</div><div>Derniere visite</div><div>Statut</div><div></div>
      </div>
      <div v-for="(p, i) in list" :key="p.ID" class="pt-row" style="grid-template-columns:44px 2fr 1fr 1fr 1fr 100px;cursor:pointer" @click="router.push('/patient/' + p.ID)">
        <div><div :class="'avatar ' + avatarColors[i % 6]">{{ initials(p) }}</div></div>
        <div><div class="pt-name">{{ p.LastName }} {{ p.FirstName }}</div><div class="pt-code">{{ p.Code }}</div></div>
        <div><span :class="'risk ' + riskClass(p.RiskScore)"><span class="risk-dot"></span> {{ riskLabel(p.RiskScore) }}</span></div>
        <div class="pt-cell">{{ p.UpdatedAt ? new Date(p.UpdatedAt).toLocaleDateString('fr-FR') : '—' }}</div>
        <div><span :class="'pill ' + statusClass(p.Status)">{{ statusLabel(p.Status) }}</span></div>
        <div class="pt-acts"><button class="icon-btn" @click.stop="router.push('/new-apt')"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="4" width="18" height="18" rx="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg></button></div>
      </div>
    </div>
    <div style="text-align:center;padding:12px;color:var(--gray-400);font-size:.82rem">{{ total }} patients</div>
  </template>
</template>
