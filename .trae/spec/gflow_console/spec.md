# GFlow 后台管理页面开发规范

## 1. 项目概述

### 1.1 项目背景

GFlow 是一个基于 BPMN 2.0 规范的工作流引擎系统，目前提供了完整的 REST API 和 Swagger UI，但缺少一个友好的可视化后台管理界面。本项目旨在开发一个功能完善的后台管理系统（GFlow Console）。

### 1.2 技术栈

| 类别 | 技术选型 |
|------|----------|
| 前端框架 | Vue 3+ |
| 构建工具 | Vite |
| 语言 | TypeScript |
| UI 组件库 | shadcn-vue |
| 样式方案 | Tailwind CSS |
| 状态管理 | Pinia |
| 路由管理 | Vue Router |
| 图表库 | ECharts / Recharts |
| 流程设计器 | bpmn-js |
| HTTP 客户端 | Axios |
| 表单验证 | VeeValidate + Zod |
| 代码规范 | ESLint + Prettier |

### 1.3 系统架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        GFlow Console                                     │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                      前端应用层                                   │    │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐  │    │
│  │  │ 仪表盘  │ │流程管理 │ │系统管理 │ │安全管理 │ │个人中心 │  │    │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘  │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                              │                                          │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                      基础设施层                                   │    │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐  │    │
│  │  │ API封装 │ │状态管理 │ │路由管理 │ │权限控制 │ │工具函数 │  │    │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘  │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                              │                                          │
│                              ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                      GFlow Server REST API                        │    │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐               │    │
│  │  │ BpmnRPC │ │SystemRPC│ │AdminRPC │ │ AuthRPC │               │    │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘               │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
```

### 1.4 shadcn-vue 简介

shadcn-vue 是一套基于 Radix Vue 和 Tailwind CSS 构建的可复制粘贴组件集合。它不是一个传统的组件库，而是一组可重用的组件代码，开发者可以完全控制和自定义。

**核心特点**：
- **可复制粘贴**：组件代码直接复制到项目中，完全可控
- **基于 Radix Vue**：提供无障碍访问的底层原语
- **Tailwind CSS**：使用 Tailwind 进行样式设计
- **类型安全**：完整的 TypeScript 支持
- **主题定制**：支持深色模式和自定义主题

**常用组件**：
- Button, Input, Select, Checkbox, Radio
- Dialog, Sheet, Popover, DropdownMenu
- Table, Card, Tabs, Accordion
- Form, Label, Toast, Alert
- Command, Data Table 等

## 2. 功能模块详细规范

### 2.1 认证模块

#### 2.1.1 登录页面

**功能描述**：用户通过用户名和密码登录系统

**API 接口**：
- `POST /v1/login` - 用户登录
- `GET /v1/users/self` - 获取当前用户信息

**页面元素**：
- 用户名输入框（使用 Input 组件）
- 密码输入框（使用 Input 组件）
- 记住我复选框（使用 Checkbox 组件）
- 登录按钮（使用 Button 组件）
- 错误提示区域（使用 Alert 组件）

**交互逻辑**：
1. 用户输入用户名和密码
2. 点击登录按钮
3. 调用登录 API
4. 成功：存储 Token，跳转到首页
5. 失败：显示错误提示

#### 2.1.2 Token 管理

**存储方式**：localStorage

**Token 结构**：
```typescript
interface Token {
  id: number;
  text: string;      // JWT Token 字符串
  expire_at: number; // 过期时间戳
  enable: number;    // 是否有效
  user_id: number;
  role_id: number;
}
```

**刷新机制**：
- Token 过期前 5 分钟自动刷新
- 请求返回 401 时跳转登录页

### 2.2 仪表盘模块

#### 2.2.1 系统概览

**功能描述**：展示系统整体运行状态

**API 接口**：
- `GET /v1/ping` - 系统健康检查
- `GET /v1/processes` - 获取流程统计
- `GET /v1/runners` - 获取 Runner 状态

**页面元素**：
- 系统状态卡片（使用 Card 组件）
- 流程统计卡片（总数、运行中、成功、失败）
- Runner 状态卡片（在线、离线）
- 今日执行量卡片

#### 2.2.2 统计图表

**功能描述**：可视化展示系统数据

**图表类型**：
1. **流程执行趋势图**（折线图）
   - X轴：时间（按天/小时）
   - Y轴：执行数量
   - 数据来源：Process 列表聚合

2. **流程状态分布图**（饼图）
   - 成功占比
   - 失败占比
   - 运行中占比
   - 数据来源：Process 状态统计

3. **Runner 负载图**（柱状图）
   - X轴：Runner 名称
   - Y轴：任务执行量
   - 数据来源：Runner 统计信息

### 2.3 流程管理模块

#### 2.3.1 流程定义列表

**功能描述**：展示所有已部署的流程定义

**API 接口**：
- `GET /v1/definitions?page=1&size=10` - 分页获取流程定义列表

**页面元素**：
- 搜索栏（使用 Input 组件）
- 流程定义卡片列表（使用 Card 组件）
- 分页组件
- 部署流程按钮（使用 Button 组件）

**数据结构**：
```typescript
interface Definitions {
  id: number;
  name: string;
  uid: string;
  description: string;
  metadata: Record<string, string>;
  content: string;      // BPMN XML
  version: number;
  isExecute: boolean;
}
```

#### 2.3.2 流程定义详情

**功能描述**：查看流程定义的详细信息

**API 接口**：
- `GET /v1/definitions/{uid}` - 获取流程定义详情
- `GET /v1/definitions/{uid}?version=N` - 获取指定版本

**页面元素**：
- 基本信息区域（使用 Card 组件）
- BPMN 图形预览区域
- 版本历史列表
- 操作按钮（使用 Button 组件）

#### 2.3.3 部署流程

**功能描述**：上传 BPMN 文件部署新流程

**API 接口**：
- `POST /v1/definitions` - 部署流程定义

**请求参数**：
```typescript
interface DeployDefinitionsRequest {
  metadata: Record<string, string>;
  content: Uint8Array;    // BPMN XML 字节
  description: string;
}
```

**页面元素**：
- 文件上传组件
- 元数据配置表单（使用 Form 组件）
- 描述输入框
- 部署按钮

#### 2.3.4 流程实例列表

**功能描述**：展示流程实例执行情况

**API 接口**：
- `GET /v1/processes` - 分页获取流程实例列表

**查询参数**：
```typescript
interface ListProcessRequest {
  page: number;
  size: number;
  definitions_uid?: string;
  definitions_version?: number;
  process_status?: number;  // ProcessStatus 枚举
  process_stage?: number;   // ProcessStage 枚举
}
```

**页面元素**：
- 筛选栏（使用 Select 组件）
- 流程实例表格（使用 DataTable 组件）
- 分页组件
- 执行流程按钮

#### 2.3.5 流程实例详情

**功能描述**：查看流程实例的执行详情

**API 接口**：
- `GET /v1/processes/{id}` - 获取流程实例详情

**页面元素**：
- 基本信息区域
- 流程执行图（高亮当前节点）
- 节点执行列表
- 输入输出数据展示
- 错误信息展示

#### 2.3.6 执行流程

**功能描述**：手动触发流程执行

**API 接口**：
- `POST /v1/processes/execute` - 执行流程

**请求参数**：
```typescript
interface ExecuteProcessRequest {
  name: string;
  definitions_uid: string;
  definitions_version?: number;
  priority?: number;
  mode?: TransitionMode;
  headers?: Record<string, string>;
  properties?: Record<string, Value>;
  dataObjects?: Record<string, Value>;
}
```

**页面元素**：
- 流程选择下拉框（使用 Select 组件）
- 执行参数配置表单（使用 Form 组件）
- 输入数据配置
- 执行按钮

#### 2.3.7 流程设计器

**功能描述**：可视化设计 BPMN 流程

**技术实现**：bpmn-js

**功能点**：
1. **工具栏**
   - 开始事件
   - 结束事件
   - 任务节点（ServiceTask、SendTask、ScriptTask 等）
   - 网关（Exclusive、Parallel、Inclusive）
   - 中间事件

2. **画布**
   - 拖拽添加节点
   - 连线操作
   - 缩放平移
   - 撤销重做

3. **属性面板**
   - 节点名称
   - 节点类型
   - 执行器配置
   - 输入输出映射
   - 超时重试设置

4. **操作功能**
   - 保存流程
   - 导入 BPMN XML
   - 导出 BPMN XML
   - 部署流程
   - 流程预览

### 2.4 系统管理模块

#### 2.4.1 Runner 管理

**功能描述**：管理执行器 Runner

**API 接口**：
- `GET /v1/runners` - 获取 Runner 列表
- `GET /v1/runners/{id}` - 获取 Runner 详情
- `POST /v1/runners/register` - 注册 Runner
- `POST /v1/runners/disregister` - 注销 Runner

**数据结构**：
```typescript
interface Runner {
  id: number;
  uid: string;
  listen_url: string;
  version: string;
  heartbeat_ms: number;
  hostname: string;
  metadata: Record<string, string>;
  features: Record<string, string>;
  transport: TransportType;
  cpu: number;
  memory: number;
  online_timestamp: number;
  offline_timestamp: number;
  online: number;
  state: State;
}
```

**页面元素**：
- Runner 列表表格（使用 DataTable 组件）
- Runner 详情弹窗（使用 Sheet 组件）
- 注册 Runner 表单（使用 Dialog 组件）
- 心跳状态指示器（使用 Badge 组件）

#### 2.4.2 端点管理

**功能描述**：管理任务端点

**API 接口**：
- `GET /v1/endpoints` - 获取端点列表
- `POST /v1/endpoints` - 添加端点
- `POST /v1/endpoints/openapi` - 从 OpenAPI 导入
- `POST /v1/endpoints/grpc` - 从 gRPC 导入
- `POST /v1/endpoints/{id}/convert` - 转换端点格式

**数据结构**：
```typescript
interface Endpoint {
  id: number;
  task_type: FlowNodeType;
  type: string;
  name: string;
  description: string;
  mode: TransitionMode;
  http_url: string;
  metadata: Record<string, string>;
  targets: string[];
  headers: Record<string, string>;
  properties: Record<string, Value>;
  dataObjects: Record<string, Value>;
  results: Record<string, Value>;
}
```

**页面元素**：
- 端点列表表格（使用 DataTable 组件）
- 添加端点表单（使用 Dialog 组件）
- OpenAPI 导入组件
- gRPC 导入组件
- 端点详情弹窗

#### 2.4.3 路由管理

**功能描述**：管理路由规则

**API 接口**：
- `GET /v1/admin/routes` - 获取路由列表
- `GET /v1/admin/routes/{id}` - 获取路由详情
- `POST /v1/admin/routes` - 创建路由
- `POST /v1/admin/routes/{id}` - 删除路由

**页面元素**：
- 路由列表表格（使用 DataTable 组件）
- 创建路由表单（使用 Dialog 组件）
- 路由详情弹窗

### 2.5 安全管理模块

#### 2.5.1 用户管理

**功能描述**：管理系统用户

**API 接口**：
- `GET /v1/admin/users` - 获取用户列表
- `GET /v1/admin/users/{id}` - 获取用户详情
- `POST /v1/admin/users` - 创建用户
- `PATCH /v1/admin/users/{id}` - 更新用户
- `DELETE /v1/admin/users/{id}` - 删除用户

**数据结构**：
```typescript
interface User {
  id: number;
  uid: string;
  username: string;
  email: string;
  description: string;
  metadata: Record<string, string>;
  role_id: number;
}
```

**页面元素**：
- 用户列表表格（使用 DataTable 组件）
- 创建用户表单（使用 Dialog 组件）
- 编辑用户表单
- 重置密码功能
- 删除确认弹窗（使用 AlertDialog 组件）

#### 2.5.2 角色管理

**功能描述**：管理系统角色

**API 接口**：
- `GET /v1/admin/roles` - 获取角色列表
- `GET /v1/admin/roles/{id}` - 获取角色详情
- `POST /v1/admin/roles` - 创建角色
- `PATCH /v1/admin/roles/{id}` - 更新角色
- `DELETE /v1/admin/roles/{id}` - 删除角色

**数据结构**：
```typescript
enum RoleType {
  Unknown = 0,
  Admin = 1,
  System = 2,
  Operator = 3,
}

