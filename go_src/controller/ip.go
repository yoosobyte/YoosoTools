package controller

import (
	"YoosoTools/go_src/entity"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var (
	udpPort       = 49151            // 统一端口
	udpBufferSize = 1024             // 包大小
	udpInterval   = 30 * time.Second // 自己多久发一次
	udpTimeout    = 90 * time.Second // 多久没更新算过期
	IPBox         = map[string]string{}
	PeerBox       = struct {
		sync.RWMutex
		M map[string]string
	}{M: make(map[string]string)}
)

func PingURL(url string) (int, string) {
	client := http.Client{
		Timeout: 20 * time.Second, // 设置超时时间
	}
	resp, err := client.Get(url)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	// 如果状态码是 2xx 或 3xx，认为服务可用
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, err.Error()
	}
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "utf-8") {
		bodyString := string(bodyBytes)
		return resp.StatusCode, bodyString
	} else {
		// 如果不是 UTF-8 编码，可以尝试其他编码处理
		// 这里假设服务器返回的是 GBK 编码
		bodyString := decodeGBK(bodyBytes)
		return resp.StatusCode, bodyString
	}
}

// decodeGBK 将GBK编码的字节转换为UTF-8字符串
func decodeGBK(s []byte) string {
	reader := transform.NewReader(
		bytes.NewReader(s),
		simplifiedchinese.GBK.NewDecoder(),
	)
	d, err := io.ReadAll(reader)
	if err != nil {
		return err.Error()
	}
	return string(d)
}

func GetOutboundIP() (string, error) {
	// 随便连一个公网地址即可触发出网路由
	conn, err := net.Dial("udp4", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// LocalAddr() 就是本机在那张网卡上的地址
	ip := conn.LocalAddr().(*net.UDPAddr).IP
	return ip.String(), nil
}

func GetIpInfo() string {
	code, resp := getRemoteIP()
	if code != 200 {
		fmt.Println("请求失败结果:", resp)
		return `{"ret":"ok","err":"` + resp + `","data":{"ip":"x.x.x.x","localIp":"x.x.x.x","location":["国家","省份","市区","","运营商"]}}`
	}
	localIp, err := GetOutboundIP()
	if err != nil {
		return `{"ret":"ok","err":"本地IP获取异常","data":{"ip":"x.x.x.x","localIp":"x.x.x.x","location":["国家","省份","市区","","运营商"]}}`
	}
	// 定义一个 map 来存储解析后的 JSON 数据
	var data map[string]interface{}
	// 解析 JSON 字符串
	err1 := json.Unmarshal([]byte(resp), &data)
	if err1 != nil {
		return `{"ret":"ok","err":"远程结果序列化异常","data":{"ip":"x.x.x.x","localIp":"x.x.x.x","location":["国家","省份","市区","","运营商"]}}`
	}
	// 插入新的键值对
	if dataMap, ok := data["data"].(map[string]interface{}); ok {
		// 插入新的键值对
		dataMap["localIp"] = localIp
		peerData := GetPeerIPs()
		var ips []map[string]string
		for _, v := range peerData {
			var tmp map[string]interface{}
			if err := json.Unmarshal([]byte(v), &tmp); err == nil {
				from, _ := tmp["from"].(string)
				ip, _ := tmp["ip"].(string)

				innerTmp := map[string]string{
					"localIp":  from,
					"remoteIp": ip,
				}
				ips = append(ips, innerTmp)
			}
		}
		dataMap["ipList"] = ips
	}
	// 将更新后的数据重新编码为 JSON 字符串
	updatedJsonStr, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return `{"ret":"ok","err":"远程结果反序列化异常","data":{"ip":"x.x.x.x","localIp":"x.x.x.x","location":["国家","省份","市区","","运营商"]}}`
	}
	return string(updatedJsonStr)
}

// 启动后：1.定时探测并广播本机外网IP；2.监听局域网其它节点的广播
func StartUDPBroadcastAll(ctx context.Context) {
	go broadcastLoop()
	go listenLoop(ctx)
	go expireLoop()
}

// ---------- 1. 定时探测 + 广播 ----------
func broadcastLoop() {
	// 第一次立刻发一次
	sendMyExternalIP()

	tick := time.NewTicker(udpInterval)
	defer tick.Stop()
	for range tick.C {
		sendMyExternalIP()
	}
}
func getRemoteIP() (int, string) {
	code := 0
	resp := IPBox[time.Now().Format("15:04")]
	if resp != "" {
		code = 200
	} else {
		code, resp = PingURL("https://myip.ipip.net/json")
		if !strings.Contains(resp, "err") {
			IPBox[time.Now().Format("15:04")] = resp
		}
	}
	return code, resp
}

