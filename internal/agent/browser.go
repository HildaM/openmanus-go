package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sync"

	"github.com/HildaM/openmanus-go/internal/llm"
	"github.com/HildaM/openmanus-go/internal/tool"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/sirupsen/logrus"
)

// BrowserAgent 浏览器智能体
type BrowserAgent struct {
	*BaseAgent
	log   *logrus.Logger
	tools *tool.ToolCollection
}

// NewBrowserAgent 创建浏览器智能体
func NewBrowserAgent(llmClient llm.LLMClient) *BrowserAgent {
	return &BrowserAgent{
		BaseAgent: NewBaseAgent("browser", llmClient, tool.NewToolCollection()),
	}
}

// Run 运行智能体
func (a *BrowserAgent) Run(ctx context.Context) error {
	hlog.Infof("BrowserAgent 智能体开始运行")
	a.State = Running
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(step int) {
			defer wg.Done()
			if err := a.Step(ctx, step); err != nil {
				hlog.Errorf("BrowserAgent 第 %d 步失败: %v", step, err)
			}
		}(i)
	}
	wg.Wait()
	a.State = Finished
	return nil
}

// Step 执行单步
func (a *BrowserAgent) Step(ctx context.Context, step int) error {
	hlog.Infof("BrowserAgent 执行第 %d 步", step)
	if err := a.Think(ctx, step); err != nil {
		return err
	}
	_, err := a.Act(ctx, fmt.Sprintf("执行第 %d 步", step))
	return err
}

// Think 思考
func (a *BrowserAgent) Think(ctx context.Context, step int) error {
	hlog.Infof("BrowserAgent 思考第 %d 步", step)
	return nil
}

// Act 行动
func (a *BrowserAgent) Act(ctx context.Context, prompt string) (string, error) {
	hlog.Infof("BrowserAgent 行动中...")

	// 1. 构造系统提示词
	systemPrompt := `你是一个网页操作专家，擅长使用浏览器工具完成任务。
你可以：
1. 抓取网页内容
2. 提取特定信息
3. 分析网页结构
4. 执行网页操作

请根据用户需求，选择合适的工具并执行。`

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
func (a *BrowserAgent) parseLLMAction(llmOutput string) (string, map[string]interface{}, error) {
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
func (a *BrowserAgent) GetName() string {
	return a.Name
}
