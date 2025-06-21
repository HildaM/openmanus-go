package tool

import (
	"context"
	"sync"
)

var (
	registry     *ToolRegistry
	registryOnce sync.Once
)

// GetRegistry 获取工具注册表实例
func GetRegistry() *ToolRegistry {
	registryOnce.Do(func() {
		registry = NewToolRegistry()
	})
	return registry
}

// RegisterDefaultTools 注册默认工具
func RegisterDefaultTools() {
	reg := GetRegistry()

	// 注册文件操作工具
	reg.Register(NewFileOpsTool())

	// 注册字符串处理工具
	reg.Register(NewStrReplaceTool())

	// 注册 Python 执行工具
	// 注意：Python 执行工具需要 Python 服务 URL
	// 这里暂时不注册，等待 Python 服务配置完成后再注册
}

// ExecuteTool 执行指定工具
func ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	reg := GetRegistry()
	tool, ok := reg.Get(toolName)
	if !ok {
		return nil, ErrInvalidOperation
	}
	return tool.Execute(ctx, args)
}

// ListTools 列出所有已注册的工具
func ListTools() []ITool {
	reg := GetRegistry()
	return reg.List()
}
