package flow

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/HildaM/openmanus-go/internal/agent"
	"github.com/HildaM/openmanus-go/internal/config"
	"github.com/HildaM/openmanus-go/internal/flow"
	"github.com/HildaM/openmanus-go/internal/llm"
	"github.com/HildaM/openmanus-go/internal/tool"
)

var (
	timeout  = flag.Int("timeout", 30, "执行超时时间(秒)")
	debug    = flag.Bool("debug", false, "是否开启调试模式")
	language = flag.String("language", "zh", "提示词语言(zh/en)")
)

func Run() error {
	// 解析命令行参数
	flag.Parse()

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("正在关闭任务流...")
		cancel()
	}()

	// 加载配置
	cfg, err := config.LoadConfig("config/config.toml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建 LLM 客户端和工具集合
	llmClient := llm.NewClient(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.MaxTokens, cfg.LLM.Temperature)

	tools := tool.NewToolCollection()

	// 创建智能体
	mcpAgent := agent.NewMCPAgent()
	manusAgent := agent.NewManusAgent(llmClient, tools)
	reactAgent := agent.NewReactAgent(llmClient, tools)
	sweAgent := agent.NewSWEAgent(llmClient)
	browserAgent := agent.NewBrowserAgent(llmClient)
	dataAnalysisAgent := agent.NewDataAnalysisAgent(llmClient)

	// 组合智能体
	agents := map[string]agent.Agent{
		"mcp":           mcpAgent,
		"manus":         manusAgent,
		"react":         reactAgent,
		"swe":           sweAgent,
		"browser":       browserAgent,
		"data_analysis": dataAnalysisAgent,
	}

	// 创建任务流
	taskFlow := flow.NewBaseFlow(agents, llmClient)

	// 执行任务流
	log.Println("开始执行多智能体协作任务...")
	result, err := taskFlow.Execute(ctx, "执行多智能体协作任务")
	if err != nil {
		return err
	}

	log.Printf("任务流执行完成: %s", result)
	return nil
}
