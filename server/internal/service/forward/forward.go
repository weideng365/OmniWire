// ==========================================================================
// OmniWire - 高性能端口转发服务层
// ==========================================================================

package forward

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"omniwire/api/v1/forward"
	"omniwire/internal/model/entity"
)

// RuleInput 规则输入
type RuleInput struct {
	Name          string
	Protocol      string
	ListenPort    int
	TargetAddr    string
	TargetPort    int
	Enabled       bool
	MaxConn       int
	UploadLimit   int64 // bytes/s, 0=无限制
	DownloadLimit int64 // bytes/s, 0=无限制
	Description   string
}

// ForwardStats 转发统计
type ForwardStats struct {
	TotalConn         int64
	CurrentConn       int32
	BytesReceived     int64
	BytesSent         int64
	LastBytesReceived int64
	LastBytesSent     int64
	UploadSpeed       int64
	DownloadSpeed     int64
	StartTime         time.Time
}

// ForwardRule 转发规则运行时
type ForwardRule struct {
	Id            int
	Name          string
	Protocol      string
	ListenPort    int
	TargetAddr    string
	TargetPort    int
	MaxConn       int
	UploadLimit   int64 // bytes/s
	DownloadLimit int64 // bytes/s
	UsePool       bool  // 是否使用连接池
	pool          *ConnPool
	listener      net.Listener
	udpConn       *net.UDPConn
	running       bool
	stopChan      chan struct{}
	stats         *ForwardStats
	mu            sync.RWMutex
}

var (
	runningRules = make(map[int]*ForwardRule)
	rulesMutex   sync.RWMutex
	// 缓冲池 - 提高内存效率
	bufferPool = sync.Pool{
		New: func() interface{} {
			buf := make([]byte, 64*1024) // 64KB 缓冲区
			return &buf
		},
	}
)

// GetTotalActiveConnections 获取所有规则的活跃连接总数
func GetTotalActiveConnections() int {
	rulesMutex.RLock()
	defer rulesMutex.RUnlock()
	var total int32
	for _, rule := range runningRules {
		total += atomic.LoadInt32(&rule.stats.CurrentConn)
	}
	return int(total)
}

// GetList 获取转发规则列表
func GetList(ctx context.Context, page, pageSize int) ([]*forward.RuleInfo, int, error) {
	rules := make([]*forward.RuleInfo, 0)
	model := g.Model("forward_rule")
	total, err := model.Count()
	if err != nil {
		return rules, 0, nil
	}

	var entityRules []*entity.ForwardRule
	err = model.Page(page, pageSize).OrderDesc("id").Scan(&entityRules)
	if err != nil {
		return rules, 0, nil
	}

	rulesMutex.RLock()
	defer rulesMutex.RUnlock()

	for _, er := range entityRules {
		rule := &forward.RuleInfo{
			Id: er.Id, Name: er.Name, Protocol: er.Protocol,
			ListenPort: er.ListenPort, TargetAddr: er.TargetAddr, TargetPort: er.TargetPort,
			Enabled: er.Enabled == 1, MaxConn: er.MaxConn, Description: er.Description,
			UploadLimit: er.UploadLimit, DownloadLimit: er.DownloadLimit,
			TotalUpload: er.TotalUpload, TotalDownload: er.TotalDownload,
			CreatedAt: er.CreatedAt.String(), UpdatedAt: er.UpdatedAt.String(),
		}
		if rr, ok := runningRules[er.Id]; ok {
			rule.Running = rr.running
			rule.CurrentConn = int(atomic.LoadInt32(&rr.stats.CurrentConn))

			// 合并运行时流量到总流量
			currentUpload := atomic.LoadInt64(&rr.stats.BytesSent)
			currentDownload := atomic.LoadInt64(&rr.stats.BytesReceived)
			rule.TotalUpload += currentUpload
			rule.TotalDownload += currentDownload

			// 获取实时速度
			rule.UploadSpeed = atomic.LoadInt64(&rr.stats.UploadSpeed)
			rule.DownloadSpeed = atomic.LoadInt64(&rr.stats.DownloadSpeed)
		}
		rules = append(rules, rule)
	}
	return rules, total, nil
}

