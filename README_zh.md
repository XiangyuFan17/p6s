# p6s - PostgreSQL 终端管理工具

<div align="center">
  <img src="assets/p6s.png" alt="p6s" width="200" height="auto">
</div>

<div align="center">
  <a href="README.md">English</a> | <a href="README_zh.md">中文</a>
</div>

## 简介

p6s 是一个基于终端的 PostgreSQL 数据库管理工具，使用 Go 语言开发，提供直观的文本用户界面 (TUI) 来管理和监控 PostgreSQL 数据库连接。

## 主要功能

- **内置数据库监控**：通过内置功能快速访问全面的数据库信息：
  - 实时查看所有/活跃/阻塞的数据库连接状态
  - 显示数据库表统计信息，包括大小、行数和索引信息
  - 一键过滤和切换不同的连接视图
- **Kubernetes 原生集成**：无缝连接 Kubernetes 集群中的 PostgreSQL 实例：
  - 通过本地 kubeconfig 自动发现 PostgreSQL pods
  - 直接选择 pods、容器和相关的 secrets 以获取凭据信息
- **连接管理**：连接到 PostgreSQL 数据库服务器，支持配置和保存连接信息
- **数据库切换**：快速切换不同的数据库
- **命令模式**：支持在应用程序内执行命令
- **自定义 SQL 查询**：执行自定义 SQL 查询并显示结果

## 截图

### Kubernetes 配置

<div align="center">
  <img src="assets/page_configk8s.png" alt="Kubernetes 配置" width="1200" height="auto">
</div>

### 使用 SQL 查询数据

<div align="center">
  <img src="assets/page_query.png" alt="查询数据" width="1200" height="auto">
</div>

## 使用方法

### 从源码构建

1. **克隆仓库**
   ```bash
   git clone https://github.com/your-username/p6s.git
   cd p6s
   ```

2. **构建并运行可执行文件**
   ```bash
   # 使用 Makefile（推荐）
   make build
   
   # 或手动构建
   go build -o p6s cmd/p6s/main.go
   
   # 运行应用程序
   ./p6s
   ```

3. **跨平台构建**
   ```bash
   # 构建所有平台（推荐）
   make build-all
   
   # 或构建特定平台
   make build-linux-amd64
   make build-linux-arm64
   make build-windows-amd64
   make build-darwin-amd64
   make build-darwin-arm64
   
   # 手动构建（替代方案）
   GOOS=linux GOARCH=amd64 go build -o build/p6s-linux-amd64 cmd/p6s/main.go
   GOOS=linux GOARCH=arm64 go build -o build/p6s-linux-arm64 cmd/p6s/main.go
   GOOS=windows GOARCH=amd64 go build -o build/p6s-windows-amd64.exe cmd/p6s/main.go
   GOOS=darwin GOARCH=amd64 go build -o build/p6s-darwin-amd64 cmd/p6s/main.go
   GOOS=darwin GOARCH=arm64 go build -o build/p6s-darwin-arm64 cmd/p6s/main.go
   ```

### 基本操作

- **配置连接**：启动应用程序后，配置数据库连接信息（主机、端口、用户名、密码、数据库名、SSL 模式）
- **切换数据库**：使用 `\c` 命令或菜单选项切换到不同的数据库
- **查看连接**：主界面显示当前数据库的所有连接信息
- **过滤连接**：使用菜单选择不同的过滤条件（所有连接、活跃连接、阻塞连接）
- **命令模式**：按 `:` 进入命令模式执行特定命令

### 键盘快捷键

- `:` - 进入命令行模式
- `\c` - 切换数据库
- `\config` - 配置连接信息
- `\configk8s` - 通过Kubernetes 配置访问数据库
- `1` - 显示所有连接
- `2` - 显示活跃连接
- `3` - 显示阻塞连接
- `4` - 显示表大小统计
- `5` - 显示 SQL 查询窗口

## 配置文件

连接配置信息保存在用户主目录下的 `.p6s/config.json` 文件中，包含数据库连接设置和 Kubernetes 集成参数：

```json
{
  "host": "",
  "port": "",
  "username": "",
  "password": "",
  "database": "",
  "sslmode": "",
  "namespace": "",
  "pod": "",
  "container": "",
  "port_name": "",
  "secret": "",
  "secret_key": ""
}
```

## 依赖项

- [github.com/gdamore/tcell/v2](https://github.com/gdamore/tcell) - 终端界面库
- [github.com/rivo/tview](https://github.com/rivo/tview) - 基于 tcell 的终端 UI 组件库
- [github.com/lib/pq](https://github.com/lib/pq) - PostgreSQL 驱动
- [k8s.io/client-go](https://github.com/kubernetes/client-go) - Kubernetes 客户端库

## 许可证

MIT License

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=XiangyuFan17/p6s&type=Date)](https://www.star-history.com/#XiangyuFan17/p6s&Date)