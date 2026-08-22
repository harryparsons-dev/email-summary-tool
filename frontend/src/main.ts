import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import ui from '@nuxt/ui/vue-plugin'
import App from './App.vue'
import LoginView from './views/LoginView.vue'
import SignupView from './views/SignupView.vue'
import './style.css'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/login' },
    { path: '/login', component: LoginView },
    { path: '/signup', component: SignupView },
  ],
})

createApp(App).use(router).use(ui).mount('#app')
