package prompt

import (
	"fmt"
	"sync"
)

// Language 语言类型
type Language string

const (
	// English 英语
	English Language = "en"
	// Chinese 中文
	Chinese Language = "zh"
)

// I18nPrompt 多语言提示词
type I18nPrompt struct {
	*BasePrompt
	translations map[Language]string
	mu           sync.RWMutex
}

// NewI18nPrompt 创建多语言提示词
func NewI18nPrompt(template string) *I18nPrompt {
	return &I18nPrompt{
		BasePrompt:   NewBasePrompt(template),
		translations: make(map[Language]string),
	}
}

// AddTranslation 添加翻译
func (p *I18nPrompt) AddTranslation(lang Language, translation string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.translations[lang] = translation
}

// GetTranslation 获取翻译
func (p *I18nPrompt) GetTranslation(lang Language) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if translation, ok := p.translations[lang]; ok {
		return translation, nil
	}
	return "", fmt.Errorf("未找到语言 %s 的翻译", lang)
}

// RenderWithLang 使用指定语言渲染提示词
func (p *I18nPrompt) RenderWithLang(lang Language, data interface{}) (string, error) {
	translation, err := p.GetTranslation(lang)
	if err != nil {
		return "", err
	}
	p.template = translation
	return p.Render(data)
}

// I18nPromptManager 多语言提示词管理器
type I18nPromptManager struct {
	prompts map[string]*I18nPrompt
	mu      sync.RWMutex
}

// NewI18nPromptManager 创建多语言提示词管理器
func NewI18nPromptManager() *I18nPromptManager {
	return &I18nPromptManager{
		prompts: make(map[string]*I18nPrompt),
	}
}

// RegisterPrompt 注册提示词
func (m *I18nPromptManager) RegisterPrompt(name string, prompt *I18nPrompt) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.prompts[name] = prompt
}

// GetPrompt 获取提示词
func (m *I18nPromptManager) GetPrompt(name string) (*I18nPrompt, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if prompt, ok := m.prompts[name]; ok {
		return prompt, nil
	}
	return nil, fmt.Errorf("未找到提示词 %s", name)
}

// RenderPrompt 渲染提示词
func (m *I18nPromptManager) RenderPrompt(name string, lang Language, data interface{}) (string, error) {
	prompt, err := m.GetPrompt(name)
	if err != nil {
		return "", err
	}
	return prompt.RenderWithLang(lang, data)
}
