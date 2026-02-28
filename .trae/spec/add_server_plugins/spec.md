# Server 插件开发规范

## 1. 概述

本文档描述如何在 GFlow 中添加、使用和调用 Server 端插件。Server 插件用于处理 BPMN 流程中的任务节点，支持多种任务类型如 ServiceTask、SendTask、ReceiveTask 等。

## 2. 插件架构

### 2.1 核心接口

```go
// Plugin 定义所有插件必须实现的基础接口
type Plugin interface {
    Name() string
    Do(ctx context.Context, req *Request, opts ...DoOption) (*Response, error)
}

// Factory 定义创建插件实例的工厂接口
type Factory interface {
    Name() string
    Create(opts ...Option) (Plugin, error)
}

// TaskFactory 扩展 Factory 接口，支持 Endpoint 注册
type TaskFactory interface {
    Factory
    Register(opts ...RegisterOption) error
    ListEndpoint() []types.Endpoint
}
```

### 2.2 请求和响应结构

```go
type Request struct {
    Headers     map[string]string         // HTTP 风格的请求头
    Properties  map[string]*types.Value   // 插件配置的键值属性
    DataObjects map[string]*types.Value   // 结构化数据对象
}

type Response struct {
    Results     map[string]*types.Value   // 插件执行的输出值
    DataObjects map[string]*types.Value   // 返回的结构化数据对象
    Error       string                    // 执行错误信息
}
```

### 2.3 插件类型

| 类型 | 工厂位置 | 用途 |
|------|----------|------|
| ServiceTask | `server/plugin/service/` | 服务任务，支持 gflow/http/grpc 三种调用方式 |
| SendTask | `server/plugin/send/` | 发送任务，发送消息到消息队列 |
| ReceiveTask | `server/plugin/receive/` | 接收任务，从消息队列接收消息 |

## 3. 插件开发指南

### 3.1 创建新插件类型

#### 步骤 1：定义请求和响应结构

```go
// 在 server/plugin/myplugin/types.go 中定义

type MyRequest struct {
    // 使用 json tag 映射到 Properties
    Param1 string `json:"param1"`
    Param2 int    `json:"param2"`
    
    // 使用 gflow:"hr:xxx" tag 映射到 Headers
    ContentType string `gflow:"hr:Content-Type"`
    
    // 使用 gflow:"dt:xxx" tag 映射到 DataObjects
    DataObject any `gflow:"dt:data1"`
}

type MyResponse struct {
    // 使用 json tag 映射到 Results
    Result  string `json:"result"`
    Success bool   `json:"success"`
    
    // 使用 gflow:"dt:xxx" tag 映射到 DataObjects
    OutputData any `gflow:"dt:output"`
}
```

#### 步骤 2：实现 Plugin 接口

```go
// 在 server/plugin/myplugin/plugin.go 中实现

type myPlugin struct {
    lg       *otelzap.Logger
    flowType types.FlowNodeType
    typ      string
    cfg      *config.MyPluginConfig
}

func (p *myPlugin) Name() string {
    return "my-plugin"
}

func (p *myPlugin) Do(ctx context.Context, req *plugins.Request, opts ...plugins.DoOption) (*plugins.Response, error) {
    // 1. 解析执行选项
    options := plugins.NewDoOptions(opts...)
    
    // 2. 绑定请求数据到结构体
    request := new(MyRequest)
    if err := req.ApplyTo(request); err != nil {
        return nil, fmt.Errorf("bind request: %w", err)
    }
    
    // 3. 执行插件逻辑
    result, err := p.execute(ctx, request)
    if err != nil {
        return nil, err
    }
    
    // 4. 提取响应数据
    response := &MyResponse{
        Result:  result,
        Success: true,
    }
    return plugins.ExtractResponse(reflect.ValueOf(response)), nil
}

func (p *myPlugin) execute(ctx context.Context, req *MyRequest) (string, error) {
    // 实现具体的业务逻辑
    return "executed", nil
}
```

