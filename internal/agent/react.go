package agent

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/HildaM/openmanus-go/internal/llm"
	"github.com/HildaM/openmanus-go/internal/tool"
	"github.com/HildaM/openmanus-go/pkg/logger"
	"github.com/sirupsen/logrus"
)

// ReactAgent ReAct 智能体
type ReactAgent struct {
	*BaseAgent
	log  *logrus.Logger
	Name string
}

// NewReactAgent 创建 ReAct 智能体
func NewReactAgent(llmClient llm.LLMClient, tools *tool.ToolCollection) *ReactAgent {
	return &ReactAgent{
		BaseAgent: NewBaseAgent("react", llmClient, tools),
		log:       logger.GetLogger(),
	}
}

// Run 执行 React 智能体任务
func (a *ReactAgent) Run(ctx context.Context) error {
	a.log.Info("React 智能体开始执行任务")
	// TODO: 实现 React 智能体具体逻辑
	return nil
}

func (a *ReactAgent) GetName() string {
	return a.Name
}

// buildPrompt 定制 React 智能体的 prompt
func (r *ReactAgent) buildPrompt(prompt string) string {
	// 获取可用工具列表
	toolList := r.tools.List()
	toolDescs := make([]string, 0, len(toolList))
	for name, desc := range toolList {
		toolDescs = append(toolDescs, fmt.Sprintf("- %s: %s", name, desc))
	}

	// 构造 ReAct 风格的 prompt
	return fmt.Sprintf(`你是一个采用 ReAct（Reasoning and Acting）范式的智能助手。请通过"思考-行动-观察"循环来解决问题。

可用工具：
%s

请按照以下格式输出：
Thought: 思考下一步行动
Action: 工具名称
Action Input: {"参数1": "值1", "参数2": "值2"}
Observation: 工具执行结果
... (循环直到任务完成)
Thought: 我已经完成任务
Final Answer: 最终结果

用户输入：%s`, strings.Join(toolDescs, "\n"), prompt)
}

// Act 执行任务
func (a *ReactAgent) Act(ctx context.Context, prompt string) (string, error) {
	a.log.Info("React 智能体开始执行任务")

	// 1. 构造系统提示词
	systemPrompt := `你是一个采用 ReAct（Reasoning and Acting）范式的智能助手。
请通过"思考-行动-观察"循环来解决问题。
每一步都要先思考，再行动，然后观察结果，最后决定下一步。`

	// 2. 初始化对话历史
	messages := []llm.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// 3. ReAct 循环
	maxSteps := 5 // 最大循环次数
	for step := 0; step < maxSteps; step++ {
		// 3.1 思考下一步行动
		llmResp, err := a.llmClient.GetCompletion(messages)
		if err != nil {
			return "", fmt.Errorf("LLM 调用失败: %v", err)
		}

		// 3.2 解析 LLM 输出
		thought, action, actionInput, err := a.parseReActOutput(llmResp.Choices[0].Message.Content)
		if err != nil {
			return "", fmt.Errorf("解析 LLM 输出失败: %v", err)
		}

		// 3.3 如果 LLM 认为任务完成，返回最终答案
		if action == "Final Answer" {
			return actionInput, nil
		}

		// 3.4 执行工具
		tool, err := a.tools.Get(action)
		if err != nil {
			return "", fmt.Errorf("未找到工具: %s", action)
		}

		result, err := tool.Run(ctx, actionInput)
		if err != nil {
			return "", fmt.Errorf("工具执行失败: %v", err)
		}

		// 3.5 将结果添加到对话历史
		messages = append(messages, llm.Message{
			Role:    "assistant",
			Content: fmt.Sprintf("Thought: %s\nAction: %s\nAction Input: %v", thought, action, actionInput),
		})
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: fmt.Sprintf("Observation: %v", result),
		})
	}

	return "达到最大步骤限制，任务未完成", nil
}

// parseReActOutput 解析 ReAct 格式的输出
func (a *ReactAgent) parseReActOutput(output string) (thought, action, actionInput string, err error) {
	// 提取 Thought
	thoughtRe := regexp.MustCompile(`Thought: (.*?)(?:\n|$)`)
	thoughtMatches := thoughtRe.FindStringSubmatch(output)
	if len(thoughtMatches) < 2 {
		return "", "", "", fmt.Errorf("无法解析 Thought: %s", output)
	}
	thought = thoughtMatches[1]

	// 提取 Action
	actionRe := regexp.MustCompile(`Action: (.*?)(?:\n|$)`)
	actionMatches := actionRe.FindStringSubmatch(output)
	if len(actionMatches) < 2 {
		return "", "", "", fmt.Errorf("无法解析 Action: %s", output)
	}
	action = actionMatches[1]

	// 提取 Action Input
	actionInputRe := regexp.MustCompile(`Action Input: (.*?)(?:\n|$)`)
	actionInputMatches := actionInputRe.FindStringSubmatch(output)
	if len(actionInputMatches) < 2 {
		return "", "", "", fmt.Errorf("无法解析 Action Input: %s", output)
	}
	actionInput = actionInputMatches[1]

	return thought, action, actionInput, nil
}
