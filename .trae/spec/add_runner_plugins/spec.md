# GFlow Runner 插件开发规范

## 1. 概述

本文档详细说明了如何在 GFlow Runner 中开发和注册新的插件。GFlow 支持两种插件类型：**基于结构体的插件**（struct-based）和**基于函数的插件**（function-based）。

### 1.1 插件系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Runner 插件系统架构                      │
├─────────────────────────────────────────────────────────────┤
│                                                           │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              TaskFactory                          │   │
│  │  ┌───────────────────────────────────────────┐  │   │
│  │  │  endpoints: map[string]*Endpoint        │  │   │
│  │  │  prototypes: map[string]Task            │  │   │
│  │  │  cachePool: map[string]Task            │  │   │
│  │  └───────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
│                           │                              │
│                           ▼                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              插件类型                              │   │
│  │  ┌──────────────┐    ┌──────────────┐          │   │
│  │  │ Struct-based │    │ Function-based│          │   │
│  │  │   (Task)    │    │   (Func)     │          │   │
│  │  └──────────────┘    └──────────────┘          │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                           │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 核心概念

| 概念 | 说明 |
|------|------|
| **Task** | 任务接口，定义三阶段执行模式 |
| **TaskFactory** | 任务工厂，负责注册和创建插件实例 |
| **Endpoint** | 任务端点，包含任务的元数据和输入输出定义 |
| **三阶段执行** | Commit（提交）、Rollback（回滚）、Destroy（销毁） |
| **依赖注入** | 自动将依赖注入到任务实例 |
| **数据绑定** | 通过标签自动绑定请求数据到任务参数 |

## 2. 插件类型

### 2.1 Struct-based 插件

基于结构体的插件通过实现 `Task` 接口来定义任务。

#### 2.1.1 Task 接口定义

```go
type Task interface {
    // Commit 执行主要业务逻辑
    Commit(ctx context.Context, request any) (any, error)
    
    // Rollback 回滚操作，用于事务补偿
    Rollback(ctx context.Context) error
    
    // Destroy 清理资源，销毁任务实例
    Destroy(ctx context.Context) error
    
    // String 返回任务名称
    String() string
}
```

#### 2.1.2 Struct-based 插件示例

```go
package tasks

import (
    "context"
    "fmt"
    
    "github.com/olive-io/gflow/api/types"
)

// EmailTask 发送邮件任务
type EmailTask struct {
    SMTPHost string `inject:"smtp-host"`
    SMTPPort int    `inject:"smtp-port"`
}

func (t *EmailTask) Commit(ctx context.Context, request any) (any, error) {
    req := request.(*EmailRequest)
    
    // 执行发送邮件逻辑
    fmt.Printf("Sending email to %s: %s\n", req.To, req.Subject)
    
    return &EmailResponse{
        Status:  "sent",
        Message: "Email sent successfully",
    }, nil
}

func (t *EmailTask) Rollback(ctx context.Context) error {
    // 邮件发送通常不支持回滚，返回 nil
    return nil
}

func (t *EmailTask) Destroy(ctx context.Context) error {
    // 清理资源
    return nil
}

func (t *EmailTask) String() string {
    return "EmailTask"
}

// EmailRequest 请求结构体
type EmailRequest struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
}

// EmailResponse 响应结构体
type EmailResponse struct {
    Status  string `json:"status"`
    Message string `json:"message"`
}
```

#### 2.1.3 支持克隆的插件

如果任务需要保持状态，可以实现 `TaskClone` 接口：

```go
// TaskClone 接口支持任务实例克隆
type TaskClone interface {
    Clone() TaskClone
}

// StatefulTask 有状态的任务
type StatefulTask struct {
    counter int
}

func (t *StatefulTask) Clone() runner.TaskClone {
    return &StatefulTask{
        counter: t.counter,
    }
}

func (t *StatefulTask) Commit(ctx context.Context, request any) (any, error) {
    t.counter++
    return &Response{Count: t.counter}, nil
}

func (t *StatefulTask) Rollback(ctx context.Context) error {
    return nil
}

func (t *StatefulTask) Destroy(ctx context.Context) error {
    return nil
}

func (t *StatefulTask) String() string {
    return "StatefulTask"
}
```