#### 步骤 3：实现 Factory 接口

```go
// 在 server/plugin/myplugin/factory.go 中实现

type myFactory struct {
    lg       *otelzap.Logger
    taskType types.FlowNodeType
    cfg      *config.MyPluginConfig
}

func NewFactory(lg *otelzap.Logger, cfg *config.MyPluginConfig) (plugins.Factory, error) {
    return &myFactory{
        lg:       lg,
        taskType: types.FlowNodeType_MyTask,
        cfg:      cfg,
    }, nil
}

func (f *myFactory) Name() string {
    return types.FlowNodeType_MyTask.String()
}

func (f *myFactory) Create(opts ...plugins.Option) (plugins.Plugin, error) {
    options := plugins.NewOptions(opts...)
    
    return &myPlugin{
        lg:       f.lg,
        flowType: f.taskType,
        typ:      options.Type,
        cfg:      f.cfg,
    }, nil
}
```

### 3.2 创建 TaskFactory（支持 Endpoint 注册）

```go
// 对于需要注册 Endpoint 的插件

type myTaskFactory struct {
    *plugins.TaskFactoryImpl
    lg       *otelzap.Logger
    taskType types.FlowNodeType
    cfg      *config.MyPluginConfig
}

func NewFactory(lg *otelzap.Logger, cfg *config.MyPluginConfig) (plugins.Factory, error) {
    factory := &myTaskFactory{
        TaskFactoryImpl: plugins.NewTaskFactory(types.FlowNodeType_MyTask),
        lg:              lg,
        taskType:        types.FlowNodeType_MyTask,
        cfg:             cfg,
    }
    
    // 注册 Endpoint
    if err := factory.Register(
        plugins.WithTaskName("myplugin.execute"),
        plugins.WithTaskType("myplugin"),
        plugins.WithFlowType(types.FlowNodeType_MyTask),
        plugins.WithRequest(new(MyRequest)),
        plugins.WithResponse(new(MyResponse)),
    ); err != nil {
        return nil, err
    }
    
    return factory, nil
}

func (f *myTaskFactory) Create(opts ...plugins.Option) (plugins.Plugin, error) {
    options := plugins.NewOptions(opts...)
    
    // 根据类型创建不同的插件实例
    switch options.Type {
    case "type1":
        return newType1Plugin(f.lg, f.cfg, options.Target)
    case "type2":
        return newType2Plugin(f.lg, f.cfg, options.Target)
    default:
        return newDefaultPlugin(f.lg, f.cfg, options.Target)
    }
}
```

## 4. 插件注册

### 4.1 在 Server 启动时注册

在 `server/server.go` 中注册插件工厂：

```go
func NewServer(ctx context.Context, cfg *config.Config) (*Server, error) {
    // ... 其他初始化代码 ...
    
    // 注册 MyPlugin 工厂
    myPluginFactory, err := myplugin.NewFactory(lg, cfg.Plugin.MyPlugin)
    if err != nil {
        return nil, fmt.Errorf("create myplugin factory: %w", err)
    }
    if err = plugins.Setup(myPluginFactory); err != nil {
        return nil, fmt.Errorf("setup myplugin factory: %w", err)
    }
    
    // ... 其他初始化代码 ...
}
```

### 4.2 配置文件

在 `server/config/config.go` 中添加配置：

```go
type PluginConfig struct {
    ServiceTask  *ServiceTaskPluginConfig  `mapstructure:"service_task"`
    SendTask     *SendTaskPluginConfig     `mapstructure:"send_task"`
    ReceiveTask  *ReceiveTaskPluginConfig  `mapstructure:"receive_task"`
    MyPlugin     *MyPluginConfig           `mapstructure:"my_plugin"`
}

type MyPluginConfig struct {
    Enabled  bool   `mapstructure:"enabled"`
    Endpoint string `mapstructure:"endpoint"`
    Timeout  int    `mapstructure:"timeout"`
}
```