interface Role {
  id: number;
  type: RoleType;
  name: string;
  display_name: string;
  description: string;
  metadata: Record<string, string>;
}
```

#### 2.5.3 权限管理

**功能描述**：管理系统权限策略

**API 接口**：
- `GET /v1/admin/interfaces` - 获取系统接口列表
- `GET /v1/admin/policies` - 获取权限策略列表
- `POST /v1/admin/policies` - 添加权限策略
- `DELETE /v1/admin/policy` - 删除权限策略

**数据结构**：
```typescript
interface Policy {
  subject: string;  // 主体（用户/角色）
  object: string;   // 客体（资源）
  action: string;   // 动作
}
```

**页面元素**：
- 接口列表展示（使用 DataTable 组件）
- 权限策略列表
- 添加策略表单
- 角色授权界面

#### 2.5.4 Token 管理

**功能描述**：管理用户 Token

**API 接口**：
- `GET /v1/admin/tokens` - 获取 Token 列表
- `GET /v1/admin/tokens/{id}` - 获取 Token 详情
- `DELETE /v1/admin/tokens/{id}` - 删除 Token

**页面元素**：
- Token 列表表格（使用 DataTable 组件）
- Token 详情弹窗
- 注销 Token 功能

### 2.6 个人中心模块

#### 2.6.1 个人信息

**功能描述**：查看和编辑个人信息

**API 接口**：
- `GET /v1/users/self` - 获取当前用户信息
- `PATCH /v1/users/self` - 更新个人信息

**页面元素**：
- 个人信息展示（使用 Card 组件）
- 编辑表单（使用 Form 组件）
- 保存按钮

#### 2.6.2 修改密码

**功能描述**：修改当前用户密码

**API 接口**：
- `PATCH /v1/users/self` - 更新密码

**页面元素**：
- 旧密码输入框
- 新密码输入框
- 确认密码输入框
- 提交按钮

## 3. 前端架构设计

### 3.1 目录结构

```
gflow-console/
├── public/
│   └── favicon.ico
├── src/
│   ├── api/                    # API 接口封装
│   │   ├── auth.ts            # 认证相关接口
│   │   ├── definitions.ts     # 流程定义接口
│   │   ├── processes.ts       # 流程实例接口
│   │   ├── runners.ts         # Runner 接口
│   │   ├── endpoints.ts       # 端点接口
│   │   ├── users.ts           # 用户接口
│   │   ├── roles.ts           # 角色接口
│   │   ├── policies.ts        # 权限接口
│   │   └── index.ts           # 统一导出
│   ├── components/             # 通用组件
│   │   ├── ui/                # shadcn-vue 组件
│   │   │   ├── button/
│   │   │   ├── input/
│   │   │   ├── select/
│   │   │   ├── dialog/
│   │   │   ├── table/
│   │   │   └── ...
│   │   ├── layout/            # 布局组件
│   │   │   ├── AppLayout.vue
│   │   │   ├── AppHeader.vue
│   │   │   ├── AppSidebar.vue
│   │   │   └── AppBreadcrumb.vue
│   │   ├── common/            # 业务通用组件
│   │   │   ├── StatusTag.vue
│   │   │   ├── SearchBar.vue
│   │   │   └── Pagination.vue
│   │   └── bpmn/              # BPMN 相关组件
│   │       ├── BpmnDesigner.vue
│   │       ├── BpmnViewer.vue
│   │       └── PropertiesPanel.vue
│   ├── views/                  # 页面组件
│   │   ├── login/             # 登录页
│   │   │   └── index.vue
│   │   ├── dashboard/         # 仪表盘
│   │   │   └── index.vue
│   │   ├── definitions/       # 流程定义
│   │   │   ├── index.vue
│   │   │   └── [uid].vue
│   │   ├── processes/         # 流程实例
│   │   │   ├── index.vue
│   │   │   └── [id].vue
│   │   ├── designer/          # 流程设计器
│   │   │   └── index.vue
│   │   ├── runners/           # Runner 管理
│   │   │   └── index.vue
│   │   ├── endpoints/         # 端点管理
│   │   │   └── index.vue
│   │   ├── routes/            # 路由管理
│   │   │   └── index.vue
│   │   ├── users/             # 用户管理
│   │   │   └── index.vue
│   │   ├── roles/             # 角色管理
│   │   │   └── index.vue
│   │   ├── policies/          # 权限管理
│   │   │   └── index.vue
│   │   ├── tokens/            # Token 管理
│   │   │   └── index.vue
│   │   └── profile/           # 个人中心
│   │       └── index.vue
│   ├── stores/                 # 状态管理
│   │   ├── auth.ts            # 认证状态
│   │   ├── user.ts            # 用户状态
│   │   └── app.ts             # 应用状态
│   ├── composables/            # 组合式函数
│   │   ├── useAuth.ts         # 认证逻辑
│   │   ├── usePermission.ts   # 权限逻辑
│   │   └── useRequest.ts      # 请求逻辑
│   ├── lib/                    # 工具库
│   │   ├── request.ts         # HTTP 请求封装
│   │   ├── storage.ts         # 存储工具
│   │   ├── format.ts          # 格式化工具
│   │   └── utils.ts           # 通用工具
│   ├── types/                  # 类型定义
│   │   ├── api.ts             # API 类型
│   │   ├── models.ts          # 数据模型
│   │   └── global.d.ts        # 全局类型
│   ├── router/                 # 路由配置
│   │   └── index.ts           # 路由定义
│   ├── styles/                 # 全局样式
│   │   └── globals.css        # 全局 CSS
│   ├── App.vue                 # 根组件
│   └── main.ts                # 入口文件
├── components.json             # shadcn-vue 配置
├── tailwind.config.js          # Tailwind 配置
├── tsconfig.json
├── vite.config.ts
├── package.json
└── .eslintrc.cjs
```

### 3.2 状态管理设计

```typescript
// stores/auth.ts
import { defineStore } from 'pinia';