### 2.2 Function-based 插件

基于函数的插件通过普通函数来定义任务，更加简洁。

#### 2.2.1 支持的函数签名

| 函数签名 | 说明 |
|---------|------|
| `func()` | 无参数无返回值 |
| `func() error` | 无参数，返回错误 |
| `func() (*Response, error)` | 无参数，返回响应和错误 |
| `func(ctx context.Context)` | 接收 context |
| `func(ctx context.Context) error` | 接收 context，返回错误 |
| `func(ctx context.Context, req *Request) (*Response, error)` | 接收 context 和请求 |
| `func(p0 Type0, p1 Type1, ...) (Result, error)` | 接收多个参数 |
| `func(ctx context.Context, p0 Type0, p1 Type1, ...) (Result, error)` | 接收 context 和多个参数 |

#### 2.2.2 Function-based 插件示例

```go
package tasks

import (
    "context"
    "fmt"
)

// SimpleTask 简单函数任务
func SimpleTask(ctx context.Context, name string) (string, error) {
    result := fmt.Sprintf("Hello, %s!", name)
    return result, nil
}

// CalculationTask 计算任务
func CalculationTask(a, b int) (int, error) {
    return a + b, nil
}

// DataProcessingTask 数据处理任务
func DataProcessingTask(ctx context.Context, data *DataRequest) (*DataResponse, error) {
    processed := processData(data)
    return &DataResponse{
        Result: processed,
        Status: "success",
    }, nil
}

// NoParamTask 无参数任务
func NoParamTask() error {
    fmt.Println("Executing task without parameters")
    return nil
}
```

## 3. 数据绑定

### 3.1 数据绑定标签

GFlow 支持三种数据绑定标签：

| 标签 | 用途 | 绑定目标 |
|------|------|---------|
| `json:"key"` | 属性绑定 | `Properties` |
| `gflow:"hr:key"` | 头部绑定 | `Headers` |
| `gflow:"dt:key"` | 数据对象绑定 | `DataObjects` |

### 3.2 请求结构体示例

```go
type ComplexRequest struct {
    // json 标签绑定到 Properties
    UserID    string `json:"user_id"`
    Action    string `json:"action"`
    Timestamp int64  `json:"timestamp"`
    
    // gflow 标签绑定到 Headers
    TraceID   string `gflow:"hr:trace-id"`
    RequestID string `gflow:"hr:request-id"`
    
    // gflow 标签绑定到 DataObjects
    UserData  *UserData `gflow:"dt:user-data"`
    Config    *Config   `gflow:"dt:config"`
}
```

### 3.3 响应结构体示例

```go
type ComplexResponse struct {
    // json 标签绑定到 Results
    Status    string `json:"status"`
    Message   string `json:"message"`
    Data      any    `json:"data"`
    
    // gflow 标签绑定到 Results
    TraceID   string `gflow:"hr:trace-id"`
    Timestamp int64  `gflow:"hr:timestamp"`
}
```

## 4. 依赖注入

### 4.1 依赖注入标签

| 标签 | 说明 |
|------|------|
| `inject:""` | 单例注入，按类型匹配 |
| `inject:"private"` | 私有实例，每次创建新实例 |
| `inject:"name"` | 命名注入，按名称匹配 |
| `inject:"inline"` | 内联结构体，展开字段 |

### 4.2 依赖注入示例

```go
type DatabaseTask struct {
    // 单例注入 - 所有实例共享同一个 DB 连接
    DB *sql.DB `inject:""`
    
    // 私有实例 - 每个任务实例获得独立的缓存
    Cache *Cache `inject:"private"`
    
    // 命名注入 - 按名称匹配
    Logger *Logger `inject:"app-logger"`
    
    // 内联结构体 - 展开字段
    Config Config `inject:"inline"`
}

type Config struct {
    Timeout time.Duration
    Retry   int
}
```

### 4.3 提供依赖

在 Runner 初始化时提供依赖：

