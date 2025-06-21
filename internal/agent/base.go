package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/HildaM/openmanus-go/internal/llm"
	"github.com/HildaM/openmanus-go/internal/tool"
)

// AgentState 智能体状态
type AgentState int

const (
	Idle AgentState = iota
	Initializing
	Running
	Paused
	Resuming
	Finished
	Error
	Terminated
)

// Message 智能体消息
type Message struct {
	Role    string // user, system, assistant, tool
	Content string
}

// Memory 智能体记忆
type Memory struct {
	Messages []Message
}

func (m *Memory) AddMessage(msg Message) {
	m.Messages = append(m.Messages, msg)
}

func (m *Memory) LastMessage() *Message {
	if len(m.Messages) == 0 {
		return nil
	}
	return &m.Messages[len(m.Messages)-1]
}

// AsyncAgent 异步智能体接口
type AsyncAgent interface {
	Agent
	RunAsync(ctx context.Context) <-chan error
	StepAsync(ctx context.Context) <-chan error
	ThinkAsync(ctx context.Context) <-chan error
	ActAsync(ctx context.Context) <-chan error
}

// AgentEvent 智能体事件
type AgentEvent struct {
	Type      string
	Timestamp time.Time
	Data      interface{}
}

// AgentLifecycle 生命周期钩子
type AgentLifecycle struct {
	OnInit     func(ctx context.Context) error
	OnStart    func(ctx context.Context) error
	OnPause    func(ctx context.Context) error
	OnResume   func(ctx context.Context) error
	OnStop     func(ctx context.Context) error
	OnError    func(ctx context.Context, err error) error
	OnComplete func(ctx context.Context) error
}

// Agent 定义智能体接口
type Agent interface {
	// Act 执行任务
	Act(ctx context.Context, prompt string) (string, error)
	// GetName 获取智能体名称
	GetName() string
}

// BaseAgent 通用基础智能体结构体
type BaseAgent struct {
	Name        string
	Description string
	State       AgentState
	MaxSteps    int
	CurrentStep int
	Memory      Memory
	mu          sync.Mutex
	lifecycle   *AgentLifecycle
	eventChan   chan AgentEvent
	startTime   time.Time
	endTime     time.Time
	lastError   error
	llmClient   llm.LLMClient
	tools       *tool.ToolCollection
}

// BaseAgent 默认实现
func (a *BaseAgent) Step(ctx context.Context) (string, error) {
	shouldAct, err := a.Think(ctx)
	if err != nil {
		return "", err
	}
	if !shouldAct {
		return "Thinking complete - no action needed", nil
	}
	return a.Act(ctx, "")
}

func (a *BaseAgent) Think(ctx context.Context) (bool, error) {
	return false, nil // 默认不行动，具体 Agent 重写
}

// Act 执行任务
func (a *BaseAgent) Act(ctx context.Context, prompt string) (string, error) {
	// 1. 构造 LLM prompt
	llmPrompt := a.buildPrompt(prompt)

	// 2. 调用 LLM
	llmResp, err := a.llmClient.GetCompletion([]llm.Message{
		{
			Role:    "user",
			Content: llmPrompt,
		},
	})
	if err != nil {
		return "", fmt.Errorf("LLM 调用失败: %v", err)
	}

	// 3. 解析 LLM 输出
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

	// 6. 将结果反馈给 LLM，进入下一轮
	resultStr, ok := result.(string)
	if !ok {
		resultStr = fmt.Sprintf("%v", result)
	}
	return a.handleToolResult(ctx, prompt, resultStr)
}

// buildPrompt 构造 LLM prompt
func (a *BaseAgent) buildPrompt(prompt string) string {
	// 获取可用工具列表
	toolList := a.tools.List()
	toolDescs := make([]string, 0, len(toolList))
	for name, desc := range toolList {
		toolDescs = append(toolDescs, fmt.Sprintf("- %s: %s", name, desc))
	}

	// 构造 ReAct 风格 prompt
	return fmt.Sprintf(`你是一个智能助手，可以调用工具完成任务。请根据用户输入，选择合适的工具并执行。

可用工具：
%s

请按照以下格式输出：
Thought: 思考下一步行动
Action: 工具名称
Action Input: {"参数1": "值1", "参数2": "值2"}

用户输入：%s`, strings.Join(toolDescs, "\n"), prompt)
}

