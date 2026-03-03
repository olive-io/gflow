# GFlow 前后端接口联调计划

## 概述

本文档描述了 GFlow 后台管理系统前后端接口联调的详细计划，包括后端 API 接口清单、前端 API 封装状态、联调优先级和实施步骤。

## 1. 后端 API 服务清单

### 后端 Swagger 接口文档

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | /openapi.yaml | 接口文档 |

### 1.1 AuthRPC - 认证服务

| 方法 | 路径 | 功能 | 前端状态 |
|------|------|------|----------|
| POST | /v1/login | 用户登录 | ✅ 已实现 |
| GET | /v1/users/self | 获取当前用户信息 | ✅ 已实现 |
| PATCH | /v1/users/self | 更新当前用户信息 | ✅ 已实现 |

### 1.2 BpmnRPC - BPMN 流程服务

| 方法 | 路径 | 功能 | 前端状态 |
|------|------|------|----------|
| POST | /v1/definitions | 部署流程定义 | ✅ 已实现 |
| GET | /v1/definitions | 列出流程定义 | ✅ 已实现 |
| GET | /v1/definitions/{uid} | 获取流程定义详情 | ✅ 已实现 |
| DELETE | /v1/definitions/{uid} | 删除流程定义 | ✅ 已实现 |
| POST | /v1/processes/execute | 执行流程实例 | ✅ 已实现 |
| GET | /v1/processes | 列出流程实例 | ✅ 已实现 |
| GET | /v1/processes/{id} | 获取流程实例详情 | ⏳ 待实现 |

### 1.3 SystemRPC - 系统服务

| 方法 | 路径 | 功能 | 前端状态 |
|------|------|------|----------|
| GET | /v1/ping | 健康检查 | ⏳ 待实现 |
| POST | /v1/runners/register | 注册 Runner | 🔜 不需要 |
| POST | /v1/runners/disregister | 注销 Runner | 🔜 不需要 |
| GET | /v1/runners | 列出 Runner | ✅ 已实现 |
| GET | /v1/runners/{id} | 获取 Runner 详情 | ✅ 已实现 |
| GET | /v1/endpoints | 列出 Endpoint | ⏳ 待实现 |
| POST | /v1/endpoints | 添加 Endpoint | ⏳ 待实现 |

### 1.4 AdminRPC - 管理服务

| 方法 | 路径 | 功能 | 前端状态 |
|------|------|------|----------|
| GET | /v1/admin/roles | 列出角色 | ⏳ 待实现 |
| POST | /v1/admin/roles | 创建角色 | ⏳ 待实现 |
| GET | /v1/admin/roles/{id} | 获取角色详情 | ⏳ 待实现 |
| PATCH | /v1/admin/roles/{id} | 更新角色 | ⏳ 待实现 |
| DELETE | /v1/admin/roles/{id} | 删除角色 | ⏳ 待实现 |
| GET | /v1/admin/users | 列出用户 | ⏳ 待实现 |
| POST | /v1/admin/users | 创建用户 | ⏳ 待实现 |
| GET | /v1/admin/users/{id} | 获取用户详情 | ⏳ 待实现 |
| PATCH | /v1/admin/users/{id} | 更新用户 | ⏳ 待实现 |
| DELETE | /v1/admin/users/{id} | 删除用户 | ⏳ 待实现 |
| GET | /v1/admin/tokens | 列出令牌 | ⏳ 待实现 |
| GET | /v1/admin/tokens/{id} | 获取令牌详情 | ⏳ 待实现 |
| DELETE | /v1/admin/tokens/{id} | 删除令牌 | ⏳ 待实现 |
| GET | /v1/admin/interfaces | 列出接口 | ⏳ 待实现 |
| GET | /v1/admin/policies | 列出策略 | ⏳ 待实现 |
| POST | /v1/admin/policies | 添加策略 | ⏳ 待实现 |
| DELETE | /v1/admin/policy | 删除策略 | ⏳ 待实现 |
| GET | /v1/admin/routes | 列出路由 | ⏳ 待实现 |
| GET | /v1/admin/routes/{id} | 获取路由详情 | ⏳ 待实现 |
| POST | /v1/admin/routes | 创建路由 | ⏳ 待实现 |
| POST | /v1/admin/routes/{id} | 删除路由 | ⏳ 待实现 |

## 2. 前端 API 文件结构

```
console/src/api/
├── index.ts           # API 导出入口
├── auth.ts            # 认证 API (✅ 已实现)
├── definitions.ts     # 流程定义 API (✅ 已实现)
├── processes.ts       # 流程实例 API (✅ 已实现)
├── runners.ts         # Runner API (✅ 已实现)
├── users.ts           # 用户管理 API (⏳ 待实现)
├── roles.ts           # 角色管理 API (⏳ 待实现)
├── tokens.ts          # 令牌管理 API (⏳ 待实现)
├── policies.ts        # 策略管理 API (⏳ 待实现)
├── endpoints.ts       # Endpoint API (⏳ 待实现)
└── system.ts          # 系统 API (⏳ 待实现)
```

