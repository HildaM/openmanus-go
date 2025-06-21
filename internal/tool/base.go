package tool

import (
	"context"
	"errors"
)

// ITool 定义了工具的基本接口
type ITool interface {
	// Name 返回工具名称
	Name() string
	// Description 返回工具描述
	Description() string
	// Execute 执行工具功能
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

// BaseTool 提供了工具的基础实现
type BaseTool struct {
	name        string
	description string
}

// NewBaseTool 创建一个新的基础工具
func NewBaseTool(name, description string) *BaseTool {
	return &BaseTool{
		name:        name,
		description: description,
	}
}

// Name 返回工具名称
func (t *BaseTool) Name() string {
	return t.name
}

// Description 返回工具描述
func (t *BaseTool) Description() string {
	return t.description
}

// ToolRegistry 工具注册表
type ToolRegistry struct {
	tools map[string]ITool
}

// NewToolRegistry 创建新的工具注册表
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]ITool),
	}
}

// Register 注册工具
func (r *ToolRegistry) Register(tool ITool) {
	r.tools[tool.Name()] = tool
}

// Get 获取工具
func (r *ToolRegistry) Get(name string) (ITool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// List 列出所有工具
func (r *ToolRegistry) List() []ITool {
	tools := make([]ITool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// 错误定义
var (
	ErrInvalidArgs      = errors.New("invalid arguments")
	ErrInvalidOperation = errors.New("invalid operation")
)

// 预留扩展内容
