package agent

import (
	"fmt"
	"strings"

	"github.com/HildaM/openmanus-go/internal/llm"
	"github.com/HildaM/openmanus-go/internal/tool"
)

// TaskAgent 实现任务分解智能体
type TaskAgent struct {
	*BaseAgent
}

// NewTaskAgent 创建任务分解智能体
func NewTaskAgent(llmClient llm.LLMClient, tools *tool.ToolCollection) *TaskAgent {
	return &TaskAgent{
		BaseAgent: NewBaseAgent("task", llmClient, tools),
	}
}

// buildPrompt 定制任务分解智能体的 prompt
func (t *TaskAgent) buildPrompt(prompt string) string {
	// 获取可用工具列表
	toolList := t.tools.List()
	toolDescs := make([]string, 0, len(toolList))
	for name, desc := range toolList {
		toolDescs = append(toolDescs, fmt.Sprintf("- %s: %s", name, desc))
	}

	// 构造任务分解风格的 prompt
	return fmt.Sprintf(`你是一个任务分解专家。请将复杂任务分解为可执行的子任务，并逐步完成。

可用工具：
%s

请按照以下格式输出：
Task Analysis: 分析任务并列出子任务
Current Task: 当前要执行的子任务
Action: 工具名称
Action Input: {"参数1": "值1", "参数2": "值2"}
Result: 工具执行结果
... (循环直到所有子任务完成)
Final Answer: 任务完成总结

用户输入：%s`, strings.Join(toolDescs, "\n"), prompt)
}
