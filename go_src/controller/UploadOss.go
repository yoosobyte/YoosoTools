package controller

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

const YoosoToolsDir = `D:\soft\YoosoTools`

func SaveFile(params map[string]interface{}) string {
	name := params["name"].(string)
	data := params["data"].([]interface{})
	byteData := make([]byte, len(data))
	for i, v := range data {
		byteData[i] = byte(v.(float64))
	}

	path := YoosoToolsDir + `\uploadFile`
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			panic(err)
		}
	}

	targetPath := filepath.Join(path, name)
	err := os.WriteFile(targetPath, byteData, 0644)
	if err != nil {
		return "处理文件错误: " + err.Error()
	}
	fmt.Println("开始上传文件:", name)
	ok, msg := upload2Oss(targetPath, name)
	if !ok {
		return msg
	}
	fmt.Println("上传成功,结果:", msg)
	return msg
}

func upload2Oss(filePath string, fileName string) (bool, string) {
	// 配置信息
	endpoint := "oss-cn-xxxxxxxx.aliyuncs.com"
	accessKeyID := "xxxxxxxx"
	accessKeySecret := "xxxxxxxx"
	bucketName := "xxxxxxxx"
	objectName := `xxxxxxxx/` + fileName
	localFileName := filePath
	if accessKeyID == "xxxxxxxx" || accessKeySecret == "xxxxxxxx" {
		return false, fmt.Sprintln("请下载源码,并完善OOS配置信息重新wails build")
	}
	// 创建OSSClient实例
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	log.Println("初始化OOS失败:", client)
	if err != nil {
		return false, fmt.Sprintln("创建OSS客户端失败:", err)
	}
	// 获取Bucket
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return false, fmt.Sprintln("获取Bucket失败:", err)
	}

	// 打开本地文件
	file, err := os.Open(localFileName)
	if err != nil {
		return false, fmt.Sprintln("打开本地文件失败:", err)
	}
	defer file.Close()

	// 上传文件
	err = bucket.PutObject(objectName, file)
	if err != nil {
		return false, fmt.Sprintln("上传文件失败:", err)
	}
	return true, `https://res.xxxxxxxx.com/` + objectName
}