```go
func main() {
    runner := runner.NewRunner(cfg)
    
    // 提供依赖
    db, _ := sql.Open("mysql", "dsn")
    logger := zap.NewNop()
    cache := NewCache()
    
    runner.Provide(
        db,                    // 按类型提供
        logger,                 // 按类型提供
        cache,                  // 按类型提供
        ProvideWithName(logger, "app-logger"), // 按名称提供
    )
    
    // 注册任务
    factory := runner.GetFactory(types.FlowNodeType_ServiceTask)
    factory.Register(
        plugins.WithTaskName("database-task"),
        plugins.WithTask(&DatabaseTask{}),
        plugins.WithRequest(new(DatabaseRequest)),
        plugins.WithResponse(new(DatabaseResponse)),
    )
    
    runner.Run()
}
```

## 5. 插件注册

### 5.1 注册 Struct-based 插件

```go
// 获取任务工厂
factory := runner.GetFactory(types.FlowNodeType_ServiceTask)

// 注册任务
err := factory.Register(
    plugins.WithTaskName("email-task"),
    plugins.WithTask(&EmailTask{}),
    plugins.WithRequest(new(EmailRequest)),
    plugins.WithResponse(new(EmailResponse)),
    plugins.WithDesc("Send email notification"),
    plugins.WithTaskType("gflow"),
)
if err != nil {
    log.Fatal(err)
}
```

### 5.2 注册 Function-based 插件

```go
// 获取任务工厂
factory := runner.GetFactory(types.FlowNodeType_ServiceTask)

// 注册函数任务
err := factory.Register(
    plugins.WithTaskName("simple-task"),
    plugins.WithTask(SimpleTask),
    plugins.WithRequest(new(SimpleRequest)),
    plugins.WithResponse(new(SimpleResponse)),
)
if err != nil {
    log.Fatal(err)
}
```

### 5.3 注册选项

| 选项 | 说明 | 示例 |
|------|------|------|
| `WithTaskName(name)` | 设置任务名称 | `WithTaskName("my-task")` |
| `WithTask(task)` | 设置任务实例 | `WithTask(&MyTask{})` |
| `WithRequest(req)` | 设置请求结构体 | `WithRequest(new(MyRequest))` |
| `WithResponse(resp)` | 设置响应结构体 | `WithResponse(new(MyResponse))` |
| `WithDesc(desc)` | 设置任务描述 | `WithDesc("My task description")` |
| `WithFlowType(type)` | 设置流程节点类型 | `WithFlowType(types.FlowNodeType_ServiceTask)` |
| `WithTaskType(type)` | 设置任务类型 | `WithTaskType("gflow")` |

## 6. 三阶段执行模式

### 6.1 阶段说明

| 阶段 | 说明 | 使用场景 |
|------|------|---------|
| **Echo** | 预检阶段，测试任务是否可用 | 验证任务配置 |
| **Commit** | 提交阶段，执行主要业务逻辑 | 正常任务执行 |
| **Rollback** | 回滚阶段，补偿操作 | 任务失败时回滚 |
| **Destroy** | 销毁阶段，清理资源 | 任务完成后清理 |

### 6.2 执行流程

```
┌─────────────────────────────────────────────────────────────┐
│                    三阶段执行流程                         │
└─────────────────────────────────────────────────────────────┘

1. Echo 阶段
   ┌──────────┐    Clone    ┌──────────┐    Commit    ┌──────────┐
   │ Prototype│ ─────────► │ Instance │ ──────────► │  Result  │
   └──────────┘            └──────────┘             └──────────┘
   
2. Commit 阶段
   ┌──────────┐    Clone    ┌──────────┐    Commit    ┌──────────┐
   │ Prototype│ ─────────► │ Instance │ ──────────► │  Result  │
   └──────────┘            └──────────┘             └──────────┘
                               │
                               ▼ Cache
                          ┌──────────┐
                          │ Instance │ (缓存用于 Rollback)
                          └──────────┘

3. Rollback 阶段
   ┌──────────┐  Get from Cache  ┌──────────┐    Rollback  ┌──────────┐
   │ Instance │ ◄────────────── │ Instance │ ──────────► │  Result  │
   └──────────┘                 └──────────┘             └──────────┘

4. Destroy 阶段
   ┌──────────┐  Get from Cache  ┌──────────┐    Destroy   ┌──────────┐
   │ Instance │ ◄────────────── │ Instance │ ──────────► │  Result  │
   └──────────┘                 └──────────┘             └──────────┘
                               │
                               ▼ Delete
```