interface AuthState {
  token: Token | null;
  isAuthenticated: boolean;
  loading: boolean;
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: null,
    isAuthenticated: false,
    loading: false,
  }),
  
  getters: {
    getToken: (state) => state.token,
    isLoggedIn: (state) => state.isAuthenticated,
  },
  
  actions: {
    async login(username: string, password: string) {
      // 登录逻辑
    },
    logout() {
      this.token = null;
      this.isAuthenticated = false;
    },
  },
});
```

```typescript
// stores/user.ts
import { defineStore } from 'pinia';

interface UserState {
  currentUser: User | null;
  permissions: Policy[];
  roles: Role[];
}

export const useUserStore = defineStore('user', {
  state: (): UserState => ({
    currentUser: null,
    permissions: [],
    roles: [],
  }),
  
  actions: {
    async fetchCurrentUser() {
      // 获取当前用户信息
    },
  },
});
```

```typescript
// stores/app.ts
import { defineStore } from 'pinia';

interface AppState {
  sidebarCollapsed: boolean;
  theme: 'light' | 'dark' | 'system';
  locale: string;
}

export const useAppStore = defineStore('app', {
  state: (): AppState => ({
    sidebarCollapsed: false,
    theme: 'system',
    locale: 'zh-CN',
  }),
  
  actions: {
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed;
    },
    setTheme(theme: 'light' | 'dark' | 'system') {
      this.theme = theme;
    },
  },
});
```

### 3.3 路由配置

```typescript
// router/index.ts
import { createRouter, createWebHistory } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { public: true },
  },
  {
    path: '/',
    component: () => import('@/components/layout/AppLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/index.vue'),
      },
      {
        path: 'definitions',
        name: 'Definitions',
        children: [
          {
            path: '',
            name: 'DefinitionList',
            component: () => import('@/views/definitions/index.vue'),
          },
          {
            path: ':uid',
            name: 'DefinitionDetail',
            component: () => import('@/views/definitions/[uid].vue'),
          },
        ],
      },
      {
        path: 'processes',
        name: 'Processes',
        children: [
          {
            path: '',
            name: 'ProcessList',
            component: () => import('@/views/processes/index.vue'),
          },
          {
            path: ':id',
            name: 'ProcessDetail',
            component: () => import('@/views/processes/[id].vue'),
          },
        ],
      },
      {
        path: 'designer',
        name: 'Designer',
        component: () => import('@/views/designer/index.vue'),
      },
      {
        path: 'runners',
        name: 'Runners',
        component: () => import('@/views/runners/index.vue'),
      },
      {
        path: 'endpoints',
        name: 'Endpoints',
        component: () => import('@/views/endpoints/index.vue'),
      },
      {
        path: 'routes',
        name: 'Routes',
        component: () => import('@/views/routes/index.vue'),
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/users/index.vue'),
      },
      {
        path: 'roles',
        name: 'Roles',
        component: () => import('@/views/roles/index.vue'),
      },
      {
        path: 'policies',
        name: 'Policies',
        component: () => import('@/views/policies/index.vue'),
      },
      {
        path: 'tokens',
        name: 'Tokens',
        component: () => import('@/views/tokens/index.vue'),
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/profile/index.vue'),
      },
    ],
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

