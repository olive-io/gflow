# GFlow 前后端联调规范

## 概述

本文档描述 GFlow 后台管理系统前后端联调的详细规范和实施步骤，指导开发者进行高效的接口联调工作。

## 1. 环境准备

### 1.1 后端服务

```bash
# 进入输出目录
cd _output

# 启动后端服务
./gflow_server --config gflow.toml

# 服务地址
# HTTP: http://localhost:6550
# gRPC: localhost:6551 (如果配置了)
```

### 1.2 前端服务

```bash
# 进入前端目录
cd console

# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev

# 服务地址
# http://localhost:5173
```

### 1.3 环境验证

```bash
# 验证后端服务
curl http://localhost:6550/v1/ping

# 验证登录接口
curl -X POST http://localhost:6550/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"p@ssw0rd"}'
```

## 2. 接口规范

### 2.1 请求格式

**基础 URL**：
- 开发环境：通过 Vite 代理，`/v1` 会被代理到 `http://localhost:6550`
- 生产环境：配置 `VITE_API_BASE_URL` 环境变量

**请求头**：
```
Content-Type: application/json
Authorization: Bearer <token>  # 需要认证的接口
```

**分页参数**：
```
page: 页码，从 1 开始
size: 每页数量，必须大于 0
```

### 2.2 响应格式

**成功响应**：
```json
{
  "processes": [...],
  "total": "21"
}
```

**错误响应**：
```json
{
  "code": 3,
  "message": "invalid password",
  "details": []
}
```

### 2.3 认证机制

1. 调用 `/v1/login` 获取 JWT Token
2. Token 存储在 `localStorage.setItem('token', token)`
3. 后续请求在 Header 中携带 `Authorization: Bearer <token>`
4. Token 过期时返回 401，前端自动跳转登录页

## 3. API 封装规范

### 3.1 目录结构

```
console/src/
├── api/
│   ├── index.ts          # 导出入口
│   ├── auth.ts           # 认证 API
│   ├── definitions.ts    # 流程定义 API
│   ├── processes.ts      # 流程实例 API
│   ├── runners.ts        # Runner API
│   ├── users.ts          # 用户管理 API
│   └── ...
├── types/
│   └── api.ts            # 类型定义
└── lib/
    └── http.ts           # HTTP 客户端
```

### 3.2 HTTP 客户端

```typescript
// src/lib/http.ts
class HttpClient {
  private getHeaders(): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    }
    const token = localStorage.getItem('token')
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }
    return headers
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      if (response.status === 401) {
        localStorage.removeItem('token')
        window.location.href = '/login'
        throw new Error('Unauthorized')
      }
      const error = await response.text()
      throw new Error(error || `HTTP Error: ${response.status}`)
    }
    const text = await response.text()
    return text ? JSON.parse(text) : {}
  }

  async get<T>(url: string, params?: Record<string, unknown>): Promise<T> {
    const fullURL = new URL(url, window.location.origin)
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          fullURL.searchParams.append(key, String(value))
        }
      })
    }
    const response = await fetch(fullURL.toString(), {
      method: 'GET',
      headers: this.getHeaders(),
    })
    return this.handleResponse<T>(response)
  }

  async post<T>(url: string, data?: unknown): Promise<T> {
    const response = await fetch(url, {
      method: 'POST',
      headers: this.getHeaders(),
      body: data ? JSON.stringify(data) : undefined,
    })
    return this.handleResponse<T>(response)
  }
}

export const http = new HttpClient()
```

### 3.3 API 模块示例

```typescript
// src/api/processes.ts
import { http } from '@/lib/http'
import type { Process } from '@/types/api'

export interface ListProcessParams {
  page?: number
  size?: number
  definitions_uid?: string
  process_status?: number
}

export interface ListProcessResponse {
  processes: Process[]
  total: number
}

export const processesApi = {
  list: async (params?: ListProcessParams): Promise<ListProcessResponse> => {
    return http.get<ListProcessResponse>('/v1/processes', params)
  },

  get: async (id: number): Promise<{ process: Process }> => {
    return http.get<{ process: Process }>(`/v1/processes/${id}`)
  },

  execute: async (data: ExecuteProcessRequest): Promise<{ process: Process }> => {
    return http.post<{ process: Process }>('/v1/processes/execute', data)
  },
}
```

### 3.4 类型定义规范

```typescript
// src/types/api.ts

// 使用后端 proto 定义的字段名（snake_case）
export interface Process {
  id: number
  name: string
  uid: string
  definitions_uid: string
  definitions_version: number
  status: ProcessStatus
  stage: ProcessStage
  start_at: number      // 秒级时间戳
  end_at: number        // 秒级时间戳
  err_msg: string
  create_at: number
  update_at: number
}

// 枚举值与后端 proto 保持一致
export enum ProcessStatus {
  UnknownStatus = 0,
  Waiting = 1,
  Running = 2,
  Success = 3,
  Warn = 4,
  Failed = 5,
}
```

