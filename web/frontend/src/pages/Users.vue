<script setup>
import { ref, onMounted } from 'vue'
import { users } from '../api'
import { useToast } from '../composables/useToast'

const toast = useToast()
const list = ref([])
const loading = ref(true)
const error = ref('')
const showForm = ref(false)
const saving = ref(false)
const form = ref({ full_name: '', email: '', username: '', password: '', role: '' })
const formError = ref('')

const roles = [
  { value: 'admin', label: 'Administrateur' },
  { value: 'medecin', label: 'Medecin' },
  { value: 'infirmier', label: 'Infirmier(e)' },
  { value: 'asc', label: 'Agent communautaire (ASC)' },
]

const roleColors = { admin: 'pill-danger', medecin: 'pill-info', infirmier: 'pill-warning', asc: 'pill-success' }
const statusLabels = { active: 'Actif', conge: 'En conge', desactive: 'Desactive' }
const statusColors = { active: 'pill-success', conge: 'pill-neutral', desactive: 'pill-danger' }

async function load() {
  loading.value = true
  error.value = ''
  const res = await users.list()
  loading.value = false
  if (!res.ok) { error.value = res.error; return }
  list.value = res.data || []
}

onMounted(load)

function openForm() {
  form.value = { full_name: '', email: '', username: '', password: '', role: '' }
  formError.value = ''
  showForm.value = true
}

function roleLabel(r) { return (roles.find(x => x.value === r) || {}).label || r }

function initials(name) {
  return (name || '??').split(' ').map(w => w[0]).join('').substring(0, 2).toUpperCase()
}

async function createUser() {
  formError.value = ''
  if (!form.value.full_name) { formError.value = 'Le nom est requis'; return }
  if (!form.value.username) { formError.value = "L'identifiant est requis"; return }
  if (!form.value.password || form.value.password.length < 8) { formError.value = 'Mot de passe: minimum 8 caracteres'; return }
  if (!/\d/.test(form.value.password)) { formError.value = 'Mot de passe: au moins 1 chiffre'; return }
  if (!form.value.role) { formError.value = 'Selectionnez un role'; return }

  saving.value = true
  const res = await users.create(form.value)
  saving.value = false
  if (!res.ok) { formError.value = res.error; return }
  toast.success('Utilisateur cree')
  showForm.value = false
  load()
}

async function disableUser(u) {
  if (!confirm('Desactiver ' + u.full_name + ' ?')) return
  const res = await users.disable(u.id)
  if (!res.ok) { toast.error(res.error); return }
  toast.success('Utilisateur desactive')
  load()
}
</script>

<template>
  <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:16px">
    <h3 style="font-size:1.1rem">Equipe soignante</h3>
    <button class="btn btn-primary" style="width:auto" @click="openForm">+ Ajouter</button>
  </div>

  <div v-if="error" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin-bottom:12px">{{ error }}</div>

  <!-- Add user form -->
  <div v-if="showForm" class="card" style="margin-bottom:16px">
    <div class="card-head"><h3>Nouvel utilisateur</h3></div>
    <div class="card-body">
      <div style="display:grid;grid-template-columns:1fr 1fr;gap:14px">
        <div class="form-group"><label>Nom complet *</label><input class="form-input" v-model="form.full_name" placeholder="Ex: Dr. Kamga Jean"></div>
        <div class="form-group"><label>Email</label><input class="form-input" type="email" v-model="form.email" placeholder="Ex: kamga@hopital.cm"></div>
        <div class="form-group"><label>Identifiant *</label><input class="form-input" v-model="form.username" placeholder="Ex: kamga.j"></div>
        <div class="form-group"><label>Role *</label>
          <select class="form-input" v-model="form.role">
            <option value="">Selectionner un role</option>
            <option v-for="r in roles" :key="r.value" :value="r.value">{{ r.label }}</option>
          </select>
        </div>
        <div class="form-group" style="grid-column:1/-1"><label>Mot de passe temporaire * (min 8 caracteres + 1 chiffre)</label><input class="form-input" type="password" v-model="form.password" placeholder="L'utilisateur devra le changer a sa premiere connexion"></div>
      </div>
      <div v-if="formError" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin:12px 0">{{ formError }}</div>
      <div style="display:flex;gap:10px;margin-top:14px">
        <button class="btn btn-primary" style="width:auto" @click="createUser" :disabled="saving">{{ saving ? 'Creation...' : 'Creer le compte' }}</button>
        <button class="btn btn-secondary" style="width:auto" @click="showForm = false">Annuler</button>
      </div>
    </div>
  </div>

  <!-- Users table -->
  <div class="card">
    <div class="card-body">
      <div v-if="loading" class="ms-loading">Chargement...</div>
      <div v-else-if="!list.length" class="ms-empty">Aucun utilisateur enregistre</div>
      <template v-else>
        <div class="ptable">
          <div class="pt-head" style="grid-template-columns:44px 2fr 1fr 1fr 1fr 80px">
            <div></div><div>Utilisateur</div><div>Role</div><div>Statut</div><div>Email</div><div></div>
          </div>
          <div v-for="u in list" :key="u.id" class="pt-row" style="grid-template-columns:44px 2fr 1fr 1fr 1fr 80px">
            <div><div class="avatar a1">{{ initials(u.full_name) }}</div></div>
            <div>
              <div class="pt-name">{{ u.full_name }}</div>
              <div class="pt-code">{{ u.username }}</div>
            </div>
            <div><span :class="'pill ' + (roleColors[u.role] || 'pill-neutral')">{{ roleLabel(u.role) }}</span></div>
            <div><span :class="'pill ' + (statusColors[u.status] || 'pill-neutral')">{{ statusLabels[u.status] || u.status }}</span></div>
            <div class="pt-cell">{{ u.email || '—' }}</div>
            <div class="pt-acts">
              <button class="icon-btn" title="Desactiver" @click="disableUser(u)">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                </svg>
              </button>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>