// parseLLMAction 解析 LLM 输出
func (a *BaseAgent) parseLLMAction(llmOutput string) (string, map[string]interface{}, error) {
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

// handleToolResult 处理工具执行结果
func (a *BaseAgent) handleToolResult(ctx context.Context, prompt, result string) (string, error) {
	// 构造包含工具结果的 prompt
	resultPrompt := fmt.Sprintf(`工具执行结果：%s

请根据工具执行结果，决定下一步行动。如果需要继续调用工具，请按照以下格式输出：
Thought: 思考下一步行动
Action: 工具名称
Action Input: {"参数1": "值1", "参数2": "值2"}

如果任务已完成，请直接输出最终结果。`, result)

	// 调用 LLM 处理结果
	llmResp, err := a.llmClient.GetCompletion([]llm.Message{
		{
			Role:    "user",
			Content: resultPrompt,
		},
	})
	if err != nil {
		return "", fmt.Errorf("LLM 调用失败: %v", err)
	}

	// 如果输出包含 Action，说明需要继续调用工具
	if strings.Contains(llmResp.Choices[0].Message.Content, "Action:") {
		return a.Act(ctx, llmResp.Choices[0].Message.Content)
	}

	// 否则返回最终结果
	return llmResp.Choices[0].Message.Content, nil
}

// IsStuck 检查是否卡住（重复消息）
func (a *BaseAgent) IsStuck(duplicateThreshold int) bool {
	if len(a.Memory.Messages) < 2 {
		return false
	}
	last := a.Memory.Messages[len(a.Memory.Messages)-1].Content
	count := 0
	for i := len(a.Memory.Messages) - 2; i >= 0; i-- {
		if a.Memory.Messages[i].Content == last {
			count++
			if count >= duplicateThreshold {
				return true
			}
		}
	}
	return false
}

// NewBaseAgent 创建基础智能体
func NewBaseAgent(name string, llmClient llm.LLMClient, tools *tool.ToolCollection) *BaseAgent {
	return &BaseAgent{
		Name:        name,
		State:       Idle,
		MaxSteps:    10,
		CurrentStep: 0,
		Memory:      *NewMemory(),
		lifecycle:   &AgentLifecycle{},
		eventChan:   make(chan AgentEvent, 100),
		llmClient:   llmClient,
		tools:       tools,
	}
}

// NewMemory 创建内存
func NewMemory() *Memory {
	return &Memory{
		Messages: make([]Message, 0),
	}
}

// RunAsync 异步执行智能体任务
func (a *BaseAgent) RunAsync(ctx context.Context) <-chan error {
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		if err := a.Run(ctx); err != nil {
			errChan <- err
		}
	}()
	return errChan
}

// StepAsync 异步执行单步
func (a *BaseAgent) StepAsync(ctx context.Context) <-chan error {
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		if _, err := a.Step(ctx); err != nil {
			errChan <- err
		}
	}()
	return errChan
}

// ThinkAsync 异步思考
func (a *BaseAgent) ThinkAsync(ctx context.Context) <-chan error {
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		if _, err := a.Think(ctx); err != nil {
			errChan <- err
		}
	}()
	return errChan
}

// ActAsync 异步行动
func (a *BaseAgent) ActAsync(ctx context.Context) <-chan error {
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		if _, err := a.Act(ctx, ""); err != nil {
			errChan <- err
		}
	}()
	return errChan
}

// SetLifecycle 设置生命周期钩子
func (a *BaseAgent) SetLifecycle(lifecycle *AgentLifecycle) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lifecycle = lifecycle
}

