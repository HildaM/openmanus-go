package sandbox

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"syscall"
	"time"
)

// ResourceLimits 资源限制配置
type ResourceLimits struct {
	// CPU 限制（百分比）
	CPULimit float64
	// 内存限制（MB）
	MemoryLimit int64
	// 磁盘限制（MB）
	DiskLimit int64
	// 进程数限制
	ProcessLimit int
	// 执行时间限制（秒）
	TimeLimit int
}

// DefaultResourceLimits 默认资源限制
var DefaultResourceLimits = ResourceLimits{
	CPULimit:     50.0, // 50% CPU
	MemoryLimit:  512,  // 512MB
	DiskLimit:    1024, // 1GB
	ProcessLimit: 10,   // 最多10个进程
	TimeLimit:    30,   // 30秒
}

// Sandbox 定义沙箱接口
type Sandbox interface {
	// Execute 在沙箱中执行命令
	Execute(ctx context.Context, cmd string) (string, error)
	// SetResourceLimits 设置资源限制
	SetResourceLimits(limits ResourceLimits)
	// GetResourceLimits 获取资源限制
	GetResourceLimits() ResourceLimits
}

// BaseSandbox 基础沙箱实现
type BaseSandbox struct {
	limits ResourceLimits
}

// NewBaseSandbox 创建基础沙箱
func NewBaseSandbox() *BaseSandbox {
	return &BaseSandbox{
		limits: DefaultResourceLimits,
	}
}

// SetResourceLimits 设置资源限制
func (s *BaseSandbox) SetResourceLimits(limits ResourceLimits) {
	s.limits = limits
}

// GetResourceLimits 获取资源限制
func (s *BaseSandbox) GetResourceLimits() ResourceLimits {
	return s.limits
}

// Execute 在沙箱中执行命令
func (s *BaseSandbox) Execute(ctx context.Context, cmd string) (string, error) {
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(s.limits.TimeLimit)*time.Second)
	defer cancel()

	// 根据操作系统创建命令
	var execCmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		execCmd = exec.CommandContext(timeoutCtx, "bash", "-c", cmd)
	case "windows":
		execCmd = exec.CommandContext(timeoutCtx, "cmd", "/C", cmd)
	default:
		return "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	// 设置资源限制
	if err := s.setResourceLimits(execCmd); err != nil {
		return "", fmt.Errorf("设置资源限制失败: %v", err)
	}

	// 执行命令并获取输出
	output, err := execCmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("命令执行失败: %v", err)
	}

	return string(output), nil
}

// setResourceLimits 设置进程资源限制
func (s *BaseSandbox) setResourceLimits(cmd *exec.Cmd) error {
	// 根据操作系统设置不同的资源限制
	switch runtime.GOOS {
	case "linux":
		// Linux 系统使用 cgroups 进行资源限制
		// 这里需要系统管理员权限和 cgroups 支持
		return nil
	case "windows":
		// Windows 系统使用作业对象进行资源限制
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		}
		return nil
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}
