import { defineStore } from 'pinia'
import { ref } from 'vue'
import axios from 'axios'

export const useAuthStore = defineStore('auth', () => {
  const isAuthenticated = ref(false)
  const checking = ref(false)

  const checkAuth = async () => {
    if (checking.value) return isAuthenticated.value
    checking.value = true
    try {
      // Just check health or a protected endpoint to verify session
      await axios.get('/api/proxies')
      isAuthenticated.value = true
    } catch (e) {
      isAuthenticated.value = false
    } finally {
      checking.value = false
    }
    return isAuthenticated.value
  }

  const login = async (password) => {
    try {
      await axios.post('/api/login', { password })
      isAuthenticated.value = true
      return true
    } catch (e) {
      isAuthenticated.value = false
      return false
    }
  }

  const logout = async () => {
    // Ideally call logout API if it exists, but just clearing state for now
    // Or set cookie expiration. The backend uses HTTP-only cookies.
    // If we assume the cookie is removed by browser on session end or we need an endpoint.
    // The backend doesn't have explicit logout endpoint, but we can just redirect to login.
    isAuthenticated.value = false
  }

  return { isAuthenticated, checkAuth, login, logout }
})
