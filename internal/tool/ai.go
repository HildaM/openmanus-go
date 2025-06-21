package tool

import (
	"context"
	"encoding/json"
	"fmt"
)

// AITool AI 工具
type AITool struct {
	client *PythonServiceClient
}

// NewAITool 创建 AI 工具
func NewAITool(client *PythonServiceClient) *AITool {
	return &AITool{
		client: client,
	}
}

func (a *AITool) Name() string {
	return "AITool"
}

// AIToolInput AI 工具输入
type AIToolInput struct {
	Model     string                 `json:"model"`
	Prompt    string                 `json:"prompt"`
	Options   map[string]interface{} `json:"options,omitempty"`
	Stream    bool                   `json:"stream,omitempty"`
	MaxTokens int                    `json:"max_tokens,omitempty"`
}

// AIToolOutput AI 工具输出
type AIToolOutput struct {
	Text     string                 `json:"text"`
	Usage    map[string]interface{} `json:"usage,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (a *AITool) Run(ctx context.Context, input interface{}) (interface{}, error) {
	// 解析输入
	var aiInput AIToolInput
	switch v := input.(type) {
	case string:
		aiInput = AIToolInput{
			Model:  "default",
			Prompt: v,
		}
	case map[string]interface{}:
		// 将 map 转换为 JSON 字符串
		jsonData, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("invalid input format: %v", err)
		}
		// 将 JSON 字符串解析为 AIToolInput
		if err := json.Unmarshal(jsonData, &aiInput); err != nil {
			return nil, fmt.Errorf("invalid input format: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported input type: %T", input)
	}

	// 调用 Python 服务
	result, err := a.client.Call(ctx, "generate", aiInput)
	if err != nil {
		return nil, fmt.Errorf("AI service call failed: %v", err)
	}

	// 解析输出
	var output AIToolOutput
	if err := json.Unmarshal(result.([]byte), &output); err != nil {
		return nil, fmt.Errorf("failed to parse AI service response: %v", err)
	}

	return output, nil
}

// AIToolConfig AI 工具配置
type AIToolConfig struct {
	Model       string
	MaxTokens   int
	Temperature float64
	TopP        float64
	Stream      bool
}

// NewAIToolWithConfig 使用配置创建 AI 工具
func NewAIToolWithConfig(client *PythonServiceClient, config AIToolConfig) *AITool {
	return &AITool{
		client: client,
	}
}
