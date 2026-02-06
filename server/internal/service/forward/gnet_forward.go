// ==========================================================================
// OmniWire - 基于 gnet 的高性能端口转发 (epoll + 零拷贝)
// ==========================================================================

package forward

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/panjf2000/gnet/v2"
)

// GnetForwarder 基于 gnet 的高性能转发器
type GnetForwarder struct {
	gnet.BuiltinEventEngine
	eng           gnet.Engine
	ruleId        int
	name          string
	listenPort    int
	targetAddr    string
	targetPort    int
	uploadLimit   int64
	downloadLimit int64
	stats         *ForwardStats
	connMap       sync.Map // fd -> *proxyConn
	running       int32
}

// proxyConn 代理连接
type proxyConn struct {
	clientConn gnet.Conn
	targetConn net.Conn
	buffer     []byte
}

// NewGnetForwarder 创建 gnet 转发器
func NewGnetForwarder(ruleId int, name string, listenPort int, targetAddr string, targetPort int) *GnetForwarder {
	return &GnetForwarder{
		ruleId:     ruleId,
		name:       name,
		listenPort: listenPort,
		targetAddr: targetAddr,
		targetPort: targetPort,
		stats:      &ForwardStats{StartTime: time.Now()},
	}
}

// OnBoot 引擎启动回调
func (f *GnetForwarder) OnBoot(eng gnet.Engine) gnet.Action {
	f.eng = eng
	atomic.StoreInt32(&f.running, 1)
	g.Log().Infof(context.Background(), "[gnet] 转发器 %s 已启动 (:%d -> %s:%d)", f.name, f.listenPort, f.targetAddr, f.targetPort)
	return gnet.None
}

// OnShutdown 引擎关闭回调
func (f *GnetForwarder) OnShutdown(eng gnet.Engine) {
	atomic.StoreInt32(&f.running, 0)
	g.Log().Infof(context.Background(), "[gnet] 转发器 %s 已关闭", f.name)
}

// OnOpen 新连接回调
func (f *GnetForwarder) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	atomic.AddInt64(&f.stats.TotalConn, 1)
	atomic.AddInt32(&f.stats.CurrentConn, 1)

	// 连接目标服务器
	targetConn, err := net.DialTimeout("tcp", net.JoinHostPort(f.targetAddr, fmt.Sprintf("%d", f.targetPort)), 10*time.Second)
	if err != nil {
		atomic.AddInt32(&f.stats.CurrentConn, -1)
		return nil, gnet.Close
	}

	// 设置 TCP 优化
	if tcpConn, ok := targetConn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	pc := &proxyConn{
		clientConn: c,
		targetConn: targetConn,
		buffer:     make([]byte, 64*1024),
	}
	f.connMap.Store(c.Fd(), pc)

	// 启动目标到客户端的数据传输
	go f.targetToClient(pc)

	return nil, gnet.None
}

// OnClose 连接关闭回调
func (f *GnetForwarder) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	atomic.AddInt32(&f.stats.CurrentConn, -1)

	if pcInterface, ok := f.connMap.LoadAndDelete(c.Fd()); ok {
		pc := pcInterface.(*proxyConn)
		if pc.targetConn != nil {
			pc.targetConn.Close()
		}
	}
	return gnet.None
}

// OnTraffic 数据到达回调 (零拷贝读取)
func (f *GnetForwarder) OnTraffic(c gnet.Conn) (action gnet.Action) {
	pcInterface, ok := f.connMap.Load(c.Fd())
	if !ok {
		return gnet.Close
	}
	pc := pcInterface.(*proxyConn)

	// 零拷贝读取数据
	buf, _ := c.Next(-1)
	if len(buf) == 0 {
		return gnet.None
	}

	// 统计流量
	atomic.AddInt64(&f.stats.BytesReceived, int64(len(buf)))

	// 速率限制
	if f.uploadLimit > 0 {
		// 简单的令牌桶限速
		time.Sleep(time.Duration(len(buf)) * time.Second / time.Duration(f.uploadLimit))
	}

	// 发送到目标服务器
	_, err := pc.targetConn.Write(buf)
	if err != nil {
		return gnet.Close
	}

	return gnet.None
}

