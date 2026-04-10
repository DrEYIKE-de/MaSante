<script setup>
import { useRouter, useRoute } from 'vue-router'
import { store } from '../store'
import { auth } from '../api'
import { useToast } from '../composables/useToast'

const router = useRouter()
const route = useRoute()
const toast = useToast()

const navSections = [
  { label: 'Principal', items: [
    { text: 'Tableau de bord', path: '/' },
    { text: 'Calendrier RDV', path: '/calendar' },
    { text: 'Prise de RDV', path: '/new-apt' },
  ]},
  { label: 'Patients', items: [
    { text: 'Nouveau patient', path: '/new-patient' },
    { text: 'Liste patients', path: '/patients' },
  ]},
  { label: 'Terrain', items: [
    { text: 'Rappels', path: '/reminders' },
  ]},
  { label: 'Systeme', items: [
    { text: 'Utilisateurs', path: '/users' },
    { text: 'Parametres', path: '/settings' },
    { text: "Centre d'aide", path: '/help' },
  ]},
]

const roleLabels = {
  admin: 'Administrateur',
  medecin: 'Medecin',
  infirmier: 'Infirmier(e)',
  asc: 'Agent communautaire',
}

function initials(name) {
  return (name || '').split(' ').map(w => w[0]).join('').substring(0, 2).toUpperCase() || '?'
}

async function logout() {
  await auth.logout()
  store.user = null
  toast.info('Deconnecte')
  router.push('/login')
}
</script>

<template>
  <div class="app visible">
    <aside class="sidebar">
      <!-- Brand -->
      <div class="sb-brand">
        <div class="sb-logo">
          <svg width="20" height="20" viewBox="0 0 80 80" fill="white" stroke="white" stroke-linecap="round"><circle cx="40" cy="54" r="16"/><line x1="40" y1="38" x2="40" y2="16" stroke-width="5" fill="none"/><path d="M40,24 C30,14 16,14 10,22" stroke-width="5" fill="none"/><path d="M40,24 C50,14 64,14 70,22" stroke-width="5" fill="none"/></svg>
        </div>
        <div class="sb-brand-text">
          <h3>MaSante</h3>
          <span v-if="store.centerName">{{ store.centerName }}</span>
          <span v-else>Plateforme de sante</span>
        </div>
      </div>

      <!-- Nav -->
      <nav class="sb-nav">
        <div v-for="section in navSections" :key="section.label" class="sb-section">
          <div class="sb-section-label">{{ section.label }}</div>
          <div
            v-for="item in section.items"
            :key="item.path"
            class="sb-item"
            :class="{ active: route.path === item.path }"
            @click="router.push(item.path)"
          >
            {{ item.text }}
          </div>
        </div>
      </nav>

      <!-- User -->
      <div class="sb-user">
        <div class="sb-avatar" @click="router.push('/profile')" style="cursor:pointer">
          {{ store.user ? initials(store.user.full_name) : '?' }}
        </div>
        <div class="sb-user-info" @click="router.push('/profile')" style="cursor:pointer">
          <div class="name">{{ store.user?.full_name || '' }}</div>
          <div class="role">{{ roleLabels[store.user?.role] || store.user?.role || '' }}</div>
        </div>
        <button class="icon-btn" @click="logout" title="Se deconnecter" style="color:rgba(255,255,255,.4);flex-shrink:0">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/>
          </svg>
        </button>
      </div>
    </aside>

    <main class="main">
      <header class="topbar">
        <div class="topbar-title">{{ route.meta.title || 'MaSante' }}</div>
        <div v-if="!route.meta.hideSearch" class="topbar-search">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--gray-300)" stroke-width="1.5" style="position:absolute;left:10px;top:50%;transform:translateY(-50%)">
            <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <input type="text" placeholder="Rechercher un patient..." @keydown.enter="$router.push({ path: '/patients', query: { q: $event.target.value.trim() } })">
        </div>
        <div class="topbar-right">
          <div class="topbar-badge online">
            <span class="bdot"></span>
            <span>En ligne</span>
          </div>
        </div>
      </header>

      <div class="content">
        <router-view />
      </div>
    </main>
  </div>
</template>
