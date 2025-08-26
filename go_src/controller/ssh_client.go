package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/crypto/ssh"
)

// SessionData 代表一个 SSH 会话
type SessionData struct {
	mu       sync.Mutex // 保护本实例所有字段
	Client   *ssh.Client
	Session  *ssh.Session
	InPipe   io.WriteCloser
	OutPipe  io.Reader
	LastSeen time.Time
}

var (
	sessions      = make(map[string]*SessionData) // sessionId -> *SessionData
	sessionsIndex sync.RWMutex                    // 只保护 map 本身
	appCtx        context.Context
)

// SetAppContext 由 main.go 在 OnStartup 时调用
func SetAppContext(ctx context.Context) {
	appCtx = ctx
}

type Server struct {
	ServerUrl      string `json:"serverUrl"`
	ServerPort     string `json:"serverPort"`
	ServerUserName string `json:"serverUserName"`
	ServerPassword string `json:"serverPassword"`
}

// NewCon 供前端调用：建立连接
func NewCon(params string, sessionId string, ctx context.Context, rows, cols int) string {
	SetAppContext(ctx)

	var server Server
	if err := json.Unmarshal([]byte(params), &server); err != nil {
		return fmt.Sprintf("解析 JSON 失败: %v", err)
	}

	return InitSSH(server, sessionId, rows, cols)
}

// InitSSH 初始化 SSH 会话
func InitSSH(server Server, sessionId string, rows, cols int) string {
	// 1. 如果已存在，先关闭
	if sd := delSession(sessionId); sd != nil {
		sd.close() // 内部已加锁
	}

	// 2. 建立新连接
	sd, errMsg := buildSession(server, rows, cols)
	if errMsg != "" {
		return errMsg
	}

	// 3. 写入索引
	addSession(sessionId, sd)

	// 4. 启动读输出 goroutine
	go upOut(sessionId)

	return "success"
}

