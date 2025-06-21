package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// LLMProvider LLM 提供商类型
type LLMProvider string

const (
	ProviderOpenAI   LLMProvider = "openai"
	ProviderDeepSeek LLMProvider = "deepseek"
)

// LLMConfig LLM 配置
type LLMConfig struct {
	Provider    LLMProvider `mapstructure:"provider"`
	Model       string      `mapstructure:"model"`
	BaseURL     string      `mapstructure:"base_url"`
	APIKey      string      `mapstructure:"api_key"`
	MaxTokens   int         `mapstructure:"max_tokens"`
	Temperature float64     `mapstructure:"temperature"`
}

// Config 应用配置结构
type Config struct {
	// 服务配置
	Server struct {
		Port    int    `mapstructure:"port"`
		Host    string `mapstructure:"host"`
		Timeout int    `mapstructure:"timeout"`
	} `mapstructure:"server"`

	// 日志配置
	Log struct {
		Level      string `mapstructure:"level"`
		Format     string `mapstructure:"format"`
		OutputPath string `mapstructure:"output_path"`
	} `mapstructure:"log"`

	// MCP 配置
	MCP struct {
		Enabled    bool   `mapstructure:"enabled"`
		ServiceURL string `mapstructure:"service_url"`
		MaxRetries int    `mapstructure:"max_retries"`
		RetryDelay int    `mapstructure:"retry_delay"`
	} `mapstructure:"mcp"`

	// 工具配置
	Tools struct {
		PythonService struct {
			URL     string `mapstructure:"url"`
			Timeout int    `mapstructure:"timeout"`
		} `mapstructure:"python_service"`
	} `mapstructure:"tools"`

	// 智能体配置
	Agents struct {
		MaxConcurrent int `mapstructure:"max_concurrent"`
		Timeout       int `mapstructure:"timeout"`
	} `mapstructure:"agents"`

	// LLM 配置
	LLM LLMConfig `mapstructure:"llm"`
}

var (
	config *Config
	v      *viper.Viper
)

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	if v == nil {
		v = viper.New()
	}

	// 设置默认值
	setDefaults()

	// 设置配置文件路径
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// 默认配置文件路径
		v.SetConfigName("config")
		v.SetConfigType("toml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/openmanus")
	}

	// 读取环境变量
	v.AutomaticEnv()
	v.SetEnvPrefix("OPENMANUS")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %v", err)
		}
	}

	// 解析配置
	config = &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return config, nil
}

// GetConfig 获取配置实例
func GetConfig() *Config {
	return config
}

// setDefaults 设置默认配置
func setDefaults() {
	// 服务器默认配置
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.timeout", 30)

	// 日志默认配置
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text")
	v.SetDefault("log.output_path", "logs")

	// MCP 默认配置
	v.SetDefault("mcp.enabled", true)
	v.SetDefault("mcp.max_retries", 3)
	v.SetDefault("mcp.retry_delay", 5)

	// 工具默认配置
	v.SetDefault("tools.python_service.timeout", 30)

	// 智能体默认配置
	v.SetDefault("agents.max_concurrent", 5)
	v.SetDefault("agents.timeout", 300)

	// LLM 默认配置
	v.SetDefault("llm.provider", ProviderOpenAI)
	v.SetDefault("llm.model", "gpt-4")
	v.SetDefault("llm.base_url", "https://api.openai.com/v1")
	v.SetDefault("llm.max_tokens", 4096)
	v.SetDefault("llm.temperature", 0.0)
}

// SaveConfig 保存配置到文件
func SaveConfig(configPath string) error {
	if v == nil {
		return fmt.Errorf("配置未初始化")
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 保存配置
	return v.WriteConfigAs(configPath)
}
