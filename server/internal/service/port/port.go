// ==========================================================================
// OmniWire - 端口管理服务层
// ==========================================================================

package port

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"omniwire/api/v1/port"
)

// PortCheckInfo 端口检查信息
type PortCheckInfo struct {
	InUse   bool
	Process string
	PID     int
}

// FirewallStatus 防火墙状态
type FirewallStatus struct {
	Enabled bool
	Rules   []*port.FirewallRule
}

// Scan 扫描端口
func Scan(ctx context.Context, startPort, endPort int) ([]*port.PortInfo, error) {
	ports := make([]*port.PortInfo, 0)

	concurrency := g.Cfg().MustGet(ctx, "port.scanConcurrency", 100).Int()
	timeout := g.Cfg().MustGet(ctx, "port.scanTimeout", 100).Int()

	var wg sync.WaitGroup
	results := make(chan *port.PortInfo, endPort-startPort+1)
	sem := make(chan struct{}, concurrency)

	for p := startPort; p <= endPort; p++ {
		wg.Add(1)
		go func(portNum int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// 检查 TCP
			addr := fmt.Sprintf("127.0.0.1:%d", portNum)
			conn, err := net.DialTimeout("tcp", addr, time.Duration(timeout)*time.Millisecond)
			if err == nil {
				conn.Close()
				results <- &port.PortInfo{
					Port:     portNum,
					Protocol: "tcp",
					State:    "listen",
				}
			}
		}(p)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for p := range results {
		ports = append(ports, p)
	}

	return ports, nil
}

// Check 检查端口占用
func Check(ctx context.Context, portNum int) (*PortCheckInfo, error) {
	info := &PortCheckInfo{}

	// 尝试连接端口
	addr := fmt.Sprintf("127.0.0.1:%d", portNum)
	conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
	if err == nil {
		conn.Close()
		info.InUse = true
	}

	// 获取进程信息
	if runtime.GOOS == "linux" {
		cmd := exec.Command("sh", "-c", fmt.Sprintf("lsof -i :%d -t 2>/dev/null | head -1", portNum))
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			pidStr := strings.TrimSpace(string(output))
			info.PID, _ = strconv.Atoi(pidStr)

			// 获取进程名
			cmd = exec.Command("ps", "-p", pidStr, "-o", "comm=")
			output, err = cmd.Output()
			if err == nil {
				info.Process = strings.TrimSpace(string(output))
			}
		}
	}

	return info, nil
}

// GetListeningPorts 获取所有监听端口
func GetListeningPorts(ctx context.Context) ([]*port.PortInfo, error) {
	ports := make([]*port.PortInfo, 0)

	if runtime.GOOS == "linux" {
		// 使用 ss 命令获取监听端口
		cmd := exec.Command("ss", "-tlnp")
		output, err := cmd.Output()
		if err != nil {
			// 尝试使用 netstat
			cmd = exec.Command("netstat", "-tlnp")
			output, err = cmd.Output()
			if err != nil {
				return ports, nil
			}
		}

		lines := strings.Split(string(output), "\n")
		for i, line := range lines {
			if i == 0 || line == "" {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) < 4 {
				continue
			}

			// 解析地址和端口
			addrPort := fields[3]
			parts := strings.Split(addrPort, ":")
			if len(parts) < 2 {
				continue
			}

			portNum, err := strconv.Atoi(parts[len(parts)-1])
			if err != nil {
				continue
			}

			p := &port.PortInfo{
				Port:     portNum,
				Protocol: "tcp",
				State:    "listen",
				Address:  strings.Join(parts[:len(parts)-1], ":"),
			}

			// 解析进程信息
			if len(fields) >= 6 {
				procInfo := fields[5]
				if strings.Contains(procInfo, "users:") {
					// 格式: users:(("nginx",pid=1234,fd=5))
					if idx := strings.Index(procInfo, "pid="); idx != -1 {
						pidStr := procInfo[idx+4:]
						if endIdx := strings.IndexAny(pidStr, ",)"); endIdx != -1 {
							pidStr = pidStr[:endIdx]
							p.PID, _ = strconv.Atoi(pidStr)
						}
					}
					if idx := strings.Index(procInfo, "((\""); idx != -1 {
						nameStr := procInfo[idx+3:]
						if endIdx := strings.Index(nameStr, "\""); endIdx != -1 {
							p.Process = nameStr[:endIdx]
						}
					}
				}
			}

			ports = append(ports, p)
		}
	}

	return ports, nil
}

