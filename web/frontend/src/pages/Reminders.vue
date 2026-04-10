<script setup>
import { ref, onMounted } from 'vue'
import { reminders } from '../api'
import { useToast } from '../composables/useToast'

const toast = useToast()
const stats = ref({ total: 0, sent: 0, pending: 0, failed: 0 })
const queue = ref([])
const templates = ref([])
const loading = ref(true)
const error = ref('')
const savingTemplate = ref(null)

onMounted(async () => {
  error.value = ''
  const [sR, qR, tR] = await Promise.all([reminders.stats(), reminders.list(), reminders.templates()])
  loading.value = false
  if (!sR.ok) { error.value = sR.error; return }
  stats.value = sR.data || stats.value
  if (qR.ok) queue.value = qR.data || []
  if (tR.ok) templates.value = (tR.data || []).map(t => ({ ...t, _body: t.Body || '' }))
})

function statusLabel(s) { return { sent: 'Envoye', pending: 'En attente', failed: 'Echoue', cancelled: 'Annule' }[s] || s }
function statusPill(s) { return { sent: 'pill-success', pending: 'pill-warning', failed: 'pill-danger', cancelled: 'pill-neutral' }[s] || 'pill-neutral' }
function fmtDate(d) { return d ? new Date(d).toLocaleDateString('fr-FR', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' }) : '—' }

async function saveTemplate(tpl) {
  savingTemplate.value = tpl.ID
  error.value = ''
  const res = await reminders.updateTemplate(tpl.ID, { body: tpl._body })
  savingTemplate.value = null
  if (!res.ok) { error.value = res.error; return }
  tpl.Body = tpl._body
  toast.success('Modele sauvegarde')
}

async function sendAll() {
  error.value = ''
  const res = await reminders.sendAll()
  if (!res.ok) { error.value = res.error; return }
  toast.success('Envoi des rappels lance')
}
</script>

<template>
  <div v-if="loading" class="ms-loading">Chargement...</div>
  <template v-else>
    <div v-if="error" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin-bottom:12px">{{ error }}</div>

    <!-- Stats -->
    <div class="stats-row" style="margin-bottom:20px">
      <div class="stat"><div class="stat-icon blue">&#128172;</div><div class="stat-val">{{ stats.total }}</div><div class="stat-label">Total rappels</div></div>
      <div class="stat"><div class="stat-icon green">&#9989;</div><div class="stat-val">{{ stats.sent }}</div><div class="stat-label">Envoyes</div></div>
      <div class="stat"><div class="stat-icon amber">&#9203;</div><div class="stat-val">{{ stats.pending }}</div><div class="stat-label">En attente</div></div>
      <div class="stat"><div class="stat-icon red">&#10060;</div><div class="stat-val">{{ stats.failed }}</div><div class="stat-label">Echoues</div></div>
    </div>

    <div class="grid-2" style="align-items:start">
      <!-- Queue -->
      <div class="card">
        <div class="card-head">
          <h3>File d'envoi</h3>
          <button class="btn btn-sm btn-primary" @click="sendAll">Envoyer tout</button>
        </div>
        <div class="card-body">
          <div v-if="!queue.length" class="ms-empty">Aucun rappel en file</div>
          <div v-for="r in queue" :key="r.ID" class="list-item">
            <div class="item-info">
              <div class="item-name">{{ r.PatientName || 'Patient' }}</div>
              <div class="item-sub">{{ r.Channel || 'sms' }} — {{ fmtDate(r.ScheduledAt) }}</div>
            </div>
            <span :class="'pill ' + statusPill(r.Status)">{{ statusLabel(r.Status) }}</span>
          </div>
        </div>
      </div>

      <!-- Templates -->
      <div class="card">
        <div class="card-head"><h3>Modeles de message</h3></div>
        <div class="card-body">
          <div v-if="!templates.length" class="ms-empty">Aucun modele</div>
          <div v-for="tpl in templates" :key="tpl.ID" style="margin-bottom:16px;padding-bottom:16px;border-bottom:1px solid var(--gray-50)">
            <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:6px">
              <div style="font-weight:600;font-size:.85rem">{{ tpl.Name || tpl.Type || 'Modele' }}</div>
              <span class="pill pill-info" style="font-size:.7rem">{{ tpl.Channel || 'sms' }}</span>
            </div>
            <div class="form-group" style="margin-bottom:8px">
              <textarea class="form-input" v-model="tpl._body" rows="3" style="resize:vertical;font-size:.82rem"></textarea>
            </div>
            <button class="btn btn-sm btn-secondary" @click="saveTemplate(tpl)" :disabled="savingTemplate === tpl.ID || tpl._body === tpl.Body">{{ savingTemplate === tpl.ID ? 'Sauvegarde...' : 'Sauvegarder' }}</button>
          </div>
        </div>
      </div>
    </div>
  </template>
</template>
