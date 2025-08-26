package controller

import (
	"YoosoTools/go_src/entity"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func PostNewPath(nowPath string, sessionId string, isInit int) string {
	for isInit == 1 {
		sessionData, exists := sessions[sessionId]

		if !exists || sessionData.Client == nil {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	cmd := fmt.Sprintf("ls -lshF %s --time-style=long-iso -t | sed -E 's|([*@|=>]$)$||'", nowPath)
	output, err := ExecuteBackendCommand(cmd, sessionId)
	if err != nil {
		return entity.ErrorOnlyMsgStr("命令执行失败")
	}
	listing, err := parseDirectoryListing(output, nowPath)
	if err != nil {
		return entity.ErrorOnlyMsgStr("解析结果失败")
	}
	return entity.SuccessOnlyDataStr(listing)
}

func parseDirectoryListing(input string, basePath string) ([]entity.PathItem, error) {
	lines := strings.Split(input, "\n")
	var items []entity.PathItem

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "total") {
			continue
		}

		// 分割行，使用多个空格作为分隔符
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		// 提取相关信息
		fileSize := fields[0] + "B"
		dateStr := fields[6]
		timeStr := fields[7]
		fileName := fields[8]

		// 去除文件名末尾的斜杠（如果是目录）
		isFolder := 0
		if strings.HasSuffix(fileName, "/") {
			isFolder = 1
			fileName = strings.TrimSuffix(fileName, "/")
		}

		// 解析时间
		editTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", dateStr, timeStr))
		if err != nil {
			fmt.Printf("failed to parse time: %v", err)
			return nil, fmt.Errorf("failed to parse time: %v", err)
		}

		// 构建完整路径
		fullPath := basePath
		if basePath != "" && !strings.HasSuffix(basePath, "/") {
			fullPath += "/"
		}
		fullPath += fileName

		// 检测危险文件夹
		isDanger := 0
		if isDangerousPath(fullPath) {
			isDanger = 1
		}

		item := entity.PathItem{
			FileName: fileName,
			Path:     basePath,
			IsFolder: isFolder,
			EditTime: editTime.Format("2006/01/02 15:04"),
			FileSize: fileSize,
			FullPath: fullPath,
			IsDanger: isDanger,
		}

		items = append(items, item)
	}

	return items, nil
}

// 下载文件（带进度）
func DownloadFile(sessionId, filePath string) string {
	// 发送初始进度
	runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId, "0|开始检查文件...")

	// 检查文件是否存在且是普通文件
	checkCmd := fmt.Sprintf("[ -f %q ] && echo 'exists' || echo 'not exists'", filePath)
	checkOutput, err := ExecuteBackendCommand(checkCmd, sessionId)
	if err != nil || strings.TrimSpace(string(checkOutput)) != "exists" {
		runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId, "1|错误：文件不存在")
		return entity.ErrorOnlyMsgStr("文件不存在")
	}

	runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId, "10|文件存在，准备下载...")

	// 获取文件大小
	sizeCmd := fmt.Sprintf("stat -c %%s %q", filePath)
	sizeOutput, err := ExecuteBackendCommand(sizeCmd, sessionId)
	if err != nil {
		runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId, "1|错误：获取文件大小失败")
		return entity.ErrorOnlyMsgStr("获取文件大小失败")
	}

	fileSize, err := strconv.ParseInt(strings.TrimSpace(sizeOutput), 10, 64)
	if err != nil {
		runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId, "1|错误：解析文件大小失败")
		return entity.ErrorOnlyMsgStr("解析文件大小失败")
	}

	runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId,
		fmt.Sprintf("20|文件大小: %s", formatFileSize(fileSize)))

	// 获取文件名
	fileName := filepath.Base(filePath)
	runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId,
		fmt.Sprintf("30|文件名: %s", fileName))

	// 分块读取文件（为了显示进度）
	runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId, "40|开始读取文件内容...")

	// 分块下载的前端没兼容
	//if fileSize > 10*1024*1024 { // 大于10MB的文件使用分块读取
	if false { // 大于10MB的文件使用分块读取
		return downloadFileInChunks(sessionId, filePath, fileName, fileSize)
	} else {
		return downloadFileFull(sessionId, filePath, fileName, fileSize)
	}
}

