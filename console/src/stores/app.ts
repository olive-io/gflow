import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export interface SidebarState {
  collapsed: boolean
}

export type Theme = 'light' | 'dark' | 'system'

export const useAppStore = defineStore('app', () => {
  const sidebarCollapsed = ref(false)
  const loading = ref(false)
  const theme = ref<Theme>((localStorage.getItem('theme') as Theme) || 'system')

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setSidebarCollapsed(collapsed: boolean) {
    sidebarCollapsed.value = collapsed
  }

  function setLoading(value: boolean) {
    loading.value = value
  }

  function setTheme(newTheme: Theme) {
    theme.value = newTheme
    localStorage.setItem('theme', newTheme)
    applyTheme(newTheme)
  }

  function applyTheme(themeValue: Theme) {
    const root = document.documentElement
    root.classList.remove('light', 'dark')

    if (themeValue === 'system') {
      const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      root.classList.add(systemDark ? 'dark' : 'light')
    } else {
      root.classList.add(themeValue)
    }
  }

  function initTheme() {
    applyTheme(theme.value)

    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
      if (theme.value === 'system') {
        applyTheme('system')
      }
    })
  }

  watch(theme, (newTheme) => {
    applyTheme(newTheme)
  })

  return {
    sidebarCollapsed,
    loading,
    theme,
    toggleSidebar,
    setSidebarCollapsed,
    setLoading,
    setTheme,
    initTheme,
  }
})
