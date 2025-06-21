package mcp

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/HildaM/openmanus-go/internal/mcp"
)

var (
	port    = flag.Int("port", 8080, "MCP 服务端口")
	timeout = flag.Int("timeout", 30, "执行超时时间(秒)")
	debug   = flag.Bool("debug", false, "是否开启调试模式")
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
		log.Println("正在关闭 MCP 服务...")
		cancel()
	}()

	// 创建 MCP 服务
	server := mcp.NewServer()

	// 启动服务
	log.Printf("MCP 服务启动中... 监听端口: %d", *port)
	return server.Run(ctx)
}
