<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { calendar, appointments } from '../api'
import { useToast } from '../composables/useToast'

const router = useRouter()
const toast = useToast()

// Modal state.
const showModal = ref(false)
const modalApt = ref(null)
const modalStatus = ref('')
const modalNotes = ref('')
const modalSaving = ref(false)

function openAptModal(apt) {
  modalApt.value = apt
  modalStatus.value = ''
  modalNotes.value = apt.Notes || ''
  showModal.value = true
}

async function updateAptStatus() {
  if (!modalStatus.value || !modalApt.value) return
  modalSaving.value = true
  let res
  if (modalStatus.value === 'termine') {
    res = await appointments.complete(modalApt.value.ID, { notes: modalNotes.value, follow_up_freq: 'trimestriel' })
  } else if (modalStatus.value === 'manque') {
    res = await appointments.missed(modalApt.value.ID, { notes: modalNotes.value, reschedule: false, send_reminder: true })
  } else if (modalStatus.value === 'annule') {
    res = await appointments.cancel(modalApt.value.ID)
  } else if (modalStatus.value === 'reporte') {
    // Reschedule needs a new date — for now just mark as missed with reschedule
    res = await appointments.missed(modalApt.value.ID, { notes: modalNotes.value, reschedule: true, reschedule_days: 7 })
  }
  modalSaving.value = false
  if (res && !res.ok) { toast.error(res.error); return }
  toast.success('RDV mis a jour')
  showModal.value = false
  loadData()
}
const view = ref('week')
const currentDate = ref(new Date())
const aptData = ref([])
const loading = ref(true)
const error = ref('')
const selectedDay = ref(null)

function localDate(d) { return d.getFullYear()+'-'+String(d.getMonth()+1).padStart(2,'0')+'-'+String(d.getDate()).padStart(2,'0') }
function aptDate(a) { return (a.Date || '').slice(0, 10) }
function aptHour(a) { return (a.Time || '').slice(0, 2) }

const hours = [8, 9, 10, 11, 12, 13, 14, 15, 16]

const weekStart = computed(() => {
  const d = new Date(currentDate.value)
  const day = d.getDay() || 7
  d.setDate(d.getDate() - day + 1)
  d.setHours(0,0,0,0)
  return d
})

const weekDays = computed(() => {
  return Array.from({length:7}, (_,i) => {
    const d = new Date(weekStart.value)
    d.setDate(d.getDate() + i)
    return d
  })
})

const weekLabel = computed(() => {
  const s = weekDays.value[0], e = weekDays.value[6]
  return s.toLocaleDateString('fr-FR',{day:'numeric',month:'short'}) + ' - ' + e.toLocaleDateString('fr-FR',{day:'numeric',month:'short',year:'numeric'})
})

const monthLabel = computed(() => currentDate.value.toLocaleDateString('fr-FR',{month:'long',year:'numeric'}))

const calendarGrid = computed(() => {
  const year = currentDate.value.getFullYear(), month = currentDate.value.getMonth()
  const first = new Date(year, month, 1)
  const lastDay = new Date(year, month+1, 0).getDate()
  const startDay = (first.getDay() || 7) - 1
  const weeks = []
  let week = new Array(startDay).fill(null)
  for (let d=1; d<=lastDay; d++) {
    week.push(d)
    if (week.length === 7) { weeks.push(week); week = [] }
  }
  if (week.length) { while(week.length<7) week.push(null); weeks.push(week) }
  return weeks
})

function aptsForDate(dateStr) { return aptData.value.filter(a => aptDate(a) === dateStr) }
function aptsForDayHour(d, h) { const ds = localDate(d); const hs = String(h).padStart(2,'0'); return aptData.value.filter(a => aptDate(a) === ds && aptHour(a) === hs) }
function aptCountForDay(day) { if(!day) return 0; return aptsForDate(localDate(new Date(currentDate.value.getFullYear(), currentDate.value.getMonth(), day))).length }
function selectedDayApts() { if(!selectedDay.value) return []; return aptsForDate(localDate(new Date(currentDate.value.getFullYear(), currentDate.value.getMonth(), selectedDay.value))) }

