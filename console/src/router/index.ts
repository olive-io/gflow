import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/designer',
    name: 'Designer',
    component: () => import('@/views/Designer.vue'),
    meta: { requiresAuth: true, title: '流程设计器' },
  },
  {
    path: '/designer/:id',
    name: 'DesignerEdit',
    component: () => import('@/views/Designer.vue'),
    meta: { requiresAuth: true, title: '编辑流程' },
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
        name: 'Definitions',
        component: () => import('@/views/Definitions.vue'),
        meta: { title: '流程定义' },
      },
      {
        path: 'definitions/:uid',
        name: 'DefinitionDetail',
        component: () => import('@/views/DefinitionDetail.vue'),
        meta: { title: '流程定义详情' },
      },
      {
        path: 'processes',
        name: 'Process',
        component: () => import('@/views/Process.vue'),
        meta: { title: '流程实例' },
      },
      {
        path: 'processes/:id',
        name: 'ProcessDetail',
        component: () => import('@/views/ProcessDetail.vue'),
        meta: { title: '流程实例详情' },
      },
      {
        path: 'processes/:id',
        name: 'ProcessDetail',
        component: () => import('@/views/ProcessDetail.vue'),
        meta: { title: '流程实例详情' },
      },
      {
        path: 'system',
        name: 'System',
        component: () => import('@/views/System.vue'),
        meta: { title: '组件管理' },
      },
      {
        path: 'endpoints',
        name: 'Endpoints',
        component: () => import('@/views/Endpoints.vue'),
        meta: { title: '接口管理' },
      },
      {
        path: 'endpoints/:id',
        name: 'EndpointDetail',
        component: () => import('@/views/EndpointDetail.vue'),
        meta: { title: '接口详情' },
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/Users.vue'),
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