// sendMyExternalIP 把公网 IP 发到局域网广播地址
func sendMyExternalIP() error {
	// 1. 拿到公网 IP（你已有的 getRemoteIP）
	_, resp := getRemoteIP()
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		return err
	}
	ip, _ := data["data"].(map[string]interface{})["ip"].(string)
	if ip == "" {
		return fmt.Errorf("no ip in response")
	}

	// 2. 拿到本机连路由器的 IPv4
	localIPStr, err := GetOutboundIP()
	if err != nil {
		return fmt.Errorf("get local ip: %w", err)
	}
	localIP := net.ParseIP(localIPStr).To4()
	if localIP == nil {
		return fmt.Errorf("invalid local ip")
	}

	// 3. 找到这块网卡的广播地址
	bcast, err := broadcastForIP(localIP)
	if err != nil {
		return fmt.Errorf("get broadcast: %w", err)
	}

	// 4. 组装数据
	payload := map[string]string{"ip": ip}
	b, _ := json.Marshal(payload)

	// 5. 明确从 localIP 发到本网段广播地址
	conn, err := net.DialUDP("udp4",
		&net.UDPAddr{IP: localIP, Port: 0}, // 源地址
		&net.UDPAddr{IP: bcast, Port: udpPort})
	if err != nil {
		return fmt.Errorf("dial udp: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write(b)
	return err
}

// broadcastForIP 根据本地 IPv4 计算本网段广播地址
func broadcastForIP(ip net.IP) (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			ipNet, ok := a.(*net.IPNet)
			if !ok || ipNet.IP.To4() == nil {
				continue
			}
			if ipNet.Contains(ip) {
				mask := ipNet.Mask
				bcast := make(net.IP, 4)
				for i := 0; i < 4; i++ {
					bcast[i] = ipNet.IP.To4()[i] | ^mask[i]
				}
				return bcast, nil
			}
		}
	}
	return nil, fmt.Errorf("no matching interface")
}

// 把 listenLoop() 里写 PeerBox
func listenLoop(ctx context.Context) {
	addr := net.UDPAddr{Port: udpPort, IP: net.IPv4zero}
	conn, err := net.ListenUDP("udp4", &addr)
	if err != nil {
		fmt.Println("UDP listen err:", err)
		return
	}
	fmt.Println("监听到:", addr)
	defer conn.Close()

	buf := make([]byte, udpBufferSize)
	for {
		n, peer, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		var payload struct {
			IP string `json:"ip"`
		}
		fmt.Printf("收到来自 %s 的广播: %s\n", peer.IP.String(), string(buf[:n]))
		if json.Unmarshal(buf[:n], &payload) != nil || payload.IP == "" {
			continue
		}

		record := map[string]interface{}{
			"ip":   payload.IP,
			"from": peer.IP.String(),
			"ts":   time.Now().Unix(),
		}
		j, _ := json.Marshal(record)

		// 写进 PeerBox，而不是 IPBox
		PeerBox.Lock()
		PeerBox.M[peer.IP.String()] = string(j)
		PeerBox.Unlock()
		if ctx != nil {
			runtime.EventsEmit(ctx, "new_peer_data")
		}
	}
}

// expireLoop 也只清理 PeerBox
func expireLoop() {
	for range time.Tick(udpTimeout / 2) {
		now := time.Now()
		PeerBox.Lock()
		for k, v := range PeerBox.M {
			var tmp map[string]interface{}
			if json.Unmarshal([]byte(v), &tmp) != nil {
				delete(PeerBox.M, k)
				continue
			}
			ts, _ := tmp["ts"].(float64)
			if now.Sub(time.Unix(int64(ts), 0)) > udpTimeout {
				delete(PeerBox.M, k)
			}
		}
		PeerBox.Unlock()
	}
}

// 如果外部需要读取邻居表，可以封装一个线程安全的函数
func GetPeerIPs() map[string]string {
	PeerBox.RLock()
	defer PeerBox.RUnlock()
	out := make(map[string]string, len(PeerBox.M))
	for k, v := range PeerBox.M {
		out[k] = v
	}
	return out
}

func PostRadio() string {
	sendMyExternalIP()
	return entity.SuccessStr()
}
