package tool

import (
	"bytes"
	"context"
	"os/exec"
)

type ShellTool struct{}

func (s *ShellTool) Name() string {
	return "ShellTool"
}

func (s *ShellTool) Run(ctx context.Context, input interface{}) (interface{}, error) {
	cmdStr, ok := input.(string)
	if !ok {
		return nil, ErrInvalidInput
	}
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return out.String(), err
	}
	return out.String(), nil
}

var ErrInvalidInput = &ToolError{"ShellTool: 输入必须为字符串"}
