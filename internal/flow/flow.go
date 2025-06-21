package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/HildaM/openmanus-go/internal/agent"
	"github.com/HildaM/openmanus-go/internal/llm"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// Flow 定义任务流接口
type Flow interface {
	Execute(ctx context.Context, prompt string) (string, error)
}

// BaseFlow 基础任务流实现
type BaseFlow struct {
	agents    map[string]agent.Agent
	llmClient llm.LLMClient
}

// NewBaseFlow 创建基础任务流
func NewBaseFlow(agents map[string]agent.Agent, llmClient llm.LLMClient) *BaseFlow {
	return &BaseFlow{
		agents:    agents,
		llmClient: llmClient,
	}
}

// Execute 执行任务流
func (f *BaseFlow) Execute(ctx context.Context, prompt string) (string, error) {
	hlog.Infof("开始执行任务流，提示: %s", prompt)

	// 1. 分析任务，确定需要哪些智能体参与
	agents, err := f.analyzeTask(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("任务分析失败: %v", err)
	}

	// 2. 创建结果通道
	resultChan := make(chan agentResult, len(agents))
	errorChan := make(chan error, len(agents))

	// 3. 启动智能体协程
	var wg sync.WaitGroup
	for _, agentName := range agents {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			hlog.Infof("智能体 %s 开始执行", name)
			result, err := f.agents[name].Act(ctx, prompt)
			if err != nil {
				errorChan <- fmt.Errorf("智能体 %s 执行失败: %v", name, err)
				return
			}
			resultChan <- agentResult{
				agentName: name,
				result:    result,
			}
		}(agentName)
	}

	// 4. 等待所有智能体完成
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// 5. 收集结果和错误
	var results []agentResult
	var errors []error

	for result := range resultChan {
		results = append(results, result)
	}
	for err := range errorChan {
		errors = append(errors, err)
	}

	// 6. 如果有错误，返回第一个错误
	if len(errors) > 0 {
		return "", errors[0]
	}

	// 7. 整合结果
	finalResult, err := f.integrateResults(ctx, prompt, results)
	if err != nil {
		return "", fmt.Errorf("结果整合失败: %v", err)
	}

	return finalResult, nil
}

// agentResult 智能体执行结果
type agentResult struct {
	agentName string
	result    string
}

// analyzeTask 分析任务，确定需要哪些智能体参与
func (f *BaseFlow) analyzeTask(ctx context.Context, prompt string) ([]string, error) {
	// 构造系统提示词
	systemPrompt := `你是一个任务分析专家，需要分析用户任务并选择合适的智能体来执行。
请仔细分析任务的特点和需求，选择最合适的智能体组合。

可用的智能体及其特点：
1. manus（基础智能体）
   - 擅长使用工具解决复杂问题
   - 适合需要多工具协作的任务
   - 可以处理文件操作、代码生成等任务
   - 适合需要精确执行的任务

2. react（ReAct 智能体）
   - 采用"思考-行动-观察"循环
   - 适合需要多步骤推理的任务
   - 可以处理需要反复尝试的任务
   - 适合需要探索性解决的任务

3. browser（网页操作智能体）
   - 擅长网页内容抓取和分析
   - 适合需要获取网页信息的任务
   - 可以处理网页交互和自动化操作
   - 适合需要网页数据提取的任务

请分析任务，并返回需要参与的智能体列表（JSON 格式）。
返回格式示例：
{
    "agents": ["manus", "react"],
    "reason": "任务需要多步骤推理和工具使用"
}`

	// 调用 LLM 分析任务
	llmResp, err := f.llmClient.GetCompletion([]llm.Message{
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
		return nil, fmt.Errorf("LLM 调用失败: %v", err)
	}

	// 解析 LLM 输出
	var response struct {
		Agents []string `json:"agents"`
		Reason string   `json:"reason"`
	}
	if err := json.Unmarshal([]byte(llmResp.Choices[0].Message.Content), &response); err != nil {
		return nil, fmt.Errorf("解析智能体列表失败: %v", err)
	}

	// 验证智能体是否存在
	for _, agentName := range response.Agents {
		if _, ok := f.agents[agentName]; !ok {
			return nil, fmt.Errorf("未知的智能体: %s", agentName)
		}
	}

	// 记录选择原因
	hlog.Infof("智能体选择原因: %s", response.Reason)

	return response.Agents, nil
}

// integrateResults 整合多个智能体的执行结果
func (f *BaseFlow) integrateResults(ctx context.Context, prompt string, results []agentResult) (string, error) {
	// 构造系统提示词
	systemPrompt := `你是一个结果整合专家，需要整合多个智能体的执行结果，给出最终答案。
请仔细分析每个智能体的输出，提取有用信息，并给出一个完整的、连贯的回答。`

	// 构造结果提示词
	var resultsText strings.Builder
	for _, r := range results {
		resultsText.WriteString(fmt.Sprintf("%s 的结果：\n%s\n\n", r.agentName, r.result))
	}

	// 调用 LLM 整合结果
	llmResp, err := f.llmClient.GetCompletion([]llm.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("原始任务：%s\n\n各智能体执行结果：\n%s", prompt, resultsText.String()),
		},
	})
	if err != nil {
		return "", fmt.Errorf("LLM 调用失败: %v", err)
	}

	return llmResp.Choices[0].Message.Content, nil
}