// Create 创建转发规则
func Create(ctx context.Context, input *RuleInput) (*forward.RuleInfo, error) {
	now := time.Now()
	insertData := g.Map{
		"name": input.Name, "protocol": input.Protocol, "listen_port": input.ListenPort,
		"target_addr": input.TargetAddr, "target_port": input.TargetPort,
		"enabled": boolToInt(input.Enabled), "max_conn": input.MaxConn,
		"upload_limit": input.UploadLimit, "download_limit": input.DownloadLimit,
		"description": input.Description, "created_at": now, "updated_at": now,
	}
	result, err := g.Model("forward_rule").Insert(insertData)
	if err != nil {
		return nil, fmt.Errorf("创建规则失败: %v", err)
	}
	id, _ := result.LastInsertId()
	if input.Enabled {
		go Start(context.Background(), int(id))
	}
	return &forward.RuleInfo{
		Id: int(id), Name: input.Name, Protocol: input.Protocol,
		ListenPort: input.ListenPort, TargetAddr: input.TargetAddr, TargetPort: input.TargetPort,
		Enabled: input.Enabled, MaxConn: input.MaxConn, Description: input.Description,
		UploadLimit: input.UploadLimit, DownloadLimit: input.DownloadLimit,
		CreatedAt: now.String(), UpdatedAt: now.String(),
	}, nil
}

// Update 更新转发规则
func Update(ctx context.Context, id int, input *RuleInput) error {
	Stop(ctx, id)
	updateData := g.Map{"updated_at": time.Now(), "enabled": boolToInt(input.Enabled)}
	if input.Name != "" {
		updateData["name"] = input.Name
	}
	if input.Protocol != "" {
		updateData["protocol"] = input.Protocol
	}
	if input.ListenPort > 0 {
		updateData["listen_port"] = input.ListenPort
	}
	if input.TargetAddr != "" {
		updateData["target_addr"] = input.TargetAddr
	}
	if input.TargetPort > 0 {
		updateData["target_port"] = input.TargetPort
	}
	if input.MaxConn > 0 {
		updateData["max_conn"] = input.MaxConn
	}
	updateData["upload_limit"] = input.UploadLimit
	updateData["download_limit"] = input.DownloadLimit
	updateData["description"] = input.Description
	_, err := g.Model("forward_rule").Where("id", id).Update(updateData)
	if err != nil {
		return fmt.Errorf("更新规则失败: %v", err)
	}
	if input.Enabled {
		go Start(context.Background(), id)
	}
	return nil
}

// Delete 删除转发规则
func Delete(ctx context.Context, id int) error {
	Stop(ctx, id)
	_, err := g.Model("forward_rule").Where("id", id).Delete()
	return err
}

// Start 启动转发规则
func Start(ctx context.Context, id int) error {
	var rule entity.ForwardRule
	if err := g.Model("forward_rule").Where("id", id).Scan(&rule); err != nil {
		return fmt.Errorf("规则不存在")
	}
	fr := &ForwardRule{
		Id: rule.Id, Name: rule.Name, Protocol: rule.Protocol,
		ListenPort: rule.ListenPort, TargetAddr: rule.TargetAddr, TargetPort: rule.TargetPort,
		MaxConn: rule.MaxConn, UploadLimit: rule.UploadLimit, DownloadLimit: rule.DownloadLimit,
		stopChan: make(chan struct{}),
		stats:    &ForwardStats{StartTime: time.Now()},
	}
	var err error
	if rule.Protocol == "tcp" {
		err = startTCPForward(fr)
	} else {
		err = startUDPForward(fr)
	}
	if err != nil {
		return err
	}

	// 启动速度监控
	go startSpeedMonitor(fr)

	rulesMutex.Lock()
	runningRules[id] = fr
	rulesMutex.Unlock()
	g.Log().Infof(ctx, "[端口转发] 规则 %s 已启动 (:%d -> %s:%d)", rule.Name, rule.ListenPort, rule.TargetAddr, rule.TargetPort)
	return nil
}

// startSpeedMonitor 启动速度监控
func startSpeedMonitor(fr *ForwardRule) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-fr.stopChan:
			return
		case <-ticker.C:
			currentReceived := atomic.LoadInt64(&fr.stats.BytesReceived)
			currentSent := atomic.LoadInt64(&fr.stats.BytesSent)

			// 计算速度
			downloadSpeed := currentReceived - fr.stats.LastBytesReceived
			uploadSpeed := currentSent - fr.stats.LastBytesSent

			// 更新速度
			atomic.StoreInt64(&fr.stats.DownloadSpeed, downloadSpeed)
			atomic.StoreInt64(&fr.stats.UploadSpeed, uploadSpeed)

			// 更新上次数据
			fr.stats.LastBytesReceived = currentReceived
			fr.stats.LastBytesSent = currentSent
		}
	}
}

