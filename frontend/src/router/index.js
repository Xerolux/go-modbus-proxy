 import { createRouter, createWebHashHistory } from 'vue-router'
 import { useAuthStore } from '../stores/auth'

 // Lazy load components for better performance
 const Dashboard = () => import('../views/Dashboard.vue')
 const Login = () => import('../views/Login.vue')
 const Control = () => import('../views/Control.vue')
 const Config = () => import('../views/Config.vue')
 const Logs = () => import('../views/Logs.vue')
 const Devices = () => import('../views/Devices.vue')
 const SystemInfo = () => import('../views/SystemInfo.vue')
 const Layout = () => import('../components/Layout.vue')

 const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { preload: true }
  },
  {
    path: '/',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: Dashboard,
        meta: { preload: true }
      },
      {
        path: '/control',
        name: 'Control',
        component: Control
      },
      {
        path: '/devices',
        name: 'Devices',
        component: Devices
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
      },
      {
        path: '/system',
        name: 'System',
        component: SystemInfo
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