function fmtType(t) { return {consultation:'Consultation',retrait_medicaments:'Retrait med.',bilan_sanguin:'Bilan sanguin',club_adherence:"Club d'adh."}[t]||t }
function statusBg(s) { return {confirme:'var(--success-bg)',en_attente:'var(--warning-bg)',manque:'var(--danger-bg)',termine:'var(--info-bg)'}[s]||'var(--primary-subtle)' }
function statusFg(s) { return {confirme:'var(--success)',en_attente:'var(--warning)',manque:'var(--danger)',termine:'var(--info)'}[s]||'var(--primary)' }
function statusLabel(s) { return {confirme:'Confirme',en_attente:'En attente',manque:'Manque',termine:'Termine',annule:'Annule',reporte:'Reporte'}[s]||s }
function dayName(d) { return d.toLocaleDateString('fr-FR',{weekday:'short'}) }
function isToday(d) { const t=new Date(); return d.getDate()===t.getDate()&&d.getMonth()===t.getMonth()&&d.getFullYear()===t.getFullYear() }
function isTodayNum(day) { if(!day) return false; const t=new Date(); return day===t.getDate()&&currentDate.value.getMonth()===t.getMonth()&&currentDate.value.getFullYear()===t.getFullYear() }

async function loadData() {
  loading.value = true
  error.value = ''
  aptData.value = []

  if (view.value === 'week') {
    const res = await calendar.week(localDate(weekStart.value))
    if (!res.ok) { error.value = res.error } else { aptData.value = res.data || [] }
  } else {
    // Load all weeks of the month.
    const year = currentDate.value.getFullYear(), month = currentDate.value.getMonth()
    const first = new Date(year, month, 1)
    const firstMon = new Date(first)
    firstMon.setDate(first.getDate() - ((first.getDay()+6)%7))
    const all = []
    for (let w=0; w<6; w++) {
      const ws = new Date(firstMon)
      ws.setDate(firstMon.getDate() + w*7)
      if (w > 0 && ws.getMonth() > month && ws.getFullYear() >= year) break
      const res = await calendar.week(localDate(ws))
      if (res.ok && res.data) all.push(...res.data)
    }
    aptData.value = all
  }
  loading.value = false
}

function prev() { const d=new Date(currentDate.value); if(view.value==='week')d.setDate(d.getDate()-7);else d.setMonth(d.getMonth()-1); currentDate.value=d; selectedDay.value=null }
function next() { const d=new Date(currentDate.value); if(view.value==='week')d.setDate(d.getDate()+7);else d.setMonth(d.getMonth()+1); currentDate.value=d; selectedDay.value=null }
function goToday() { currentDate.value=new Date(); selectedDay.value=null }

onMounted(loadData)
watch([currentDate, view], loadData)
</script>

