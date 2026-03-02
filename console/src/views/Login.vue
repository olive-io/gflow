<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useUserStore } from '@/stores'
import { authApi } from '@/api'
import { Workflow } from 'lucide-vue-next'

const router = useRouter()
const userStore = useUserStore()

const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function handleLogin() {
  if (!username.value || !password.value) {
    error.value = '请输入用户名和密码'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const response = await authApi.login(username.value, password.value)
    
    if (response.token?.text) {
      userStore.setToken(response.token)
      
      const selfResponse = await authApi.getSelf()
      if (selfResponse.user) {
        userStore.setUser(selfResponse.user)
        userStore.setRole(selfResponse.role)
      }
      
      router.push('/dashboard')
    } else {
      error.value = '登录失败，未获取到令牌'
    }
  } catch (e: unknown) {
    const err = e as Error
    error.value = err.message || '登录失败，请检查用户名和密码'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex">
    <!-- Left side - Brand section -->
    <div class="hidden lg:flex lg:w-1/2 gflow-gradient relative overflow-hidden">
      <!-- Background pattern -->
      <div class="absolute inset-0 opacity-10">
        <svg class="w-full h-full" xmlns="http://www.w3.org/2000/svg">
          <defs>
            <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
              <path d="M 40 0 L 0 0 0 40" fill="none" stroke="white" stroke-width="1"/>
            </pattern>
          </defs>
          <rect width="100%" height="100%" fill="url(#grid)" />
        </svg>
      </div>
      
      <div class="relative z-10 flex flex-col justify-center items-center w-full px-12 text-white">
        <!-- Logo -->
        <div class="flex items-center gap-3 mb-8">
          <div class="w-16 h-16 bg-white/20 rounded-2xl flex items-center justify-center backdrop-blur-sm">
            <Workflow class="w-10 h-10" />
          </div>
          <span class="text-4xl font-bold">GFlow</span>
        </div>
        
        <!-- Title -->
        <h1 class="text-4xl font-bold text-center mb-4">
          工作流管理平台
        </h1>
        
        <!-- Description -->
        <p class="text-xl text-white/80 text-center max-w-md">
          高效、可靠的工作流引擎，助力企业数字化转型
        </p>
        
        <!-- Features -->
        <div class="mt-12 grid grid-cols-2 gap-6 max-w-lg">
          <div class="flex items-center gap-3 bg-white/10 rounded-xl p-4 backdrop-blur-sm">
            <div class="w-10 h-10 bg-white/20 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <span class="font-medium">高性能</span>
          </div>
          <div class="flex items-center gap-3 bg-white/10 rounded-xl p-4 backdrop-blur-sm">
            <div class="w-10 h-10 bg-white/20 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
            </div>
            <span class="font-medium">安全可靠</span>
          </div>
          <div class="flex items-center gap-3 bg-white/10 rounded-xl p-4 backdrop-blur-sm">
            <div class="w-10 h-10 bg-white/20 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
              </svg>
            </div>
            <span class="font-medium">可视化设计</span>
          </div>
          <div class="flex items-center gap-3 bg-white/10 rounded-xl p-4 backdrop-blur-sm">
            <div class="w-10 h-10 bg-white/20 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 4a2 2 0 114 0v1a1 1 0 001 1h3a1 1 0 011 1v3a1 1 0 01-1 1h-1a2 2 0 100 4h1a1 1 0 011 1v3a1 1 0 01-1 1h-3a1 1 0 01-1-1v-1a2 2 0 10-4 0v1a1 1 0 01-1 1H7a1 1 0 01-1-1v-3a1 1 0 00-1-1H4a2 2 0 110-4h1a1 1 0 001-1V7a1 1 0 011-1h3a1 1 0 001-1V4z" />
              </svg>
            </div>
            <span class="font-medium">插件扩展</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Right side - Login form -->
    <div class="w-full lg:w-1/2 flex items-center justify-center p-8 bg-background">
      <Card class="w-full max-w-md border-0 shadow-none">
        <CardHeader class="space-y-1 pb-8">
          <div class="flex items-center gap-2 mb-2 lg:hidden">
            <div class="w-10 h-10 bg-gflow-primary rounded-xl flex items-center justify-center text-white">
              <Workflow class="w-6 h-6" />
            </div>
            <span class="text-2xl font-bold">GFlow</span>
          </div>
          <CardTitle class="text-2xl font-bold">欢迎回来</CardTitle>
          <CardDescription>
            请输入您的账号信息登录系统
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form @submit.prevent="handleLogin" class="space-y-4">
            <div class="space-y-2">
              <Label for="username">用户名</Label>
              <Input
                id="username"
                v-model="username"
                type="text"
                placeholder="请输入用户名"
                autocomplete="username"
              />
            </div>
            <div class="space-y-2">
              <Label for="password">密码</Label>
              <Input
                id="password"
                v-model="password"
                type="password"
                placeholder="请输入密码"
                autocomplete="current-password"
              />
            </div>
            
            <div v-if="error" class="text-sm text-destructive bg-destructive/10 rounded-lg p-3">
              {{ error }}
            </div>
            
            <Button type="submit" class="w-full" :disabled="loading" size="lg">
              {{ loading ? '登录中...' : '登录' }}
            </Button>
          </form>
          
          <div class="mt-6 text-center text-sm text-muted-foreground">
            <p>默认账号: admin / p@ssw0rd</p>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
