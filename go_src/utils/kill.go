package utils

import (
	"fmt"
	"os"
	"strconv"

	"github.com/mitchellh/go-ps"
	"github.com/shirou/gopsutil/net"
)

func KillByPort(strPort string) string {
	port, err := strconv.ParseInt(strPort, 10, 64)
	connections, err := net.Connections("all")
	if err != nil {
		return fmt.Sprintf("未找到连接: %v;", err)
	}
	for _, conn := range connections {
		if conn.Laddr.Port == uint32(port) {
			pid := int(conn.Pid)
			if pid == 0 {
				return fmt.Sprintf("未找到进程正在使用此端口 %d;", port)
			}

			process, err := ps.FindProcess(pid)
			if err != nil || process == nil {
				return fmt.Sprintf("未找到此进程 %d: %v;", pid, err)
			}

			osProcess, err := os.FindProcess(pid)
			if err != nil {
				return fmt.Sprintf("未能获取进程 %d: %v;", pid, err)
			}

			err = osProcess.Kill()
			if err != nil {
				return fmt.Sprintf("未能终止进程 %d: %v;", pid, err)
			}

			return fmt.Sprintf("终止进程 %d (%s) 使用的端口 %d",
				pid, process.Executable(), port)
		}
	}

	return fmt.Sprintf("未找到使用次端口 %d 的进程;", port)
}