// 完整下载小文件
func downloadFileFull(sessionId, filePath, fileName string, fileSize int64) string {
	runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId, "50|完整读取文件...")

	// 获取文件 Base64 编码
	base64Cmd := fmt.Sprintf("base64 -w 0 %q", filePath)
	output, err := ExecuteBackendCommand(base64Cmd, sessionId)
	if err != nil {
		runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId, "1|错误：文件读取失败")
		return entity.ErrorOnlyMsgStr("文件读取失败")
	}

	runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId, "80|文件编码完成...")

	// 返回 Base64 编码的文件内容和文件名
	result := map[string]interface{}{
		"fileName": fileName,
		"content":  string(output),
		"encoding": "base64",
		"fileSize": fileSize,
		"filePath": filePath,
	}

	runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId, "100|文件下载完成！")
	return entity.SuccessOnlyDataStr(result)
}

// 分块下载大文件
func downloadFileInChunks(sessionId, filePath, fileName string, fileSize int64) string {
	// 计算分块大小（每块1MB）
	chunkSize := int64(1 * 1024 * 1024)
	totalChunks := (fileSize + chunkSize - 1) / chunkSize

	runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId,
		fmt.Sprintf("45|大文件分块下载，共 %d 块", totalChunks))

	var chunks []string
	var downloadedSize int64

	for i := int64(0); i < totalChunks; i++ {
		offset := i * chunkSize
		currentChunkSize := chunkSize
		if offset+chunkSize > fileSize {
			currentChunkSize = fileSize - offset
		}

		// 计算进度 (50-95%)
		progress := 50 + int(float64(i)/float64(totalChunks)*45)
		runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId,
			fmt.Sprintf("%d|下载第 %d/%d 块", progress, i+1, totalChunks))

		// 读取文件块
		chunkCmd := fmt.Sprintf("dd if=%q bs=1 skip=%d count=%d 2>/dev/null | base64 -w 0",
			filePath, offset, currentChunkSize)

		chunkOutput, err := ExecuteBackendCommand(chunkCmd, sessionId)
		if err != nil {
			runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId, "1|错误：下载文件块失败")
			return entity.ErrorOnlyMsgStr("下载文件块失败")
		}

		chunks = append(chunks, string(chunkOutput))
		downloadedSize += currentChunkSize

		runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId,
			fmt.Sprintf("%d|已下载: %s/%s", progress, formatFileSize(downloadedSize), formatFileSize(fileSize)))
	}

	runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId, "95|合并文件块...")

	// 合并所有块
	fullContent := strings.Join(chunks, "")

	// 返回结果
	result := map[string]interface{}{
		"fileName":    fileName,
		"content":     fullContent,
		"encoding":    "base64",
		"fileSize":    fileSize,
		"filePath":    filePath,
		"chunked":     true,
		"totalChunks": totalChunks,
	}

	runtime.EventsEmit(appCtx, "download_rate_call_"+sessionId, "100|文件下载完成！")
	return entity.SuccessOnlyDataStr(result)
}

// 获取文件大小（用于进度显示）
func GetFileSize(sessionId, filePath string) string {

	sizeCmd := fmt.Sprintf("stat -c %%s %q", filePath)
	output, err := ExecuteBackendCommand(sizeCmd, sessionId)
	if err != nil {
		return entity.ErrorOnlyMsgStr("获取文件大小失败")
	}

	return entity.SuccessOnlyDataStr(map[string]string{
		"size": strings.TrimSpace(string(output)),
	})
}

func NewFolder(filePath, fileName, sessionId string) string {
	newPath := fmt.Sprintf("%s/%s", filePath, fileName)
	cmd := fmt.Sprintf("mkdir %s", newPath)
	_, err := ExecuteBackendCommand(cmd, sessionId)
	if err != nil {
		return entity.ErrorOnlyMsgStr("创建文件夹失败")
	}
	return entity.SuccessOnlyDataStr(newPath)
}

// 安全删除文件或文件夹
func DeleteItem(path, sessionId string) string {

	// 1. 检查路径是否存在
	checkCmd := fmt.Sprintf("[ -e %q ] && echo 'exists' || echo 'not exists'", path)
	checkOutput, err := ExecuteBackendCommand(checkCmd, sessionId)
	if err != nil || strings.TrimSpace(string(checkOutput)) != "exists" {
		return entity.ErrorOnlyMsgStr("路径不存在")
	}

	// 2. 检查是否是危险系统路径（安全保护）
	if isDangerousPath(path) {
		return entity.ErrorOnlyMsgStr("禁止删除系统保护路径")
	}

	// 3. 判断是文件还是文件夹，使用合适的命令
	var deleteCmd string
	checkTypeCmd := fmt.Sprintf("[ -f %q ] && echo 'file' || echo 'folder'", path)
	typeOutput, err := ExecuteBackendCommand(checkTypeCmd, sessionId)
	if err != nil {
		return entity.ErrorOnlyMsgStr("无法判断路径类型")
	}

	pathType := strings.TrimSpace(string(typeOutput))
	if pathType == "file" {
		deleteCmd = fmt.Sprintf("rm -f %q", path) // 删除文件
	} else {
		deleteCmd = fmt.Sprintf("rm -rf %q", path) // 删除文件夹
	}

	// 4. 执行删除
	output, err := ExecuteBackendCommand(deleteCmd, sessionId)
	if err != nil {
		fmt.Printf("删除失败: %v, 输出: %s\n", err, string(output))
		return entity.ErrorOnlyMsgStr("删除失败")
	}

	return entity.SuccessOnlyMsgStr("删除成功")
}

