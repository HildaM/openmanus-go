package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// FileTool 文件工具
type FileTool struct {
	basePath string
}

// NewFileTool 创建文件工具
func NewFileTool(basePath string) *FileTool {
	return &FileTool{
		basePath: basePath,
	}
}

func (f *FileTool) Name() string {
	return "FileTool"
}

// FileOperation 文件操作类型
type FileOperation struct {
	Action  string                 `json:"action"`
	Path    string                 `json:"path"`
	Content interface{}            `json:"content,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// FileResult 文件操作结果
type FileResult struct {
	Success  bool                   `json:"success"`
	Path     string                 `json:"path"`
	Content  interface{}            `json:"content,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Error    string                 `json:"error,omitempty"`
}

func (f *FileTool) Run(ctx context.Context, input interface{}) (interface{}, error) {
	// 解析输入
	var op FileOperation
	switch v := input.(type) {
	case string:
		// 如果是字符串，假设是文件路径，执行读取操作
		op = FileOperation{
			Action: "read",
			Path:   v,
		}
	case map[string]interface{}:
		// 将 map 转换为 FileOperation
		jsonData, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("invalid input format: %v", err)
		}
		if err := json.Unmarshal(jsonData, &op); err != nil {
			return nil, fmt.Errorf("invalid input format: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported input type: %T", input)
	}

	// 构建完整路径
	fullPath := filepath.Join(f.basePath, op.Path)

	// 执行操作
	var result FileResult
	switch op.Action {
	case "read":
		content, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("read file failed: %v", err)
		}
		result = FileResult{
			Success: true,
			Path:    fullPath,
			Content: string(content),
			Metadata: map[string]interface{}{
				"size": len(content),
				"time": time.Now(),
			},
		}

	case "write":
		var content []byte
		switch v := op.Content.(type) {
		case string:
			content = []byte(v)
		case []byte:
			content = v
		default:
			content, _ = json.Marshal(v)
		}

		if err := ioutil.WriteFile(fullPath, content, os.ModePerm); err != nil {
			return nil, fmt.Errorf("write file failed: %v", err)
		}
		result = FileResult{
			Success: true,
			Path:    fullPath,
			Metadata: map[string]interface{}{
				"size": len(content),
				"time": time.Now(),
			},
		}

	case "delete":
		if err := os.Remove(fullPath); err != nil {
			return nil, fmt.Errorf("delete file failed: %v", err)
		}
		result = FileResult{
			Success: true,
			Path:    fullPath,
			Metadata: map[string]interface{}{
				"time": time.Now(),
			},
		}

	case "list":
		files, err := ioutil.ReadDir(fullPath)
		if err != nil {
			return nil, fmt.Errorf("list directory failed: %v", err)
		}
		fileList := make([]map[string]interface{}, 0)
		for _, file := range files {
			fileList = append(fileList, map[string]interface{}{
				"name":    file.Name(),
				"size":    file.Size(),
				"mode":    file.Mode(),
				"modTime": file.ModTime(),
				"isDir":   file.IsDir(),
			})
		}
		result = FileResult{
			Success: true,
			Path:    fullPath,
			Content: fileList,
			Metadata: map[string]interface{}{
				"count": len(fileList),
				"time":  time.Now(),
			},
		}

	case "mkdir":
		if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
			return nil, fmt.Errorf("create directory failed: %v", err)
		}
		result = FileResult{
			Success: true,
			Path:    fullPath,
			Metadata: map[string]interface{}{
				"time": time.Now(),
			},
		}

	case "move":
		target, ok := op.Options["target"].(string)
		if !ok {
			return nil, fmt.Errorf("target path not specified")
		}
		targetPath := filepath.Join(f.basePath, target)
		if err := os.Rename(fullPath, targetPath); err != nil {
			return nil, fmt.Errorf("move file failed: %v", err)
		}
		result = FileResult{
			Success: true,
			Path:    targetPath,
			Metadata: map[string]interface{}{
				"time": time.Now(),
			},
		}

	case "copy":
		target, ok := op.Options["target"].(string)
		if !ok {
			return nil, fmt.Errorf("target path not specified")
		}
		targetPath := filepath.Join(f.basePath, target)
		content, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("read source file failed: %v", err)
		}
		if err := ioutil.WriteFile(targetPath, content, os.ModePerm); err != nil {
			return nil, fmt.Errorf("write target file failed: %v", err)
		}
		result = FileResult{
			Success: true,
			Path:    targetPath,
			Metadata: map[string]interface{}{
				"size": len(content),
				"time": time.Now(),
			},
		}

	default:
		return nil, fmt.Errorf("unsupported file operation: %s", op.Action)
	}

	return result, nil
}

// 预定义的文件操作
const (
	FileOpRead   = "read"   // 读取文件
	FileOpWrite  = "write"  // 写入文件
	FileOpDelete = "delete" // 删除文件
	FileOpList   = "list"   // 列出目录
	FileOpMkdir  = "mkdir"  // 创建目录
	FileOpMove   = "move"   // 移动文件
	FileOpCopy   = "copy"   // 复制文件
)