// GetConnections 获取端口连接信息
func GetConnections(ctx context.Context, portNum int) ([]*port.ConnectionInfo, error) {
	conns := make([]*port.ConnectionInfo, 0)

	if runtime.GOOS == "linux" {
		cmd := exec.Command("ss", "-tn", fmt.Sprintf("sport = :%d or dport = :%d", portNum, portNum))
		output, err := cmd.Output()
		if err != nil {
			return conns, nil
		}

		lines := strings.Split(string(output), "\n")
		for i, line := range lines {
			if i == 0 || line == "" {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) < 5 {
				continue
			}

			c := &port.ConnectionInfo{
				State:    fields[0],
				Protocol: "tcp",
			}

			// 解析本地地址
			localParts := strings.Split(fields[3], ":")
			if len(localParts) >= 2 {
				c.LocalAddr = strings.Join(localParts[:len(localParts)-1], ":")
				c.LocalPort, _ = strconv.Atoi(localParts[len(localParts)-1])
			}

			// 解析远程地址
			remoteParts := strings.Split(fields[4], ":")
			if len(remoteParts) >= 2 {
				c.RemoteAddr = strings.Join(remoteParts[:len(remoteParts)-1], ":")
				c.RemotePort, _ = strconv.Atoi(remoteParts[len(remoteParts)-1])
			}

			conns = append(conns, c)
		}
	}

	return conns, nil
}

// GetFirewallStatus 获取防火墙状态
func GetFirewallStatus(ctx context.Context) (*FirewallStatus, error) {
	status := &FirewallStatus{
		Rules: make([]*port.FirewallRule, 0),
	}

	if runtime.GOOS == "linux" {
		// 检查 iptables
		cmd := exec.Command("iptables", "-L", "-n")
		_, err := cmd.Output()
		status.Enabled = err == nil
	}

	return status, nil
}

// OpenPort 开放端口
func OpenPort(ctx context.Context, portNum int, protocol string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("仅支持 Linux 系统")
	}

	protocols := []string{}
	if protocol == "both" {
		protocols = []string{"tcp", "udp"}
	} else {
		protocols = []string{protocol}
	}

	for _, proto := range protocols {
		cmd := exec.Command("iptables", "-I", "INPUT", "-p", proto, "--dport", strconv.Itoa(portNum), "-j", "ACCEPT")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("开放端口失败: %v", err)
		}
	}

	g.Log().Infof(ctx, "端口 %d/%s 已开放", portNum, protocol)
	return nil
}

// ClosePort 关闭端口
func ClosePort(ctx context.Context, portNum int, protocol string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("仅支持 Linux 系统")
	}

	protocols := []string{}
	if protocol == "both" {
		protocols = []string{"tcp", "udp"}
	} else {
		protocols = []string{protocol}
	}

	for _, proto := range protocols {
		cmd := exec.Command("iptables", "-D", "INPUT", "-p", proto, "--dport", strconv.Itoa(portNum), "-j", "ACCEPT")
		cmd.Run() // 忽略错误，规则可能不存在

		// 添加 DROP 规则
		cmd = exec.Command("iptables", "-I", "INPUT", "-p", proto, "--dport", strconv.Itoa(portNum), "-j", "DROP")
		cmd.Run()
	}

	g.Log().Infof(ctx, "端口 %d/%s 已关闭", portNum, protocol)
	return nil
}
