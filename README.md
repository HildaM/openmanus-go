# OpenManus Go

[中文文档](README_zh.md) | English

OpenManus Go is a Go language rewrite of OpenManus, supporting multi-agent systems, tool calling, flow control, and secure sandboxing.

## Installation Guide

### Prerequisites
- Go 1.21 or higher
- OpenAI API key configuration

### Installation Steps
1. Clone the repository:
   ```bash
   git clone https://github.com/HildaM/openmanus-go.git
   cd openmanus-go
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Configure API key:
   Edit `config/config.toml` and replace `api_key` with your OpenAI API key.

## Usage Examples

Run the main program:
```bash
go run cmd/main.go
```

Run task flow:
```bash
go run cmd/flow/main.go
```

Run MCP service:
```bash
go run cmd/mcp/main.go
```

## Project Structure
- `cmd/`: Command line entry points
  - `main.go`: Main program entry
  - `flow/`: Task flow entry
  - `mcp/`: MCP service entry
- `internal/`: Internal packages
  - `agent/`: Agent implementations (ReAct, SWE, Browser, DataAnalysis, etc.)
  - `tool/`: Tool collection (file operations, Shell, Python execution, etc.)
  - `flow/`: Flow control
  - `llm/`: LLM client
  - `mcp/`: MCP protocol implementation
  - `sandbox/`: Secure sandbox
  - `prompt/`: Prompt management
  - `config/`: Configuration management
- `pkg/`: Public packages
  - `logger/`: Logging utilities
- `config/`: Configuration files

## Features

### Multi-Agent System
- **ReAct Agent**: Reasoning-Acting loop agent
- **SWE Agent**: Software engineering agent
- **Browser Agent**: Browser automation agent
- **Data Analysis Agent**: Data analysis agent
- **MCP Agent**: MCP protocol agent
- **Manus Agent**: Core agent

### Tool System
- File operation tools
- Shell command execution
- Python code execution
- Web API calls
- Data analysis tools

### Security Features
- Sandbox environment execution
- Permission control
- Secure code execution

## Contributing
Welcome to submit Issues or Pull Requests!

## License
MIT