// 检查是否是危险路径
func isDangerousPath(path string) bool {
	// 清理路径（去除末尾斜杠）
	cleanPath := strings.TrimRight(path, "/")
	if cleanPath == "" {
		cleanPath = "/"
	}

	// 危险系统路径列表
	dangerousPaths := []string{
		"/", "/bin", "/boot", "/dev", "/etc", "/home",
		"/lib", "/lib64", "/opt", "/proc", "/root",
		"/sbin", "/sys", "/usr", "/var",
	}

	for _, dangerousPath := range dangerousPaths {
		if cleanPath == dangerousPath {
			return true
		}
	}

	// 检查是否包含危险模式
	dangerousPatterns := []string{
		"*/bin/*", "*/etc/*", "*/lib/*", "*/sbin/*", "*/usr/*", "*/var/*",
	}

	for _, pattern := range dangerousPatterns {
		matched, _ := filepath.Match(pattern, cleanPath)
		if matched {
			return true
		}
	}

	return false
}

func MoveOrCopyFolderAndFile(nowDir, copyDir, newName, sessionId string, isCopy int) string {
	// 参数检查
	if nowDir == "" || copyDir == "" || newName == "" || sessionId == "" {
		return entity.ErrorOnlyMsgStr("参数不能为空")
	}

	// 清理路径
	nowDir = strings.TrimSpace(nowDir)
	copyDir = strings.TrimSpace(copyDir)
	newName = strings.TrimSpace(newName)

	// 去除末尾斜杠
	nowDir = strings.TrimSuffix(nowDir, "/")
	copyDir = strings.TrimSuffix(copyDir, "/")

	// 检查源路径是否存在
	checkCmd := fmt.Sprintf("[ -e %q ] && echo 'exists' || echo 'not exists'", copyDir)
	checkOutput, err := ExecuteBackendCommand(checkCmd, sessionId)
	if err != nil || strings.TrimSpace(checkOutput) != "exists" {
		return entity.ErrorOnlyMsgStr("源路径不存在")
	}

	// 构建目标路径
	targetPath := fmt.Sprintf("%s/%s", nowDir, newName)
	targetPath = strings.ReplaceAll(targetPath, "//", "/")

	// 检查目标路径是否已存在
	checkTargetCmd := fmt.Sprintf("[ -e %q ] && echo 'exists' || echo 'not exists'", targetPath)
	checkTargetOutput, err := ExecuteBackendCommand(checkTargetCmd, sessionId)
	if err == nil && strings.TrimSpace(checkTargetOutput) == "exists" {
		return entity.ErrorOnlyMsgStr("目标路径已存在")
	}

	var cmd string
	var operation string

	if isCopy == 1 {
		// 复制操作
		operation = "复制"

		// 判断是文件还是文件夹
		checkTypeCmd := fmt.Sprintf("[ -f %q ] && echo 'file' || echo 'folder'", copyDir)
		typeOutput, err := ExecuteBackendCommand(checkTypeCmd, sessionId)
		if err != nil {
			return entity.ErrorOnlyMsgStr("无法判断源路径类型")
		}

		pathType := strings.TrimSpace(typeOutput)
		if pathType == "file" {
			// 复制文件
			cmd = fmt.Sprintf("cp %q %q", copyDir, targetPath)
		} else {
			// 复制文件夹（递归复制）
			cmd = fmt.Sprintf("cp -r %q %q", copyDir, targetPath)
		}
	} else {
		// 移动操作
		operation = "移动"
		cmd = fmt.Sprintf("mv %q %q", copyDir, targetPath)
	}

	// 执行操作
	output, err := ExecuteBackendCommand(cmd, sessionId)
	if err != nil {
		fmt.Printf("%s失败: %v, 输出: %s\n", operation, err, output)
		return entity.ErrorOnlyMsgStr(operation + "失败")
	}

	// 验证操作是否成功
	verifyCmd := fmt.Sprintf("[ -e %q ] && echo 'success' || echo 'failed'", targetPath)
	verifyOutput, err := ExecuteBackendCommand(verifyCmd, sessionId)
	if err != nil || strings.TrimSpace(verifyOutput) != "success" {
		return entity.ErrorOnlyMsgStr(operation + "验证失败")
	}

	// 返回成功信息
	result := map[string]string{
		"operation": operation,
		"source":    copyDir,
		"target":    targetPath,
	}

	return entity.SuccessOnlyDataStr(result)
}