// Stop 停止转发规则
func Stop(ctx context.Context, id int) error {
	rulesMutex.Lock()
	rr, ok := runningRules[id]
	if ok {
		delete(runningRules, id)
	}
	rulesMutex.Unlock()
	if !ok {
		return nil
	}

	// 保存流量统计到数据库
	if rr.stats != nil {
		g.Model("forward_rule").Where("id", id).Update(g.Map{
			"total_upload":   g.DB().Raw("total_upload + ?", atomic.LoadInt64(&rr.stats.BytesSent)),
			"total_download": g.DB().Raw("total_download + ?", atomic.LoadInt64(&rr.stats.BytesReceived)),
		})
	}

	close(rr.stopChan)
	rr.running = false
	if rr.listener != nil {
		rr.listener.Close()
	}
	if rr.udpConn != nil {
		rr.udpConn.Close()
	}
	g.Log().Infof(ctx, "[端口转发] 规则 ID=%d 已停止", id)
	return nil
}

// GetStats 获取转发统计
func GetStats(ctx context.Context, id int) (*forward.RuleStats, error) {
	rulesMutex.RLock()
	rr, ok := runningRules[id]
	rulesMutex.RUnlock()
	stats := &forward.RuleStats{Id: id}
	if ok && rr.stats != nil {
		stats.TotalConn = atomic.LoadInt64(&rr.stats.TotalConn)
		stats.CurrentConn = int(atomic.LoadInt32(&rr.stats.CurrentConn))
		stats.BytesReceived = atomic.LoadInt64(&rr.stats.BytesReceived)
		stats.BytesSent = atomic.LoadInt64(&rr.stats.BytesSent)
		stats.StartTime = rr.stats.StartTime.String()
		stats.Uptime = int64(time.Since(rr.stats.StartTime).Seconds())
	}
	return stats, nil
}

// ==================== TCP 转发 (高性能优化) ====================

func startTCPForward(fr *ForwardRule) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", fr.ListenPort))
	if err != nil {
		return err
	}
	fr.listener = listener
	fr.running = true

	go func() {
		for {
			select {
			case <-fr.stopChan:
				return
			default:
				// 设置 Accept 超时，避免阻塞
				if tcpListener, ok := listener.(*net.TCPListener); ok {
					tcpListener.SetDeadline(time.Now().Add(time.Second))
				}
				conn, err := listener.Accept()
				if err != nil {
					if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
						continue
					}
					continue
				}

				// 检查最大连接数
				currentConn := atomic.LoadInt32(&fr.stats.CurrentConn)
				if int(currentConn) >= fr.MaxConn {
					conn.Close()
					continue
				}

				atomic.AddInt64(&fr.stats.TotalConn, 1)
				atomic.AddInt32(&fr.stats.CurrentConn, 1)

				// 使用 goroutine 池处理连接
				go handleTCPConn(fr, conn)
			}
		}
	}()
	return nil
}

func handleTCPConn(fr *ForwardRule, src net.Conn) {
	defer func() {
		src.Close()
		atomic.AddInt32(&fr.stats.CurrentConn, -1)
	}()

	// 设置 TCP 优化选项
	if tcpConn, ok := src.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	var dst net.Conn
	var err error

	// 使用连接池或直接连接
	if fr.UsePool && fr.pool != nil {
		pooledConn, poolErr := fr.pool.Get(context.Background())
		if poolErr != nil {
			// 连接池失败，回退到直接连接
			dst, err = net.DialTimeout("tcp", net.JoinHostPort(fr.TargetAddr, fmt.Sprintf("%d", fr.TargetPort)), 10*time.Second)
		} else {
			dst = pooledConn
			err = nil
		}
	} else {
		dst, err = net.DialTimeout("tcp", net.JoinHostPort(fr.TargetAddr, fmt.Sprintf("%d", fr.TargetPort)), 10*time.Second)
	}

	if err != nil {
		return
	}
	defer dst.Close()

	// 设置目标连接的 TCP 优化 (非连接池连接)
	if tcpConn, ok := dst.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	// 双向数据传输 (使用缓冲池)
	done := make(chan struct{}, 2)

	// 客户端 -> 服务器 (上传)
	go func() {
		defer func() { done <- struct{}{} }()
		n := copyWithStats(dst, src, fr.UploadLimit)
		atomic.AddInt64(&fr.stats.BytesSent, n)
	}()

	// 服务器 -> 客户端 (下载)
	go func() {
		defer func() { done <- struct{}{} }()
		n := copyWithStats(src, dst, fr.DownloadLimit)
		atomic.AddInt64(&fr.stats.BytesReceived, n)
	}()

	// 等待任一方向完成
	<-done

	// 半关闭连接，让另一方优雅完成
	if tcpConn, ok := src.(*net.TCPConn); ok {
		tcpConn.CloseRead()
	}
	if tcpConn, ok := dst.(*net.TCPConn); ok {
		tcpConn.CloseRead()
	}
}

