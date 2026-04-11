<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { setup as setupApi } from '../api'

// Direct fetch for settings endpoints not in the api module.
async function apiGet(path) {
  try {
    const res = await fetch('/api/v1' + path, { credentials: 'same-origin' })
    const data = await res.json().catch(() => null)
    if (!res.ok) return { ok: false, error: (data && data.error) || 'Erreur' }
    return { ok: true, data }
  } catch (e) { return { ok: false, error: 'Connexion impossible' } }
}
async function apiPut(path, body) {
  try {
    const res = await fetch('/api/v1' + path, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, credentials: 'same-origin', body: JSON.stringify(body) })
    const data = await res.json().catch(() => null)
    if (!res.ok) return { ok: false, error: (data && data.error) || 'Erreur' }
    return { ok: true, data }
  } catch (e) { return { ok: false, error: 'Connexion impossible' } }
}
import { useToast } from '../composables/useToast'

const router = useRouter()
const toast = useToast()

// SMS config.
const smsLoading = ref(true)
const smsForm = ref({ enabled: false, provider: '', api_key: '', api_secret: '', sender_id: '' })
const smsSaving = ref(false)
const smsError = ref('')
const smsHasKey = ref(false)

const providers = [
  { value: 'africastalking', label: "Africa's Talking" },
  { value: 'mtn', label: 'MTN' },
  { value: 'orange', label: 'Orange' },
  { value: 'twilio', label: 'Twilio' },
  { value: 'infobip', label: 'Infobip' },
]

onMounted(async () => {
  const res = await apiGet('/settings/sms')
  smsLoading.value = false
  if (res.ok && res.data) {
    smsForm.value.enabled = res.data.enabled
    smsForm.value.provider = res.data.provider || ''
    smsForm.value.sender_id = res.data.sender_id || ''
    smsHasKey.value = res.data.has_key
  }
})

async function saveSMS() {
  smsError.value = ''
  if (smsForm.value.enabled && !smsForm.value.provider) { smsError.value = 'Selectionnez un fournisseur'; return }
  smsSaving.value = true
  const res = await apiPut('/settings/sms', smsForm.value)
  smsSaving.value = false
  if (!res.ok) { smsError.value = res.error; return }
  toast.success('Configuration SMS sauvegardee')
  smsHasKey.value = !!smsForm.value.api_key
}

const reports = [
  { label: 'Rapport mensuel', desc: 'Patients, RDV et indicateurs du mois', excel: '/api/v1/export/monthly/excel', pdf: '/api/v1/export/monthly/pdf' },
  { label: 'Tous les patients', desc: 'Liste complete avec statut et risque', excel: '/api/v1/export/patients/excel', pdf: '/api/v1/export/patients/pdf' },
  { label: 'Patients actifs', desc: 'Patients actuellement suivis', excel: '/api/v1/export/patients/excel?status=active', pdf: '/api/v1/export/patients/pdf?status=active' },
  { label: 'Patients a surveiller', desc: 'Risque eleve de RDV manque', excel: '/api/v1/export/patients/excel?status=a_surveiller', pdf: '/api/v1/export/patients/pdf?status=a_surveiller' },
  { label: 'Patients perdus de vue', desc: 'Sans visite depuis 90+ jours', excel: '/api/v1/export/patients/excel?status=perdu_de_vue', pdf: '/api/v1/export/patients/pdf?status=perdu_de_vue' },
  { label: 'Patients sortis', desc: 'Deces, transferts, abandons, guerisons', excel: '/api/v1/export/patients/excel?status=sorti', pdf: '/api/v1/export/patients/pdf?status=sorti' },
]
</script>

