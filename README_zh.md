# OpenManus Go

OpenManus Go 是 OpenManus 的 Go 语言重写版本，支持多智能体、工具调用、流程控制和安全沙箱等功能。

## 安装指南

### 前置条件
- Go 1.21 或更高版本
- 配置 OpenAI API 密钥

### 安装步骤
1. 克隆仓库：
   ```bash
   git clone https://github.com/HildaM/openmanus-go.git
   cd openmanus-go
   ```

2. 安装依赖：
   ```bash
   go mod tidy
   ```

3. 配置 API 密钥：
   编辑 `config/config.toml`，替换 `api_key` 为你的 OpenAI API 密钥。

## 使用示例

运行主程序：
```bash
go run cmd/main.go
```

运行任务流：
```bash
go run cmd/flow/main.go
```

运行 MCP 服务：
```bash
go run cmd/mcp/main.go
```

## 项目结构
- `cmd/`: 命令行入口
  - `main.go`: 主程序入口
  - `flow/`: 任务流程入口
  - `mcp/`: MCP 服务入口
- `internal/`: 内部包
  - `agent/`: 智能体实现（ReAct、SWE、Browser、DataAnalysis 等）
  - `tool/`: 工具集合（文件操作、Shell、Python 执行等）
  - `flow/`: 流程控制
  - `llm/`: LLM 客户端
  - `mcp/`: MCP 协议实现
  - `sandbox/`: 安全沙箱
  - `prompt/`: 提示词管理
  - `config/`: 配置管理
- `pkg/`: 公共包
  - `logger/`: 日志工具
- `config/`: 配置文件

## 功能特性

### 多智能体系统
- **ReAct Agent**: 推理-行动循环智能体
- **SWE Agent**: 软件工程智能体
- **Browser Agent**: 浏览器自动化智能体
- **Data Analysis Agent**: 数据分析智能体
- **MCP Agent**: MCP 协议智能体
- **Manus Agent**: 核心智能体

### 工具系统
- 文件操作工具
- Shell 命令执行
- Python 代码执行
- Web API 调用
- 数据分析工具

### 安全特性
- 沙箱环境执行
- 权限控制
- 安全的代码执行

## 贡献指南
欢迎提交 Issue 或 Pull Request！

## 许可证
MIT