<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { patients, appointments } from '../api'
import { useToast } from '../composables/useToast'

const router = useRouter()
const toast = useToast()

const searchQuery = ref('')
const searchResults = ref([])
const searching = ref(false)
const selectedPatient = ref(null)
const aptType = ref('')
const aptDate = ref('')
const slots = ref([])
const selectedSlot = ref('')
const loadingSlots = ref(false)
const notes = ref('')
const saving = ref(false)
const error = ref('')

function localDate(d) { return d.getFullYear()+'-'+String(d.getMonth()+1).padStart(2,'0')+'-'+String(d.getDate()).padStart(2,'0') }

const types = [
  { value: 'consultation', label: 'Consultation' },
  { value: 'retrait_medicaments', label: 'Retrait medicaments' },
  { value: 'bilan_sanguin', label: 'Bilan sanguin' },
  { value: 'club_adherence', label: "Club d'adherence" },
]

let searchTimer = null
function onSearch() {
  clearTimeout(searchTimer)
  if (!searchQuery.value || searchQuery.value.length < 2) { searchResults.value = []; return }
  searchTimer = setTimeout(async () => {
    searching.value = true
    const res = await patients.search(searchQuery.value)
    searching.value = false
    if (res.ok) searchResults.value = res.data || []
    else { error.value = res.error; searchResults.value = [] }
  }, 300)
}

function pickPatient(p) {
  selectedPatient.value = p
  searchQuery.value = ''
  searchResults.value = []
}

function clearPatient() {
  selectedPatient.value = null
  aptType.value = ''
  aptDate.value = ''
  slots.value = []
  selectedSlot.value = ''
}

async function loadSlots() {
  if (!aptDate.value) return
  loadingSlots.value = true
  selectedSlot.value = ''
  error.value = ''
  const res = await appointments.slots(aptDate.value)
  loadingSlots.value = false
  if (!res.ok) { error.value = res.error; slots.value = []; return }
  slots.value = res.data || []
}

async function submit() {
  error.value = ''
  if (!selectedPatient.value) { error.value = 'Veuillez selectionner un patient'; return }
  if (!aptType.value) { error.value = 'Veuillez choisir le type de RDV'; return }
  if (!aptDate.value) { error.value = 'Veuillez choisir une date'; return }
  if (!selectedSlot.value) { error.value = 'Veuillez selectionner un creneau'; return }

  saving.value = true
  const res = await appointments.create({
    patient_id: selectedPatient.value.ID,
    type: aptType.value,
    date: aptDate.value,
    time: selectedSlot.value,
    notes: notes.value,
  })
  saving.value = false
  if (!res.ok) { error.value = res.error; return }
  toast.success('RDV programme avec succes')
  router.push('/calendar')
}

function initials(p) { return ((p.LastName || '')[0] + (p.FirstName || '')[0]).toUpperCase() }
</script>

<template>
  <div class="apt-grid">
    <div>
      <!-- Step 1: Patient search -->
      <div class="card" style="margin-bottom:16px">
        <div class="card-head"><h3>1. Patient</h3></div>
        <div class="card-body">
          <template v-if="!selectedPatient">
            <div class="form-group">
              <label>Rechercher un patient</label>
              <input class="form-input" v-model="searchQuery" @input="onSearch" placeholder="Nom, prenom ou code...">
            </div>
            <div v-if="searching" style="font-size:.82rem;color:var(--gray-400);padding:8px 0">Recherche...</div>
            <div v-if="searchResults.length" style="border:1px solid var(--gray-100);border-radius:var(--radius);max-height:200px;overflow-y:auto">
              <div v-for="p in searchResults" :key="p.ID" class="list-item" style="cursor:pointer" @click="pickPatient(p)">
                <div class="avatar a1">{{ initials(p) }}</div>
                <div class="item-info">
                  <div class="item-name">{{ p.LastName }} {{ p.FirstName }}</div>
                  <div class="item-sub">{{ p.Code }}</div>
                </div>
              </div>
            </div>
            <div v-if="searchQuery.length >= 2 && !searching && !searchResults.length" class="ms-empty" style="padding:12px 0;font-size:.82rem">Aucun patient trouve</div>
          </template>
          <template v-else>
            <div class="list-item" style="background:var(--primary-bg);border-radius:var(--radius);padding:10px 12px">
              <div class="avatar a1">{{ initials(selectedPatient) }}</div>
              <div class="item-info">
                <div class="item-name">{{ selectedPatient.LastName }} {{ selectedPatient.FirstName }}</div>
                <div class="item-sub">{{ selectedPatient.Code }}</div>
              </div>
              <button class="btn btn-sm btn-secondary" style="width:auto;flex-shrink:0" @click="clearPatient">Changer</button>
            </div>
          </template>
        </div>
      </div>

      <!-- Step 2: Type -->
      <div class="card">
        <div class="card-head"><h3>2. Type de RDV</h3></div>
        <div class="card-body">
          <div class="remind-opts">
            <div v-for="t in types" :key="t.value" class="r-opt" :class="{ on: aptType === t.value }" @click="aptType = t.value">{{ t.label }}</div>
          </div>
        </div>
      </div>
    </div>

    <div>
      <!-- Step 3: Date & slot -->
      <div class="card" style="margin-bottom:16px">
        <div class="card-head"><h3>3. Date et creneau</h3></div>
        <div class="card-body">
          <div class="form-group">
            <label>Date du RDV</label>
            <input class="form-input" type="date" v-model="aptDate" @change="loadSlots">
          </div>
          <div v-if="loadingSlots" style="font-size:.82rem;color:var(--gray-400);padding:8px 0">Chargement des creneaux...</div>
          <div v-if="slots.length" style="margin-top:12px">
            <label style="font-size:.82rem;font-weight:600;margin-bottom:8px;display:block">Creneaux disponibles</label>
            <div class="slots-grid">
              <button
                v-for="s in slots"
                :key="s.Time"
                class="slot"
                :class="{ picked: selectedSlot === s.Time, off: !s.Available }"
                :disabled="!s.Available"
                @click="s.Available && (selectedSlot = s.Time)"
              >{{ s.Time }}</button>
            </div>
          </div>
          <div v-if="aptDate && !loadingSlots && !slots.length" class="ms-empty" style="padding:12px 0;font-size:.82rem">Aucun creneau disponible</div>
        </div>
      </div>

      <!-- Notes -->
      <div class="card" style="margin-bottom:16px">
        <div class="card-head"><h3>4. Notes</h3></div>
        <div class="card-body">
          <div class="form-group">
            <label>Observations ou instructions particulieres</label>
            <textarea class="form-input" v-model="notes" rows="3" placeholder="Ex: Patient a jeun pour bilan sanguin, apporter le carnet de sante..."></textarea>
          </div>
        </div>
      </div>

      <!-- Submit -->
      <div v-if="error" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin-bottom:12px">{{ error }}</div>
      <button class="btn btn-primary" @click="submit" :disabled="saving">{{ saving ? 'Programmation...' : 'Programmer le RDV' }}</button>
    </div>
  </div>
</template>