<template>
  <!-- Controls -->
  <div class="card" style="margin-bottom:16px">
    <div class="card-head" style="flex-wrap:wrap;gap:8px">
      <div style="display:flex;gap:6px">
        <button class="fbtn" :class="{on:view==='week'}" @click="view='week'">Semaine</button>
        <button class="fbtn" :class="{on:view==='month'}" @click="view='month'">Mois</button>
      </div>
      <h3 style="flex:1;text-align:center;text-transform:capitalize">{{ view==='week'?weekLabel:monthLabel }}</h3>
      <div style="display:flex;gap:6px">
        <button class="btn btn-sm btn-secondary" style="width:auto" @click="prev">&larr;</button>
        <button class="btn btn-sm btn-secondary" style="width:auto" @click="goToday">Aujourd'hui</button>
        <button class="btn btn-sm btn-secondary" style="width:auto" @click="next">&rarr;</button>
        <button class="btn btn-sm btn-primary" style="width:auto" @click="router.push('/new-apt')">+ RDV</button>
      </div>
    </div>
  </div>

  <div v-if="error" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin-bottom:12px">{{ error }}</div>
  <div v-if="loading" class="ms-loading">Chargement...</div>

  <!-- WEEK -->
  <template v-else-if="view==='week'">
    <div class="card">
      <div style="overflow-x:auto;padding:0">
        <table style="width:100%;border-collapse:collapse;font-size:.82rem;min-width:700px">
          <thead>
            <tr>
              <th style="width:56px;padding:8px;border-bottom:1px solid var(--gray-100);text-align:left;font-size:.72rem;color:var(--gray-400)">Heure</th>
              <th v-for="d in weekDays" :key="localDate(d)" style="padding:8px;border-bottom:1px solid var(--gray-100);text-align:center;min-width:80px" :style="isToday(d)?'background:var(--primary-bg)':''">
                <div style="font-weight:600;text-transform:capitalize;font-size:.78rem">{{ dayName(d) }}</div>
                <div style="font-size:1.1rem;font-weight:700">{{ d.getDate() }}</div>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="h in hours" :key="h">
              <td style="padding:4px 8px;border-bottom:1px solid var(--gray-50);color:var(--gray-300);font-size:.72rem;vertical-align:top">{{ String(h).padStart(2,'0') }}:00</td>
              <td v-for="d in weekDays" :key="localDate(d)+h" style="padding:3px;border-bottom:1px solid var(--gray-50);border-left:1px solid var(--gray-50);vertical-align:top;min-height:40px" :style="isToday(d)?'background:var(--primary-bg)':''">
                <div v-for="apt in aptsForDayHour(d,h)" :key="apt.ID" @click="openAptModal(apt)" style="padding:2px 6px;border-radius:4px;margin-bottom:2px;font-size:.7rem;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;cursor:pointer" :style="'background:' + statusBg(apt.Status) + ';color:' + statusFg(apt.Status)" :title="apt.PatientName+' — '+fmtType(apt.Type)">
                  {{ apt.Time ? apt.Time.slice(0,5) : '' }} {{ (apt.PatientName||'').split(' ')[0] }}
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </template>

  <!-- MONTH -->
  <template v-else>
    <div class="grid-2" style="align-items:start">
      <div class="card">
        <div style="padding:12px">
          <table style="width:100%;border-collapse:collapse">
            <thead>
              <tr>
                <th v-for="dn in ['L','M','M','J','V','S','D']" :key="dn" style="padding:8px;text-align:center;color:var(--gray-400);font-size:.72rem;font-weight:600">{{ dn }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(week,wi) in calendarGrid" :key="wi">
                <td v-for="(day,di) in week" :key="di"
                  style="padding:4px;text-align:center;cursor:pointer;vertical-align:top"
                  @click="day && (selectedDay = selectedDay===day ? null : day)">
                  <div v-if="day"
                    style="width:40px;height:40px;border-radius:8px;display:flex;flex-direction:column;align-items:center;justify-content:center;margin:0 auto;transition:.15s"
                    :style="[
                      selectedDay===day ? 'background:var(--primary);color:#fff' : isTodayNum(day) ? 'background:var(--primary-bg);font-weight:700' : 'hover:background:var(--gray-50)'
                    ]">
                    <span style="font-size:.88rem;line-height:1">{{ day }}</span>
                    <span v-if="aptCountForDay(day)" style="font-size:.6rem;margin-top:1px;opacity:.7">{{ aptCountForDay(day) }} rdv</span>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="card">
        <div class="card-head" style="padding:12px 16px"><h3>{{ selectedDay ? selectedDay + ' ' + monthLabel : 'Selectionnez un jour' }}</h3></div>
        <div class="card-body" style="padding:8px 16px">
          <div v-if="!selectedDay" class="ms-empty" style="padding:16px">Cliquez sur un jour du calendrier</div>
          <template v-else>
            <div v-if="!selectedDayApts().length" class="ms-empty" style="padding:16px">Aucun RDV ce jour</div>
            <div v-for="apt in selectedDayApts()" :key="apt.ID" class="list-item" style="padding:8px 0;cursor:pointer" @click="openAptModal(apt)">
              <span style="font-size:.82rem;font-weight:600;color:var(--gray-500);min-width:44px">{{ apt.Time || '--:--' }}</span>
              <div class="item-info">
                <div class="item-name">{{ apt.PatientName || 'Patient' }}</div>
                <div class="item-sub">{{ fmtType(apt.Type) }}</div>
              </div>
              <span class="pill" :class="{'pill-success':apt.Status==='confirme','pill-warning':apt.Status==='en_attente','pill-danger':apt.Status==='manque','pill-info':apt.Status==='termine'}">{{ apt.Status }}</span>
            </div>
          </template>
        </div>
      </div>
    </div>
  </template>

  <!-- RDV Detail Modal -->
  <teleport to="body">
    <template v-if="showModal && modalApt">
      <div style="position:fixed;inset:0;background:rgba(0,0,0,.3);z-index:2000" @click="showModal=false"></div>
      <div style="position:fixed;top:50%;left:50%;transform:translate(-50%,-50%);width:480px;max-height:85vh;overflow-y:auto;background:var(--white);border-radius:var(--radius-xl);box-shadow:0 20px 60px rgba(0,0,0,.15);z-index:2001">
        <div class="card-head" style="padding:16px 20px;border-bottom:1px solid var(--gray-100)">
          <h3>Rendez-vous</h3>
          <button class="icon-btn" @click="showModal=false">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>

        <div style="padding:20px">
          <!-- Patient info -->
          <div style="display:flex;align-items:center;gap:12px;padding:12px;background:var(--gray-25);border-radius:var(--radius);margin-bottom:16px">
            <div class="avatar a1" style="width:40px;height:40px">{{ (modalApt.PatientName||'??').split(' ').map(w=>w[0]).join('').substring(0,2).toUpperCase() }}</div>
            <div>
              <div style="font-weight:600;font-size:.95rem">{{ modalApt.PatientName || 'Patient' }}</div>
              <div style="font-size:.82rem;color:var(--gray-400)">{{ fmtType(modalApt.Type) }} — {{ modalApt.Time || '--:--' }}, {{ (modalApt.Date||'').slice(0,10) }}</div>
            </div>
            <span class="pill" :class="{'pill-success':modalApt.Status==='confirme','pill-warning':modalApt.Status==='en_attente','pill-danger':modalApt.Status==='manque','pill-info':modalApt.Status==='termine'}" style="margin-left:auto">{{ statusLabel(modalApt.Status) }}</span>
          </div>

          <!-- Change status -->
          <div class="form-group">
            <label>Changer le statut</label>
            <div style="display:grid;grid-template-columns:1fr 1fr;gap:8px;margin-top:6px">
              <div v-for="[val,lbl,color] in [['termine','Termine','var(--success)'],['manque','Manque','var(--danger)'],['reporte','Reporte (+7 jours)','var(--warning)'],['annule','Annule','var(--gray-500)']]" :key="val"
                style="padding:10px 14px;border-radius:var(--radius);cursor:pointer;text-align:center;font-size:.85rem;font-weight:500;transition:.15s"
                :style="modalStatus===val ? 'background:'+color+';color:#fff;border:1.5px solid '+color : 'border:1.5px solid var(--gray-200);color:var(--gray-600)'"
                @click="modalStatus=val">
                {{ lbl }}
              </div>
            </div>
          </div>

          <!-- Notes -->
          <div class="form-group" style="margin-top:12px">
            <label>Notes</label>
            <textarea class="form-input" v-model="modalNotes" rows="3" placeholder="Observations, raison de l'absence..."></textarea>
          </div>

          <!-- Actions -->
          <div style="display:flex;gap:8px;margin-top:16px">
            <button class="btn btn-secondary" style="width:auto" @click="showModal=false">Fermer</button>
            <button class="btn btn-primary" style="width:auto" @click="updateAptStatus" :disabled="!modalStatus || modalSaving">{{ modalSaving ? 'Mise a jour...' : 'Valider' }}</button>
            <button class="btn btn-secondary" style="width:auto;margin-left:auto" @click="router.push('/patient/'+modalApt.PatientID)">Voir la fiche</button>
          </div>
        </div>
      </div>
    </template>
  </teleport>
</template>
