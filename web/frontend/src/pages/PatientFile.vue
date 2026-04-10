<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { patients } from '../api'
import { useToast } from '../composables/useToast'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const patient = ref(null)
const loading = ref(true)
const error = ref('')
const showExitModal = ref(false)
const exitReason = ref('')
const exiting = ref(false)

const exitReasons = [
  'Transfert vers un autre centre',
  'Deces',
  'Perdu de vue confirme',
  'Arret volontaire du traitement',
  'Guerison / fin de suivi',
  'Autre',
]

onMounted(async () => {
  const res = await patients.get(route.params.id)
  loading.value = false
  if (!res.ok) { error.value = res.error; return }
  patient.value = res.data
})

function riskLabel(s) { return s <= 3 ? 'Faible' : s <= 6 ? 'Moyen' : 'Eleve' }
function riskClass(s) { return s <= 3 ? 'low' : s <= 6 ? 'med' : 'high' }
function statusLabel(s) { return { active: 'Actif', a_surveiller: 'A surveiller', perdu_de_vue: 'Perdu de vue', sorti: 'Sorti' }[s] || s }
function statusClass(s) { return { active: 'pill-success', a_surveiller: 'pill-warning', perdu_de_vue: 'pill-danger', sorti: 'pill-neutral' }[s] || 'pill-neutral' }
function initials(p) { return ((p.LastName || '')[0] + (p.FirstName || '')[0]).toUpperCase() }
function fmtDate(d) { return d ? new Date(d).toLocaleDateString('fr-FR') : '—' }

async function confirmExit() {
  if (!exitReason.value) { error.value = 'Veuillez selectionner un motif de sortie'; return }
  exiting.value = true
  error.value = ''
  const res = await patients.exit(route.params.id, { reason: exitReason.value })
  exiting.value = false
  if (!res.ok) { error.value = res.error; return }
  toast.success('Patient sorti du programme')
  showExitModal.value = false
  patient.value.Status = 'sorti'
}
</script>

<template>
  <div v-if="loading" class="ms-loading">Chargement...</div>
  <div v-else-if="error && !patient" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem">{{ error }}</div>
  <template v-else-if="patient">
    <div class="grid-2" style="align-items:start">
      <!-- Profile card -->
      <div class="card">
        <div class="card-head"><h3>Dossier patient</h3><span :class="'pill ' + statusClass(patient.Status)">{{ statusLabel(patient.Status) }}</span></div>
        <div class="card-body">
          <div style="display:flex;align-items:center;gap:16px;margin-bottom:20px">
            <div class="avatar a1" style="width:56px;height:56px;font-size:1.2rem">{{ initials(patient) }}</div>
            <div>
              <div style="font-size:1.1rem;font-weight:700">{{ patient.LastName }} {{ patient.FirstName }}</div>
              <div style="font-size:.82rem;color:var(--gray-400)">Code: {{ patient.Code }}</div>
            </div>
          </div>
          <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;font-size:.85rem">
            <div><span style="color:var(--gray-400)">Sexe:</span> {{ patient.Sex === 'M' ? 'Masculin' : 'Feminin' }}</div>
            <div><span style="color:var(--gray-400)">Quartier:</span> {{ patient.District || '—' }}</div>
            <div><span style="color:var(--gray-400)">Telephone:</span> {{ patient.Phone || '—' }}</div>
            <div><span style="color:var(--gray-400)">Langue:</span> {{ patient.Language || '—' }}</div>
            <div><span style="color:var(--gray-400)">Date naissance:</span> {{ fmtDate(patient.DateOfBirth) }}</div>
            <div><span style="color:var(--gray-400)">Inscription:</span> {{ fmtDate(patient.CreatedAt) }}</div>
          </div>
          <div style="margin-top:16px">
            <span style="color:var(--gray-400);font-size:.85rem">Score de risque:</span>
            <span :class="'risk ' + riskClass(patient.RiskScore)" style="margin-left:8px"><span class="risk-dot"></span> {{ patient.RiskScore || 0 }} — {{ riskLabel(patient.RiskScore || 0) }}</span>
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div>
        <div class="card" style="margin-bottom:16px">
          <div class="card-head"><h3>Actions</h3></div>
          <div class="card-body" style="display:flex;flex-direction:column;gap:10px">
            <button class="btn btn-primary" @click="router.push('/new-apt?patient=' + patient.ID)">Programmer un RDV</button>
            <button class="btn btn-secondary" @click="router.push('/patients')">Retour a la liste</button>
            <button v-if="patient.Status !== 'sorti'" class="btn" style="background:var(--danger-bg);color:var(--danger);border:1px solid var(--danger)" @click="showExitModal = true">Sortir du programme</button>
          </div>
        </div>

        <!-- Contact info -->
        <div class="card">
          <div class="card-head"><h3>Contact de confiance</h3></div>
          <div class="card-body" style="font-size:.85rem">
            <div v-if="patient.ContactName">
              <div><span style="color:var(--gray-400)">Nom:</span> {{ patient.ContactName }}</div>
              <div style="margin-top:4px"><span style="color:var(--gray-400)">Tel:</span> {{ patient.ContactPhone || '—' }}</div>
              <div style="margin-top:4px"><span style="color:var(--gray-400)">Lien:</span> {{ patient.ContactRelation || '—' }}</div>
            </div>
            <div v-else class="ms-empty" style="padding:8px 0">Aucun contact renseigne</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Exit Modal -->
    <template v-if="showExitModal">
      <div style="position:fixed;inset:0;background:rgba(0,0,0,.4);z-index:9990;display:flex;align-items:center;justify-content:center" @click.self="showExitModal = false">
        <div class="card" style="width:100%;max-width:460px;margin:16px">
          <div class="card-head"><h3>Sortie du programme</h3></div>
          <div class="card-body">
            <p style="font-size:.85rem;color:var(--gray-500);margin-bottom:14px">Selectionnez le motif de sortie pour {{ patient.LastName }} {{ patient.FirstName }}:</p>
            <div class="form-group">
              <select class="form-input" v-model="exitReason">
                <option value="">Choisir un motif</option>
                <option v-for="r in exitReasons" :key="r" :value="r">{{ r }}</option>
              </select>
            </div>
            <div v-if="error" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin-bottom:12px">{{ error }}</div>
            <div style="display:flex;gap:10px;margin-top:16px">
              <button class="btn btn-secondary" style="flex:1" @click="showExitModal = false">Annuler</button>
              <button class="btn" style="flex:1;background:var(--danger);color:#fff" @click="confirmExit" :disabled="exiting">{{ exiting ? 'Sortie...' : 'Confirmer la sortie' }}</button>
            </div>
          </div>
        </div>
      </div>
    </template>
  </template>
</template>
