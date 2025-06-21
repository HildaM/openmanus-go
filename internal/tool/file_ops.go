package tool

import (
	"context"
	"io/ioutil"
	"os"
)

// FileOpsTool 文件操作工具
type FileOpsTool struct {
	name        string
	description string
}

// NewFileOpsTool 创建文件操作工具
func NewFileOpsTool() *FileOpsTool {
	return &FileOpsTool{
		name:        "file_ops",
		description: "文件操作工具，支持读写文件、创建目录等操作",
	}
}

// Name 返回工具名称
func (t *FileOpsTool) Name() string {
	return t.name
}

// Description 返回工具描述
func (t *FileOpsTool) Description() string {
	return t.description
}

// Execute 执行工具功能
func (t *FileOpsTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return t.Run(ctx, args)
}

// Run 执行文件操作
func (t *FileOpsTool) Run(ctx context.Context, input interface{}) (interface{}, error) {
	args, ok := input.(map[string]interface{})
	if !ok {
		return nil, ErrInvalidArgs
	}

	op, ok := args["operation"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	switch op {
	case "read":
		return t.readFile(args)
	case "write":
		return t.writeFile(args)
	case "mkdir":
		return t.createDir(args)
	case "list":
		return t.listDir(args)
	case "delete":
		return t.deleteFile(args)
	default:
		return nil, ErrInvalidOperation
	}
}

// readFile 读取文件
func (t *FileOpsTool) readFile(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return string(data), nil
}

// writeFile 写入文件
func (t *FileOpsTool) writeFile(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	content, ok := args["content"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	err := ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return nil, err
	}

	return "success", nil
}

// createDir 创建目录
func (t *FileOpsTool) createDir(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, err
	}

	return "success", nil
}

// listDir 列出目录内容
func (t *FileOpsTool) listDir(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for _, file := range files {
		result = append(result, map[string]interface{}{
			"name":    file.Name(),
			"size":    file.Size(),
			"mode":    file.Mode(),
			"modTime": file.ModTime(),
			"isDir":   file.IsDir(),
		})
	}

	return result, nil
}

// deleteFile 删除文件或目录
func (t *FileOpsTool) deleteFile(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, ErrInvalidArgs
	}

	err := os.RemoveAll(path)
	if err != nil {
		return nil, err
	}

	return "success", nil
}
