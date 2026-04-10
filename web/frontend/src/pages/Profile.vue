<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { profile } from '../api'
import { store } from '../store'
import { useToast } from '../composables/useToast'

const router = useRouter()
const toast = useToast()
const loading = ref(true)
const error = ref('')
const saving = ref(false)
const form = ref({ full_name: '', email: '', phone: '' })

const pwForm = ref({ current: '', newPw: '', confirm: '' })
const pwError = ref('')
const pwSaving = ref(false)

onMounted(async () => {
  const res = await profile.get()
  loading.value = false
  if (!res.ok) { error.value = res.error; return }
  const d = res.data || {}
  form.value = { full_name: d.full_name || '', email: d.email || '', phone: d.phone || '' }
})

async function save() {
  error.value = ''
  if (!form.value.full_name) { error.value = 'Le nom est requis'; return }
  saving.value = true
  const res = await profile.update(form.value)
  saving.value = false
  if (!res.ok) { error.value = res.error; return }
  store.user.full_name = form.value.full_name
  toast.success('Profil mis a jour')
}

async function changePassword() {
  pwError.value = ''
  if (!pwForm.value.current) { pwError.value = 'Mot de passe actuel requis'; return }
  if (!pwForm.value.newPw || pwForm.value.newPw.length < 8) { pwError.value = 'Minimum 8 caracteres'; return }
  if (!/\d/.test(pwForm.value.newPw)) { pwError.value = 'Au moins 1 chiffre requis'; return }
  if (pwForm.value.newPw !== pwForm.value.confirm) { pwError.value = 'Les mots de passe ne correspondent pas'; return }

  pwSaving.value = true
  const res = await profile.changePassword(pwForm.value.current, pwForm.value.newPw)
  pwSaving.value = false
  if (!res.ok) { pwError.value = res.error; return }
  toast.success('Mot de passe modifie — redirection...')
  setTimeout(() => { store.user = null; router.push('/login') }, 2000)
}
</script>

<template>
  <div v-if="loading" class="ms-loading">Chargement...</div>
  <template v-else>
    <div class="grid-2" style="align-items:start">
      <div class="card">
        <div class="card-head"><h3>Mon profil</h3></div>
        <div class="card-body">
          <div class="form-group"><label>Nom complet</label><input class="form-input" v-model="form.full_name" placeholder="Votre nom"></div>
          <div class="form-group"><label>Email</label><input class="form-input" type="email" v-model="form.email" placeholder="email@exemple.cm"></div>
          <div class="form-group"><label>Telephone</label><input class="form-input" type="tel" v-model="form.phone" placeholder="+237 6XX XXX XXX"></div>
          <div v-if="error" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin:12px 0">{{ error }}</div>
          <button class="btn btn-primary" style="width:auto;margin-top:8px" @click="save" :disabled="saving">{{ saving ? 'Sauvegarde...' : 'Enregistrer' }}</button>
        </div>
      </div>
      <div class="card">
        <div class="card-head"><h3>Changer le mot de passe</h3></div>
        <div class="card-body">
          <div class="form-group"><label>Mot de passe actuel</label><input class="form-input" type="password" v-model="pwForm.current"></div>
          <div class="form-group"><label>Nouveau (min 8 + 1 chiffre)</label><input class="form-input" type="password" v-model="pwForm.newPw"></div>
          <div class="form-group"><label>Confirmer</label><input class="form-input" type="password" v-model="pwForm.confirm"></div>
          <div v-if="pwError" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin:12px 0">{{ pwError }}</div>
          <button class="btn btn-secondary" style="width:auto;margin-top:8px" @click="changePassword" :disabled="pwSaving">{{ pwSaving ? 'Modification...' : 'Modifier le mot de passe' }}</button>
        </div>
      </div>
    </div>
  </template>
</template>
