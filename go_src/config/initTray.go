package config

import (
	"context"
	_ "embed"
	"sync"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed app.ico
var trayIcon []byte
var trayOnce sync.Once

func InitTray(ctx context.Context) {
	// 保证只真正执行一次
	systray.Run(func() {
		systray.SetIcon(trayIcon)
		systray.SetTitle("YoosoTools")
		systray.SetTooltip("YoosoTools")

		mShow := systray.AddMenuItem("显示窗口", "")
		mQuit := systray.AddMenuItem("退出", "")

		for {
			select {
			case <-mShow.ClickedCh:
				runtime.WindowShow(ctx)
			case <-mQuit.ClickedCh:
				runtime.Quit(ctx)
				return
			}
		}
	}, nil)
}