// targetToClient 目标服务器到客户端的数据传输
func (f *GnetForwarder) targetToClient(pc *proxyConn) {
	defer func() {
		pc.clientConn.Close()
		pc.targetConn.Close()
	}()

	for atomic.LoadInt32(&f.running) == 1 {
		n, err := pc.targetConn.Read(pc.buffer)
		if err != nil {
			return
		}

		// 统计流量
		atomic.AddInt64(&f.stats.BytesSent, int64(n))

		// 速率限制
		if f.downloadLimit > 0 {
			time.Sleep(time.Duration(n) * time.Second / time.Duration(f.downloadLimit))
		}

		// 发送到客户端 (使用 gnet 的异步写入)
		err = pc.clientConn.AsyncWrite(pc.buffer[:n], nil)
		if err != nil {
			return
		}
	}
}

// Start 启动转发器
func (f *GnetForwarder) Start() error {
	addr := fmt.Sprintf("tcp://:%d", f.listenPort)

	// gnet 配置 - 使用 epoll/kqueue
	opts := []gnet.Option{
		gnet.WithMulticore(true),                      // 多核支持
		gnet.WithReusePort(true),                      // SO_REUSEPORT
		gnet.WithTCPKeepAlive(30 * time.Second),       // TCP KeepAlive
		gnet.WithTCPNoDelay(gnet.TCPNoDelay),          // TCP NoDelay
		gnet.WithLoadBalancing(gnet.LeastConnections), // 最少连接负载均衡
		gnet.WithLockOSThread(true),                   // 锁定 OS 线程
	}

	go func() {
		err := gnet.Run(f, addr, opts...)
		if err != nil {
			g.Log().Errorf(context.Background(), "[gnet] 启动失败: %v", err)
		}
	}()

	return nil
}

// Stop 停止转发器
func (f *GnetForwarder) Stop() error {
	atomic.StoreInt32(&f.running, 0)
	return f.eng.Stop(context.Background())
}

// GetStats 获取统计信息
func (f *GnetForwarder) GetStats() *ForwardStats {
	return f.stats
}

// IsRunning 检查是否运行中
func (f *GnetForwarder) IsRunning() bool {
	return atomic.LoadInt32(&f.running) == 1
}

// ==================== gnet 转发器管理 ====================

var (
	gnetForwarders = make(map[int]*GnetForwarder)
	gnetMutex      sync.RWMutex
)

// StartGnetForward 启动 gnet 转发
func StartGnetForward(ruleId int, name string, listenPort int, targetAddr string, targetPort int) error {
	gnetMutex.Lock()
	defer gnetMutex.Unlock()

	// 检查是否已存在
	if _, ok := gnetForwarders[ruleId]; ok {
		return fmt.Errorf("转发器已存在")
	}

	forwarder := NewGnetForwarder(ruleId, name, listenPort, targetAddr, targetPort)
	if err := forwarder.Start(); err != nil {
		return err
	}

	gnetForwarders[ruleId] = forwarder
	return nil
}

// StopGnetForward 停止 gnet 转发
func StopGnetForward(ruleId int) error {
	gnetMutex.Lock()
	defer gnetMutex.Unlock()

	if forwarder, ok := gnetForwarders[ruleId]; ok {
		delete(gnetForwarders, ruleId)
		return forwarder.Stop()
	}
	return nil
}

// GetGnetForwarderStats 获取统计
func GetGnetForwarderStats(ruleId int) *ForwardStats {
	gnetMutex.RLock()
	defer gnetMutex.RUnlock()

	if forwarder, ok := gnetForwarders[ruleId]; ok {
		return forwarder.GetStats()
	}
	return nil
}