// 路由守卫
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore();
  
  if (!to.meta.public && !authStore.isAuthenticated) {
    next({ name: 'Login' });
  } else {
    next();
  }
});

export default router;
```

### 3.4 API 请求封装

```typescript
// lib/request.ts
import axios from 'axios';
import type { AxiosInstance, AxiosRequestConfig } from 'axios';

const request: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000,
});

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
request.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      // 跳转登录页
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default request;
```

### 3.5 shadcn-vue 配置

```json
// components.json
{
  "$schema": "https://shadcn-vue.com/schema.json",
  "style": "default",
  "typescript": true,
  "tsConfigPath": "./tsconfig.json",
  "tailwind": {
    "config": "tailwind.config.js",
    "css": "src/styles/globals.css",
    "baseColor": "slate",
    "cssVariables": true
  },
  "framework": "vite",
  "aliases": {
    "components": "@/components",
    "utils": "@/lib/utils"
  }
}
```

### 3.6 Tailwind CSS 配置

```javascript
// tailwind.config.js
/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: ['class'],
  content: [
    './index.html',
    './src/**/*.{vue,js,ts,jsx,tsx}',
  ],
  theme: {
    container: {
      center: true,
      padding: '2rem',
      screens: {
        '2xl': '1400px',
      },
    },
    extend: {
      colors: {
        border: 'hsl(var(--border))',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        primary: {
          DEFAULT: 'hsl(var(--primary))',
          foreground: 'hsl(var(--primary-foreground))',
        },
        secondary: {
          DEFAULT: 'hsl(var(--secondary))',
          foreground: 'hsl(var(--secondary-foreground))',
        },
        destructive: {
          DEFAULT: 'hsl(var(--destructive))',
          foreground: 'hsl(var(--destructive-foreground))',
        },
        muted: {
          DEFAULT: 'hsl(var(--muted))',
          foreground: 'hsl(var(--muted-foreground))',
        },
        accent: {
          DEFAULT: 'hsl(var(--accent))',
          foreground: 'hsl(var(--accent-foreground))',
        },
        popover: {
          DEFAULT: 'hsl(var(--popover))',
          foreground: 'hsl(var(--popover-foreground))',
        },
        card: {
          DEFAULT: 'hsl(var(--card))',
          foreground: 'hsl(var(--card-foreground))',
        },
      },
      borderRadius: {
        lg: 'var(--radius)',
        md: 'calc(var(--radius) - 2px)',
        sm: 'calc(var(--radius) - 4px)',
      },
    },
  },
  plugins: [require('tailwindcss-animate')],
};
```

## 4. 开发规范

### 4.0 包管理器规范

**使用 pnpm 作为包管理器**

本项目统一使用 pnpm 进行依赖管理，不使用 npm 或 yarn。

**常用命令对照**：

| 操作 | npm | pnpm |
|------|-----|------|
| 安装依赖 | npm install | pnpm install |
| 添加依赖 | npm install package | pnpm add package |
| 添加开发依赖 | npm install -D package | pnpm add -D package |
| 运行脚本 | npm run dev | pnpm dev |
| 构建项目 | npm run build | pnpm build |
| 删除依赖 | npm uninstall package | pnpm remove package |

**pnpm 配置**：

```toml
# .npmrc
shamefully-hoist=true
strict-peer-dependencies=false
auto-install-peers=true
```

**优势**：
- 更快的安装速度
- 节省磁盘空间（符号链接共享依赖）
- 更严格的依赖管理
- 支持 monorepo

### 4.1 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| Vue 组件文件 | PascalCase | `UserList.vue` |
| 组合式函数 | camelCase + use 前缀 | `useAuth.ts` |
| 样式文件 | camelCase | `userList.css` |
| 常量 | UPPER_SNAKE_CASE | `API_BASE_URL` |
| 接口类型 | PascalCase | `User` |
| 枚举 | PascalCase | `ProcessStatus` |
| Pinia Store | camelCase + use 前缀 | `useAuthStore` |

### 4.2 组件规范

```vue
<!-- UserList.vue -->
<script setup lang="ts">
import { ref, onMounted } from 'vue';
import type { User } from '@/types/models';

