package tool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// PythonServiceConfig Python 服务配置
type PythonServiceConfig struct {
	Address     string
	ServiceName string
	Timeout     time.Duration
	MaxRetries  int
}

// PythonServiceClient Python 服务客户端
type PythonServiceClient struct {
	conn   *grpc.ClientConn
	config PythonServiceConfig
	mu     sync.RWMutex
}

// NewPythonServiceClient 创建 Python 服务客户端
func NewPythonServiceClient(config PythonServiceConfig) (*PythonServiceClient, error) {
	// 设置 gRPC 连接选项
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	// 建立连接
	conn, err := grpc.Dial(config.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Python service: %v", err)
	}

	return &PythonServiceClient{
		conn:   conn,
		config: config,
	}, nil
}

// Close 关闭连接
func (c *PythonServiceClient) Close() error {
	return c.conn.Close()
}

// Call 调用 Python 服务
func (c *PythonServiceClient) Call(ctx context.Context, method string, request interface{}) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	// 实现重试逻辑
	var lastErr error
	for i := 0; i < c.config.MaxRetries; i++ {
		// TODO: 实现具体的 gRPC 调用
		// 这里需要根据实际的 protobuf 定义来实现
		return nil, fmt.Errorf("not implemented")
	}

	return nil, lastErr
}

// ServiceRegistry 服务注册表
type ServiceRegistry struct {
	services map[string]*PythonServiceClient
	mu       sync.RWMutex
}

// NewServiceRegistry 创建服务注册表
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]*PythonServiceClient),
	}
}

// Register 注册服务
func (r *ServiceRegistry) Register(name string, client *PythonServiceClient) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[name] = client
}

// Get 获取服务
func (r *ServiceRegistry) Get(name string) (*PythonServiceClient, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if client, ok := r.services[name]; ok {
		return client, nil
	}
	return nil, fmt.Errorf("service %s not found", name)
}

// Remove 移除服务
func (r *ServiceRegistry) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.services, name)
}

// List 列出所有服务
func (r *ServiceRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}
	return names
}
