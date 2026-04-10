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
    { icon: 'chart', text: 'Tableau de bord', path: '/' },
    { icon: 'calendar', text: 'Calendrier RDV', path: '/calendar' },
    { icon: 'plus', text: 'Prise de RDV', path: '/new-apt' },
  ]},
  { label: 'Patients', items: [
    { icon: 'user-plus', text: 'Nouveau patient', path: '/new-patient' },
    { icon: 'users', text: 'Liste patients', path: '/patients' },
  ]},
  { label: 'Terrain', items: [
    { icon: 'bell', text: 'Rappels', path: '/reminders' },
  ]},
  { label: 'Systeme', items: [
    { icon: 'shield', text: 'Utilisateurs', path: '/users' },
    { icon: 'settings', text: 'Parametres', path: '/settings' },
    { icon: 'help-circle', text: "Centre d'aide", path: '/help' },
  ]},
]

function initials(name) {
  return (name || '??').split(' ').map(w => w[0]).join('').substring(0, 2).toUpperCase()
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
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="sb-brand">
        <div class="sb-logo">&#127807;</div>
        <div class="sb-brand-text">
          <h3>MaSante</h3>
          <span>Plateforme de sante</span>
        </div>
      </div>

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

      <div class="sb-user">
        <div class="sb-avatar" @click="router.push('/profile')" style="cursor:pointer">
          {{ store.user ? initials(store.user.full_name) : '?' }}
        </div>
        <div class="sb-user-info" @click="router.push('/profile')" style="cursor:pointer">
          <div class="name">{{ store.user?.full_name || '' }}</div>
          <div class="role">{{ store.user?.role || '' }}</div>
        </div>
        <button class="icon-btn" @click="logout" title="Deconnecter" style="color:rgba(255,255,255,.35);flex-shrink:0">
          &#x2192;
        </button>
      </div>
    </aside>

    <!-- Main -->
    <main class="main">
      <header class="topbar">
        <div class="topbar-title">{{ route.meta.title || 'MaSante' }}</div>
        <div v-if="!route.meta.hideSearch" class="topbar-search">
          <input type="text" placeholder="Rechercher un patient..." @keydown.enter="searchPatient($event)">
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

<script>
export default {
  methods: {
    searchPatient(e) {
      const q = e.target.value.trim()
      if (q) {
        this.$router.push({ path: '/patients', query: { q } })
      }
    }
  }
}
</script>