### 6.3 实现建议

#### 6.3.1 Commit 阶段

```go
func (t *MyTask) Commit(ctx context.Context, request any) (any, error) {
    req := request.(*MyRequest)
    
    // 1. 参数验证
    if req.UserID == "" {
        return nil, errors.New("user_id is required")
    }
    
    // 2. 执行业务逻辑
    result, err := t.process(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // 3. 返回结果
    return &MyResponse{
        Status:  "success",
        Data:    result,
    }, nil
}
```

#### 6.3.2 Rollback 阶段

```go
func (t *MyTask) Rollback(ctx context.Context) error {
    // 1. 获取之前保存的状态
    state := t.getState()
    
    // 2. 执行回滚逻辑
    if state != nil {
        err := t.undo(ctx, state)
        if err != nil {
            return fmt.Errorf("rollback failed: %w", err)
        }
    }
    
    return nil
}
```

#### 6.3.3 Destroy 阶段

```go
func (t *MyTask) Destroy(ctx context.Context) error {
    // 1. 关闭连接
    if t.conn != nil {
        t.conn.Close()
    }
    
    // 2. 清理缓存
    t.cache = nil
    
    // 3. 释放资源
    return nil
}
```

## 7. 错误处理

### 7.1 错误返回规范

```go
// Commit 阶段返回错误
func (t *MyTask) Commit(ctx context.Context, request any) (any, error) {
    // 业务错误
    if err := validate(request); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // 系统错误
    if err := t.process(ctx); err != nil {
        return nil, fmt.Errorf("process failed: %w", err)
    }
    
    return result, nil
}

// Rollback 阶段返回错误
func (t *MyTask) Rollback(ctx context.Context) error {
    if err := t.undo(ctx); err != nil {
        return fmt.Errorf("rollback failed: %w", err)
    }
    return nil
}

// Destroy 阶段返回错误
func (t *MyTask) Destroy(ctx context.Context) error {
    if err := t.cleanup(ctx); err != nil {
        return fmt.Errorf("cleanup failed: %w", err)
    }
    return nil
}
```

### 7.2 Panic 恢复

系统会自动捕获 panic 并转换为错误：

```go
// 系统自动处理 panic，无需手动处理
func (t *MyTask) Commit(ctx context.Context, request any) (any, error) {
    // 即使发生 panic，系统也会捕获并返回错误
    panic("something went wrong")
}
```

## 8. 测试

### 8.1 单元测试示例

```go
package tasks

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
)

func TestEmailTask_Commit(t *testing.T) {
    task := &EmailTask{
        SMTPHost: "localhost",
        SMTPPort: 25,
    }
    
    req := &EmailRequest{
        To:      "test@example.com",
        Subject: "Test Email",
        Body:    "This is a test email",
    }
    
    resp, err := task.Commit(context.Background(), req)
    assert.NoError(t, err)
    assert.NotNil(t, resp)
    
    emailResp := resp.(*EmailResponse)
    assert.Equal(t, "sent", emailResp.Status)
}

func TestEmailTask_Rollback(t *testing.T) {
    task := &EmailTask{}
    err := task.Rollback(context.Background())
    assert.NoError(t, err)
}

func TestEmailTask_Destroy(t *testing.T) {
    task := &EmailTask{}
    err := task.Destroy(context.Background())
    assert.NoError(t, err)
}
```

### 8.2 集成测试示例

