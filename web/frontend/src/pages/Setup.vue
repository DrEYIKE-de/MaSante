<script setup>
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { setup as setupApi } from '../api'
import { store } from '../store'
import { useToast } from '../composables/useToast'

const router = useRouter()
const toast = useToast()

const step = ref(store.setupStep + 1 || 1)
const total = 5
const error = ref('')
const saving = ref(false)

// Step 1 — Center.
const centerName = ref('')
const centerType = ref('centre_sante')
const country = ref('Cameroun')
const city = ref('')
const district = ref('')
const lat = ref('')
const lng = ref('')
const gpsStatus = ref('')

// Step 2 — Admin.
const adminName = ref('')
const adminEmail = ref('')
const adminUsername = ref('')
const adminTitle = ref('Medecin referent')
const adminPwd = ref('')
const adminPwd2 = ref('')

// Step 3 — Schedule.
const days = ref([true, true, true, true, true, false, false])
const startTime = ref('08:00')
const endTime = ref('16:00')
const slotDuration = ref('30')
const maxPatients = ref('40')

// Step 4 — SMS.
const smsEnabled = ref(false)
const smsProvider = ref("Africa's Talking")
const smsKey = ref('')
const smsSecret = ref('')
const smsSender = ref('')

const progress = computed(() => (step.value / total) * 100)

const canProceed = computed(() => {
  if (step.value === 1) return centerName.value.trim() && city.value.trim()
  if (step.value === 2) return adminName.value.trim() && adminEmail.value.trim() && adminUsername.value.trim() && adminPwd.value.length >= 8 && adminPwd.value === adminPwd2.value
  return true
})

function detectGPS() {
  if (!navigator.geolocation) { gpsStatus.value = 'Non supporte'; return }
  gpsStatus.value = 'Detection...'
  navigator.geolocation.getCurrentPosition(
    (pos) => { lat.value = pos.coords.latitude.toFixed(6); lng.value = pos.coords.longitude.toFixed(6); gpsStatus.value = 'Position detectee' },
    (err) => { gpsStatus.value = err.code === 1 ? 'Acces refuse' : 'Indisponible' },
    { timeout: 10000 }
  )
}

function validatePwd(pwd) {
  if (pwd.length < 8) return 'Minimum 8 caracteres'
  if (!/\d/.test(pwd)) return 'Au moins 1 chiffre requis'
  return null
}

async function next() {
  error.value = ''
  saving.value = true
  let res

  try {
    if (step.value === 1) {
      const typeMap = { 'Hopital public': 'hopital_public', 'Centre de sante': 'centre_sante', 'Clinique privee': 'clinique_privee' }
      const data = { name: centerName.value, type: typeMap[centerType.value] || 'centre_sante', country: country.value, city: city.value, district: district.value }
      if (lat.value) data.lat = parseFloat(lat.value)
      if (lng.value) data.lng = parseFloat(lng.value)
      res = await setupApi.center(data)
    } else if (step.value === 2) {
      const pwdErr = validatePwd(adminPwd.value)
      if (pwdErr) { error.value = pwdErr; saving.value = false; return }
      if (adminPwd.value !== adminPwd2.value) { error.value = 'Les mots de passe ne correspondent pas'; saving.value = false; return }
      res = await setupApi.admin({ full_name: adminName.value, email: adminEmail.value, username: adminUsername.value, password: adminPwd.value, title: adminTitle.value })
    } else if (step.value === 3) {
      const activeDays = days.value.map((d, i) => d ? i + 1 : null).filter(Boolean).join(',')
      res = await setupApi.schedule({ consultation_days: activeDays, start_time: startTime.value, end_time: endTime.value, slot_duration: parseInt(slotDuration.value), max_patients_day: parseInt(maxPatients.value) })
    } else if (step.value === 4) {
      const provMap = { "Africa's Talking": 'africastalking', 'MTN': 'mtn', 'Orange': 'orange', 'Twilio': 'twilio', 'Infobip': 'infobip' }
      res = await setupApi.sms({ enabled: smsEnabled.value, provider: smsEnabled.value ? (provMap[smsProvider.value] || '') : '', api_key: smsKey.value, api_secret: smsSecret.value, sender_id: smsSender.value })
    } else if (step.value === 5) {
      res = await setupApi.complete()
      if (res.ok) {
        toast.success('Configuration terminee !')
        store.setupDone = true
        router.push('/login')
        return
      }
    }

    if (res && !res.ok) { error.value = res.error; saving.value = false; return }
    step.value++
  } finally {
    saving.value = false
  }
}

function prev() { if (step.value > 1) step.value-- }

