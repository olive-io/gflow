# GFlow 项目分析计划

## 项目概述

**GFlow** 是一个基于 BPMN 2.0 规范的工作流引擎系统，采用 Go 语言编写（go 1.25.0）。该项目实现了分布式工作流调度和执行能力，支持多种任务类型和插件扩展机制。

## 架构分析

### 1. 整体架构

GFlow 采用 **客户端-服务器** 分布式架构，包含两个核心组件：

```
┌─────────────────────────────────────────────────────────────────┐
│                         GFlow System                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐         ┌─────────────────────────────┐   │
│  │  gflow-server   │◄───────►│       gflow-runner          │   │
│  │  (调度服务器)    │  gRPC   │      (执行器)                │   │
│  └────────┬────────┘         └──────────────┬──────────────┘   │
│           │                                  │                  │
│           ▼                                  ▼                  │
│  ┌─────────────────┐         ┌─────────────────────────────┐   │
│  │    Database     │         │      Plugin System          │   │
│  │  (MySQL/SQLite/ │         │  (ServiceTask/SendTask/     │   │
│  │   PostgreSQL)   │         │   ReceiveTask/ScriptTask)   │   │
│  └─────────────────┘         └─────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 2. 核心组件

#### 2.1 gflow-server（调度服务器）

**位置**: `cmd/gflow-server/main.go` → `server/server.go`

**职责**:
- 提供 gRPC 和 REST API 服务
- 管理 BPMN 流程定义（Definitions）
- 调度流程实例（Process）执行
- 管理执行器（Runner）注册和心跳
- 权限控制（基于 Casbin）

**关键模块**:
| 模块 | 文件 | 功能 |
|------|------|------|
| Scheduler | `server/scheduler/scheduler.go` | 流程调度引擎，使用优先级队列和协程池 |
| Dispatcher | `server/dispatch/dispatch.go` | 任务分发器，管理 Runner 连接 |
| DAO | `server/dao/*.go` | 数据访问层，使用 GORM |
| Plugin | `server/plugin/*/` | 服务端插件实现 |

#### 2.2 gflow-runner（执行器）

**位置**: `cmd/gflow-runner/main.go` → `runner/runner.go`

**职责**:
- 向 Server 注册并维持心跳
- 接收并执行任务
- 管理本地插件实例
- 上报执行状态和资源使用情况

**关键模块**:
| 模块 | 文件 | 功能 |
|------|------|------|
| Task | `runner/task.go` | 任务代理和函数代理实现 |
| Plugin | `runner/plugin.go` | Runner 端插件管理 |

### 3. 数据模型

**位置**: `api/types/bpmn.proto`

核心数据结构：

```
┌──────────────────┐     ┌──────────────────┐
│   Definitions    │────►│     Process      │
│  (流程定义)       │     │   (流程实例)      │
└──────────────────┘     └────────┬─────────┘
                                  │
                                  ▼
                         ┌──────────────────┐
                         │    FlowNode      │
                         │   (流程节点)      │
                         └──────────────────┘
