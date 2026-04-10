<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { calendar } from '../api'

const router = useRouter()
const view = ref('week')
const currentDate = ref(new Date())
const weekData = ref([])
const loading = ref(true)
const error = ref('')
const selectedDay = ref(null)

function localDate(d) { return d.getFullYear()+'-'+String(d.getMonth()+1).padStart(2,'0')+'-'+String(d.getDate()).padStart(2,'0') }

const hours = Array.from({ length: 9 }, (_, i) => 8 + i) // 8:00 - 16:00

const weekStart = computed(() => {
  const d = new Date(currentDate.value)
  const day = d.getDay() || 7
  d.setDate(d.getDate() - day + 1)
  d.setHours(0, 0, 0, 0)
  return d
})

const weekDays = computed(() => {
  const days = []
  for (let i = 0; i < 7; i++) {
    const d = new Date(weekStart.value)
    d.setDate(d.getDate() + i)
    days.push(d)
  }
  return days
})

const monthLabel = computed(() => {
  const opts = { month: 'long', year: 'numeric' }
  return currentDate.value.toLocaleDateString('fr-FR', opts)
})

const weekLabel = computed(() => {
  const s = weekDays.value[0]
  const e = weekDays.value[6]
  return s.toLocaleDateString('fr-FR', { day: 'numeric', month: 'short' }) + ' - ' + e.toLocaleDateString('fr-FR', { day: 'numeric', month: 'short', year: 'numeric' })
})

const calendarGrid = computed(() => {
  const first = new Date(currentDate.value.getFullYear(), currentDate.value.getMonth(), 1)
  const lastDay = new Date(currentDate.value.getFullYear(), currentDate.value.getMonth() + 1, 0).getDate()
  const startDay = (first.getDay() || 7) - 1
  const weeks = []
  let week = new Array(startDay).fill(null)
  for (let d = 1; d <= lastDay; d++) {
    week.push(d)
    if (week.length === 7) { weeks.push(week); week = [] }
  }
  if (week.length) { while (week.length < 7) week.push(null); weeks.push(week) }
  return weeks
})

function aptCountForDay(day) {
  if (!day) return 0
  const dateStr = localDate(new Date(currentDate.value.getFullYear(), currentDate.value.getMonth(), day))
  return weekData.value.filter(a => a.Date === dateStr).length
}

function aptsForDayHour(date, hour) {
  const dateStr = localDate(date)
  const hourStr = String(hour).padStart(2, '0')
  return weekData.value.filter(a => a.Date === dateStr && a.Time && a.Time.startsWith(hourStr))
}

function selectedDayApts() {
  if (!selectedDay.value) return []
  const dateStr = localDate(new Date(currentDate.value.getFullYear(), currentDate.value.getMonth(), selectedDay.value))
  return weekData.value.filter(a => a.Date === dateStr)
}

function fmtType(t) { return { consultation: 'Consultation', retrait_medicaments: 'Retrait med.', bilan_sanguin: 'Bilan sanguin', club_adherence: "Club d'adh." }[t] || t }
function dayName(d) { return d.toLocaleDateString('fr-FR', { weekday: 'short' }) }
function isToday(d) { const t = new Date(); return d.getDate() === t.getDate() && d.getMonth() === t.getMonth() && d.getFullYear() === t.getFullYear() }
function isTodayNum(day) { if (!day) return false; const t = new Date(); return day === t.getDate() && currentDate.value.getMonth() === t.getMonth() && currentDate.value.getFullYear() === t.getFullYear() }

async function loadWeek() {
  loading.value = true
  error.value = ''
  const res = await calendar.week(localDate(weekStart.value))
  loading.value = false
  if (!res.ok) { error.value = res.error; return }
  weekData.value = res.data || []
}

function prev() {
  const d = new Date(currentDate.value)
  if (view.value === 'week') d.setDate(d.getDate() - 7)
  else d.setMonth(d.getMonth() - 1)
  currentDate.value = d
  selectedDay.value = null
}

function next() {
  const d = new Date(currentDate.value)
  if (view.value === 'week') d.setDate(d.getDate() + 7)
  else d.setMonth(d.getMonth() + 1)
  currentDate.value = d
  selectedDay.value = null
}

function goToday() {
  currentDate.value = new Date()
  selectedDay.value = null
}

function selectDay(day) {
  if (!day) return
  selectedDay.value = selectedDay.value === day ? null : day
}

onMounted(loadWeek)
watch([currentDate, view], loadWeek)
</script>

