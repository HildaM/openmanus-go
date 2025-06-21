package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// WebScraperTool Web 抓取工具
type WebScraperTool struct {
	client *http.Client
}

// NewWebScraperTool 创建 Web 抓取工具
func NewWebScraperTool() *WebScraperTool {
	return &WebScraperTool{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (w *WebScraperTool) Name() string {
	return "WebScraperTool"
}

// WebRequest Web 请求参数
type WebRequest struct {
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Headers map[string]string      `json:"headers,omitempty"`
	Body    interface{}            `json:"body,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// WebResponse Web 响应结果
type WebResponse struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       interface{}            `json:"body"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

func (w *WebScraperTool) Run(ctx context.Context, input interface{}) (interface{}, error) {
	// 解析输入
	var req WebRequest
	switch v := input.(type) {
	case string:
		// 如果是字符串，假设是 URL，执行 GET 请求
		req = WebRequest{
			Method: "GET",
			URL:    v,
		}
	case map[string]interface{}:
		// 将 map 转换为 WebRequest
		jsonData, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("invalid input format: %v", err)
		}
		if err := json.Unmarshal(jsonData, &req); err != nil {
			return nil, fmt.Errorf("invalid input format: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported input type: %T", input)
	}

	// 验证 URL
	if _, err := url.ParseRequestURI(req.URL); err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	// 创建请求
	var bodyReader *strings.Reader
	if req.Body != nil {
		bodyData, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %v", err)
		}
		bodyReader = strings.NewReader(string(bodyData))
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// 发送请求
	startTime := time.Now()
	resp, err := w.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// 解析响应头
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = values[0]
	}

	// 构建响应
	result := WebResponse{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       string(body),
		Metadata: map[string]interface{}{
			"duration": time.Since(startTime).String(),
			"time":     time.Now(),
		},
	}

	// 检查状态码
	if resp.StatusCode >= 400 {
		return result, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	return result, nil
}

// WebScraperConfig Web 抓取工具配置
type WebScraperConfig struct {
	Timeout          time.Duration
	MaxRedirects     int
	FollowRedirects  bool
	VerifySSL        bool
	DefaultUserAgent string
	DefaultHeaders   map[string]string
}

// NewWebScraperToolWithConfig 使用配置创建 Web 抓取工具
func NewWebScraperToolWithConfig(config WebScraperConfig) *WebScraperTool {
	client := &http.Client{
		Timeout: config.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !config.FollowRedirects {
				return http.ErrUseLastResponse
			}
			if len(via) >= config.MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", config.MaxRedirects)
			}
			return nil
		},
	}

	return &WebScraperTool{
		client: client,
	}
}

// 预定义的 HTTP 方法
const (
	MethodGet     = "GET"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodDelete  = "DELETE"
	MethodPatch   = "PATCH"
	MethodHead    = "HEAD"
	MethodOptions = "OPTIONS"
)
