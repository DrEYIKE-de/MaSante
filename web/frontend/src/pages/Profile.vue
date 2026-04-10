<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { profile } from '../api'
import { useToast } from '../composables/useToast'

const router = useRouter()
const toast = useToast()
const loading = ref(true)
const error = ref('')
const saving = ref(false)
const form = ref({ name: '', email: '', phone: '' })

const pwForm = ref({ current: '', newPw: '', confirm: '' })
const pwError = ref('')
const pwSaving = ref(false)

onMounted(async () => {
  const res = await profile.get()
  loading.value = false
  if (!res.ok) { error.value = res.error; return }
  const d = res.data || {}
  form.value = { name: d.Name || '', email: d.Email || '', phone: d.Phone || '' }
})

async function save() {
  error.value = ''
  if (!form.value.name) { error.value = 'Le nom est requis'; return }
  saving.value = true
  const res = await profile.update(form.value)
  saving.value = false
  if (!res.ok) { error.value = res.error; return }
  toast.success('Profil mis a jour')
}

async function changePassword() {
  pwError.value = ''
  if (!pwForm.value.current) { pwError.value = 'Le mot de passe actuel est requis'; return }
  if (!pwForm.value.newPw || pwForm.value.newPw.length < 8) { pwError.value = 'Le nouveau mot de passe doit contenir au moins 8 caracteres'; return }
  if (!/\d/.test(pwForm.value.newPw)) { pwError.value = 'Le nouveau mot de passe doit contenir au moins un chiffre'; return }
  if (pwForm.value.newPw !== pwForm.value.confirm) { pwError.value = 'Les mots de passe ne correspondent pas'; return }

  pwSaving.value = true
  const res = await profile.changePassword(pwForm.value.current, pwForm.value.newPw)
  pwSaving.value = false
  if (!res.ok) { pwError.value = res.error; return }
  toast.success('Mot de passe modifie. Veuillez vous reconnecter.')
  setTimeout(() => { window.location.hash = '#/login' }, 1500)
}
</script>

<template>
  <div v-if="loading" class="ms-loading">Chargement...</div>
  <template v-else>
    <div class="apt-grid">
      <div>
        <div class="card">
          <div class="card-head"><h3>Mon profil</h3></div>
          <div class="card-body">
            <div class="form-group"><label>Nom complet</label><input class="form-input" v-model="form.name" placeholder="Votre nom"></div>
            <div class="form-group"><label>Email</label><input class="form-input" type="email" v-model="form.email" placeholder="email@exemple.com"></div>
            <div class="form-group"><label>Telephone</label><input class="form-input" v-model="form.phone" placeholder="+237 6XX XXX XXX"></div>
            <div v-if="error" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin:12px 0">{{ error }}</div>
            <button class="btn btn-primary" style="margin-top:8px" @click="save" :disabled="saving">{{ saving ? 'Sauvegarde...' : 'Sauvegarder' }}</button>
          </div>
        </div>
      </div>
      <div>
        <div class="card">
          <div class="card-head"><h3>Changer le mot de passe</h3></div>
          <div class="card-body">
            <div class="form-group"><label>Mot de passe actuel</label><input class="form-input" type="password" v-model="pwForm.current"></div>
            <div class="form-group"><label>Nouveau mot de passe</label><input class="form-input" type="password" v-model="pwForm.newPw" placeholder="Min. 8 car. + 1 chiffre"></div>
            <div class="form-group"><label>Confirmer</label><input class="form-input" type="password" v-model="pwForm.confirm"></div>
            <div v-if="pwError" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin:12px 0">{{ pwError }}</div>
            <button class="btn btn-primary" style="margin-top:8px" @click="changePassword" :disabled="pwSaving">{{ pwSaving ? 'Modification...' : 'Modifier le mot de passe' }}</button>
          </div>
        </div>
      </div>
    </div>
  </template>
</template>