## 4. 联调流程

### 4.1 新增接口联调步骤

#### 步骤 1：查阅后端 Proto 定义

```bash
# 查看接口定义
cat api/rpc/bpmn.proto

# 查看类型定义
cat api/types/bpmn.proto
```

#### 步骤 2：添加类型定义

```typescript
// src/types/api.ts
export interface NewType {
  id: number
  name: string
  // ...
}
```

#### 步骤 3：创建 API 模块

```typescript
// src/api/newmodule.ts
import { http } from '@/lib/http'
import type { NewType } from '@/types/api'

export const newModuleApi = {
  list: async () => {
    return http.get<NewType[]>('/v1/newmodule')
  },
}
```

#### 步骤 4：导出 API

```typescript
// src/api/index.ts
export * from './newmodule'
```

#### 步骤 5：在页面中使用

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { newModuleApi } from '@/api'
import type { NewType } from '@/types/api'

const items = ref<NewType[]>([])
const loading = ref(false)

async function fetchData() {
  loading.value = true
  try {
    const response = await newModuleApi.list()
    items.value = response
  } catch (error) {
    console.error('Failed to fetch:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>
```

#### 步骤 6：测试验证

```bash
# 1. 确保后端服务运行
curl http://localhost:6550/v1/ping

# 2. 测试接口
TOKEN=$(curl -s -X POST http://localhost:6550/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"p@ssw0rd"}' \
  | grep -o '"text":"[^"]*"' | cut -d'"' -f4)

curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:6550/v1/newmodule

# 3. 在浏览器中测试前端页面
# 打开 http://localhost:5173
```

### 4.2 调试技巧

#### 使用浏览器开发者工具

1. 打开 Network 面板
2. 筛选 XHR 请求
3. 查看请求头、请求体、响应体
4. 检查 HTTP 状态码

#### 使用 Vue DevTools

1. 安装 Vue DevTools 浏览器扩展
2. 查看组件状态
3. 查看 Pinia Store 状态

#### 日志调试

```typescript
// 在 API 调用处添加日志
async function fetchData() {
  console.log('[API] Fetching data...')
  try {
    const response = await api.list()
    console.log('[API] Response:', response)
  } catch (error) {
    console.error('[API] Error:', error)
  }
}
```

## 5. 常见问题处理

### 5.1 CORS 跨域问题

**开发环境**：使用 Vite 代理

```typescript
// vite.config.ts
export default defineConfig({
  server: {
    proxy: {
      '/v1': {
        target: 'http://localhost:6550',
        changeOrigin: true,
      },
    },
  },
})
```

**生产环境**：后端配置 CORS 头

### 5.2 Token 过期处理

```typescript
// src/lib/http.ts
private async handleResponse<T>(response: Response): Promise<T> {
  if (response.status === 401) {
    // 清除本地存储
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    // 跳转登录页
    window.location.href = '/login'
    throw new Error('Unauthorized')
  }
  // ...
}
```

### 5.3 数字类型转换

后端返回的数字可能是字符串格式：

```typescript
// 转换示例
const total = Number(response.total)
const id = Number(response.id)
```

### 5.4 时间戳转换

后端时间戳为秒级，前端需要转换：

```typescript
function formatDateTime(timestamp: number): string {
  if (!timestamp) return '-'
  // 秒级转毫秒级
  return new Date(timestamp * 1000).toLocaleString('zh-CN')
}
```

### 5.5 分页参数错误

```
错误：invalid ListProcessRequest.Size: value must be greater than 0

解决：确保 page 和 size 参数都大于 0
```

## 6. 测试清单

### 6.1 接口测试

- [ ] 接口返回正确的数据结构
- [ ] 分页参数正确传递
- [ ] 筛选参数正确传递
- [ ] 错误响应正确处理

### 6.2 认证测试

- [ ] 未登录访问受保护接口返回 401
- [ ] Token 正确携带在请求头
- [ ] Token 过期自动跳转登录页

### 6.3 UI 测试

- [ ] 加载状态正确显示
- [ ] 错误状态正确显示
- [ ] 空数据状态正确显示
- [ ] 数据正确渲染

## 7. 相关资源

- 后端 Proto 定义：`api/rpc/*.proto`
- 后端类型定义：`api/types/*.proto`
- OpenAPI 文档：`http://localhost:6550/openapi.yaml`
- 前端 API 封装：`console/src/api/`
- 前端类型定义：`console/src/types/api.ts`