<template>
  <div class="grid-2" style="align-items:start">
    <!-- Left column -->
    <div>
      <!-- SMS Config -->
      <div class="card" style="margin-bottom:16px">
        <div class="card-head" style="padding:12px 16px"><h3>Configuration SMS</h3></div>
        <div class="card-body" style="padding:16px">
          <div v-if="smsLoading" class="ms-loading" style="padding:12px">Chargement...</div>
          <template v-else>
            <div class="form-group" style="margin-bottom:12px">
              <label>Activer les rappels SMS</label>
              <div style="display:flex;gap:8px;margin-top:6px">
                <div class="r-opt" :class="{ on: smsForm.enabled }" @click="smsForm.enabled = true">Oui</div>
                <div class="r-opt" :class="{ on: !smsForm.enabled }" @click="smsForm.enabled = false">Non</div>
              </div>
            </div>

            <template v-if="smsForm.enabled">
              <div class="form-group" style="margin-bottom:10px">
                <label>Fournisseur</label>
                <select class="form-input" v-model="smsForm.provider">
                  <option value="">Selectionner</option>
                  <option v-for="p in providers" :key="p.value" :value="p.value">{{ p.label }}</option>
                </select>
              </div>
              <div class="form-group" style="margin-bottom:10px">
                <label>Cle API {{ smsHasKey ? '(deja configuree)' : '' }}</label>
                <input class="form-input" type="password" v-model="smsForm.api_key" :placeholder="smsHasKey ? 'Laisser vide pour garder la cle actuelle' : 'Collez votre cle API'">
              </div>
              <div class="form-group" style="margin-bottom:10px">
                <label>Secret API</label>
                <input class="form-input" type="password" v-model="smsForm.api_secret" :placeholder="smsHasKey ? 'Laisser vide pour garder' : 'Collez votre secret'">
              </div>
              <div class="form-group" style="margin-bottom:10px">
                <label>Nom expediteur</label>
                <input class="form-input" v-model="smsForm.sender_id" placeholder="Ex: MaSante">
              </div>
            </template>

            <div v-if="smsError" style="padding:8px 12px;background:var(--danger-bg);color:var(--danger);border-radius:6px;font-size:.82rem;margin:10px 0">{{ smsError }}</div>
            <button class="btn btn-primary" style="width:auto;margin-top:4px" @click="saveSMS" :disabled="smsSaving">{{ smsSaving ? 'Sauvegarde...' : 'Sauvegarder' }}</button>
          </template>
        </div>
      </div>

      <!-- Administration -->
      <div class="card">
        <div class="card-head" style="padding:12px 16px"><h3>Administration</h3></div>
        <div style="padding:0">
          <div style="display:flex;align-items:center;gap:12px;padding:10px 16px;border-bottom:1px solid var(--gray-50);cursor:pointer" @click="router.push('/users')">
            <div style="width:34px;height:34px;border-radius:8px;background:var(--gray-50);display:flex;align-items:center;justify-content:center;flex-shrink:0">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--gray-400)" stroke-width="1.5"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/></svg>
            </div>
            <div style="flex:1"><div style="font-weight:600;font-size:.85rem">Gestion des utilisateurs</div><div style="font-size:.75rem;color:var(--gray-400)">Ajouter, modifier ou desactiver des comptes</div></div>
          </div>
          <div style="display:flex;align-items:center;gap:12px;padding:10px 16px;cursor:pointer" @click="router.push('/profile')">
            <div style="width:34px;height:34px;border-radius:8px;background:var(--gray-50);display:flex;align-items:center;justify-content:center;flex-shrink:0">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--gray-400)" stroke-width="1.5"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
            </div>
            <div style="flex:1"><div style="font-weight:600;font-size:.85rem">Mon profil</div><div style="font-size:.75rem;color:var(--gray-400)">Modifier vos informations</div></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Right column — Exports -->
    <div class="card">
      <div class="card-head" style="padding:12px 16px"><h3>Rapports et exports</h3></div>
      <div style="padding:0">
        <div v-for="r in reports" :key="r.label" style="display:flex;align-items:center;gap:12px;padding:10px 16px;border-bottom:1px solid var(--gray-50)">
          <div style="width:34px;height:34px;border-radius:8px;background:var(--gray-50);display:flex;align-items:center;justify-content:center;flex-shrink:0">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--gray-400)" stroke-width="1.5" stroke-linecap="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
          </div>
          <div style="flex:1;min-width:0"><div style="font-weight:600;font-size:.85rem">{{ r.label }}</div><div style="font-size:.75rem;color:var(--gray-400)">{{ r.desc }}</div></div>
          <div style="display:flex;gap:4px;flex-shrink:0">
            <a :href="r.excel" target="_blank" rel="noopener" class="btn btn-sm btn-secondary" style="width:auto;text-decoration:none;padding:5px 10px;font-size:.75rem">Excel</a>
            <a :href="r.pdf" target="_blank" rel="noopener" class="btn btn-sm btn-secondary" style="width:auto;text-decoration:none;padding:5px 10px;font-size:.75rem">PDF</a>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
