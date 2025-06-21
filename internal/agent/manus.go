package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/HildaM/openmanus-go/internal/llm"
	"github.com/HildaM/openmanus-go/internal/tool"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/sirupsen/logrus"
)

// ManusAgent Manus 智能体
type ManusAgent struct {
	*BaseAgent
	log   *logrus.Logger
	tools *tool.ToolCollection
}

// NewManusAgent 创建 Manus 智能体
func NewManusAgent(llmClient llm.LLMClient, tools *tool.ToolCollection) *ManusAgent {
	return &ManusAgent{
		BaseAgent: NewBaseAgent("manus", llmClient, tools),
	}
}

// Run 运行智能体
func (a *ManusAgent) Run(ctx context.Context) error {
	hlog.Infof("Manus 智能体开始运行")
	a.State = Running

	// 模拟任务调度
	taskChan := make(chan string, 3)
	taskChan <- "task1"
	taskChan <- "task2"
	taskChan <- "task3"
	close(taskChan)

	var wg sync.WaitGroup
	for task := range taskChan {
		wg.Add(1)
		go func(task string) {
			defer wg.Done()
			hlog.Infof("执行任务: %s", task)
			if _, err := a.Step(ctx); err != nil {
				hlog.Errorf("任务 %s 执行失败: %v", task, err)
			}
		}(task)
	}
	wg.Wait()

	a.State = Finished
	return nil
}

// Step 执行单步
func (a *ManusAgent) Step(ctx context.Context) (string, error) {
	hlog.Infof("Manus 智能体执行单步")
	shouldAct, err := a.Think(ctx)
	if err != nil {
		return "", err
	}
	if !shouldAct {
		return "Thinking complete - no action needed", nil
	}
	return a.Act(ctx, "")
}

// Think 思考
func (a *ManusAgent) Think(ctx context.Context) (bool, error) {
	hlog.Infof("Manus 智能体思考中...")
	// 模拟思考过程
	return true, nil
}

// Act 行动
func (a *ManusAgent) Act(ctx context.Context, prompt string) (string, error) {
	hlog.Infof("Manus 智能体行动中...")

	// 1. 构造系统提示词
	systemPrompt := `你是一个强大的智能助手，擅长使用工具解决复杂问题。
请仔细分析用户需求，选择合适的工具并执行。
如果任务需要多个步骤，请一步一步来，确保每个步骤都正确执行。`

	// 2. 调用 LLM 分析任务
	llmResp, err := a.llmClient.GetCompletion([]llm.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	})
	if err != nil {
		return "", fmt.Errorf("LLM 调用失败: %v", err)
	}

	// 3. 解析 LLM 输出，提取需要使用的工具和参数
	action, actionInput, err := a.parseLLMAction(llmResp.Choices[0].Message.Content)
	if err != nil {
		return "", fmt.Errorf("解析 LLM 输出失败: %v", err)
	}

	// 4. 查找并调用工具
	tool, err := a.tools.Get(action)
	if err != nil {
		return "", fmt.Errorf("未找到工具: %s", action)
	}

	// 5. 执行工具
	result, err := tool.Run(ctx, actionInput)
	if err != nil {
		return "", fmt.Errorf("工具执行失败: %v", err)
	}

	// 6. 将结果反馈给 LLM，获取最终答案
	finalResp, err := a.llmClient.GetCompletion([]llm.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("任务: %s\n工具执行结果: %v\n请给出最终答案。", prompt, result),
		},
	})
	if err != nil {
		return "", fmt.Errorf("LLM 调用失败: %v", err)
	}

	return finalResp.Choices[0].Message.Content, nil
}

// parseLLMAction 解析 LLM 输出，提取工具名称和参数
func (a *ManusAgent) parseLLMAction(llmOutput string) (string, map[string]interface{}, error) {
	// 提取 Action 和 Action Input
	re := regexp.MustCompile(`Action: (\w+)\nAction Input: (.*)`)
	matches := re.FindStringSubmatch(llmOutput)
	if len(matches) < 3 {
		return "", nil, fmt.Errorf("无法解析 LLM 输出: %s", llmOutput)
	}

	action := matches[1]
	actionInput := matches[2]

	// 解析 JSON 参数
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(actionInput), &params); err != nil {
		return "", nil, fmt.Errorf("解析工具参数失败: %v", err)
	}

	return action, params, nil
}

// GetName 获取智能体名称
func (a *ManusAgent) GetName() string {
	return a.Name
}

// RunAsync 异步运行智能体
func (a *ManusAgent) RunAsync(ctx context.Context) <-chan error {
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		hlog.Infof("Manus 智能体开始异步运行")
		a.SetState(Running)

		// 模拟任务调度
		taskChan := make(chan string, 3)
		taskChan <- "task1"
		taskChan <- "task2"
		taskChan <- "task3"
		close(taskChan)

		var wg sync.WaitGroup
		stepErrChan := make(chan error, len(taskChan))

		for task := range taskChan {
			wg.Add(1)
			go func(task string) {
				defer wg.Done()
				hlog.Infof("异步执行任务: %s", task)
				if err := <-a.StepAsync(ctx); err != nil {
					hlog.Errorf("任务 %s 执行失败: %v", task, err)
					stepErrChan <- err
				}
			}(task)
		}

		// 等待所有任务完成
		wg.Wait()
		close(stepErrChan)

		// 检查是否有错误
		for err := range stepErrChan {
			if err != nil {
				errChan <- err
				return
			}
		}

		a.SetState(Finished)
	}()
	return errChan
}

// StepAsync 异步执行单步
func (a *ManusAgent) StepAsync(ctx context.Context) <-chan error {
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		hlog.Infof("Manus 智能体异步执行单步")

		// 异步思考
		if err := <-a.ThinkAsync(ctx); err != nil {
			errChan <- err
			return
		}

		// 异步行动
		if err := <-a.ActAsync(ctx); err != nil {
			errChan <- err
			return
		}
	}()
	return errChan
}

// ThinkAsync 异步思考
func (a *ManusAgent) ThinkAsync(ctx context.Context) <-chan error {
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		hlog.Infof("Manus 智能体异步思考中...")
		// 模拟思考过程
	}()
	return errChan
}

// ActAsync 异步行动
func (a *ManusAgent) ActAsync(ctx context.Context) <-chan error {
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		hlog.Infof("Manus 智能体异步行动中...")

		// 模拟工具调用
		shellTool, err := a.tools.Get("shell")
		if err != nil {
			errChan <- err
			return
		}

		result, err := shellTool.Run(ctx, "echo 'Hello, World!'")
		if err != nil {
			errChan <- err
			return
		}
		hlog.Infof("工具执行结果: %s", result)
	}()
	return errChan
}

// buildPrompt 定制 Manus 智能体的 prompt
func (m *ManusAgent) buildPrompt(prompt string) string {
	// 获取可用工具列表
	toolList := m.tools.List()
	toolDescs := make([]string, 0, len(toolList))
	for name, desc := range toolList {
		toolDescs = append(toolDescs, fmt.Sprintf("- %s: %s", name, desc))
	}

	// 构造 Manus 风格的 prompt
	return fmt.Sprintf(`你是一个强大的智能助手，擅长使用工具解决复杂问题。请根据用户输入，选择合适的工具并执行。

可用工具：
%s

请按照以下格式输出：
Thought: 思考下一步行动
Action: 工具名称
Action Input: {"参数1": "值1", "参数2": "值2"}

用户输入：%s`, strings.Join(toolDescs, "\n"), prompt)
}
