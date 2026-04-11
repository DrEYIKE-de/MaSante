<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { reminders, setup as setupApi } from '../api'
import { useToast } from '../composables/useToast'

const router = useRouter()
const toast = useToast()
const stats = ref({ DeliveryRate: 0, PendingCount: 0, FailedCount: 0 })
const queue = ref([])
const templates = ref([])
const loading = ref(true)
const smsConfigured = ref(false)

const statusLabels = { planifie: 'Planifie', envoye: 'Envoye', recu: 'Recu', echec: 'Echec' }
const statusPills = { planifie: 'pill-warning', envoye: 'pill-success', recu: 'pill-info', echec: 'pill-danger' }
const typeLabels = { j7: 'Rappel J-7', j2: 'Rappel J-2', j0: 'Rappel jour J', retard: 'Rappel retard' }
const channelLabels = { sms: 'SMS', whatsapp: 'WhatsApp', voice: 'Appel vocal' }

const savingTemplate = ref(null)
const showHelp = ref(false)

onMounted(async () => {
  // Check if SMS is configured.
  const setupRes = await setupApi.status()
  if (setupRes.ok) {
    // If setup is done, check if we have a provider configured.
    // We can infer from stats — if delivery rate is null and no pending, likely not configured.
  }

  const [sR, qR, tR] = await Promise.all([reminders.stats(), reminders.list(), reminders.templates()])
  loading.value = false

  if (sR.ok) {
    stats.value = sR.data || stats.value
    // If everything is 0 and no queue, SMS might not be configured.
    const s = sR.data
    smsConfigured.value = !!(s && (s.DeliveryRate > 0 || s.PendingCount > 0 || s.FailedCount > 0))
  }
  if (qR.ok) {
    queue.value = qR.data || []
    if (queue.value.length > 0) smsConfigured.value = true
  }
  if (tR.ok) {
    templates.value = (tR.data || []).map(t => ({ ...t, _body: t.Body || '' }))
    if (templates.value.length > 0) smsConfigured.value = true // templates exist = setup was done
  }
})

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

    <!-- Header with help toggle -->
    <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px">
      <div></div>
      <button @click="showHelp = !showHelp" style="background:none;border:1px solid var(--gray-200);border-radius:50%;width:28px;height:28px;display:flex;align-items:center;justify-content:center;cursor:pointer;color:var(--gray-400);font-size:.85rem;font-weight:600;transition:.15s" :style="showHelp ? 'background:var(--primary);color:#fff;border-color:var(--primary)' : ''" title="Comment fonctionnent les rappels ?">?</button>
    </div>

    <!-- Help panel (toggle) -->
    <div v-if="showHelp" style="padding:14px 18px;background:var(--gray-25);border-radius:var(--radius);font-size:.82rem;color:var(--gray-500);line-height:1.6;margin-bottom:14px">
      <p>Les rappels SMS sont envoyes <strong>automatiquement</strong> aux patients avant leurs rendez-vous :</p>
      <p style="margin-top:6px">
        <strong>J-7</strong> = 7 jours avant &middot;
        <strong>J-2</strong> = 2 jours avant &middot;
        <strong>Jour J</strong> = le matin du RDV &middot;
        <strong>Retard</strong> = apres un RDV manque
      </p>
      <p style="margin-top:6px">Le systeme verifie toutes les 5 minutes. Configurez votre fournisseur SMS dans <strong style="cursor:pointer;text-decoration:underline" @click="$router.push('/settings')">Parametres</strong>.</p>
    </div>

    <!-- SMS not configured warning -->
    <div v-if="!smsConfigured && !queue.length" style="padding:10px 14px;background:var(--warning-bg);color:var(--warning);border-radius:var(--radius);font-size:.82rem;margin-bottom:12px">
      Aucun fournisseur SMS configure. <strong style="cursor:pointer;text-decoration:underline" @click="$router.push('/settings')">Configurer</strong>
    </div>

    <!-- Stats -->
    <div class="stats-row" style="margin-bottom:16px">
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
        <div class="stat-label">En attente d'envoi</div>
      </div>
      <div class="stat">
        <div class="stat-icon red">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
        </div>
        <div class="stat-val">{{ stats.FailedCount }}</div>
        <div class="stat-label">Echecs d'envoi</div>
      </div>
    </div>

    <div class="grid-2" style="align-items:start">
      <!-- Queue -->
      <div class="card">
        <div class="card-head" style="padding:12px 16px">
          <h3>File d'envoi</h3>
          <button v-if="queue.length" class="btn btn-sm btn-primary" style="width:auto" @click="sendAll">Envoyer maintenant</button>
        </div>
        <div v-if="!queue.length" class="ms-empty" style="padding:24px">
          <div style="margin-bottom:8px">Aucun rappel en attente</div>
          <div style="font-size:.78rem;color:var(--gray-300)">Les rappels sont generes automatiquement quand des RDV approchent</div>
        </div>
        <div v-else style="padding:0">
          <div v-for="r in queue" :key="r.ID" style="display:flex;align-items:center;gap:12px;padding:10px 16px;border-bottom:1px solid var(--gray-50)">
            <div style="width:34px;height:34px;border-radius:8px;display:flex;align-items:center;justify-content:center;flex-shrink:0;background:var(--info-bg);color:var(--info)">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
            </div>
            <div style="flex:1;min-width:0">
              <div style="font-weight:600;font-size:.85rem">{{ r.PatientName || 'Patient' }}</div>
              <div style="font-size:.75rem;color:var(--gray-400)">{{ typeLabels[r.Type] || r.Type }} &middot; {{ channelLabels[r.Channel] || r.Channel }}</div>
            </div>
            <span :class="'pill ' + (statusPills[r.Status] || 'pill-neutral')">{{ statusLabels[r.Status] || r.Status }}</span>
          </div>
        </div>
      </div>

      <!-- Templates -->
      <div class="card">
        <div class="card-head" style="padding:12px 16px"><h3>Modeles de message</h3></div>
        <div v-if="!templates.length" class="ms-empty">Aucun modele</div>
        <div v-else style="padding:0">
          <div v-for="tpl in templates" :key="tpl.ID" style="padding:14px 16px;border-bottom:1px solid var(--gray-50)">
            <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:8px">
              <div style="font-weight:600;font-size:.85rem">{{ typeLabels[tpl.Name] || tpl.Name }}</div>
              <span class="pill pill-neutral" style="font-size:.68rem">{{ tpl.Language || 'fr' }}</span>
            </div>
            <textarea class="form-input" v-model="tpl._body" rows="3" style="resize:vertical;font-size:.82rem;margin-bottom:8px"></textarea>
            <div style="display:flex;align-items:center;gap:8px">
              <button class="btn btn-sm btn-secondary" style="width:auto" @click="saveTemplate(tpl)" :disabled="savingTemplate === tpl.ID || tpl._body === tpl.Body">
                {{ savingTemplate === tpl.ID ? 'Sauvegarde...' : 'Sauvegarder' }}
              </button>
              <span style="font-size:.72rem;color:var(--gray-300);flex:1">
                {prenom} {date} {heure} {centre} seront remplaces automatiquement
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </template>
</template>