func RenameFolderOrFile(filePath, newFileName, sessionId string) string {
	// 参数检查
	if filePath == "" || newFileName == "" || sessionId == "" {
		return entity.ErrorOnlyMsgStr("参数不能为空")
	}

	// 清理参数
	filePath = strings.TrimSpace(filePath)
	newFileName = strings.TrimSpace(newFileName)

	// 去除路径末尾的斜杠并确保使用正确的路径分隔符
	filePath = strings.TrimSuffix(filePath, "/")
	filePath = strings.ReplaceAll(filePath, "\\", "/") // 确保使用正斜杠

	// 检查文件名合法性
	if !isValidFileName(newFileName) {
		return entity.ErrorOnlyMsgStr("文件名包含非法字符")
	}

	// 检查源文件/文件夹是否存在
	checkExistCmd := fmt.Sprintf("[ -e %q ] && echo 'exists' || echo 'not exists'", filePath)
	checkOutput, err := ExecuteBackendCommand(checkExistCmd, sessionId)
	if err != nil || strings.TrimSpace(checkOutput) != "exists" {
		return entity.ErrorOnlyMsgStr("源路径不存在: " + filePath)
	}

	// 获取源路径的目录
	dirPath := filepath.Dir(filePath)
	dirPath = strings.ReplaceAll(dirPath, "\\", "/") // 确保使用正斜杠
	oldFileName := filepath.Base(filePath)

	// 如果新文件名和原文件名相同，直接返回成功
	if oldFileName == newFileName {
		return entity.SuccessOnlyMsgStr("文件名未改变")
	}

	// 构建目标路径（确保正确的路径分隔符）
	targetPath := fmt.Sprintf("%s/%s", dirPath, newFileName)
	targetPath = strings.ReplaceAll(targetPath, "\\", "/") // 替换所有反斜杠
	targetPath = strings.ReplaceAll(targetPath, "//", "/") // 去除重复斜杠

	// 检查目标路径是否已存在
	checkTargetCmd := fmt.Sprintf("[ -e %q ] && echo 'exists' || echo 'not exists'", targetPath)
	checkTargetOutput, err := ExecuteBackendCommand(checkTargetCmd, sessionId)
	if err == nil && strings.TrimSpace(checkTargetOutput) == "exists" {
		return entity.ErrorOnlyMsgStr("目标名称已存在: " + targetPath)
	}

	// 调试信息
	fmt.Printf("重命名: %s -> %s\n", filePath, targetPath)

	// 执行重命名命令
	cmd := fmt.Sprintf("mv -- %q %q", filePath, targetPath)
	output, err := ExecuteBackendCommand(cmd, sessionId)
	if err != nil {
		fmt.Printf("重命名失败: %v, 输出: %s, 命令: %s\n", err, output, cmd)
		return entity.ErrorOnlyMsgStr("重命名失败: " + err.Error())
	}

	// 验证重命名是否成功
	verifySourceCmd := fmt.Sprintf("[ ! -e %q ] && echo 'source_gone' || echo 'source_exists'", filePath)
	verifyTargetCmd := fmt.Sprintf("[ -e %q ] && echo 'target_exists' || echo 'target_missing'", targetPath)

	sourceOutput, _ := ExecuteBackendCommand(verifySourceCmd, sessionId)
	targetOutput, _ := ExecuteBackendCommand(verifyTargetCmd, sessionId)

	if strings.TrimSpace(sourceOutput) != "source_gone" || strings.TrimSpace(targetOutput) != "target_exists" {
		return entity.ErrorOnlyMsgStr("重命名验证失败")
	}

	// 返回成功信息
	result := map[string]string{
		"operation": "重命名",
		"oldPath":   filePath,
		"newPath":   targetPath,
		"oldName":   oldFileName,
		"newName":   newFileName,
		"message":   "重命名成功",
	}

	return entity.SuccessOnlyDataStr(result)
}

// 更安全的路径构建函数
func buildSafePath(dirPath, fileName string) string {
	// 清理路径
	dirPath = strings.TrimSpace(dirPath)
	fileName = strings.TrimSpace(fileName)

	// 确保使用正斜杠
	dirPath = strings.ReplaceAll(dirPath, "\\", "/")

	// 去除末尾斜杠
	dirPath = strings.TrimSuffix(dirPath, "/")

	// 构建路径
	if dirPath == "" || dirPath == "/" {
		return "/" + fileName
	}

	return dirPath + "/" + fileName
}

