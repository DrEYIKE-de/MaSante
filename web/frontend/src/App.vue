<script setup>
import { onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { store } from './store'
import { setup as setupApi, auth } from './api'
import AppShell from './components/AppShell.vue'

const router = useRouter()
const route = useRoute()

onMounted(async () => {
  const res = await setupApi.status()
  if (res.ok && !res.data.setup_complete) {
    store.setupDone = false
    store.setupStep = res.data.current_step || 0
    router.push('/setup')
    return
  }
  store.setupDone = true

  const me = await auth.me()
  if (me.ok) {
    store.user = me.data
    if (route.path === '/login' || route.path === '/setup') {
      router.push('/')
    }
  } else {
    router.push('/login')
  }
})
</script>

<template>
  <template v-if="store.setupDone === null">
    <div class="ms-loading">Chargement...</div>
  </template>
  <template v-else-if="$route.meta.fullPage">
    <router-view />
  </template>
  <template v-else-if="store.user">
    <AppShell />
  </template>
  <template v-else>
    <router-view />
  </template>

  <!-- Toasts -->
  <div id="ms-toasts"></div>
</template>