// EmitEvent 发送事件
func (a *BaseAgent) EmitEvent(eventType string, data interface{}) {
	a.eventChan <- AgentEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// GetEvents 获取事件通道
func (a *BaseAgent) GetEvents() <-chan AgentEvent {
	return a.eventChan
}

// SetState 设置状态（线程安全）
func (a *BaseAgent) SetState(state AgentState) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 状态转换规则
	switch a.State {
	case Idle:
		if state != Initializing && state != Terminated {
			return fmt.Errorf("invalid state transition from Idle to %v", state)
		}
	case Initializing:
		if state != Running && state != Error {
			return fmt.Errorf("invalid state transition from Initializing to %v", state)
		}
	case Running:
		if state != Paused && state != Finished && state != Error {
			return fmt.Errorf("invalid state transition from Running to %v", state)
		}
	case Paused:
		if state != Resuming && state != Terminated {
			return fmt.Errorf("invalid state transition from Paused to %v", state)
		}
	case Resuming:
		if state != Running && state != Error {
			return fmt.Errorf("invalid state transition from Resuming to %v", state)
		}
	case Finished, Error, Terminated:
		return fmt.Errorf("cannot transition from terminal state %v", a.State)
	}

	// 执行状态转换
	oldState := a.State
	a.State = state

	// 触发状态转换事件
	a.EmitEvent("state_change", map[string]interface{}{
		"old_state": oldState,
		"new_state": state,
	})

	// 执行生命周期钩子
	switch state {
	case Initializing:
		if a.lifecycle.OnInit != nil {
			if err := a.lifecycle.OnInit(context.Background()); err != nil {
				a.lastError = err
				return err
			}
		}
	case Running:
		if a.lifecycle.OnStart != nil {
			if err := a.lifecycle.OnStart(context.Background()); err != nil {
				a.lastError = err
				return err
			}
		}
		a.startTime = time.Now()
	case Paused:
		if a.lifecycle.OnPause != nil {
			if err := a.lifecycle.OnPause(context.Background()); err != nil {
				a.lastError = err
				return err
			}
		}
	case Resuming:
		if a.lifecycle.OnResume != nil {
			if err := a.lifecycle.OnResume(context.Background()); err != nil {
				a.lastError = err
				return err
			}
		}
	case Finished:
		if a.lifecycle.OnComplete != nil {
			if err := a.lifecycle.OnComplete(context.Background()); err != nil {
				a.lastError = err
				return err
			}
		}
		a.endTime = time.Now()
	case Error:
		if a.lifecycle.OnError != nil {
			if err := a.lifecycle.OnError(context.Background(), a.lastError); err != nil {
				return err
			}
		}
	case Terminated:
		if a.lifecycle.OnStop != nil {
			if err := a.lifecycle.OnStop(context.Background()); err != nil {
				return err
			}
		}
		a.endTime = time.Now()
	}

	return nil
}

// GetState 获取状态（线程安全）
func (a *BaseAgent) GetState() AgentState {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.State
}

// GetExecutionTime 获取执行时间
func (a *BaseAgent) GetExecutionTime() time.Duration {
	if a.startTime.IsZero() {
		return 0
	}
	end := a.endTime
	if end.IsZero() {
		end = time.Now()
	}
	return end.Sub(a.startTime)
}

// GetLastError 获取最后的错误
func (a *BaseAgent) GetLastError() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.lastError
}

// Run 执行智能体任务
func (a *BaseAgent) Run(ctx context.Context) error {
	if err := a.SetState(Initializing); err != nil {
		return err
	}
	if err := a.SetState(Running); err != nil {
		return err
	}

	for a.CurrentStep = 0; a.CurrentStep < a.MaxSteps; a.CurrentStep++ {
		select {
		case <-ctx.Done():
			return a.SetState(Terminated)
		default:
			shouldAct, err := a.Think(ctx)
			if err != nil {
				a.lastError = err
				return a.SetState(Error)
			}
			if !shouldAct {
				continue
			}
			result, err := a.Act(ctx, "")
			if err != nil {
				a.lastError = err
				return a.SetState(Error)
			}
			// 写入 memory
			a.Memory.AddMessage(Message{
				Role:    "assistant",
				Content: result,
			})
			// 检查是否卡住
			if a.IsStuck(2) {
				return a.SetState(Finished)
			}
		}
	}
	return a.SetState(Finished)
}

// GetName 获取智能体名称
func (a *BaseAgent) GetName() string {
	return a.Name
}
