import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: '/dashboard',
      },
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表盘' },
      },
      {
        path: 'definitions',
        name: 'DefinitionList',
        component: () => import('@/views/DefinitionList.vue'),
        meta: { title: '流程定义' },
      },
      {
        path: 'instances',
        name: 'ProcessList',
        component: () => import('@/views/ProcessList.vue'),
        meta: { title: '流程实例' },
      },
      {
        path: 'designer',
        name: 'Designer',
        component: () => import('@/views/Designer.vue'),
        meta: { title: '流程设计器' },
      },
      {
        path: 'designer/:id',
        name: 'DesignerEdit',
        component: () => import('@/views/Designer.vue'),
        meta: { title: '编辑流程' },
      },
      {
        path: 'runners',
        name: 'RunnerList',
        component: () => import('@/views/RunnerList.vue'),
        meta: { title: 'Runner 管理' },
      },
      {
        path: 'users',
        name: 'UserList',
        component: () => import('@/views/UserList.vue'),
        meta: { title: '用户管理' },
      },
      {
        path: 'audit-logs',
        name: 'AuditLogs',
        component: () => import('@/views/AuditLogs.vue'),
        meta: { title: '审计日志' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Navigation guard
router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  
  if (to.meta.requiresAuth !== false && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
