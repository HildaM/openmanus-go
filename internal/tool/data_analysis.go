package tool

import (
	"context"
	"encoding/json"
	"fmt"
)

// DataAnalysisTool 数据分析工具
type DataAnalysisTool struct {
	client *PythonServiceClient
}

// NewDataAnalysisTool 创建数据分析工具
func NewDataAnalysisTool(client *PythonServiceClient) *DataAnalysisTool {
	return &DataAnalysisTool{
		client: client,
	}
}

func (d *DataAnalysisTool) Name() string {
	return "DataAnalysisTool"
}

// DataAnalysisInput 数据分析工具输入
type DataAnalysisInput struct {
	Operation string                 `json:"operation"`
	Data      interface{}            `json:"data"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// DataAnalysisOutput 数据分析工具输出
type DataAnalysisOutput struct {
	Result        interface{}            `json:"result"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Visualization map[string]interface{} `json:"visualization,omitempty"`
}

func (d *DataAnalysisTool) Run(ctx context.Context, input interface{}) (interface{}, error) {
	// 解析输入
	var daInput DataAnalysisInput
	switch v := input.(type) {
	case string:
		// 尝试将字符串解析为 JSON
		if err := json.Unmarshal([]byte(v), &daInput); err != nil {
			// 如果不是 JSON，则作为数据直接使用
			daInput = DataAnalysisInput{
				Operation: "analyze",
				Data:      v,
			}
		}
	case map[string]interface{}:
		// 将 map 转换为 JSON 字符串
		jsonData, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("invalid input format: %v", err)
		}
		// 将 JSON 字符串解析为 DataAnalysisInput
		if err := json.Unmarshal(jsonData, &daInput); err != nil {
			return nil, fmt.Errorf("invalid input format: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported input type: %T", input)
	}

	// 调用 Python 服务
	result, err := d.client.Call(ctx, "analyze", daInput)
	if err != nil {
		return nil, fmt.Errorf("data analysis service call failed: %v", err)
	}

	// 解析输出
	var output DataAnalysisOutput
	if err := json.Unmarshal(result.([]byte), &output); err != nil {
		return nil, fmt.Errorf("failed to parse data analysis service response: %v", err)
	}

	return output, nil
}

// DataAnalysisConfig 数据分析工具配置
type DataAnalysisConfig struct {
	DefaultOperation     string
	VisualizationEnabled bool
	CacheEnabled         bool
	CacheTTL             int
}

// NewDataAnalysisToolWithConfig 使用配置创建数据分析工具
func NewDataAnalysisToolWithConfig(client *PythonServiceClient, config DataAnalysisConfig) *DataAnalysisTool {
	return &DataAnalysisTool{
		client: client,
	}
}

// 预定义的分析操作
const (
	OpAnalyze    = "analyze"    // 基础分析
	OpVisualize  = "visualize"  // 数据可视化
	OpForecast   = "forecast"   // 预测分析
	OpCluster    = "cluster"    // 聚类分析
	OpCorrelate  = "correlate"  // 相关性分析
	OpRegression = "regression" // 回归分析
)