// Props
interface Props {
  pageSize?: number;
}

const props = withDefaults(defineProps<Props>(), {
  pageSize: 10,
});

// Emits
const emit = defineEmits<{
  (e: 'pageChange', page: number): void;
}>();

// State
const users = ref<User[]>([]);
const loading = ref(false);

// Methods
const fetchUsers = async () => {
  loading.value = true;
  try {
    // 获取用户列表
  } finally {
    loading.value = false;
  }
};

const handlePageChange = (page: number) => {
  emit('pageChange', page);
};

// Lifecycle
onMounted(() => {
  fetchUsers();
});
</script>

<template>
  <div class="user-list">
    <DataTable :data="users" :loading="loading" />
  </div>
</template>
```

### 4.3 组合式函数规范

```typescript
// composables/useAuth.ts
import { computed } from 'vue';
import { useAuthStore } from '@/stores/auth';

export function useAuth() {
  const authStore = useAuthStore();

  const isAuthenticated = computed(() => authStore.isAuthenticated);
  const token = computed(() => authStore.token);

  const login = async (username: string, password: string) => {
    return authStore.login(username, password);
  };

  const logout = () => {
    authStore.logout();
  };

  return {
    isAuthenticated,
    token,
    login,
    logout,
  };
}
```

### 4.4 Git 提交规范

| 类型 | 说明 |
|------|------|
| feat | 新功能 |
| fix | 修复 Bug |
| docs | 文档更新 |
| style | 代码格式调整 |
| refactor | 重构 |
| test | 测试相关 |
| chore | 构建/工具相关 |

## 5. 部署配置

### 5.1 环境变量

```bash
# .env.development
VITE_API_BASE_URL=http://localhost:8080