// copyWithStats 带速率限制的数据复制
func copyWithStats(dst io.Writer, src io.Reader, rateLimit int64) int64 {
	// 从缓冲池获取缓冲区
	bufPtr := bufferPool.Get().(*[]byte)
	defer bufferPool.Put(bufPtr)
	buf := *bufPtr

	var total int64
	var lastTime = time.Now()
	var bytesSinceLastCheck int64

	for {
		nr, readErr := src.Read(buf)
		if nr > 0 {
			nw, writeErr := dst.Write(buf[:nr])
			if nw > 0 {
				total += int64(nw)
				bytesSinceLastCheck += int64(nw)
			}
			if writeErr != nil {
				break
			}
			if nr != nw {
				break
			}

			// 速率限制
			if rateLimit > 0 {
				elapsed := time.Since(lastTime)
				if elapsed >= 100*time.Millisecond {
					expectedBytes := rateLimit * int64(elapsed) / int64(time.Second)
					if bytesSinceLastCheck > expectedBytes {
						sleepTime := time.Duration(bytesSinceLastCheck-expectedBytes) * time.Second / time.Duration(rateLimit)
						if sleepTime > 0 && sleepTime < time.Second {
							time.Sleep(sleepTime)
						}
					}
					lastTime = time.Now()
					bytesSinceLastCheck = 0
				}
			}
		}
		if readErr != nil {
			break
		}
	}
	return total
}

// ==================== UDP 转发 (高性能优化) ====================

func startUDPForward(fr *ForwardRule) error {
	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", fr.ListenPort))
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	// 设置 UDP 缓冲区大小
	conn.SetReadBuffer(4 * 1024 * 1024)  // 4MB 接收缓冲区
	conn.SetWriteBuffer(4 * 1024 * 1024) // 4MB 发送缓冲区

	fr.udpConn = conn
	fr.running = true

	// 客户端会话管理
	type udpSession struct {
		conn       *net.UDPConn
		lastActive time.Time
	}
	clientMap := sync.Map{}

	// 定期清理过期会话
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-fr.stopChan:
				return
			case <-ticker.C:
				now := time.Now()
				clientMap.Range(func(key, value interface{}) bool {
					session := value.(*udpSession)
					if now.Sub(session.lastActive) > 2*time.Minute {
						session.conn.Close()
						clientMap.Delete(key)
						atomic.AddInt32(&fr.stats.CurrentConn, -1)
					}
					return true
				})
			}
		}
	}()

	// 多个 worker 处理 UDP 数据包
	numWorkers := 4
	for i := 0; i < numWorkers; i++ {
		go func() {
			buf := make([]byte, 65535)
			for {
				select {
				case <-fr.stopChan:
					return
				default:
					conn.SetReadDeadline(time.Now().Add(time.Second))
					n, srcAddr, err := conn.ReadFromUDP(buf)
					if err != nil {
						continue
					}

					atomic.AddInt64(&fr.stats.BytesReceived, int64(n))
					key := srcAddr.String()

					// 获取或创建会话
					sessionInterface, loaded := clientMap.Load(key)
					var session *udpSession

					if !loaded {
						// 解析目标地址
						targetAddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(fr.TargetAddr, fmt.Sprintf("%d", fr.TargetPort)))
						if err != nil {
							continue
						}

						// 创建到目标的连接
						newConn, err := net.DialUDP("udp", nil, targetAddr)
						if err != nil {
							continue
						}

						session = &udpSession{conn: newConn, lastActive: time.Now()}
						clientMap.Store(key, session)
						atomic.AddInt64(&fr.stats.TotalConn, 1)
						atomic.AddInt32(&fr.stats.CurrentConn, 1)

						// 启动响应处理
						go func(sa *net.UDPAddr, s *udpSession) {
							respBuf := make([]byte, 65535)
							for {
								s.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
								n, err := s.conn.Read(respBuf)
								if err != nil {
									return
								}
								s.lastActive = time.Now()
								atomic.AddInt64(&fr.stats.BytesSent, int64(n))
								conn.WriteToUDP(respBuf[:n], sa)
							}
						}(srcAddr, session)
					} else {
						session = sessionInterface.(*udpSession)
						session.lastActive = time.Now()
					}

					// 发送数据到目标
					session.conn.Write(buf[:n])
				}
			}
		}()
	}

	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// InitForwardRules 初始化转发规则
func InitForwardRules(ctx context.Context) {
	var rules []*entity.ForwardRule
	g.Model("forward_rule").Where("enabled", 1).Scan(&rules)
	for _, r := range rules {
		Start(ctx, r.Id)
	}
	g.Log().Infof(ctx, "[端口转发] 已启动 %d 条规则", len(rules))
}