在配置文件 `config/server.toml` 中：

```toml
[plugin.my_plugin]
enabled = true
endpoint = "localhost:8080"
timeout = 30000
```

## 5. 插件调用流程

### 5.1 调用链路

```
BPMN Engine
    │
    ▼
Scheduler.doTask()
    │
    ├── 1. 获取工厂: plugins.Get(taskType)
    │
    ├── 2. 创建插件: factory.Create(opts...)
    │
    ├── 3. 构建请求: plugins.Request{...}
    │
    ├── 4. 执行插件: plugin.Do(ctx, req, doOpts...)
    │
    └── 5. 处理响应: resp.Results, resp.DataObjects
```

### 5.2 调用示例

```go
func (sch *Scheduler) doTask(ctx context.Context, node *types.FlowNode) error {
    // 1. 根据任务类型获取工厂
    taskType := node.FlowType.String()
    factory, err := plugins.Get(taskType)
    if err != nil {
        return fmt.Errorf("get factory: %w", err)
    }
    
    // 2. 创建插件实例
    options := []plugins.Option{
        plugins.WithType(node.Type),
        plugins.WithTarget(node.Target),
    }
    plugin, err := factory.Create(options...)
    if err != nil {
        return fmt.Errorf("create plugin: %w", err)
    }
    
    // 3. 构建请求
    req := &plugins.Request{
        Headers:     node.Headers,
        Properties:  node.Properties,
        DataObjects: node.DataObjects,
    }
    
    // 4. 构建执行选项
    doOptions := []plugins.DoOption{
        plugins.DoWithProcess(node.ProcessId),
        plugins.DoWithTaskStage(types.CallTaskStage_Commit),
        plugins.DoWithTimeout(30000),
    }
    
    // 5. 执行插件
    resp, err := plugin.Do(ctx, req, doOptions...)
    if err != nil {
        return fmt.Errorf("plugin do: %w", err)
    }
    
    // 6. 处理响应
    for k, v := range resp.Results {
        fmt.Printf("Result %s: %v\n", k, v)
    }
    
    return nil
}
```

## 6. 数据绑定

### 6.1 标签说明

| 标签 | 格式 | 映射目标 | 示例 |
|------|------|----------|------|
| `json` | `json:"name"` | Properties | `json:"topic"` |
| `gflow:"hr:xxx"` | `gflow:"hr:Header-Name"` | Headers | `gflow:"hr:Content-Type"` |
| `gflow:"dt:xxx"` | `gflow:"dt:DataObjectName"` | DataObjects | `gflow:"dt:request_body"` |

### 6.2 数据绑定示例

```go
type ExampleRequest struct {
    // 从 Properties 获取
    Topic    string `json:"topic"`
    Body     string `json:"body"`
    
    // 从 Headers 获取
    AuthToken string `gflow:"hr:Authorization"`
    TraceID   string `gflow:"hr:X-Trace-ID"`
    
    // 从 DataObjects 获取
    RequestData map[string]any `gflow:"dt:request_data"`
}

type ExampleResponse struct {
    // 输出到 Results
    Result  string `json:"result"`
    Success bool   `json:"success"`
    
    // 输出到 DataObjects
    ResponseData map[string]any `gflow:"dt:response_data"`
}
```

## 7. 现有插件参考

### 7.1 ServiceTask 插件

支持三种调用方式：

- **gflow**: 通过 Dispatcher 调用远程 Runner
- **http**: 调用 HTTP API
- **grpc**: 调用 gRPC 服务

### 7.2 SendTask 插件

发送消息到 RabbitMQ：

```go
type RabbitRequest struct {
    ContentType string `gflow:"hr:Content-Type"`
    Topic       string `json:"topic"`
    Body        string `json:"body"`
}
```

### 7.3 ReceiveTask 插件

从 RabbitMQ 接收消息：

