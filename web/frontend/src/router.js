import { createRouter, createWebHashHistory } from 'vue-router'
import { store } from './store'

import SetupPage from './pages/Setup.vue'
import LoginPage from './pages/Login.vue'
import DashboardPage from './pages/Dashboard.vue'
import CalendarPage from './pages/Calendar.vue'
import NewAptPage from './pages/NewApt.vue'
import NewPatientPage from './pages/NewPatient.vue'
import PatientsPage from './pages/Patients.vue'
import PatientFilePage from './pages/PatientFile.vue'
import RemindersPage from './pages/Reminders.vue'
import UsersPage from './pages/Users.vue'
import ProfilePage from './pages/Profile.vue'
import SettingsPage from './pages/Settings.vue'
import HelpPage from './pages/Help.vue'

const routes = [
  { path: '/setup', component: SetupPage, meta: { public: true, fullPage: true } },
  { path: '/login', component: LoginPage, meta: { public: true, fullPage: true } },
  { path: '/', component: DashboardPage, meta: { title: 'Tableau de bord' } },
  { path: '/calendar', component: CalendarPage, meta: { title: 'Calendrier', hideSearch: true } },
  { path: '/new-apt', component: NewAptPage, meta: { title: 'Prise de RDV', hideSearch: true } },
  { path: '/new-patient', component: NewPatientPage, meta: { title: 'Nouveau patient', hideSearch: true } },
  { path: '/patients', component: PatientsPage, meta: { title: 'Liste patients' } },
  { path: '/patient/:id', component: PatientFilePage, meta: { title: 'Fiche patient' } },
  { path: '/reminders', component: RemindersPage, meta: { title: 'Rappels' } },
  { path: '/users', component: UsersPage, meta: { title: 'Utilisateurs', hideSearch: true } },
  { path: '/profile', component: ProfilePage, meta: { title: 'Mon profil', hideSearch: true } },
  { path: '/settings', component: SettingsPage, meta: { title: 'Parametres', hideSearch: true } },
  { path: '/help', component: HelpPage, meta: { title: "Centre d'aide", hideSearch: true } },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

router.beforeEach((to) => {
  if (!to.meta.public && !store.user) {
    return '/login'
  }
})

export default router
