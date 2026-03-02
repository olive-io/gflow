import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'
import type { User, Token, Role, Policy } from '@/types/api'

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const token = ref<Token | null>(null)
  const role = ref<Role | null>(null)
  const policies = ref<Policy[]>([])

  const isLoggedIn = computed(() => !!token.value?.text)
  const username = computed(() => user.value?.username || '')
  const tokenText = computed(() => token.value?.text || '')
  const isAdmin = computed(() => role.value?.type === 1)

  function setUser(userData: User) {
    user.value = userData
    localStorage.setItem('user', JSON.stringify(userData))
  }

  function setToken(newToken: Token) {
    token.value = newToken
    localStorage.setItem('token', newToken.text)
  }

  function setRole(newRole: Role) {
    role.value = newRole
    localStorage.setItem('role', JSON.stringify(newRole))
  }

  function setPolicies(newPolicies: Policy[]) {
    policies.value = newPolicies
    localStorage.setItem('policies', JSON.stringify(newPolicies))
  }

  function logout() {
    user.value = null
    token.value = null
    role.value = null
    policies.value = []
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    localStorage.removeItem('role')
    localStorage.removeItem('policies')
  }

  async function login(username: string, password: string) {
    const response = await authApi.login(username, password)
    setToken(response.token)
    return response
  }

  async function fetchCurrentUser() {
    const response = await authApi.getSelf()
    setUser(response.user)
    setRole(response.role)
    setPolicies(response.policies)
    return response
  }

  function loadFromStorage() {
    const storedToken = localStorage.getItem('token')
    const storedUser = localStorage.getItem('user')
    const storedRole = localStorage.getItem('role')
    const storedPolicies = localStorage.getItem('policies')
    
    if (storedToken) {
      token.value = {
        id: 0,
        text: storedToken,
        expire_at: 0,
        enable: 1,
        user_id: 0,
        role_id: 0,
        create_at: 0,
      }
    }
    if (storedUser) {
      try {
        user.value = JSON.parse(storedUser)
      } catch {
        // ignore
      }
    }
    if (storedRole) {
      try {
        role.value = JSON.parse(storedRole)
      } catch {
        // ignore
      }
    }
    if (storedPolicies) {
      try {
        policies.value = JSON.parse(storedPolicies)
      } catch {
        // ignore
      }
    }
  }

  loadFromStorage()

  return {
    user,
    token,
    role,
    policies,
    isLoggedIn,
    username,
    tokenText,
    isAdmin,
    setUser,
    setToken,
    setRole,
    setPolicies,
    logout,
    login,
    fetchCurrentUser,
  }
})