// 文件名合法性检查
func isValidFileName(filename string) bool {
	if filename == "" || filename == "." || filename == ".." {
		return false
	}

	// 禁止使用的字符（Windows和Linux都禁止的字符）
	illegalChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range illegalChars {
		if strings.Contains(filename, char) {
			return false
		}
	}

	// 禁止以空格开头或结尾
	if strings.HasPrefix(filename, " ") || strings.HasSuffix(filename, " ") {
		return false
	}

	// 禁止使用系统保留名称（Windows）
	reservedNames := []string{"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}

	upperName := strings.ToUpper(filename)
	for _, reserved := range reservedNames {
		if upperName == reserved {
			return false
		}
	}

	// 文件名长度限制
	if len(filename) > 255 {
		return false
	}

	return true
}

// 或者使用更简单的版本
func RenameFolderOrFileSimple(filePath, newFileName, sessionId string) string {
	// 参数清理
	filePath = strings.TrimSpace(filePath)
	newFileName = strings.TrimSpace(newFileName)
	filePath = strings.ReplaceAll(filePath, "\\", "/")

	// 获取目录和文件名
	dirPath := filepath.Dir(filePath)
	dirPath = strings.ReplaceAll(dirPath, "\\", "/")

	// 构建目标路径
	targetPath := buildSafePath(dirPath, newFileName)

	fmt.Printf("重命名: %s -> %s\n", filePath, targetPath)

	// 执行命令
	cmd := fmt.Sprintf("mv -- \"%s\" \"%s\"", filePath, targetPath)
	output, err := ExecuteBackendCommand(cmd, sessionId)
	if err != nil {
		return entity.ErrorOnlyMsgStr(fmt.Sprintf("重命名失败: %v, 输出: %s", err, output))
	}

	return entity.SuccessOnlyMsgStr("重命名成功")
}

func SaveNewFile(fileName, fileContent, filePath, sessionId string) string {
	// 参数检查
	if fileName == "" || filePath == "" || sessionId == "" {
		return entity.ErrorOnlyMsgStr("文件名、文件路径和会话ID不能为空")
	}

	// 清理参数
	fileName = strings.TrimSpace(fileName)
	filePath = strings.TrimSpace(filePath)
	filePath = strings.TrimSuffix(filePath, "/")
	filePath = strings.ReplaceAll(filePath, "\\", "/") // 统一路径分隔符

	// 检查文件名合法性
	if !isValidFileName(fileName) {
		return entity.ErrorOnlyMsgStr("文件名包含非法字符")
	}

	// 构建完整文件路径
	fullPath := fmt.Sprintf("%s/%s", filePath, fileName)
	fullPath = strings.ReplaceAll(fullPath, "//", "/")

	// 检查目标目录是否存在，如果不存在则创建
	dirPath := filepath.Dir(fullPath)
	checkDirCmd := fmt.Sprintf("[ -d %q ] && echo 'exists' || echo 'not exists'", dirPath)
	dirOutput, err := ExecuteBackendCommand(checkDirCmd, sessionId)
	if err != nil || strings.TrimSpace(dirOutput) != "exists" {
		// 创建目录
		mkdirCmd := fmt.Sprintf("mkdir -p %q", dirPath)
		_, err := ExecuteBackendCommand(mkdirCmd, sessionId)
		if err != nil {
			return entity.ErrorOnlyMsgStr("创建目录失败: " + dirPath)
		}
	}

	// 检查文件是否已存在
	checkFileCmd := fmt.Sprintf("[ -f %q ] && echo 'exists' || echo 'not exists'", fullPath)
	fileOutput, err := ExecuteBackendCommand(checkFileCmd, sessionId)
	if err == nil && strings.TrimSpace(fileOutput) == "exists" {
		return entity.ErrorOnlyMsgStr("文件已存在: " + fileName)
	}

	// 将文件内容进行Base64编码，避免特殊字符问题
	encodedContent := base64.StdEncoding.EncodeToString([]byte(fileContent))

	// 创建临时文件并写入内容
	tempFile := fmt.Sprintf("/tmp/temp_file_%d.txt", time.Now().UnixNano())

	// 使用echo和base64解码写入文件（更可靠的方式）
	createTempCmd := fmt.Sprintf("echo %s | base64 -d > %q", encodedContent, tempFile)
	_, err = ExecuteBackendCommand(createTempCmd, sessionId)
	if err != nil {
		return entity.ErrorOnlyMsgStr("创建临时文件失败")
	}

	// 移动临时文件到目标位置
	moveCmd := fmt.Sprintf("mv %q %q", tempFile, fullPath)
	_, err = ExecuteBackendCommand(moveCmd, sessionId)
	if err != nil {
		// 清理临时文件
		cleanupCmd := fmt.Sprintf("rm -f %q", tempFile)
		ExecuteBackendCommand(cleanupCmd, sessionId)
		return entity.ErrorOnlyMsgStr("保存文件失败: " + err.Error())
	}

	// 设置文件权限（可选）
	chmodCmd := fmt.Sprintf("chmod 644 %q", fullPath)
	ExecuteBackendCommand(chmodCmd, sessionId)

	// 验证文件是否创建成功
	verifyCmd := fmt.Sprintf("[ -f %q ] && echo 'success' || echo 'failed'", fullPath)
	verifyOutput, err := ExecuteBackendCommand(verifyCmd, sessionId)
	if err != nil || strings.TrimSpace(verifyOutput) != "success" {
		return entity.ErrorOnlyMsgStr("文件创建验证失败")
	}

	// 获取文件信息
	infoCmd := fmt.Sprintf("stat -c '%%s|%%y' %q", fullPath)
	infoOutput, err := ExecuteBackendCommand(infoCmd, sessionId)
	if err != nil {
		// 即使获取信息失败，文件还是创建成功了
		infoOutput = "unknown"
	}

	infoParts := strings.SplitN(strings.TrimSpace(infoOutput), "|", 2)
	fileSize := "unknown"
	modifyTime := "unknown"
	if len(infoParts) == 2 {
		fileSize = infoParts[0]
		modifyTime = infoParts[1]
	}

	// 返回成功信息
	result := map[string]interface{}{
		"operation":     "创建文件",
		"fileName":      fileName,
		"filePath":      fullPath,
		"fileSize":      fileSize,
		"modifyTime":    modifyTime,
		"contentLength": len(fileContent),
	}

	return entity.SuccessOnlyDataStr(result)
}
func UploadDirOrFile(targetUploadPath, willUploadPath, sessionId string) string {
	// 详细的调试信息
	fmt.Printf("UploadDirOrFile 参数接收: targetUploadPath=%s (type: %T)\n", targetUploadPath, targetUploadPath)
	fmt.Printf("UploadDirOrFile 参数接收: willUploadPath=%s (type: %T)\n", willUploadPath, willUploadPath)
	fmt.Printf("UploadDirOrFile 参数接收: sessionId=%s (type: %T)\n", sessionId, sessionId)

	// 参数检查
	if targetUploadPath == "" {
		fmt.Println("错误: targetUploadPath 为空")
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "1|错误：目标上传路径不能为空")
		return entity.ErrorOnlyMsgStr("目标上传路径不能为空")
	}
	if willUploadPath == "" {
		fmt.Println("错误: willUploadPath 为空")
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "1|错误：目标Linux路径不能为空")
		return entity.ErrorOnlyMsgStr("目标Linux路径不能为空")
	}
	if sessionId == "" {
		fmt.Println("错误: sessionId 为空")
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "1|错误：会话ID不能为空")
		return entity.ErrorOnlyMsgStr("会话ID不能为空")
	}

	// 检查会话是否存在
	sessionData, exists := sessions[sessionId]
	if !exists {
		fmt.Printf("错误: 会话不存在, sessionId=%s\n", sessionId)
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "1|错误：会话不存在或已过期")
		return entity.ErrorOnlyMsgStr("会话不存在或已过期")
	}

	if sessionData.Client == nil {
		fmt.Printf("错误: SSH客户端为nil, 会话ID: %s\n", sessionId)
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "1|错误：SSH客户端未初始化")
		return entity.ErrorOnlyMsgStr("SSH客户端未初始化")
	}

	// 发送初始进度
	runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "0|开始检查文件...")

	// 清理路径
	willUploadPath = strings.TrimSpace(willUploadPath)
	targetUploadPath = strings.TrimSpace(targetUploadPath)
	targetUploadPath = strings.ReplaceAll(targetUploadPath, "\\", "/")
	targetUploadPath = strings.TrimSuffix(targetUploadPath, "/")

	// 检查本地文件/文件夹是否存在
	if _, err := os.Stat(willUploadPath); os.IsNotExist(err) {
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "1|错误：本地路径不存在")
		return entity.ErrorOnlyMsgStr("本地路径不存在: " + willUploadPath)
	}

	// 获取文件信息判断是文件还是文件夹
	fileInfo, err := os.Stat(willUploadPath)
	if err != nil {
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "1|错误：无法获取文件信息")
		return entity.ErrorOnlyMsgStr("无法获取本地文件信息: " + err.Error())
	}

	// 创建远程目录
	runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "5|创建远程目录...")
	mkdirCmd := fmt.Sprintf("mkdir -p '%s'", targetUploadPath)
	_, err = ExecuteBackendCommand(mkdirCmd, sessionId)
	if err != nil {
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "1|错误：创建远程目录失败")
		return entity.ErrorOnlyMsgStr("创建远程目录失败: " + err.Error())
	}

	var result map[string]interface{}
	var uploadType string

	if fileInfo.IsDir() {
		// 上传文件夹
		uploadType = "文件夹"
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "10|开始上传文件夹...")
		result = uploadDirectory(willUploadPath, targetUploadPath, sessionId, sessionData)
	} else {
		// 上传文件
		uploadType = "文件"
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "10|开始上传文件...")
		result = uploadSingleFile(willUploadPath, targetUploadPath, sessionId, sessionData)
	}

	if result["success"].(bool) {
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "100|"+uploadType+"上传完成！")
		return entity.SuccessOnlyDataStr(map[string]interface{}{
			"operation":  "上传" + uploadType,
			"localPath":  willUploadPath,
			"remotePath": targetUploadPath,
			"uploadType": uploadType,
			"message":    uploadType + "上传成功",
			"details":    result,
		})
	} else {
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "1|错误："+uploadType+"上传失败")
		return entity.ErrorOnlyMsgStr(uploadType + "上传失败: " + result["error"].(string))
	}
}

