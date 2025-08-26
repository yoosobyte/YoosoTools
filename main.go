package main

import (
	"YoosoTools/go_src/config"
	"embed"
	"errors"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 1. 单实例检查
	if err := config.TrySingleInstance(); err != nil {
		if errors.Is(err, config.ErrAlreadyRunning) {
			return // 老进程已唤醒，自己退出
		}
		println("single instance check error:", err.Error())
		return
	}
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "悦速工具箱 - YoosoTools",
		Width:  1204,
		Height: 788,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour:  &options.RGBA{R: 255, G: 255, B: 255, A: 0},
		HideWindowOnClose: true,
		OnStartup:         app.startup,
		Bind: []interface{}{
			app,
		},
		// 开发模式配置
		//Debug: options.Debug{
		//	OpenInspectorOnStartup: false, // 启动时自动打开开发者工具
		//},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
