<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore, useUserStore } from '@/stores'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  LayoutDashboard,
  FileText,
  GitBranch,
  Pencil,
  Server,
  Users,
  ScrollText,
  ChevronLeft,
  ChevronRight,
  LogOut,
  User,
  Settings,
  Workflow,
} from 'lucide-vue-next'
import { cn } from '@/lib/utils'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()
const userStore = useUserStore()

const menuItems = [
  { path: '/dashboard', name: '仪表盘', icon: LayoutDashboard },
  { path: '/definitions', name: '流程定义', icon: FileText },
  { path: '/processes', name: '流程实例', icon: GitBranch },
  { path: '/runners', name: 'Runner 管理', icon: Server },
  { path: '/users', name: '用户管理', icon: Users },
  { path: '/audit-logs', name: '审计日志', icon: ScrollText },
]

const breadcrumbs = computed(() => {
  const matched = route.matched.filter(r => r.meta?.title)
  return matched.map(r => ({
    title: r.meta?.title as string,
    path: r.path,
  }))
})

function handleLogout() {
  userStore.logout()
  router.push('/login')
}

function getInitials(name: string) {
  return name.slice(0, 2).toUpperCase()
}
</script>

<template>
  <div class="min-h-screen flex bg-background">
    <!-- Sidebar -->
    <aside
      :class="cn(
        'fixed left-0 top-0 z-40 h-screen bg-card border-r transition-all duration-300',
        appStore.sidebarCollapsed ? 'w-16' : 'w-64'
      )"
    >
      <!-- Logo -->
      <div class="h-16 flex items-center justify-between px-4 border-b">
        <div class="flex items-center gap-3 overflow-hidden">
          <div class="w-10 h-10 bg-gflow-primary rounded-xl flex items-center justify-center text-white shrink-0">
            <Workflow class="w-6 h-6" />
          </div>
          <span
            v-if="!appStore.sidebarCollapsed"
            class="text-xl font-bold whitespace-nowrap"
          >
            GFlow
          </span>
        </div>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 py-4 px-3 space-y-1 overflow-y-auto">
        <router-link
          v-for="item in menuItems"
          :key="item.path"
          :to="item.path"
          :class="cn(
            'flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors',
            route.path === item.path || route.path.startsWith(item.path + '/')
              ? 'bg-gflow-primary/10 text-gflow-primary'
              : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
          )"
        >
          <component :is="item.icon" class="w-5 h-5 shrink-0" />
          <span v-if="!appStore.sidebarCollapsed" class="whitespace-nowrap">
            {{ item.name }}
          </span>
        </router-link>
      </nav>

      <!-- Collapse button -->
      <div class="absolute bottom-4 left-0 right-0 px-3">
        <Button
          variant="ghost"
          size="icon"
          class="w-full h-10"
          @click="appStore.toggleSidebar"
        >
          <ChevronLeft v-if="!appStore.sidebarCollapsed" class="w-5 h-5" />
          <ChevronRight v-else class="w-5 h-5" />
        </Button>
      </div>
    </aside>

    <!-- Main content -->
    <div
      :class="cn(
        'flex-1 flex flex-col transition-all duration-300',
        appStore.sidebarCollapsed ? 'ml-16' : 'ml-64'
      )"
    >
      <!-- Header -->
      <header class="h-16 bg-card border-b flex items-center justify-between px-6 sticky top-0 z-30">
        <!-- Breadcrumbs -->
        <nav class="flex items-center gap-2 text-sm">
          <span class="text-muted-foreground">首页</span>
          <template v-for="(crumb, index) in breadcrumbs" :key="crumb.path">
            <span class="text-muted-foreground">/</span>
            <span
              :class="index === breadcrumbs.length - 1 ? 'text-foreground font-medium' : 'text-muted-foreground'"
            >
              {{ crumb.title }}
            </span>
          </template>
        </nav>

        <!-- User menu -->
        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <Button variant="ghost" class="flex items-center gap-2 px-2">
              <Avatar class="w-8 h-8">
                <AvatarImage :src="userStore.user?.avatar" />
                <AvatarFallback class="bg-gflow-primary text-white text-xs">
                  {{ getInitials(userStore.username || 'U') }}
                </AvatarFallback>
              </Avatar>
              <span class="hidden md:inline">{{ userStore.username }}</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" class="w-48">
            <DropdownMenuLabel>
              <div class="flex flex-col">
                <span>{{ userStore.username }}</span>
                <span class="text-xs text-muted-foreground font-normal">
                  {{ userStore.user?.email }}
                </span>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem>
              <User class="w-4 h-4 mr-2" />
              个人信息
            </DropdownMenuItem>
            <DropdownMenuItem>
              <Settings class="w-4 h-4 mr-2" />
              系统设置
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem @click="handleLogout" class="text-destructive">
              <LogOut class="w-4 h-4 mr-2" />
              退出登录
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </header>

      <!-- Page content -->
      <main class="flex-1 p-6 overflow-auto bg-muted/30">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