const countries = ['Cameroun', "Cote d'Ivoire", 'RD Congo', 'Senegal', 'Tchad', 'Gabon', 'Congo', 'Centrafrique', 'Burkina Faso', 'Mali', 'Niger', 'Guinee', 'Togo', 'Benin', 'Rwanda', 'Burundi', 'Kenya', 'Madagascar']
const titles = ['Medecin referent', 'Chef de service', 'Directeur', 'Coordinateur programme', "Infirmier(e) chef"]
const providers = ["Africa's Talking", 'MTN', 'Orange', 'Twilio']
const dayLabels = ['L', 'M', 'M', 'J', 'V', 'S', 'D']
</script>

<template>
  <div style="position:fixed;inset:0;display:flex;flex-direction:column;background:var(--white);z-index:100">
    <!-- Header -->
    <div class="setup-header">
      <div class="sh-logo"><svg width="18" height="18" viewBox="0 0 80 80" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M40 8C22.3 8 8 22.3 8 40c0 17.7 14.3 32 32 32s32-14.3 32-32" stroke="white" stroke-width="8" stroke-linecap="round"/><circle cx="40" cy="40" r="10" fill="white"/></svg></div>
      <h2>Configuration de MaSante</h2>
      <span class="sh-step">Etape {{ step }} sur {{ total }}</span>
    </div>
    <div class="setup-progress"><div class="setup-progress-fill" :style="{ width: progress + '%' }"></div></div>

    <!-- Body -->
    <div class="setup-body">
      <div class="setup-content">

        <!-- Step 1 -->
        <template v-if="step === 1">
          <h3>Votre etablissement</h3>
          <p class="setup-desc">Ces informations identifient votre centre de sante.</p>
          <div class="setup-grid">
            <div class="form-group setup-full">
              <label>Nom de l'etablissement *</label>
              <input class="form-input" v-model="centerName" placeholder="Ex: Hopital Laquintinie">
            </div>
            <div class="form-group">
              <label>Type</label>
              <div class="type-options">
                <div v-for="t in ['Hopital public','Centre de sante','Clinique privee']" :key="t" class="type-opt" :class="{ on: centerType === t }" @click="centerType = t">{{ t }}</div>
              </div>
            </div>
            <div class="form-group">
              <label>Pays *</label>
              <select class="form-input" v-model="country">
                <option v-for="c in countries" :key="c">{{ c }}</option>
              </select>
            </div>
            <div class="form-group">
              <label>Ville *</label>
              <input class="form-input" v-model="city" placeholder="Ex: Douala">
            </div>
            <div class="form-group">
              <label>Quartier</label>
              <input class="form-input" v-model="district" placeholder="Ex: Akwa">
            </div>
            <div class="form-group setup-full">
              <label>GPS (optionnel)</label>
              <div style="display:flex;gap:10px;align-items:center">
                <input class="form-input" v-model="lat" placeholder="Latitude" style="flex:1">
                <input class="form-input" v-model="lng" placeholder="Longitude" style="flex:1">
                <button class="btn btn-secondary" type="button" style="width:auto;flex-shrink:0" @click="detectGPS">Detecter</button>
              </div>
              <span v-if="gpsStatus" style="font-size:.78rem;color:var(--gray-400)">{{ gpsStatus }}</span>
            </div>
          </div>
        </template>

        <!-- Step 2 -->
        <template v-if="step === 2">
          <h3>Compte administrateur</h3>
          <p class="setup-desc">Ce sera le premier utilisateur avec tous les droits.</p>
          <div class="setup-grid">
            <div class="form-group"><label>Nom complet *</label><input class="form-input" v-model="adminName" placeholder="Ex: Dr. Adele Mbarga"></div>
            <div class="form-group"><label>Email *</label><input class="form-input" type="email" v-model="adminEmail" placeholder="Ex: adele@hopital.cm"></div>
            <div class="form-group"><label>Identifiant *</label><input class="form-input" v-model="adminUsername" placeholder="Choisir un identifiant"></div>
            <div class="form-group">
              <label>Fonction</label>
              <select class="form-input" v-model="adminTitle">
                <option v-for="t in titles" :key="t">{{ t }}</option>
              </select>
            </div>
            <div class="form-group setup-full"><label>Mot de passe * (min 8 + 1 chiffre)</label><input class="form-input" type="password" v-model="adminPwd" placeholder="Minimum 8 caracteres"></div>
            <div class="form-group setup-full"><label>Confirmer *</label><input class="form-input" type="password" v-model="adminPwd2" placeholder="Retapez le mot de passe"></div>
            <p v-if="adminPwd && adminPwd2 && adminPwd !== adminPwd2" style="color:var(--danger);font-size:.82rem;grid-column:1/-1">Les mots de passe ne correspondent pas</p>
          </div>
        </template>

        <!-- Step 3 -->
        <template v-if="step === 3">
          <h3>Horaires de consultation</h3>
          <p class="setup-desc">Definissez les creneaux. Modifiable a tout moment.</p>
          <div class="form-group">
            <label>Jours</label>
            <div class="day-checks">
              <div v-for="(d, i) in dayLabels" :key="i" class="day-check" :class="{ on: days[i] }" @click="days[i] = !days[i]">{{ d }}</div>
            </div>
          </div>
          <div class="setup-grid" style="margin-top:14px">
            <div class="form-group"><label>Debut</label><input class="form-input" type="time" v-model="startTime"></div>
            <div class="form-group"><label>Fin</label><input class="form-input" type="time" v-model="endTime"></div>
            <div class="form-group">
              <label>Duree creneau</label>
              <select class="form-input" v-model="slotDuration">
                <option v-for="d in ['15','30','45','60']" :key="d" :value="d">{{ d }} min</option>
              </select>
            </div>
            <div class="form-group"><label>Max patients/jour</label><input class="form-input" type="number" v-model="maxPatients" min="1"></div>
          </div>
        </template>

        <!-- Step 4 -->
        <template v-if="step === 4">
          <h3>Rappels SMS (optionnel)</h3>
          <p class="setup-desc">Les rappels reduisent les RDV manques de 25 a 50%.</p>
          <div class="type-options" style="margin-bottom:16px">
            <div class="type-opt" :class="{ on: smsEnabled }" @click="smsEnabled = true">Oui, activer</div>
            <div class="type-opt" :class="{ on: !smsEnabled }" @click="smsEnabled = false">Plus tard</div>
          </div>
          <template v-if="smsEnabled">
            <div class="form-group">
              <label>Fournisseur</label>
              <select class="form-input" v-model="smsProvider">
                <option v-for="p in providers" :key="p">{{ p }}</option>
              </select>
            </div>
            <div class="setup-grid">
              <div class="form-group"><label>Cle API</label><input class="form-input" type="password" v-model="smsKey" placeholder="Votre cle API"></div>
              <div class="form-group"><label>Secret</label><input class="form-input" type="password" v-model="smsSecret" placeholder="Votre secret"></div>
              <div class="form-group setup-full"><label>Expediteur</label><input class="form-input" v-model="smsSender" placeholder="Ex: MaSante"></div>
            </div>
          </template>
        </template>

        <!-- Step 5 -->
        <template v-if="step === 5">
          <div style="text-align:center;padding:40px 0">
            <div style="width:72px;height:72px;border-radius:50%;background:var(--success-bg);color:var(--success);display:flex;align-items:center;justify-content:center;margin:0 auto 20px;"><svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg></div>
            <h3 style="margin-bottom:10px">Votre plateforme est prete</h3>
            <p style="color:var(--gray-400)">Cliquez sur "Lancer MaSante" pour commencer.</p>
            <div class="setup-recap" style="text-align:left;margin-top:24px">
              <div class="sr-row"><span class="sr-label">Etablissement</span><span class="sr-val">{{ centerName || '—' }}</span></div>
              <div class="sr-row"><span class="sr-label">Localisation</span><span class="sr-val">{{ [city, district, country].filter(Boolean).join(', ') || '—' }}</span></div>
              <div class="sr-row"><span class="sr-label">Administrateur</span><span class="sr-val">{{ adminName || '—' }}</span></div>
              <div class="sr-row"><span class="sr-label">Identifiant</span><span class="sr-val">{{ adminUsername || '—' }}</span></div>
              <div class="sr-row"><span class="sr-label">SMS</span><span class="sr-val">{{ smsEnabled ? 'Active' : 'Non configure' }}</span></div>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- Error -->
    <div v-if="error" style="padding:10px 40px">
      <div style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem">{{ error }}</div>
    </div>

    <!-- Footer -->
    <div class="setup-footer">
      <button class="btn btn-secondary" style="width:auto" :style="{ visibility: step === 1 ? 'hidden' : 'visible' }" @click="prev">&#8592; Precedent</button>
      <button class="btn btn-primary" style="width:auto" :disabled="!canProceed || saving" @click="next" :style="{ opacity: canProceed && !saving ? 1 : 0.5 }">
        {{ step === total ? 'Lancer MaSante &#10003;' : 'Suivant &#8594;' }}
      </button>
    </div>
  </div>
</template>
