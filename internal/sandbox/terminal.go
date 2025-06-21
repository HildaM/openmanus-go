package sandbox

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"sync"
)

// Terminal 终端模拟器
type Terminal struct {
	*BaseSandbox
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	mu     sync.Mutex
}

// NewTerminal 创建终端模拟器
func NewTerminal() *Terminal {
	return &Terminal{
		BaseSandbox: NewBaseSandbox(),
	}
}

// Start 启动终端
func (t *Terminal) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 根据操作系统创建命令
	switch runtime.GOOS {
	case "linux":
		t.cmd = exec.CommandContext(ctx, "bash")
	case "windows":
		t.cmd = exec.CommandContext(ctx, "cmd")
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	// 设置标准输入输出
	var err error
	t.stdin, err = t.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("创建标准输入管道失败: %v", err)
	}

	t.stdout, err = t.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建标准输出管道失败: %v", err)
	}

	t.stderr, err = t.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建标准错误管道失败: %v", err)
	}

	// 启动命令
	if err := t.cmd.Start(); err != nil {
		return fmt.Errorf("启动命令失败: %v", err)
	}

	return nil
}

// Write 写入命令
func (t *Terminal) Write(cmd string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cmd == nil || t.stdin == nil {
		return fmt.Errorf("终端未启动")
	}

	_, err := t.stdin.Write([]byte(cmd + "\n"))
	return err
}

// Read 读取输出
func (t *Terminal) Read() (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cmd == nil || t.stdout == nil {
		return "", fmt.Errorf("终端未启动")
	}

	buf := make([]byte, 1024)
	n, err := t.stdout.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	return string(buf[:n]), nil
}

// ReadError 读取错误输出
func (t *Terminal) ReadError() (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cmd == nil || t.stderr == nil {
		return "", fmt.Errorf("终端未启动")
	}

	buf := make([]byte, 1024)
	n, err := t.stderr.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	return string(buf[:n]), nil
}

// Close 关闭终端
func (t *Terminal) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cmd == nil {
		return nil
	}

	if t.stdin != nil {
		t.stdin.Close()
	}
	if t.stdout != nil {
		t.stdout.Close()
	}
	if t.stderr != nil {
		t.stderr.Close()
	}

	return t.cmd.Wait()
}

// ExecuteCommand 执行命令并获取输出
func (t *Terminal) ExecuteCommand(ctx context.Context, cmd string) (string, error) {
	if err := t.Start(ctx); err != nil {
		return "", err
	}
	defer t.Close()

	if err := t.Write(cmd); err != nil {
		return "", err
	}

	output, err := t.Read()
	if err != nil {
		return "", err
	}

	errorOutput, err := t.ReadError()
	if err != nil {
		return "", err
	}

	if errorOutput != "" {
		return output + "\nError: " + errorOutput, nil
	}

	return output, nil
}