```

**主要实体**:
- `Definitions`: BPMN 流程定义，存储 XML 内容
- `Process`: 流程实例，包含执行状态、阶段、上下文
- `FlowNode`: 流程节点，记录任务执行详情
- `Endpoint`: 任务端点，定义任务输入输出规范
- `Runner`: 执行器信息，包含资源、心跳等

### 4. API 层

**位置**: `api/rpc/*.proto`

| 服务 | 功能 |
|------|------|
| BpmnRPC | 流程定义部署、流程执行、流程查询 |
| SystemRPC | Runner 注册、端点管理、健康检查 |
| AuthRPC | 用户认证、权限验证 |
| AdminRPC | 用户管理、角色管理、路由管理 |

**协议支持**:
- gRPC（主要）
- REST API（通过 grpc-gateway）
- WebSocket（用于流式调用）

### 5. 插件系统

**位置**: `plugins/plugin.go`

**插件类型**（对应 BPMN 任务类型）:
| 类型 | 说明 | 实现位置 |
|------|------|----------|
| ServiceTask | 服务任务 | `server/plugin/service/gflow.go` |
| SendTask | 发送任务 | `server/plugin/send/rabbitmq.go` |
| ReceiveTask | 接收任务 | `server/plugin/receive/rabbitmq.go` |
| ScriptTask | 脚本任务 | 待实现 |

**插件接口**:
```go
type Plugin interface {
    Name() string
    Do(ctx context.Context, req *Request, opts ...DoOption) (*Response, error)
}

type Factory interface {
    Name() string
    Create(opts ...Option) (Plugin, error)
}
```

### 6. 流程执行引擎

**核心依赖**: `github.com/olive-io/bpmn/v2`

**执行流程**:
1. 解析 BPMN XML 定义
2. 创建流程实例
3. 订阅执行追踪事件（Trace）
4. 遇到任务节点时调用插件执行
5. 支持 Commit/Rollback/Destroy 三阶段

**阶段模式**:
- `Simple`: 简单模式，仅执行任务
- `Transition`: 事务模式，支持回滚和销毁

### 7. 技术栈

| 类别 | 技术 |
|------|------|
| 语言 | Go 1.25.0 |
| RPC | gRPC + grpc-gateway |
| 数据库 | GORM（支持 MySQL/PostgreSQL/SQLite/SQLServer） |
| 消息队列 | RabbitMQ |
| 日志 | Zap + OpenTelemetry |
| 监控 | Prometheus + OpenTelemetry |
| 权限 | Casbin |
| 流程引擎 | olive-io/bpmn |
| 协程池 | panjf2000/ants |
| 配置 | Viper + TOML/YAML |

## 目录结构

```
gflow/
├── api/                    # API 定义
│   ├── rpc/               # gRPC 服务定义
│   └── types/             # 数据类型定义
├── build/                  # 构建脚本和配置
├── clientgo/              # Go 客户端库
├── cmd/                    # 入口程序
│   ├── gflow-server/      # 服务器入口
│   └── gflow-runner/      # 执行器入口
├── pkg/                    # 公共包
│   ├── casbin/            # Casbin 权限
│   ├── dbutil/            # 数据库工具
│   ├── logutil/           # 日志工具
│   ├── metrics/           # 指标收集
│   ├── signalutil/        # 信号处理
│   └── trace/             # 链路追踪
├── plugins/               # 插件系统核心
├── runner/                # Runner 实现
├── server/                # Server 实现
│   ├── config/           # 配置
│   ├── dao/              # 数据访问层
│   ├── dispatch/         # 任务分发
│   ├── docs/             # OpenAPI 文档
│   ├── plugin/           # 服务端插件
│   └── scheduler/        # 调度器
└── third-party/          # 第三方 proto 文件
```

## 构建和测试

```bash
# 构建
go build ./cmd/gflow-server
go build ./cmd/gflow-runner

# 测试
make test          # 运行测试
make vet           # 静态检查
make lint          # Lint 检查

# 生成 API
make apis          # 从 proto 生成代码
```

## 关键设计特点

1. **分布式执行**: Server 负责调度，Runner 负责执行，支持水平扩展
2. **插件化架构**: 通过 Factory/Plugin 接口实现任务类型扩展
3. **事务支持**: Transition 模式支持任务回滚和资源清理
4. **可观测性**: 集成 OpenTelemetry、Prometheus、结构化日志
5. **多协议支持**: gRPC、REST、WebSocket
6. **多数据库支持**: 通过 GORM 支持 MySQL/PostgreSQL/SQLite/SQLServer

## 待分析项

- [ ] 详细分析 BPMN 流程执行机制
- [ ] 分析 Runner 与 Server 的通信协议
- [ ] 分析权限模型和 Casbin 配置
- [ ] 分析消息队列集成方式
- [ ] 分析测试覆盖率和测试策略
