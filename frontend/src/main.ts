import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import ui from '@nuxt/ui/vue-plugin'
import App from './App.vue'
import Login from './pages/Login.vue'
import Profile from './pages/Profile.vue'
import ProjectDetail from './pages/ProjectDetail.vue'
import Projects from './pages/Projects.vue'
import Signup from './pages/Signup.vue'
import { isAuthenticated, restoreAuthentication } from './services/authService'
import './style.css'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/projects' },
    { path: '/login', component: Login },
    { path: '/signup', component: Signup },
    { path: '/account', component: Profile, meta: { requiresAuthentication: true } },
    { path: '/projects', component: Projects, meta: { requiresAuthentication: true } },
    { path: '/projects/:id', component: ProjectDetail, meta: { requiresAuthentication: true } },
  ],
})

router.beforeEach(async (to) => {
  await restoreAuthentication()

  if (to.path === '/login' && isAuthenticated.value) {
    return '/projects'
  }

  if (to.meta.requiresAuthentication && !isAuthenticated.value) {
    return '/login'
  }
})

createApp(App).use(router).use(ui).mount('#app')
