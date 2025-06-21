# OpenManus Go

OpenManus Go 是 OpenManus 的 Go 语言重写版本，支持多智能体、工具调用、流程控制和安全沙箱等功能。

## 安装指南

### 前置条件
- Go 1.21 或更高版本
- 配置 OpenAI API 密钥

### 安装步骤
1. 克隆仓库：
   ```bash
   git clone https://github.com/yourusername/openmanus-go.git
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

## 项目结构
- `cmd/`: 命令行入口
- `internal/`: 内部包（智能体、工具、流程、沙箱等）
- `pkg/`: 公共包（日志、工具函数等）
- `config/`: 配置文件

## 贡献指南
欢迎提交 Issue 或 Pull Request！

## 许可证
MIT