## 3. 联调优先级

### P0 - 核心功能 (已完成)
- [x] 用户登录/登出
- [x] 获取当前用户信息
- [x] 流程定义列表
- [x] 流程实例列表
- [x] Runner 列表

### P1 - 基础管理功能
- [ ] 流程定义详情
- [ ] 流程实例详情
- [ ] 流程实例执行
- [ ] 用户管理 CRUD
- [ ] 流程设计器保存/部署

### P2 - 高级管理功能
- [ ] 角色管理 CRUD
- [ ] 令牌管理
- [ ] 策略管理
- [ ] Endpoint 管理

### P3 - 系统功能
- [ ] 健康检查
- [ ] 路由管理
- [ ] 接口列表

## 4. 实施计划

### 阶段一：核心功能联调 (1-2 天)

#### 4.1.1 认证模块
- [x] 登录接口联调
- [x] Token 存储和自动刷新
- [x] 用户信息获取
- [ ] 密码修改功能

#### 4.1.2 流程定义模块
- [x] 列表查询（分页、筛选）
- [ ] 详情获取
- [ ] 部署功能
- [x] 删除功能

#### 4.1.3 流程实例模块
- [x] 列表查询（分页、筛选）
- [ ] 详情获取（包含执行节点）
- [ ] 执行流程
- [ ] 停止/重试流程

### 阶段二：管理功能联调 (2-3 天)

#### 4.2.1 用户管理模块
- [ ] 用户列表
- [ ] 创建用户
- [ ] 编辑用户
- [ ] 删除用户
- [ ] 用户详情

#### 4.2.2 Runner 管理模块
- [x] Runner 列表
- [x] Runner 详情
- [ ] Endpoint 列表

#### 4.2.3 角色权限模块
- [ ] 角色列表
- [ ] 创建角色
- [ ] 编辑角色
- [ ] 删除角色
- [ ] 策略列表
- [ ] 添加/删除策略

### 阶段三：高级功能联调 (1-2 天)

#### 4.3.1 流程设计器
- [ ] BPMN 文件保存
- [ ] BPMN 文件加载
- [ ] 流程部署

#### 4.3.2 系统设置
- [ ] 令牌管理
- [ ] 路由管理
- [ ] 接口列表

## 5. 技术要点

### 5.1 请求封装

```typescript
// HTTP 客户端封装
class HttpClient {
  private baseURL: string
  
  // 自动添加 Authorization 头
  private getHeaders(): HeadersInit
  
  // 统一错误处理
  private async handleResponse<T>(response: Response): Promise<T>
  
  // HTTP 方法
  async get<T>(url: string, params?): Promise<T>
  async post<T>(url: string, data?): Promise<T>
  async put<T>(url: string, data?): Promise<T>
  async patch<T>(url: string, data?): Promise<T>
  async delete<T>(url: string): Promise<T>
}
```

### 5.2 响应格式

后端响应格式（JSON）：
```json
{
  "processes": [...],
  "total": "21"
}
```

注意：后端数字类型字段返回字符串格式，前端需要转换。

### 5.3 认证机制

- 使用 JWT Token 认证
- Token 存储在 localStorage
- 请求头添加 `Authorization: Bearer <token>`
- 401 响应自动跳转登录页

### 5.4 错误处理

```typescript
// 后端错误响应格式
{
  "code": 3,
  "message": "invalid password",
  "details": []
}

// 前端错误处理
try {
  const response = await api.call()
} catch (error) {
  // 显示错误提示
  showError(error.message)
}
```

## 6. 测试清单

### 6.1 认证测试
- [ ] 正确账号登录
- [ ] 错误密码登录
- [ ] Token 过期处理
- [ ] 登出功能

### 6.2 流程定义测试
- [ ] 列表分页
- [ ] 搜索筛选
- [ ] 详情查看
- [ ] 部署流程
- [ ] 删除流程

### 6.3 流程实例测试
- [ ] 列表分页
- [ ] 状态筛选
- [ ] 详情查看
- [ ] 执行流程
- [ ] 停止流程

### 6.4 用户管理测试
- [ ] 列表分页
- [ ] 创建用户
- [ ] 编辑用户
- [ ] 删除用户
- [ ] 权限验证

## 7. 注意事项

1. **字段命名差异**：后端使用 snake_case，前端 TypeScript 使用 camelCase，需要在类型定义中注意映射。

2. **数字类型**：后端返回的数字可能是字符串格式，前端需要 `Number()` 转换。

3. **时间戳**：后端时间戳为秒级，前端需要乘以 1000 转换为毫秒。

4. **分页参数**：后端分页从 1 开始，page 和 size 参数必须大于 0。

5. **CORS 配置**：开发环境通过 Vite 代理解决跨域问题。

## 8. 相关文件

- 后端 Proto 定义：`api/rpc/*.proto`
- 后端类型定义：`api/types/*.proto`
- 前端 API 封装：`console/src/api/*.ts`
- 前端类型定义：`console/src/types/api.ts`
- HTTP 客户端：`console/src/lib/http.ts`
