package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PythonExecTool Python 执行工具
type PythonExecTool struct {
	*BaseTool
	serverURL string
	client    *http.Client
}

// NewPythonExecTool 创建 Python 执行工具
func NewPythonExecTool(serverURL string) *PythonExecTool {
	return &PythonExecTool{
		BaseTool: NewBaseTool(
			"python_exec",
			"Python 执行工具，通过 HTTP 调用 Python 服务",
		),
		serverURL: serverURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Execute 执行 Python 代码
func (t *PythonExecTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	code, ok := args["code"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	// 准备请求数据
	reqData := map[string]interface{}{
		"code": code,
	}

	// 如果有输入数据，添加到请求中
	if input, ok := args["input"]; ok {
		reqData["input"] = input
	}

	// 发送请求到 Python 服务
	resp, err := t.sendRequest(ctx, reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to execute Python code: %v", err)
	}

	return resp, nil
}

// sendRequest 发送请求到 Python 服务
func (t *PythonExecTool) sendRequest(ctx context.Context, data map[string]interface{}) (interface{}, error) {
	// 将数据转换为 JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", t.serverURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Python service returned error: %s", resp.Status)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 检查是否有错误
	if err, ok := result["error"]; ok {
		return nil, fmt.Errorf("Python execution error: %v", err)
	}

	return result["result"], nil
}

// PythonService Python 服务配置
type PythonService struct {
	URL     string
	Timeout time.Duration
}

// NewPythonService 创建 Python 服务配置
func NewPythonService(url string, timeout time.Duration) *PythonService {
	return &PythonService{
		URL:     url,
		Timeout: timeout,
	}
}

// Start 启动 Python 服务
func (s *PythonService) Start() error {
	// 这里可以添加启动 Python 服务的逻辑
	// 例如：检查服务是否可用，启动子进程等
	return nil
}

// Stop 停止 Python 服务
func (s *PythonService) Stop() error {
	// 这里可以添加停止 Python 服务的逻辑
	return nil
}
