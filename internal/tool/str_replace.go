package tool

import (
	"context"
	"regexp"
	"strings"
)

// StrReplaceTool 字符串处理工具
type StrReplaceTool struct {
	name        string
	description string
}

// NewStrReplaceTool 创建字符串处理工具
func NewStrReplaceTool() *StrReplaceTool {
	return &StrReplaceTool{
		name:        "str_replace",
		description: "字符串处理工具，支持替换、分割、合并等操作",
	}
}

// Name 返回工具名称
func (t *StrReplaceTool) Name() string {
	return t.name
}

// Description 返回工具描述
func (t *StrReplaceTool) Description() string {
	return t.description
}

// Execute 执行工具功能
func (t *StrReplaceTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return t.Run(ctx, args)
}

// Run 执行字符串处理
func (t *StrReplaceTool) Run(ctx context.Context, input interface{}) (interface{}, error) {
	args, ok := input.(map[string]interface{})
	if !ok {
		return nil, ErrInvalidArgs
	}

	op, ok := args["operation"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	switch op {
	case "replace":
		return t.replace(args)
	case "split":
		return t.split(args)
	case "join":
		return t.join(args)
	case "regex_replace":
		return t.regexReplace(args)
	default:
		return nil, ErrInvalidOperation
	}
}

// replace 替换字符串
func (t *StrReplaceTool) replace(args map[string]interface{}) (interface{}, error) {
	text, ok := args["text"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	old, ok := args["old"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	new, ok := args["new"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	return strings.ReplaceAll(text, old, new), nil
}

// split 分割字符串
func (t *StrReplaceTool) split(args map[string]interface{}) (interface{}, error) {
	text, ok := args["text"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	sep, ok := args["separator"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	return strings.Split(text, sep), nil
}

// join 合并字符串
func (t *StrReplaceTool) join(args map[string]interface{}) (interface{}, error) {
	texts, ok := args["texts"].([]interface{})
	if !ok {
		return nil, ErrInvalidArgs
	}

	sep, ok := args["separator"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	strTexts := make([]string, len(texts))
	for i, text := range texts {
		strTexts[i] = text.(string)
	}

	return strings.Join(strTexts, sep), nil
}

// regexReplace 正则替换
func (t *StrReplaceTool) regexReplace(args map[string]interface{}) (interface{}, error) {
	text, ok := args["text"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	pattern, ok := args["pattern"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	replacement, ok := args["replacement"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return re.ReplaceAllString(text, replacement), nil
}
