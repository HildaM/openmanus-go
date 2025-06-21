package agent

import (
	"fmt"
	"strings"

	"github.com/HildaM/openmanus-go/internal/llm"
	"github.com/HildaM/openmanus-go/internal/tool"
)

// FlowAgent 实现工作流智能体
type FlowAgent struct {
	*BaseAgent
}

// NewFlowAgent 创建工作流智能体
func NewFlowAgent(llmClient llm.LLMClient, tools *tool.ToolCollection) *FlowAgent {
	return &FlowAgent{
		BaseAgent: NewBaseAgent("flow", llmClient, tools),
	}
}

// buildPrompt 定制工作流智能体的 prompt
func (f *FlowAgent) buildPrompt(prompt string) string {
	// 获取可用工具列表
	toolList := f.tools.List()
	toolDescs := make([]string, 0, len(toolList))
	for name, desc := range toolList {
		toolDescs = append(toolDescs, fmt.Sprintf("- %s: %s", name, desc))
	}

	// 构造工作流风格的 prompt
	return fmt.Sprintf(`你是一个工作流专家。请设计并执行工作流程，协调多个步骤来完成复杂任务。

可用工具：
%s

请按照以下格式输出：
Workflow Design: 设计工作流程
Current Step: 当前执行的步骤
Action: 工具名称
Action Input: {"参数1": "值1", "参数2": "值2"}
Result: 工具执行结果
Next Step: 下一步计划
... (循环直到工作流完成)
Final Answer: 工作流执行总结

用户输入：%s`, strings.Join(toolDescs, "\n"), prompt)
}
