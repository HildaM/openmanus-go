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

// DataAnalysisAgent DataAnalysis 智能体
type DataAnalysisAgent struct {
	*BaseAgent
	log   *logrus.Logger
	tools *tool.ToolCollection
}

// NewDataAnalysisAgent 创建 DataAnalysis 智能体
func NewDataAnalysisAgent(llmClient llm.LLMClient) *DataAnalysisAgent {
	base := NewBaseAgent("DataAnalysisAgent", llmClient, tool.NewToolCollection())
	agent := &DataAnalysisAgent{
		BaseAgent: base,
		log:       logrus.New(),
		tools:     tool.NewToolCollection(),
	}
	// 注册数据分析相关工具
	agent.tools.Register("python", &tool.AITool{})
	agent.tools.Register("data_analysis", &tool.DataAnalysisTool{})
	return agent
}

// Run 运行智能体
func (a *DataAnalysisAgent) Run(ctx context.Context) error {
	hlog.Infof("DataAnalysisAgent 智能体开始运行")
	a.State = Running
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(step int) {
			defer wg.Done()
			if err := a.Step(ctx, step); err != nil {
				hlog.Errorf("DataAnalysisAgent 第 %d 步失败: %v", step, err)
			}
		}(i)
	}
	wg.Wait()
	a.State = Finished
	return nil
}

// Step 执行单步
func (a *DataAnalysisAgent) Step(ctx context.Context, step int) error {
	hlog.Infof("DataAnalysisAgent 执行第 %d 步", step)
	if err := a.Think(ctx, step); err != nil {
		return err
	}
	_, err := a.Act(ctx, fmt.Sprintf("执行第 %d 步", step))
	return err
}

// Think 思考
func (a *DataAnalysisAgent) Think(ctx context.Context, step int) error {
	hlog.Infof("DataAnalysisAgent 思考第 %d 步", step)
	return nil
}

// Act 行动
func (a *DataAnalysisAgent) Act(ctx context.Context, prompt string) (string, error) {
	hlog.Infof("DataAnalysisAgent 行动中...")
	// 示例：调用数据分析工具
	dataTool, err := a.tools.Get("data_analysis")
	if err != nil {
		return "", err
	}
	result, err := dataTool.Run(ctx, map[string]interface{}{"operation": "basic_analysis", "data": "示例数据"})
	if err != nil {
		return "", err
	}
	hlog.Infof("DataAnalysisAgent 工具执行结果: %v", result)
	return fmt.Sprintf("DataAnalysisAgent 执行结果: %v", result), nil
}

// GetName 获取智能体名称
func (a *DataAnalysisAgent) GetName() string {
	return a.Name
}
