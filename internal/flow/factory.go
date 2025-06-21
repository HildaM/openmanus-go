package flow

import (
	"fmt"

	"github.com/HildaM/openmanus-go/internal/agent"
	"github.com/HildaM/openmanus-go/internal/llm"
)

// FlowFactory 流程工厂接口
type FlowFactory interface {
	CreateFlow(agents map[string]agent.Agent) (Flow, error)
}

// DefaultFlowFactory 默认流程工厂实现
type DefaultFlowFactory struct{}

// NewDefaultFlowFactory 创建默认流程工厂
func NewDefaultFlowFactory() *DefaultFlowFactory {
	return &DefaultFlowFactory{}
}

// CreateFlow 创建流程
func (f *DefaultFlowFactory) CreateFlow(agents map[string]agent.Agent, llmClient llm.LLMClient) (Flow, error) {
	if agents == nil {
		return nil, fmt.Errorf("agents 不能为空")
	}
	return NewBaseFlow(agents, llmClient), nil
}
