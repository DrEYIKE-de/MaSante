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
const form = ref({ name: '', username: '', password: '', role: '' })
const formError = ref('')

const roles = [
  { value: 'admin', label: 'Administrateur' },
  { value: 'medecin', label: 'Medecin' },
  { value: 'infirmier', label: 'Infirmier' },
  { value: 'pharmacien', label: 'Pharmacien' },
  { value: 'conseiller', label: 'Conseiller' },
]

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
  form.value = { name: '', username: '', password: '', role: '' }
  formError.value = ''
  showForm.value = true
}

function roleLabel(r) { return (roles.find(x => x.value === r) || {}).label || r }

async function createUser() {
  formError.value = ''
  if (!form.value.name) { formError.value = 'Le nom est requis'; return }
  if (!form.value.username) { formError.value = "Le nom d'utilisateur est requis"; return }
  if (!form.value.password || form.value.password.length < 8) { formError.value = 'Le mot de passe doit contenir au moins 8 caracteres'; return }
  if (!/\d/.test(form.value.password)) { formError.value = 'Le mot de passe doit contenir au moins un chiffre'; return }
  if (!form.value.role) { formError.value = 'Le role est requis'; return }

  saving.value = true
  const res = await users.create(form.value)
  saving.value = false
  if (!res.ok) { formError.value = res.error; return }
  toast.success('Utilisateur cree')
  showForm.value = false
  load()
}

async function disableUser(u) {
  if (!confirm('Desactiver ' + u.Name + ' ?')) return
  error.value = ''
  const res = await users.disable(u.ID)
  if (!res.ok) { error.value = res.error; return }
  toast.success('Utilisateur desactive')
  load()
}

function fmtDate(d) { return d ? new Date(d).toLocaleDateString('fr-FR') : '—' }
function initials(n) { return (n || '??').split(' ').map(w => w[0]).join('').substring(0, 2).toUpperCase() }
</script>

<template>
  <div v-if="error" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin-bottom:12px">{{ error }}</div>

  <!-- Add user form -->
  <div v-if="showForm" class="card" style="margin-bottom:16px">
    <div class="card-head"><h3>Nouvel utilisateur</h3></div>
    <div class="card-body">
      <div style="display:grid;grid-template-columns:1fr 1fr;gap:14px">
        <div class="form-group"><label>Nom complet *</label><input class="form-input" v-model="form.name" placeholder="Ex: Dr. Kamga"></div>
        <div class="form-group"><label>Nom d'utilisateur *</label><input class="form-input" v-model="form.username" placeholder="Ex: kamga"></div>
        <div class="form-group"><label>Mot de passe *</label><input class="form-input" type="password" v-model="form.password" placeholder="Min. 8 car. + 1 chiffre"></div>
        <div class="form-group"><label>Role *</label>
          <select class="form-input" v-model="form.role">
            <option value="">Selectionner</option>
            <option v-for="r in roles" :key="r.value" :value="r.value">{{ r.label }}</option>
          </select>
        </div>
      </div>
      <div v-if="formError" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin:12px 0">{{ formError }}</div>
      <div style="display:flex;gap:10px;margin-top:14px">
        <button class="btn btn-primary" @click="createUser" :disabled="saving">{{ saving ? 'Creation...' : 'Creer' }}</button>
        <button class="btn btn-secondary" @click="showForm = false">Annuler</button>
      </div>
    </div>
  </div>

  <div class="card">
    <div class="card-head"><h3>Utilisateurs</h3><button class="btn btn-sm btn-primary" @click="openForm">+ Ajouter</button></div>
    <div class="card-body">
      <div v-if="loading" class="ms-loading">Chargement...</div>
      <div v-else-if="!list.length" class="ms-empty">Aucun utilisateur</div>
      <template v-else>
        <div class="ptable">
          <div class="pt-head" style="grid-template-columns:44px 2fr 1fr 1fr 1fr 80px">
            <div></div><div>Nom</div><div>Identifiant</div><div>Role</div><div>Cree le</div><div></div>
          </div>
          <div v-for="u in list" :key="u.ID" class="pt-row" style="grid-template-columns:44px 2fr 1fr 1fr 1fr 80px">
            <div><div class="avatar a2">{{ initials(u.Name) }}</div></div>
            <div><div class="pt-name">{{ u.Name }}</div></div>
            <div class="pt-cell">{{ u.Username }}</div>
            <div><span class="pill pill-info">{{ roleLabel(u.Role) }}</span></div>
            <div class="pt-cell">{{ fmtDate(u.CreatedAt) }}</div>
            <div class="pt-acts"><button class="icon-btn" title="Desactiver" @click="disableUser(u)">&#128683;</button></div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>
