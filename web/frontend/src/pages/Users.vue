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
  { value: 'asc', label: 'Agent communautaire' },
]
const roleColors = { admin: 'pill-danger', medecin: 'pill-info', infirmier: 'pill-warning', asc: 'pill-success' }
const statusLabels = { active: 'Actif', conge: 'En conge', desactive: 'Desactive' }
const statusColors = { active: 'pill-success', conge: 'pill-neutral', desactive: 'pill-danger' }

async function load() {
  loading.value = true
  const res = await users.list()
  loading.value = false
  if (!res.ok) { error.value = res.error; return }
  list.value = res.data || []
}
onMounted(load)

function roleLabel(r) { return (roles.find(x => x.value === r) || {}).label || r }
function initials(name) { return (name || '??').split(' ').map(w => w[0]).join('').substring(0, 2).toUpperCase() }

function openForm() {
  form.value = { full_name: '', email: '', username: '', password: '', role: '' }
  formError.value = ''
  showForm.value = true
}

async function createUser() {
  formError.value = ''
  if (!form.value.full_name) { formError.value = 'Nom requis'; return }
  if (!form.value.username) { formError.value = 'Identifiant requis'; return }
  if (!form.value.password || form.value.password.length < 8) { formError.value = 'Mot de passe: min 8 caracteres'; return }
  if (!/\d/.test(form.value.password)) { formError.value = 'Mot de passe: au moins 1 chiffre'; return }
  if (!form.value.role) { formError.value = 'Role requis'; return }
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
  <div v-if="error" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin-bottom:12px">{{ error }}</div>

  <!-- Add form -->
  <div v-if="showForm" class="card" style="margin-bottom:16px">
    <div class="card-head"><h3>Nouvel utilisateur</h3></div>
    <div class="card-body" style="padding:16px">
      <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px">
        <div class="form-group" style="margin-bottom:0"><label>Nom complet *</label><input class="form-input" v-model="form.full_name" placeholder="Ex: Dr. Kamga Jean"></div>
        <div class="form-group" style="margin-bottom:0"><label>Email</label><input class="form-input" type="email" v-model="form.email" placeholder="kamga@hopital.cm"></div>
        <div class="form-group" style="margin-bottom:0"><label>Identifiant *</label><input class="form-input" v-model="form.username" placeholder="kamga.j"></div>
        <div class="form-group" style="margin-bottom:0">
          <label>Role *</label>
          <select class="form-input" v-model="form.role">
            <option value="">Selectionner</option>
            <option v-for="r in roles" :key="r.value" :value="r.value">{{ r.label }}</option>
          </select>
        </div>
        <div class="form-group" style="margin-bottom:0;grid-column:1/-1"><label>Mot de passe * (min 8 + 1 chiffre)</label><input class="form-input" type="password" v-model="form.password" placeholder="Mot de passe temporaire"></div>
      </div>
      <div v-if="formError" style="padding:8px 12px;background:var(--danger-bg);color:var(--danger);border-radius:6px;font-size:.82rem;margin-top:10px">{{ formError }}</div>
      <div style="display:flex;gap:8px;margin-top:12px">
        <button class="btn btn-primary" style="width:auto" @click="createUser" :disabled="saving">{{ saving ? 'Creation...' : 'Creer' }}</button>
        <button class="btn btn-secondary" style="width:auto" @click="showForm = false">Annuler</button>
      </div>
    </div>
  </div>

  <!-- Users list -->
  <div class="card">
    <div class="card-head" style="padding:12px 16px">
      <h3>Equipe soignante</h3>
      <button class="btn btn-sm btn-primary" style="width:auto" @click="openForm">+ Ajouter</button>
    </div>
    <div v-if="loading" class="ms-loading">Chargement...</div>
    <div v-else-if="!list.length" class="ms-empty">Aucun utilisateur</div>
    <div v-else style="padding:0">
      <div v-for="u in list" :key="u.id" style="display:flex;align-items:center;gap:12px;padding:10px 16px;border-bottom:1px solid var(--gray-50)">
        <div class="avatar a1" style="width:34px;height:34px;font-size:.72rem">{{ initials(u.full_name) }}</div>
        <div style="flex:1;min-width:0">
          <div style="font-weight:600;font-size:.88rem">{{ u.full_name }}</div>
          <div style="font-size:.75rem;color:var(--gray-400)">{{ u.username }}{{ u.email ? ' — ' + u.email : '' }}</div>
        </div>
        <span :class="'pill ' + (roleColors[u.role] || 'pill-neutral')" style="flex-shrink:0">{{ roleLabel(u.role) }}</span>
        <span :class="'pill ' + (statusColors[u.status] || 'pill-neutral')" style="flex-shrink:0">{{ statusLabels[u.status] || u.status }}</span>
        <button class="icon-btn" title="Desactiver" style="flex-shrink:0" @click="disableUser(u)">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="var(--gray-400)" stroke-width="1.5" stroke-linecap="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
        </button>
      </div>
    </div>
  </div>
</template>
