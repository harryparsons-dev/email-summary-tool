import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import ui from '@nuxt/ui/vue-plugin'
import App from './App.vue'
import Login from './pages/Login.vue'
import Profile from './pages/Profile.vue'
import Signup from './pages/Signup.vue'
import { isAuthenticated, restoreAuthentication } from './services/authService'
import './style.css'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/login' },
    { path: '/login', component: Login },
    { path: '/signup', component: Signup },
    { path: '/profile', component: Profile },
  ],
})

router.beforeEach(async (to) => {
  await restoreAuthentication()

  if (to.path === '/login' && isAuthenticated.value) {
    return '/profile'
  }
})

createApp(App).use(router).use(ui).mount('#app')
