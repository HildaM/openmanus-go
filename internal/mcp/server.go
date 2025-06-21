package mcp

import (
	"context"
	"net/http"

	"github.com/HildaM/openmanus-go/internal/agent"
	"github.com/HildaM/openmanus-go/pkg/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/sirupsen/logrus"
)

// Server MCP 服务端
type Server struct {
	mcp   *agent.MCPAgent
	log   *logrus.Logger
	hertz *server.Hertz
}

// NewServer 创建 MCP 服务端
func NewServer() *Server {
	h := server.New(server.WithHostPorts(":8080"))
	s := &Server{
		mcp:   agent.NewMCPAgent(),
		log:   logger.GetLogger(),
		hertz: h,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.hertz.POST("/run", s.handleRun)
}

// Run 启动 HTTP 服务
func (s *Server) Run(ctx context.Context) error {
	s.log.Info("MCP Hertz HTTP 服务启动中... 监听 :8080")
	s.hertz.Spin()
	return nil
}

// handleRun 触发 MCP 智能体运行
func (s *Server) handleRun(c context.Context, ctx *app.RequestContext) {
	go func() {
		s.log.Info("收到 /run 请求，开始执行 MCP 智能体")
		s.mcp.Run(context.Background())
	}()
	ctx.JSON(http.StatusOK, map[string]string{"status": "MCP 智能体已启动"})
}
