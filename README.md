# 安装 | 运行 | 打包

## 安装
* go下载地址 https://dl.google.com/go/go1.25.0.windows-amd64.msi
* 重新安装（确保 GOPATH/bin 在系统 PATH 中）
* 设置国内镜像源:go env -w GOPROXY=https://goproxy.cn,direct
* 安装wails:go install github.com/wailsapp/wails/v2/cmd/wails@latest

## 运行
* 终端：启动 Wails 后端
* wails dev

## 打包
* 终端：启动 Wails 后端
* wails build