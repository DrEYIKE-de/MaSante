<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { dashboard } from '../api'

const router = useRouter()
const stats = ref({ patients: {}, appointments: {} })
const todayApts = ref([])
const overdueApts = ref([])
const loading = ref(true)

const totalPatients = ref(0)
const todayCount = ref(0)
const lostCount = ref(0)
const retention = ref(0)

onMounted(async () => {
  const [sR, tR, oR] = await Promise.all([dashboard.stats(), dashboard.today(), dashboard.overdue()])
  loading.value = false
  if (sR.ok) {
    const p = sR.data.patients || {}
    const a = sR.data.appointments || {}
    totalPatients.value = Object.values(p).reduce((s, v) => s + v, 0)
    todayCount.value = Object.values(a).reduce((s, v) => s + v, 0)
    lostCount.value = p.perdu_de_vue || 0
    retention.value = totalPatients.value > 0 ? Math.round((1 - lostCount.value / totalPatients.value) * 100) : 0
  }
  if (tR.ok) todayApts.value = tR.data || []
  if (oR.ok) overdueApts.value = oR.data || []
})

function fmtType(t) { return { consultation: 'Consultation', retrait_medicaments: 'Retrait medicaments', bilan_sanguin: 'Bilan sanguin', club_adherence: "Club d'adherence" }[t] || t }
function initials(n) { return (n || '??').split(' ').map(w => w[0]).join('').substring(0, 2).toUpperCase() }
function daysLate(d) { return Math.max(1, Math.floor((Date.now() - new Date(d).getTime()) / 864e5)) }
</script>

<template>
  <div v-if="loading" class="ms-loading">Chargement...</div>
  <template v-else>
    <div class="stats-row">
      <div class="stat">
        <div class="stat-icon green"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg></div>
        <div class="stat-val">{{ totalPatients }}</div><div class="stat-label">Patients actifs</div>
      </div>
      <div class="stat">
        <div class="stat-icon blue"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="4" width="18" height="18" rx="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg></div>
        <div class="stat-val">{{ todayCount }}</div><div class="stat-label">RDV aujourd'hui</div>
      </div>
      <div class="stat">
        <div class="stat-icon amber"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg></div>
        <div class="stat-val">{{ retention }}%</div><div class="stat-label">Retention</div>
      </div>
      <div class="stat">
        <div class="stat-icon red"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg></div>
        <div class="stat-val">{{ lostCount }}</div><div class="stat-label">Perdus de vue</div>
      </div>
    </div>

    <div class="grid-2" style="align-items:start">
      <div class="card">
        <div class="card-head"><h3>RDV du jour</h3><a class="card-link" @click="router.push('/calendar')">Calendrier</a></div>
        <div class="card-body">
          <div v-if="!todayApts.length" class="ms-empty">Aucun RDV aujourd'hui</div>
          <div v-for="apt in todayApts" :key="apt.ID" class="list-item">
            <span class="time-label">{{ apt.Time || '—' }}</span>
            <div class="avatar a1">{{ initials(apt.PatientName) }}</div>
            <div class="item-info">
              <div class="item-name">{{ apt.PatientName || 'Patient' }}</div>
              <div class="item-sub">{{ fmtType(apt.Type) }}</div>
            </div>
            <span class="pill" :class="{ 'pill-success': apt.Status === 'confirme', 'pill-warning': apt.Status === 'en_attente', 'pill-danger': apt.Status === 'manque', 'pill-info': apt.Status === 'termine' }">{{ apt.Status }}</span>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-head"><h3>Patients en retard</h3></div>
        <div class="card-body">
          <div v-if="!overdueApts.length" class="ms-empty">Aucun patient en retard</div>
          <div v-for="apt in overdueApts" :key="apt.ID" class="list-item">
            <div class="overdue-avatar">{{ initials(apt.PatientName) }}</div>
            <div class="item-info">
              <div class="item-name">{{ apt.PatientName || 'Patient' }}</div>
              <div class="overdue-days">{{ daysLate(apt.Date) }} jours de retard</div>
            </div>
            <button class="btn btn-sm btn-secondary" style="width:auto;flex-shrink:0" @click="router.push('/new-apt')">Reprogrammer</button>
          </div>
        </div>
      </div>
    </div>
  </template>
</template>
