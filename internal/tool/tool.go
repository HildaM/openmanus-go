package tool

import (
	"context"
	"fmt"
	"sync"
)

type Tool interface {
	Name() string
	Run(ctx context.Context, input interface{}) (output interface{}, err error)
}

// ToolError 通用错误类型
type ToolError struct {
	msg string
}

func (e *ToolError) Error() string {
	return e.msg
}

// ToolCollection 工具集合
type ToolCollection struct {
	tools map[string]Tool
	mu    sync.Mutex
}

// NewToolCollection 创建工具集合
func NewToolCollection() *ToolCollection {
	return &ToolCollection{
		tools: make(map[string]Tool),
	}
}

// Register 注册工具
func (tc *ToolCollection) Register(name string, tool Tool) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tool == nil {
		return fmt.Errorf("工具不能为空")
	}
	tc.tools[name] = tool
	return nil
}

// Get 获取工具
func (tc *ToolCollection) Get(name string) (Tool, error) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tool, ok := tc.tools[name]; ok {
		return tool, nil
	}
	return nil, fmt.Errorf("工具 %s 不存在", name)
}

// Remove 移除工具
func (tc *ToolCollection) Remove(name string) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if _, ok := tc.tools[name]; !ok {
		return fmt.Errorf("工具 %s 不存在", name)
	}
	delete(tc.tools, name)
	return nil
}

// List 列出所有工具
func (tc *ToolCollection) List() []string {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	names := make([]string, 0, len(tc.tools))
	for name := range tc.tools {
		names = append(names, name)
	}
	return names
}
