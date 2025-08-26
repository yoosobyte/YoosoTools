package main

import (
	"YoosoTools/go_src/config"
	"YoosoTools/go_src/controller"
	"YoosoTools/go_src/service"
	"YoosoTools/go_src/utils"
	"context"
	_ "embed"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 1.让老进程能响应“再次启动”的激活消息
	config.ListenActivate(ctx)
	// 2.初始化IP广播
	controller.StartUDPBroadcastAll(ctx)
	// 3.初始化系统托盘（内部用 sync.Once 保证只跑一次）
	go config.InitTray(ctx)
	// 4.初始化数据库
	go func() {
		if err := config.InitDB(); err != nil {
			fmt.Println("初始化数据库失败:", err)
		}
	}()
}

func (a *App) SelectFolder(field string) string {
	// 使用 Wails 的运行时打开文件夹选择对话框
	folderPath, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "请选择 " + field,
	})
	if err != nil {
		return ""
	}
	return folderPath
}

// 选择文件的方法
func (a *App) SelectFile(field string) string {
	// 使用 Wails 的运行时打开文件选择对话框
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "请选择 " + field,
		// 可以添加文件过滤器
		Filters: []runtime.FileFilter{
			{
				DisplayName: "所有文件",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		fmt.Printf("选择文件失败: %v\n", err)
		return ""
	}
	return filePath
}

// 同时支持选择文件和文件夹的通用方法
func (a *App) SelectPath(field string, isFile bool) string {
	if isFile {
		return a.SelectFile(field)
	} else {
		return a.SelectFolder(field)
	}
}

func (a *App) SelectJavaExeDirect() string {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "请选择 java.exe",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Java 可执行文件",
				Pattern:     "java.exe",
			},
		},
	})
	if err != nil {
		return ""
	}
	return path
}

func (a *App) KillPort(port string) string {
	return utils.KillByPort(port)
}

func (a *App) GetIpInfo() string {
	return controller.GetIpInfo()
}

func (a *App) SaveFile(params map[string]interface{}) string {
	return controller.SaveFile(params)
}

func (a *App) NewCon(params string, sessionId string, rows, cols int) string {
	return controller.NewCon(params, sessionId, a.ctx, rows, cols)
}
func (a *App) CloseCon(sessionId string) string {
	return controller.CloseSSH(sessionId)
}

// 透传任意字节
func (a *App) SendRaw(data string, sessionId string) error {
	return controller.SendRaw(data, sessionId)
}
func (a *App) SendResize(rows, cols int, sessionId string) error {
	return controller.SendResize(rows, cols, sessionId)
}

// Enter 键发送（保持兼容）
func (a *App) SendNormal(cmd string, sessionId string) string {
	return controller.SendNormal(cmd, sessionId)
}

func (a *App) SaveServer(params string) string {
	return service.SaveServer(params)
}

func (a *App) EditServer(params string) string {
	return service.EditServer(params)
}

func (a *App) RemoveServer(serverId int) string {
	return service.RemoveServer(serverId)
}

func (a *App) GetOneServer(serverId int) string {
	return service.GetOneServer(serverId)
}

func (a *App) GetListServer() string {
	return service.GetListServer()
}

func (a *App) GetFileSize(sessionId, filePath string) string {
	return controller.GetFileSize(sessionId, filePath)
}

func (a *App) DownloadFile(sessionId, filePath string) string {
	return controller.DownloadFile(sessionId, filePath)
}
func (a *App) NewFolder(filePath, fileName, sessionId string) string {
	return controller.NewFolder(filePath, fileName, sessionId)
}
func (a *App) MoveOrCopyFolderAndFile(nowDir, copyDir, newName, sessionId string, isCopy int) string {
	return controller.MoveOrCopyFolderAndFile(nowDir, copyDir, newName, sessionId, isCopy)
}
func (a *App) RenameFolderOrFile(filePath, newFilName, sessionId string) string {
	return controller.RenameFolderOrFile(filePath, newFilName, sessionId)
}
func (a *App) SaveNewFile(fileName, fileContent, filePath, sessionId string) string {
	return controller.SaveNewFile(fileName, fileContent, filePath, sessionId)
}
func (a *App) DeleteItem(filePath, sessionId string) string {
	return controller.DeleteItem(filePath, sessionId)
}
func (a *App) UploadDirOrFile(targetUploadPath, willUploadPath, sessionId string) string {
	return controller.UploadDirOrFile(targetUploadPath, willUploadPath, sessionId)
}

func (a *App) PostNewPath(nowPath string, sessionId string, isInit int) string {
	return controller.PostNewPath(nowPath, sessionId, isInit)
}

func (a *App) PostRadio() string {
	return controller.PostRadio()
}
