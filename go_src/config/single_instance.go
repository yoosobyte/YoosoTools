package config

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// -------------- 单实例相关 --------------
const sockName = "yoosotools.sock"

var (
	ErrAlreadyRunning = errors.New("another instance is already running")
	runtimeCtx        context.Context // 存一份供 showMainWindow 使用
)

// TrySingleInstance 应在 wails.Run 之前调用
func TrySingleInstance() error {
	dir, _ := os.UserCacheDir()
	lockDir := filepath.Join(dir, "YoosoTools")
	_ = os.MkdirAll(lockDir, 0o700)

	lockFile := filepath.Join(lockDir, "pid")
	pidBytes, _ := os.ReadFile(lockFile)
	if len(pidBytes) > 0 {
		var oldPid int
		fmt.Sscanf(string(pidBytes), "%d", &oldPid)
		if ok, _ := process.PidExists(int32(oldPid)); ok {
			// 老进程在，发激活消息
			return sendActivate(filepath.Join(lockDir, sockName))
		}
	}

	// 抢占锁
	if err := os.WriteFile(lockFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0o600); err != nil {
		return err
	}
	return nil
}

// ListenActivate 启动本地 socket 监听，应在 wails.Run 之后调用（即 OnStartup 里）
func ListenActivate(ctx context.Context) {
	runtimeCtx = ctx // 保存 ctx 给 showMainWindow 使用
	dir, _ := os.UserCacheDir()
	sockPath := filepath.Join(dir, "YoosoTools", sockName)

	_ = os.Remove(sockPath) // 防止残留
	go func() {
		l, err := net.Listen("unix", sockPath)
		if err != nil {
			return
		}
		for {
			conn, _ := l.Accept()
			buf := make([]byte, 1)
			conn.Read(buf)
			conn.Close()
			showMainWindow()
		}
	}()
}

func sendActivate(sockPath string) error {
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, _ = conn.Write([]byte{1})
	return ErrAlreadyRunning
}

func showMainWindow() {
	if runtimeCtx == nil {
		return
	}
	runtime.WindowShow(runtimeCtx)
	runtime.WindowUnminimise(runtimeCtx)
}
