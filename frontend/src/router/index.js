import { createRouter, createWebHashHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

import Dashboard from '../views/Dashboard.vue'
import Login from '../views/Login.vue'
import Control from '../views/Control.vue'
import Config from '../views/Config.vue'
import Logs from '../views/Logs.vue'
import Layout from '../components/Layout.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login
  },
  {
    path: '/',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: Dashboard
      },
      {
        path: '/control',
        name: 'Control',
        component: Control
      },
      {
        path: '/config',
        name: 'Config',
        component: Config
      },
      {
        path: '/logs',
        name: 'Logs',
        component: Logs
      }
    ]
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

router.beforeEach(async (to, from, next) => {
  const auth = useAuthStore()

  if (to.meta.requiresAuth) {
     const valid = await auth.checkAuth()
     if (!valid) {
         next('/login')
         return
     }
  }
  next()
})

export default router
