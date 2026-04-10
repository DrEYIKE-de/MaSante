<script setup>
import { ref, onMounted } from 'vue'
import { reminders } from '../api'
import { useToast } from '../composables/useToast'

const toast = useToast()
const stats = ref({ DeliveryRate: 0, ConfirmRate: 0, PendingCount: 0, FailedCount: 0 })
const queue = ref([])
const templates = ref([])
const loading = ref(true)
const error = ref('')
const savingTemplate = ref(null)

onMounted(async () => {
  const [sR, qR, tR] = await Promise.all([reminders.stats(), reminders.list(), reminders.templates()])
  loading.value = false
  if (sR.ok) stats.value = sR.data || stats.value
  if (qR.ok) queue.value = qR.data || []
  if (tR.ok) templates.value = (tR.data || []).map(t => ({ ...t, _body: t.Body || '' }))
})

const statusLabels = { planifie: 'Planifie', envoye: 'Envoye', recu: 'Recu', echec: 'Echec' }
const statusPills = { planifie: 'pill-warning', envoye: 'pill-success', recu: 'pill-info', echec: 'pill-danger' }
const channelLabels = { sms: 'SMS', whatsapp: 'WhatsApp', voice: 'Appel vocal' }
const typeLabels = { j7: 'J-7', j2: 'J-2', j0: 'Jour J', retard: 'Retard' }

async function saveTemplate(tpl) {
  savingTemplate.value = tpl.ID
  const res = await reminders.updateTemplate(tpl.ID, { body: tpl._body, is_active: tpl.IsActive })
  savingTemplate.value = null
  if (!res.ok) { toast.error(res.error); return }
  tpl.Body = tpl._body
  toast.success('Modele sauvegarde')
}

async function sendAll() {
  const res = await reminders.sendAll()
  if (!res.ok) { toast.error(res.error); return }
  toast.success('Envoi des rappels lance')
}
</script>

<template>
  <div v-if="loading" class="ms-loading">Chargement...</div>
  <template v-else>
    <div v-if="error" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin-bottom:12px">{{ error }}</div>

    <!-- Stats -->
    <div class="stats-row" style="margin-bottom:20px">
      <div class="stat">
        <div class="stat-icon green">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><polyline points="20 6 9 17 4 12"/></svg>
        </div>
        <div class="stat-val">{{ stats.DeliveryRate ? stats.DeliveryRate.toFixed(1) + '%' : '—' }}</div>
        <div class="stat-label">Taux de livraison</div>
      </div>
      <div class="stat">
        <div class="stat-icon amber">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
        </div>
        <div class="stat-val">{{ stats.PendingCount }}</div>
        <div class="stat-label">En attente</div>
      </div>
      <div class="stat">
        <div class="stat-icon red">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
        </div>
        <div class="stat-val">{{ stats.FailedCount }}</div>
        <div class="stat-label">Echecs</div>
      </div>
    </div>

    <div class="grid-2" style="align-items:start">
      <!-- Queue -->
      <div class="card">
        <div class="card-head">
          <h3>File d'envoi</h3>
          <button class="btn btn-sm btn-primary" style="width:auto" @click="sendAll">Envoyer tout</button>
        </div>
        <div class="card-body">
          <div v-if="!queue.length" class="ms-empty">Aucun rappel en file</div>
          <div v-for="r in queue" :key="r.ID" class="list-item">
            <div style="width:36px;height:36px;border-radius:8px;display:flex;align-items:center;justify-content:center;flex-shrink:0;background:var(--info-bg);color:var(--info)">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
            </div>
            <div class="item-info">
              <div class="item-name">{{ r.PatientName || 'Patient' }}</div>
              <div class="item-sub">{{ typeLabels[r.Type] || r.Type }} — {{ channelLabels[r.Channel] || r.Channel }}</div>
            </div>
            <span :class="'pill ' + (statusPills[r.Status] || 'pill-neutral')">{{ statusLabels[r.Status] || r.Status }}</span>
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
              <div style="font-weight:600;font-size:.85rem">{{ tpl.Name }}</div>
              <span class="pill pill-neutral" style="font-size:.7rem">{{ tpl.Language || 'fr' }}</span>
            </div>
            <div class="form-group" style="margin-bottom:8px">
              <textarea class="form-input" v-model="tpl._body" rows="3" style="resize:vertical;font-size:.82rem"></textarea>
            </div>
            <div style="display:flex;gap:8px;align-items:center">
              <button class="btn btn-sm btn-secondary" style="width:auto" @click="saveTemplate(tpl)" :disabled="savingTemplate === tpl.ID || tpl._body === tpl.Body">
                {{ savingTemplate === tpl.ID ? 'Sauvegarde...' : 'Sauvegarder' }}
              </button>
              <div style="font-size:.75rem;color:var(--gray-400);margin-top:6px;line-height:1.5">
                Mots remplaces automatiquement dans le message :<br>
                <strong>{prenom}</strong> = prenom du patient,
                <strong>{date}</strong> = date du RDV,
                <strong>{heure}</strong> = heure du RDV,
                <strong>{centre}</strong> = nom du centre de sante
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </template>
</template>
