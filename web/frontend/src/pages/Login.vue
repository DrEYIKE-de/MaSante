<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { auth } from '../api'
import { store } from '../store'
import { useToast } from '../composables/useToast'

const router = useRouter()
const toast = useToast()

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function login() {
  error.value = ''
  if (!username.value || !password.value) { error.value = 'Identifiant et mot de passe requis'; return }
  loading.value = true
  const res = await auth.login(username.value, password.value)
  loading.value = false
  if (!res.ok) { error.value = res.error; return }
  store.user = res.data
  toast.success('Connexion reussie')
  router.push('/')
}
</script>

<template>
  <div style="position:fixed;inset:0;display:flex;z-index:100">
    <div style="flex:1;background:var(--primary);display:flex;flex-direction:column;justify-content:center;align-items:center;color:#fff">
      <div style="text-align:center">
        <div style="width:80px;height:80px;border:2px solid rgba(255,255,255,.15);border-radius:20px;display:flex;align-items:center;justify-content:center;margin:0 auto 28px;background:rgba(255,255,255,.06);"><svg width="36" height="36" viewBox="0 0 32 32" fill="none" stroke="rgba(255,255,255,.7)" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><circle cx="16" cy="16" r="12" opacity=".3"/><path d="M16 8v16M8 16h16"/><circle cx="16" cy="16" r="3" fill="rgba(255,255,255,.7)" stroke="none" opacity=".5"/></svg></div>
        <h1 style="font-size:2.8rem;font-weight:700;letter-spacing:-1.5px;margin-bottom:8px">MaSante</h1>
        <p style="font-size:1rem;color:rgba(255,255,255,.5)">Plateforme de suivi sante communautaire</p>
      </div>
    </div>
    <div style="width:440px;background:var(--white);display:flex;flex-direction:column;justify-content:center;padding:56px">
      <h2 style="font-size:1.7rem;margin-bottom:6px">Connexion</h2>
      <p style="color:var(--gray-400);margin-bottom:32px;font-size:.9rem">Entrez vos identifiants</p>

      <div v-if="error" style="padding:10px 14px;background:var(--danger-bg);color:var(--danger);border-radius:var(--radius);font-size:.85rem;margin-bottom:16px">{{ error }}</div>

      <div class="form-group">
        <label>Identifiant</label>
        <input class="form-input" v-model="username" placeholder="Votre identifiant" @keydown.enter="login">
      </div>
      <div class="form-group">
        <label>Mot de passe</label>
        <input class="form-input" type="password" v-model="password" placeholder="Votre mot de passe" @keydown.enter="login">
      </div>

      <button class="btn btn-primary" style="margin-top:8px" @click="login" :disabled="loading">
        {{ loading ? 'Connexion...' : 'Se connecter' }}
      </button>
    </div>
  </div>
</template>
