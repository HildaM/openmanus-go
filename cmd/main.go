package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/HildaM/openmanus-go/internal/agent"
	"github.com/HildaM/openmanus-go/internal/config"
	"github.com/HildaM/openmanus-go/internal/flow"
	"github.com/HildaM/openmanus-go/internal/llm"
	"github.com/HildaM/openmanus-go/internal/tool"
)

var (
	configPath  string
	mode        string
	interactive bool
)

func init() {
	flag.StringVar(&configPath, "config", "config/config.toml", "配置文件路径")
	flag.StringVar(&mode, "mode", "flow", "运行模式: flow 或 mcp")
	flag.BoolVar(&interactive, "i", false, "启用交互模式")
	flag.Parse()
}

func main() {
	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("接收到终止信号，正在关闭服务...")
		cancel()
	}()

	// 注册工具
	tool.RegisterDefaultTools()

	if interactive {
		runInteractive(ctx, cfg)
	} else {
		// 根据模式启动服务
		switch mode {
		case "flow":
			if err := runFlow(ctx, cfg, "执行多智能体协作任务"); err != nil {
				log.Fatalf("运行任务流失败: %v", err)
			}
		case "mcp":
			if err := runMCP(ctx, cfg, "执行 MCP 任务"); err != nil {
				log.Fatalf("运行 MCP 服务失败: %v", err)
			}
		default:
			log.Fatalf("不支持的运行模式: %s", mode)
		}
	}
}

// runInteractive 运行交互式命令行界面
func runInteractive(ctx context.Context, cfg *config.Config) {
	fmt.Println("欢迎使用 OpenManus！")
	fmt.Println("可用命令：")
	fmt.Println("  /mode flow    - 切换到多智能体协作模式")
	fmt.Println("  /mode mcp     - 切换到工具模式")
	fmt.Println("  /help         - 显示帮助信息")
	fmt.Println("  /exit         - 退出程序")
	fmt.Println("直接输入任务描述即可开始执行任务")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	currentMode := mode

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("读取输入失败: %v", err)
			continue
		}
		input = strings.TrimSpace(input)

		switch {
		case input == "/exit":
			fmt.Println("再见！")
			return
		case input == "/help":
			printHelp()
		case strings.HasPrefix(input, "/mode "):
			newMode := strings.TrimPrefix(input, "/mode ")
			if newMode == "flow" || newMode == "mcp" {
				currentMode = newMode
				fmt.Printf("已切换到 %s 模式\n", newMode)
			} else {
				fmt.Println("不支持的模式，可用模式：flow, mcp")
			}
		case input == "":
			continue
		default:
			// 执行任务
			if err := executeTask(ctx, cfg, currentMode, input); err != nil {
				log.Printf("执行任务失败: %v", err)
			}
		}
	}
}

// executeTask 执行任务
func executeTask(ctx context.Context, cfg *config.Config, mode string, task string) error {
	switch mode {
	case "flow":
		return runFlow(ctx, cfg, task)
	case "mcp":
		return runMCP(ctx, cfg, task)
	default:
		return fmt.Errorf("不支持的模式: %s", mode)
	}
}

// runFlow 运行任务流
func runFlow(ctx context.Context, cfg *config.Config, task string) error {
	// 创建 LLM 客户端
	llmClient := llm.NewClient(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.MaxTokens, cfg.LLM.Temperature)

	// 创建工具集合
	tools := tool.NewToolCollection()
	tool.RegisterDefaultTools()

	// 创建智能体
	agents := map[string]agent.Agent{
		"manus": agent.NewManusAgent(llmClient, tools),
		"react": agent.NewReactAgent(llmClient, tools),
	}

	// 创建任务流
	factory := flow.NewDefaultFlowFactory()
	taskFlow, err := factory.CreateFlow(agents, llmClient)
	if err != nil {
		return err
	}

	// 执行任务流
	result, err := taskFlow.Execute(ctx, task)
	if err != nil {
		return err
	}

	fmt.Printf("任务执行完成: %s\n", result)
	return nil
}

// runMCP 运行 MCP 服务
func runMCP(ctx context.Context, cfg *config.Config, task string) error {
	// TODO: 实现 MCP 服务启动逻辑
	fmt.Printf("MCP 模式执行任务: %s\n", task)
	return nil
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println("\nOpenManus 帮助信息：")
	fmt.Println("1. 多智能体协作模式 (flow)")
	fmt.Println("   - 适合复杂任务，系统会自动协调多个智能体")
	fmt.Println("   - 支持的任务类型：")
	fmt.Println("     * 代码生成和修改")
	fmt.Println("     * 文件操作")
	fmt.Println("     * 数据分析")
	fmt.Println("     * 网页操作")
	fmt.Println()
	fmt.Println("2. 工具模式 (mcp)")
	fmt.Println("   - 适合直接使用工具完成任务")
	fmt.Println("   - 可用工具：")
	fmt.Println("     * 文件操作")
	fmt.Println("     * 字符串处理")
	fmt.Println("     * Python 执行")
	fmt.Println()
	fmt.Println("示例任务：")
	fmt.Println("- 创建一个新的 Go 项目")
	fmt.Println("- 分析当前目录下的代码")
	fmt.Println("- 修改指定文件的内容")
	fmt.Println("- 执行 Python 脚本")
	fmt.Println()
}
