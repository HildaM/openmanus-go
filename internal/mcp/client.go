package mcp

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Client MCP 客户端
type Client struct {
	BaseURL string
	Timeout time.Duration
}

// NewClient 创建 MCP 客户端
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		BaseURL: baseURL,
		Timeout: timeout,
	}
}

// RunMCP 触发 MCP 服务 /run
func (c *Client) RunMCP(ctx context.Context) (string, error) {
	url := c.BaseURL + "/run"
	client := &http.Client{Timeout: c.Timeout}

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求 MCP 服务失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return string(body), fmt.Errorf("MCP 服务返回错误: %s", resp.Status)
	}

	return string(body), nil
}
