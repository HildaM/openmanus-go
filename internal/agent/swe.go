package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/HildaM/openmanus-go/internal/llm"
	"github.com/HildaM/openmanus-go/internal/tool"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/sirupsen/logrus"
)

// SWEAgent SWE 智能体
type SWEAgent struct {
	*BaseAgent
	log   *logrus.Logger
	tools *tool.ToolCollection
}

// NewSWEAgent 创建 SWE 智能体
func NewSWEAgent(llmClient llm.LLMClient) *SWEAgent {
	base := NewBaseAgent("SWEAgent", llmClient, tool.NewToolCollection())
	agent := &SWEAgent{
		BaseAgent: base,
		log:       logrus.New(),
		tools:     tool.NewToolCollection(),
	}
	// 注册代码相关工具
	agent.tools.Register("shell", &tool.ShellTool{})
	agent.tools.Register("file", &tool.FileTool{})
	// TODO: 可扩展注册 PythonExecute、StrReplaceEditor 等
	return agent
}

// Run 运行智能体
func (a *SWEAgent) Run(ctx context.Context) error {
	hlog.Infof("SWEAgent 智能体开始运行")
	a.State = Running
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(step int) {
			defer wg.Done()
			if err := a.Step(ctx, step); err != nil {
				hlog.Errorf("SWEAgent 第 %d 步失败: %v", step, err)
			}
		}(i)
	}
	wg.Wait()
	a.State = Finished
	return nil
}

// Step 执行单步
func (a *SWEAgent) Step(ctx context.Context, step int) error {
	hlog.Infof("SWEAgent 执行第 %d 步", step)
	if err := a.Think(ctx, step); err != nil {
		return err
	}
	_, err := a.Act(ctx, fmt.Sprintf("执行第 %d 步", step))
	return err
}

// Think 思考
func (a *SWEAgent) Think(ctx context.Context, step int) error {
	hlog.Infof("SWEAgent 思考第 %d 步", step)
	return nil
}

// Act 行动
func (a *SWEAgent) Act(ctx context.Context, prompt string) (string, error) {
	hlog.Infof("SWEAgent 行动中...")
	// 示例：调用 shell 工具
	shellTool, err := a.tools.Get("shell")
	if err != nil {
		return "", err
	}
	result, err := shellTool.Run(ctx, "echo 'SWEAgent 行动'")
	if err != nil {
		return "", err
	}
	hlog.Infof("SWEAgent 工具执行结果: %v", result)
	return fmt.Sprintf("SWEAgent 执行结果: %v", result), nil
}

// GetName 获取智能体名称
func (a *SWEAgent) GetName() string {
	return a.Name
}