```go
package runner_test

import (
    "context"
    "testing"
    
    "github.com/olive-io/gflow/api/types"
    "github.com/olive-io/gflow/plugins"
    "github.com/olive-io/gflow/runner"
)

func TestTaskExecution(t *testing.T) {
    r := runner.NewRunner(&runner.Config{})
    
    // 注册任务
    factory := r.GetFactory(types.FlowNodeType_ServiceTask)
    factory.Register(
        plugins.WithTaskName("test-task"),
        plugins.WithTask(&TestTask{}),
        plugins.WithRequest(new(TestRequest)),
        plugins.WithResponse(new(TestResponse)),
    )
    
    // 执行任务
    req := &types.CallTaskRequest{
        TaskType: types.FlowNodeType_ServiceTask,
        Name:     "test-task",
        Stage:    types.CallTaskStage_Commit,
        Properties: map[string]*types.Value{
            "name": types.NewValue("test"),
        },
    }
    
    resp := r.Handle(context.Background(), req)
    assert.Nil(t, resp.Error)
}
```

## 9. 最佳实践

### 9.1 命名规范

- **任务名称**：使用小写字母和连字符，如 `email-task`
- **结构体名称**：使用 PascalCase，如 `EmailTask`
- **函数名称**：使用 PascalCase，如 `SendEmail`
- **请求/响应**：使用 PascalCase + Request/Response 后缀

### 9.2 代码组织

```
tasks/
├── email/
│   ├── email.go       # 任务实现
│   ├── email_test.go  # 单元测试
│   └── types.go      # 请求/响应类型
├── database/
│   ├── database.go
│   ├── database_test.go
│   └── types.go
└── notification/
    ├── notification.go
    ├── notification_test.go
    └── types.go
```

### 9.3 性能优化

- **避免全局变量**：使用依赖注入管理共享资源
- **实现 Clone 接口**：对于有状态的任务，支持实例克隆
- **合理使用缓存**：在 Commit 阶段缓存必要状态
- **及时释放资源**：在 Destroy 阶段清理所有资源

### 9.4 安全性

- **输入验证**：在 Commit 阶段验证所有输入参数
- **错误处理**：不要暴露敏感信息在错误消息中
- **资源限制**：设置合理的超时和重试次数
- **日志记录**：记录关键操作和错误信息

## 10. 常见问题

### 10.1 如何处理并发执行？

实现 `TaskClone` 接口，系统会自动克隆任务实例：

```go
func (t *MyTask) Clone() runner.TaskClone {
    return &MyTask{
        // 复制所有字段
    }
}
```

### 10.2 如何访问原始请求数据？

使用 `GetOriginData` 函数：

```go
func (t *MyTask) Commit(ctx context.Context, request any) (any, error) {
    origin, ok := runner.GetOriginData(ctx)
    if ok {
        // 访问原始请求
        headers := origin.Headers
        properties := origin.Properties
    }
    // ...
}
```

### 10.3 如何实现任务链？

在 BPMN 流程中定义多个任务节点，通过数据对象传递数据：

```go
// Task 1
type Task1 struct{}
func (t *Task1) Commit(ctx context.Context, request any) (any, error) {
    return &Response{
        DataObjects: map[string]*types.Value{
            "intermediate": types.NewValue("data"),
        },
    }, nil
}

// Task 2
type Task2 struct{}
func (t *Task2) Commit(ctx context.Context, request any) (any, error) {
    req := request.(*Request)
    intermediate := req.DataObjects["intermediate"]
    // 处理中间数据
    return &Response{}, nil
}
```

### 10.4 如何实现任务重试？

在 Runner 配置中设置重试策略，或者在任务中实现重试逻辑：

```go
func (t *MyTask) Commit(ctx context.Context, request any) (any, error) {
    var result any
    var err error
    
    for i := 0; i < 3; i++ {
        result, err = t.execute(ctx, request)
        if err == nil {
            break
        }
        time.Sleep(time.Second * time.Duration(i+1))
    }
    
    return result, err
}
```

## 11. 参考资料

- [GFlow Runner 组件分析](../.trae/runner_task_analysis.md)
- [GFlow 插件系统分析](../.trae/plugin_system_analysis.md)
- [API 类型定义](../api/types/)
- [Runner 包文档](../runner/)