// 上传单个文件（带进度）
func uploadSingleFile(localFilePath, remoteDirPath, sessionId string, sessionData *SessionData) map[string]interface{} {
	runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "15|准备上传文件...")

	// 获取文件名
	fileName := filepath.Base(localFilePath)

	// 正确构建远程路径（统一使用正斜杠）
	remoteFilePath := fmt.Sprintf("%s/%s", remoteDirPath, fileName)
	remoteFilePath = strings.ReplaceAll(remoteFilePath, "\\", "/") // 替换所有反斜杠
	remoteFilePath = strings.ReplaceAll(remoteFilePath, "//", "/") // 去除重复斜杠

	fmt.Printf("上传文件: 本地=%s, 远程=%s\n", localFilePath, remoteFilePath)

	// 打开本地文件
	file, err := os.Open(localFilePath)
	if err != nil {
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "1|错误：打开本地文件失败")
		return map[string]interface{}{
			"success": false,
			"error":   "打开本地文件失败: " + err.Error(),
		}
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "1|错误：获取文件信息失败")
		return map[string]interface{}{
			"success": false,
			"error":   "获取文件信息失败: " + err.Error(),
		}
	}

	fileSize := fileStat.Size()
	runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, fmt.Sprintf("20|文件大小: %s", formatFileSize(fileSize)))

	// 使用简单的 Base64 方式上传（便于进度跟踪）
	runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "25|读取文件内容...")
	content, err := os.ReadFile(localFilePath)
	if err != nil {
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "1|错误：读取文件失败")
		return map[string]interface{}{
			"success": false,
			"error":   "读取文件失败: " + err.Error(),
		}
	}

	runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "40|编码文件内容...")
	encodedContent := base64.StdEncoding.EncodeToString(content)

	runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "60|上传到服务器...")

	// 正确构建命令（确保路径用引号包裹）
	createCmd := fmt.Sprintf("echo %s | base64 -d > '%s'", encodedContent, remoteFilePath)
	fmt.Printf("执行命令: %s\n", createCmd)

	output, err := ExecuteBackendCommand(createCmd, sessionId)
	if err != nil {
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "1|错误：创建远程文件失败")
		return map[string]interface{}{
			"success": false,
			"error":   "创建远程文件失败: " + err.Error(),
			"output":  output,
		}
	}

	runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "80|验证文件完整性...")
	// 验证文件是否上传成功
	verifyCmd := fmt.Sprintf("[ -f '%s' ] && echo 'success' || echo 'failed'", remoteFilePath)
	verifyOutput, err := ExecuteBackendCommand(verifyCmd, sessionId)
	if err != nil || strings.TrimSpace(verifyOutput) != "success" {
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "1|错误：文件验证失败")
		return map[string]interface{}{
			"success": false,
			"error":   "文件验证失败: " + verifyOutput,
		}
	}

	// 获取远程文件大小验证
	sizeCmd := fmt.Sprintf("stat -c %%s '%s'", remoteFilePath)
	remoteSizeOutput, err := ExecuteBackendCommand(sizeCmd, sessionId)
	if err == nil {
		remoteSize, _ := strconv.ParseInt(strings.TrimSpace(remoteSizeOutput), 10, 64)
		if remoteSize != fileSize {
			runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "90|警告：文件大小不一致，重新验证...")
			// 可以在这里添加重试逻辑
		}
	}

	runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "95|文件上传成功")
	return map[string]interface{}{
		"success":    true,
		"localFile":  localFilePath,
		"remoteFile": remoteFilePath,
		"fileSize":   fileSize,
		"uploadedAt": time.Now().Format(time.RFC3339),
	}
}