```go
type RabbitRequest struct {
    Topic       string `json:"topic"`
    ContentType string `json:"content_type"`
}

type RabbitResponse struct {
    Results map[string]any `json:"results"`
}
```

## 8. 最佳实践

### 8.1 错误处理

```go
func (p *myPlugin) Do(ctx context.Context, req *plugins.Request, opts ...plugins.DoOption) (*plugins.Response, error) {
    // 1. 参数校验
    if req == nil {
        return nil, fmt.Errorf("request is nil")
    }
    
    // 2. 绑定数据
    request := new(MyRequest)
    if err := req.ApplyTo(request); err != nil {
        return nil, fmt.Errorf("bind request: %w", err)
    }
    
    // 3. 业务逻辑
    result, err := p.execute(ctx, request)
    if err != nil {
        // 返回包含错误信息的响应
        return &plugins.Response{
            Error: err.Error(),
        }, nil
    }
    
    // 4. 返回成功响应
    response := &MyResponse{Result: result}
    return plugins.ExtractResponse(reflect.ValueOf(response)), nil
}
```

### 8.2 超时控制

```go
func (p *myPlugin) Do(ctx context.Context, req *plugins.Request, opts ...plugins.DoOption) (*plugins.Response, error) {
    options := plugins.NewDoOptions(opts...)
    
    // 创建带超时的上下文
    timeout := time.Duration(options.Timeout) * time.Millisecond
    if timeout == 0 {
        timeout = 30 * time.Second
    }
    
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    // 执行操作
    return p.executeWithContext(ctx, req)
}
```

### 8.3 日志记录

```go
func (p *myPlugin) Do(ctx context.Context, req *plugins.Request, opts ...plugins.DoOption) (*plugins.Response, error) {
    // 记录开始时间
    start := time.Now()
    
    // 执行操作
    resp, err := p.execute(ctx, req)
    
    // 记录执行日志
    p.lg.Info("plugin execution",
        zap.String("plugin", p.Name()),
        zap.Duration("duration", time.Since(start)),
        zap.Error(err),
    )
    
    return resp, err
}
```

## 9. 测试

### 9.1 单元测试

```go
func TestMyPlugin_Do(t *testing.T) {
    // 创建插件实例
    plugin := &myPlugin{
        lg:  otelzap.New(zap.NewNop()),
        cfg: &config.MyPluginConfig{},
    }
    
    // 构建请求
    req := &plugins.Request{
        Properties: map[string]*types.Value{
            "param1": types.NewValue("test"),
            "param2": types.NewValue(123),
        },
    }
    
    // 执行插件
    resp, err := plugin.Do(context.Background(), req)
    
    // 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, resp)
    assert.Contains(t, resp.Results, "result")
}
```

### 9.2 集成测试

```go
func TestMyPlugin_Integration(t *testing.T) {
    // 创建工厂
    factory, err := NewFactory(otelzap.New(zap.NewNop()), &config.MyPluginConfig{})
    require.NoError(t, err)
    
    // 创建插件
    plugin, err := factory.Create(plugins.WithType("default"))
    require.NoError(t, err)
    
    // 执行完整流程
    req := &plugins.Request{...}
    resp, err := plugin.Do(context.Background(), req)
    
    assert.NoError(t, err)
    assert.NotNil(t, resp)
}
```

## 10. 扩展点

### 10.1 自定义选项

```go
// 定义新的选项类型
type MyOption struct {
    CustomField string
}

func WithCustomField(value string) Option {
    return func(o *Options) {
        if o.Custom == nil {
            o.Custom = make(map[string]any)
        }
        o.Custom["custom_field"] = value
    }
}
```

### 10.2 自定义执行选项

```go
func DoWithRetry(retry int) DoOption {
    return func(o *DoOptions) {
        if o.metadata == nil {
            o.metadata = make(map[string]any)
        }
        o.metadata["retry"] = retry
    }
}
```