# .env.production
VITE_API_BASE_URL=https://api.gflow.example.com
```

### 5.2 构建配置

```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import path from 'path';

export default defineConfig({
  plugins: [vue()],
  base: '/console/',
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      '/v1': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
});
```

## 6. shadcn-vue 组件使用指南

### 6.1 安装组件

```bash
# 安装 Button 组件
pnpm dlx shadcn-vue@latest add button

# 安装 Input 组件
pnpm dlx shadcn-vue@latest add input

# 安装 Dialog 组件
pnpm dlx shadcn-vue@latest add dialog

# 安装 Table 组件
pnpm dlx shadcn-vue@latest add table

# 安装 DataTable 组件（带功能的表格）
pnpm dlx shadcn-vue@latest add data-table
```

### 6.2 常用组件示例

```vue
<!-- 使用 Button -->
<template>
  <Button variant="default">默认按钮</Button>
  <Button variant="destructive">删除按钮</Button>
  <Button variant="outline">轮廓按钮</Button>
  <Button variant="ghost">幽灵按钮</Button>
</template>

<script setup lang="ts">
import { Button } from '@/components/ui/button';
</script>
```

```vue
<!-- 使用 Dialog -->
<template>
  <Dialog v-model:open="open">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>确认删除</DialogTitle>
        <DialogDescription>此操作不可撤销</DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <Button variant="outline" @click="open = false">取消</Button>
        <Button variant="destructive" @click="handleDelete">确认</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';

const open = ref(false);

const handleDelete = () => {
  // 删除逻辑
  open.value = false;
};
</script>
```

### 6.3 主题定制

```css
/* styles/globals.css */
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    --primary: 222.2 47.4% 11.2%;
    --primary-foreground: 210 40% 98%;
    /* ... 其他变量 */
  }

  .dark {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    --primary: 210 40% 98%;
    --primary-foreground: 222.2 47.4% 11.2%;
    /* ... 其他变量 */
  }
}
```