func buildSession(server Server, rows, cols int) (*SessionData, string) {
	config := &ssh.ClientConfig{
		User:            server.ServerUserName,
		Auth:            []ssh.AuthMethod{ssh.Password(server.ServerPassword)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	addr := fmt.Sprintf("%s:%s", server.ServerUrl, server.ServerPort)

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Sprintf("SSH 连接失败: %v", err)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Sprintf("创建会话失败: %v", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Sprintf("获取 stdin 失败: %v", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Sprintf("获取 stdout 失败: %v", err)
	}

	if err := session.RequestPty("xterm", rows, cols, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.ICANON:        1,
		ssh.ISIG:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Sprintf("申请 pty 失败: %v", err)
	}

	if err := session.Shell(); err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Sprintf("启动 shell 失败: %v", err)
	}

	fmt.Fprintf(stdin, "export PROMPT_COMMAND='printf \"\\033]777;CWD;%%s\\007\" \"$PWD\"'\n")

	return &SessionData{
		Client:   client,
		Session:  session,
		InPipe:   stdin,
		OutPipe:  stdout,
		LastSeen: time.Now(),
	}, ""
}

// getSession 只读索引
func getSession(id string) (*SessionData, bool) {
	sessionsIndex.RLock()
	sd := sessions[id]
	sessionsIndex.RUnlock()
	return sd, sd != nil
}

// addSession 写索引
func addSession(id string, sd *SessionData) {
	sessionsIndex.Lock()
	sessions[id] = sd
	sessionsIndex.Unlock()
}

// delSession 删除并返回旧实例
func delSession(id string) *SessionData {
	sessionsIndex.Lock()
	sd := sessions[id]
	delete(sessions, id)
	sessionsIndex.Unlock()
	return sd
}

// 新增：原始字节流透传
func SendRaw(data string, sessionId string) error {
	fmt.Println("透传内容:", data)
	sd, ok := getSession(sessionId)
	if !ok {
		return fmt.Errorf("session not found")
	}
	sd.mu.Lock()
	defer sd.mu.Unlock()

	_, err := sd.InPipe.Write([]byte(data))
	sd.LastSeen = time.Now()

	return err
}

// 把原来的 UpInput 改名为 SendNormal，前端 Enter 时继续用（可选）
func SendNormal(input string, sessionId string) string {
	return UpInput(input, sessionId)
}

// UpInput 内部逻辑保持不变，仅函数名变了
func UpInput(input string, sessionId string) string {
	sd, ok := getSession(sessionId)
	if !ok {
		return "error: 未连接或连接已断开"
	}
	sd.mu.Lock()
	defer sd.mu.Unlock()

	if _, err := fmt.Fprintln(sd.InPipe, input); err != nil {
		return fmt.Sprintf("发送失败: %v", err)
	}
	sd.LastSeen = time.Now()
	return "success"
}

// upOut 持续读取远端输出并透传前端，同时解析 OSC 777 提取 CWD
func upOut(sessionId string) {
	sd, ok := getSession(sessionId)
	if !ok {
		return
	}

	const oscPrefix = "\033]777;CWD;"
	const oscSuffix = "\007"

	var buf []byte
	tmp := make([]byte, 4096)

	for {
		n, err := sd.OutPipe.Read(tmp)
		if err != nil {
			return
		}
		buf = append(buf, tmp[:n]...)

		for {
			// 1. 找完整的 OSC 777 序列
			start := bytes.Index(buf, []byte(oscPrefix))
			if start == -1 {
				// 没有 OSC，直接全部透传
				if len(buf) > 0 {
					sendOut(string(buf), sessionId) // 见下方说明
					buf = buf[:0]
				}
				break
			}

			// 把 OSC 之前的内容先透传
			if start > 0 {
				sendOut(string(buf[:start]), sessionId)
			}

			rest := buf[start+len(oscPrefix):]
			end := bytes.Index(rest, []byte(oscSuffix))
			if end == -1 {
				// 不完整，继续读
				buf = buf[start:]
				break
			}

			// 2. 拿到目录并上报
			cwd := string(rest[:end])
			sendSSHDir(cwd, sessionId)

			// 3. 去掉已解析的 OSC
			buf = rest[end+len(oscSuffix):]
		}
	}
}

// 统一透传函数
func sendOut(data string, sessionId string) {
	if data == "" {
		return
	}
	if appCtx != nil {
		runtime.EventsEmit(appCtx, "terminal_output_"+sessionId, data)
	}
}

func sendSSHDir(output string, sessionId string) {
	if appCtx != nil {
		runtime.EventsEmit(appCtx, "ssh_dir_"+sessionId, output)
	}
}

// CloseSSH 供前端调用
func CloseSSH(sessionId string) string {
	sd := delSession(sessionId)
	if sd == nil {
		return "error: 会话不存在"
	}
	sd.close()
	return "success"
}

// close 会话内部统一清理
func (sd *SessionData) close() {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	_ = sd.Session.Close()
	_ = sd.Client.Close()
	_ = sd.InPipe.Close()
}

// 专门用于后端命令执行的函数
func ExecuteBackendCommand(cmd string, sessionId string) (string, error) {
	fmt.Printf("单独执行命令:%s|sessionId:%s\n", cmd, sessionId)
	sessionData, exists := sessions[sessionId]
	if !exists || sessionData.Client == nil {
		return "", fmt.Errorf("未连接")
	}

	// 每次创建新会话执行命令，避免状态污染
	session, err := sessionData.Client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建会话失败: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), fmt.Errorf("命令执行失败: %v", err)
	}

	return string(output), nil
}

func SendResize(rows int, cols int, sessionId string) error {
	fmt.Println("Resize:", rows, cols, sessionId)
	sd, ok := getSession(sessionId)
	if !ok {
		return fmt.Errorf("no session")
	}
	err := sd.Session.WindowChange(rows, cols)
	if err != nil {
		fmt.Println("WindowChange error:", err)
	}
	return err
}