// 上传整个目录（带进度）
func uploadDirectory(localDirPath, remoteDirPath, sessionId string, sessionData *SessionData) map[string]interface{} {
	runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "15|扫描文件夹内容...")

	var uploadedFiles []map[string]interface{}
	var failedFiles []map[string]interface{}

	// 首先统计文件总数和总大小
	totalFiles := 0
	totalSize := int64(0)

	filepath.Walk(localDirPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			totalFiles++
			totalSize += info.Size()
		}
		return nil
	})

	runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId,
		fmt.Sprintf("20|发现 %d 个文件，总大小: %s", totalFiles, formatFileSize(totalSize)))

	if totalFiles == 0 {
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "100|空文件夹，无需上传")
		return map[string]interface{}{
			"success":       true,
			"uploadedCount": 0,
			"failedCount":   0,
			"message":       "空文件夹",
		}
	}

	currentFile := 0
	currentSize := int64(0)

	// 遍历本地目录并上传
	err := filepath.Walk(localDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(localDirPath, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil // 跳过根目录本身
		}

		// 正确构建远程路径（统一使用正斜杠）
		remotePath := fmt.Sprintf("%s/%s", remoteDirPath, relPath)
		remotePath = strings.ReplaceAll(remotePath, "\\", "/") // 替换所有反斜杠
		remotePath = strings.ReplaceAll(remotePath, "//", "/") // 去除重复斜杠

		if info.IsDir() {
			// 创建远程目录（确保路径用引号包裹）
			mkdirCmd := fmt.Sprintf("mkdir -p '%s'", remotePath)
			_, err := ExecuteBackendCommand(mkdirCmd, sessionId)
			if err != nil {
				failedFiles = append(failedFiles, map[string]interface{}{
					"path":  path,
					"error": "创建目录失败: " + err.Error(),
					"type":  "directory",
				})
			}
		} else {
			currentFile++
			currentSize += info.Size()

			// 计算进度 (20-95%)
			progress := 20 + int(float64(currentSize)/float64(totalSize)*75)
			runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId,
				fmt.Sprintf("%d|上传文件 %d/%d: %s", progress, currentFile, totalFiles, filepath.Base(path)))

			// 上传文件（传递正确的远程目录路径）
			remoteDir := filepath.Dir(remotePath)
			remoteDir = strings.ReplaceAll(remoteDir, "\\", "/") // 确保目录路径正确
			result := uploadSingleFile(path, remoteDir, sessionId, sessionData)

			if result["success"].(bool) {
				uploadedFiles = append(uploadedFiles, map[string]interface{}{
					"localPath":  path,
					"remotePath": remotePath,
					"size":       result["fileSize"],
				})
			} else {
				failedFiles = append(failedFiles, map[string]interface{}{
					"path":  path,
					"error": result["error"].(string),
					"type":  "file",
				})
			}
		}

		return nil
	})

	if err != nil {
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "1|错误：遍历目录失败")
		return map[string]interface{}{
			"success": false,
			"error":   "遍历目录失败: " + err.Error(),
			"details": map[string]interface{}{
				"uploadedFiles": uploadedFiles,
				"failedFiles":   failedFiles,
			},
		}
	}

	// 计算最终结果
	success := len(failedFiles) == 0
	if success {
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId, "100|文件夹上传完成！")
	} else {
		runtime.EventsEmit(appCtx, "upload_rate_call_"+sessionId,
			fmt.Sprintf("95|上传完成，有 %d 个文件失败", len(failedFiles)))
	}

	return map[string]interface{}{
		"success":       success,
		"uploadedCount": len(uploadedFiles),
		"failedCount":   len(failedFiles),
		"totalFiles":    totalFiles,
		"totalSize":     totalSize,
		"uploadedFiles": uploadedFiles,
		"failedFiles":   failedFiles,
	}
}

// 格式化文件大小
func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	} else {
		return fmt.Sprintf("%.1f GB", float64(size)/(1024*1024*1024))
	}
}