<template>
  <div class="card" style="margin-bottom:16px">
    <div class="card-head" style="flex-wrap:wrap;gap:8px">
      <div style="display:flex;gap:6px">
        <button class="fbtn" :class="{ on: view === 'week' }" @click="view = 'week'">Semaine</button>
        <button class="fbtn" :class="{ on: view === 'month' }" @click="view = 'month'">Mois</button>
      </div>
      <h3 style="flex:1;text-align:center">{{ view === 'week' ? weekLabel : monthLabel }}</h3>
      <div style="display:flex;gap:6px">
        <button class="btn btn-sm btn-secondary" @click="prev">&#8592;</button>
        <button class="btn btn-sm btn-secondary" @click="goToday">Aujourd'hui</button>
        <button class="btn btn-sm btn-secondary" @click="next">&#8594;</button>
        <button class="btn btn-sm btn-primary" @click="router.push('/new-apt')">+ RDV</button>
      </div>
    </div>
  </div>

  <div v-if="error" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin-bottom:12px">{{ error }}</div>
  <div v-if="loading" class="ms-loading">Chargement...</div>

  <!-- WEEK VIEW -->
  <template v-else-if="view === 'week'">
    <div class="card">
      <div class="card-body" style="overflow-x:auto;padding:0">
        <table style="width:100%;border-collapse:collapse;font-size:.82rem;min-width:700px">
          <thead>
            <tr>
              <th style="width:60px;padding:8px;border-bottom:1px solid var(--gray-100);text-align:left">Heure</th>
              <th v-for="d in weekDays" :key="d.toISOString()" style="padding:8px;border-bottom:1px solid var(--gray-100);text-align:center;min-width:80px" :style="isToday(d) ? 'background:var(--primary-bg)' : ''">
                <div style="font-weight:600;text-transform:capitalize">{{ dayName(d) }}</div>
                <div style="font-size:.75rem;color:var(--gray-400)">{{ d.getDate() }}</div>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="h in hours" :key="h">
              <td style="padding:6px 8px;border-bottom:1px solid var(--gray-50);color:var(--gray-400);font-size:.75rem;vertical-align:top">{{ String(h).padStart(2,'0') }}:00</td>
              <td v-for="d in weekDays" :key="d.toISOString()+h" style="padding:4px;border-bottom:1px solid var(--gray-50);border-left:1px solid var(--gray-50);vertical-align:top" :style="isToday(d) ? 'background:var(--primary-bg)' : ''">
                <div v-for="apt in aptsForDayHour(d, h)" :key="apt.ID" style="padding:3px 6px;background:var(--primary);color:#fff;border-radius:4px;margin-bottom:2px;font-size:.72rem;cursor:default" :title="apt.PatientName + ' - ' + fmtType(apt.Type)">
                  {{ apt.Time ? apt.Time.substring(0,5) : '' }} {{ (apt.PatientName || '').split(' ')[0] }}
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </template>

  <!-- MONTH VIEW -->
  <template v-else>
    <div class="grid-2" style="align-items:start">
      <div class="card">
        <div class="card-body" style="padding:8px">
          <table style="width:100%;border-collapse:collapse;font-size:.82rem">
            <thead>
              <tr>
                <th v-for="dn in ['Lun','Mar','Mer','Jeu','Ven','Sam','Dim']" :key="dn" style="padding:8px;text-align:center;color:var(--gray-400);font-size:.75rem">{{ dn }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(week, wi) in calendarGrid" :key="wi">
                <td v-for="(day, di) in week" :key="di" style="padding:4px;text-align:center;cursor:pointer;border-radius:6px" :style="[isTodayNum(day) ? 'background:var(--primary-bg);font-weight:700' : '', selectedDay === day && day ? 'background:var(--primary);color:#fff' : '']" @click="selectDay(day)">
                  <template v-if="day">
                    <div style="font-size:.85rem;padding:4px">{{ day }}</div>
                    <div v-if="aptCountForDay(day)" style="display:inline-block;background:var(--primary);color:#fff;border-radius:10px;font-size:.65rem;padding:1px 6px;min-width:18px">{{ aptCountForDay(day) }}</div>
                  </template>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="card">
        <div class="card-head"><h3>{{ selectedDay ? 'RDV du ' + selectedDay + '/' + (currentDate.getMonth()+1) : 'Selectionnez un jour' }}</h3></div>
        <div class="card-body">
          <div v-if="!selectedDay" class="ms-empty">Cliquez sur un jour pour voir les details</div>
          <template v-else>
            <div v-if="!selectedDayApts().length" class="ms-empty">Aucun RDV ce jour</div>
            <div v-for="apt in selectedDayApts()" :key="apt.ID" class="list-item">
              <span class="time-label">{{ apt.Time || '--:--' }}</span>
              <div class="item-info">
                <div class="item-name">{{ apt.PatientName || 'Patient' }}</div>
                <div class="item-sub">{{ fmtType(apt.Type) }}</div>
              </div>
              <span class="pill" :class="{ 'pill-success': apt.Status === 'confirme', 'pill-warning': apt.Status === 'en_attente', 'pill-danger': apt.Status === 'manque' }">{{ apt.Status }}</span>
            </div>
          </template>
        </div>
      </div>
    </div>
  </template>
</template>
