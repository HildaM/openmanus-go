package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// LLMClient 定义 LLM 客户端接口
type LLMClient interface {
	GetCompletion(messages []Message) (*CompletionResponse, error)
}

// Client LLM 客户端
type Client struct {
	BaseURL     string
	APIKey      string
	MaxTokens   int
	Temperature float64
}

// NewClient 创建 LLM 客户端
func NewClient(baseURL, apiKey string, maxTokens int, temperature float64) LLMClient {
	return &Client{
		BaseURL:     baseURL,
		APIKey:      apiKey,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}
}

// CompletionRequest 请求结构
type CompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionResponse 响应结构
type CompletionResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

// GetCompletion 获取 LLM 补全结果
func (c *Client) GetCompletion(messages []Message) (*CompletionResponse, error) {
	reqBody := CompletionRequest{
		Model:       "gpt-4",
		Messages:    messages,
		MaxTokens:   c.MaxTokens,
		Temperature: c.Temperature,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	var result CompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}
