package prompt

import (
	"fmt"
	"strings"
	"text/template"
)

// Prompt 定义提示词接口
type Prompt interface {
	// Render 渲染提示词模板
	Render(data interface{}) (string, error)
	// GetTemplate 获取原始模板
	GetTemplate() string
}

// BasePrompt 基础提示词实现
type BasePrompt struct {
	template string
	funcMap  map[string]interface{}
}

// NewBasePrompt 创建基础提示词
func NewBasePrompt(template string) *BasePrompt {
	return &BasePrompt{
		template: template,
		funcMap: map[string]interface{}{
			"join":  strings.Join,
			"lower": strings.ToLower,
			"upper": strings.ToUpper,
		},
	}
}

// Render 渲染提示词模板
func (p *BasePrompt) Render(data interface{}) (string, error) {
	tmpl, err := template.New("prompt").Funcs(template.FuncMap(p.funcMap)).Parse(p.template)
	if err != nil {
		return "", fmt.Errorf("解析模板失败: %v", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("渲染模板失败: %v", err)
	}

	return result.String(), nil
}

// GetTemplate 获取原始模板
func (p *BasePrompt) GetTemplate() string {
	return p.template
}

// AddFunc 添加自定义函数
func (p *BasePrompt) AddFunc(name string, fn interface{}) {
	if p.funcMap == nil {
		p.funcMap = make(map[string]interface{})
	}
	p.funcMap[name] = fn
}
