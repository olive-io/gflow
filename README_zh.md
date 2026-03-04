# GFlow

## 简介

GFlow 是一个使用 Go 语言构建的 BPMN（业务流程模型和符号）工作流引擎。它提供了一个强大且可扩展的业务流程管理和执行解决方案。

## 特性

- **BPMN 2.0 支持**：完整支持 BPMN 2.0 规范
- **高性能**：使用 Go 语言构建，支持高并发
- **分布式架构**：支持分布式部署和水平扩展
- **实时监控**：实时流程监控和管理
- **REST API & gRPC**：双 API 支持，灵活集成
- **Web 控制台**：现代化的 Web 管理控制台
- **CLI 工具**：命令行工具，便于管理

## 架构

```
┌─────────────────────────────────────────────────────┐
│                      Web 控制台                              │
└─────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────┐
│                    gflow-server                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  REST API   │  │  gRPC API   │  │   调度器    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────┐
│                    gflow-runner                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   分发器    │  │   执行器    │  │    插件     │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────┐
│                     数据库 (MySQL/PostgreSQL)                │
└─────────────────────────────────────────────────────┘
```

## 安装

#### 下载二进制文件

从 [GitHub Releases](https://github.com/olive-io/gflow/releases) 下载最新版本：

```bash
# Linux
wget https://github.com/olive-io/gflow/releases/download/v1.0.0/gflow-1.0.0-linux-amd64.tar.gz
tar -xzf gflow-1.0.0-linux-amd64.tar.gz

# macOS
wget https://github.com/olive-io/gflow/releases/download/v1.0.0/gflow-1.0.0-darwin-amd64.tar.gz
tar -xzf gflow-1.0.0-darwin-amd64.tar.gz

# Windows
# 下载 gflow-1.0.0-windows-amd64.zip 并解压
```

#### 从包安装

```bash
# Debian/Ubuntu
sudo dpkg -i gflow_1.0.0_amd64.deb

# CentOS/RHEL/Fedora
sudo rpm -i gflow-1.0.0.x86_64.rpm

# Alpine Linux
sudo apk add gflow_1.0.0_x86_64.apk
```

#### 从源码构建

```bash
git clone https://github.com/olive-io/gflow.git
cd gflow
go build -o bin/gflow-server ./cmd/gflow-server
go build -o bin/gflow-runner ./cmd/gflow-runner
go build -o bin/gfcli ./cmd/gflow-cli
```

## 快速开始

1. **启动服务器**

```bash
./gflow-server --config gflow.toml
```

2. **启动 Runner**

```bash
./gflow-runner --config runner.toml
```

3. **部署 BPMN 定义**

```bash
gfcli bpmn definitions deploy process.bpmn -d "我的流程"
```

4. **执行流程**

```bash
gfcli bpmn process execute -u <definition-uid> -n "流程实例"
```

## CLI 命令

```bash
# 认证
gfcli auth login -u admin -p admin123

# BPMN 定义管理
gfcli bpmn definitions deploy <file>      # 部署 BPMN 定义
gfcli bpmn definitions list               # 列出定义
gfcli bpmn definitions get <uid>          # 获取定义

# 流程管理
gfcli bpmn process execute                # 执行流程
gfcli bpmn process list                   # 列出流程
gfcli bpmn process get <id>               # 获取流程详情

# 系统
gfcli system ping                         # 健康检查
```

## 配置

配置文件：`~/.gflow/gfcli.yaml`

```yaml
endpoint: localhost:6550
token: "your-auth-token"
```

## 开发

#### 环境要求

- Go 1.25+
- Node.js 18+ (用于 Web 控制台)
- MySQL 8.0+ 或 PostgreSQL 14+

#### 运行开发服务器

```bash
# 后端
go run ./cmd/gflow-server --config gflow.toml

# 前端
cd console && pnpm dev
```

## 许可证

Apache License 2.0
