// +build windows

package common

import (
	"context"
	"errors"
	"os/exec"
	"strconv"
	"syscall"
)

type Result struct {
	output string
	err    error
}

// 执行shell命令，可设置执行超时时间
func Exec(ctx context.Context, command string) (string, error) {
	cmd := exec.Command("cmd", "/C", command)
	// 隐藏cmd窗口
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	var resultChan = make(chan Result)
	go func() {
		output, err := cmd.CombinedOutput()
		resultChan <- Result{string(output), err}
	}()
	select {
	case <-ctx.Done():
		if cmd.Process.Pid > 0 {
			_ = exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
			_ = cmd.Process.Kill()
		}
		return "", errors.New("timeout killed")
	case result := <-resultChan:
		return convertEncoding(result.output), result.err
	}
}

func convertEncoding(outputGBK string) string {
	// windows平台编码为gbk，需转换为utf8
	outputUTF8, ok := GBK2UTF8(outputGBK)
	if ok {
		return outputUTF8
	}
	return outputGBK
}