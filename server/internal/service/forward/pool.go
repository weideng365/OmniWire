// ==========================================================================
// OmniWire - TCP 连接池
// ==========================================================================

package forward

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrPoolClosed    = errors.New("连接池已关闭")
	ErrPoolExhausted = errors.New("连接池已耗尽")
	ErrConnInvalid   = errors.New("连接无效")
)

// PoolConfig 连接池配置
type PoolConfig struct {
	TargetAddr     string        // 目标地址
	TargetPort     int           // 目标端口
	InitialSize    int           // 初始连接数
	MaxSize        int           // 最大连接数
	MaxIdleSize    int           // 最大空闲连接数
	ConnectTimeout time.Duration // 连接超时
	IdleTimeout    time.Duration // 空闲超时
	MaxLifetime    time.Duration // 连接最大存活时间
	WaitTimeout    time.Duration // 等待获取连接超时
}

// DefaultPoolConfig 默认配置
func DefaultPoolConfig(targetAddr string, targetPort int) *PoolConfig {
	return &PoolConfig{
		TargetAddr:     targetAddr,
		TargetPort:     targetPort,
		InitialSize:    5,   // 初始 5 个连接
		MaxSize:        100, // 最大 100 个连接
		MaxIdleSize:    20,  // 最大空闲 20 个
		ConnectTimeout: 10 * time.Second,
		IdleTimeout:    5 * time.Minute,
		MaxLifetime:    30 * time.Minute,
		WaitTimeout:    5 * time.Second,
	}
}

// PooledConn 池化连接
type PooledConn struct {
	conn      net.Conn
	pool      *ConnPool
	createdAt time.Time
	usedAt    time.Time
}

// Read 实现 net.Conn 接口
func (pc *PooledConn) Read(b []byte) (int, error) {
	return pc.conn.Read(b)
}

// Write 实现 net.Conn 接口
func (pc *PooledConn) Write(b []byte) (int, error) {
	return pc.conn.Write(b)
}

// Close 归还连接到池
func (pc *PooledConn) Close() error {
	if pc.pool != nil {
		return pc.pool.Put(pc)
	}
	return pc.conn.Close()
}

// ForceClose 强制关闭连接
func (pc *PooledConn) ForceClose() error {
	return pc.conn.Close()
}

// LocalAddr 实现 net.Conn 接口
func (pc *PooledConn) LocalAddr() net.Addr {
	return pc.conn.LocalAddr()
}

// RemoteAddr 实现 net.Conn 接口
func (pc *PooledConn) RemoteAddr() net.Addr {
	return pc.conn.RemoteAddr()
}

// SetDeadline 实现 net.Conn 接口
func (pc *PooledConn) SetDeadline(t time.Time) error {
	return pc.conn.SetDeadline(t)
}

// SetReadDeadline 实现 net.Conn 接口
func (pc *PooledConn) SetReadDeadline(t time.Time) error {
	return pc.conn.SetReadDeadline(t)
}

// SetWriteDeadline 实现 net.Conn 接口
func (pc *PooledConn) SetWriteDeadline(t time.Time) error {
	return pc.conn.SetWriteDeadline(t)
}

// ConnPool TCP 连接池
type ConnPool struct {
	config    *PoolConfig
	mu        sync.Mutex
	conns     chan *PooledConn // 空闲连接通道
	numOpen   int32            // 当前打开的连接数
	closed    int32            // 是否已关闭
	waitQueue chan struct{}    // 等待队列信号
}

// NewConnPool 创建连接池
func NewConnPool(config *PoolConfig) (*ConnPool, error) {
	pool := &ConnPool{
		config:    config,
		conns:     make(chan *PooledConn, config.MaxIdleSize),
		waitQueue: make(chan struct{}, config.MaxSize),
	}

	// 预创建初始连接
	for i := 0; i < config.InitialSize; i++ {
		conn, err := pool.createConn()
		if err != nil {
			// 初始化失败不阻止创建，只记录日志
			continue
		}
		select {
		case pool.conns <- conn:
		default:
			conn.ForceClose()
		}
	}

	// 启动后台清理任务
	go pool.cleanupLoop()

	return pool, nil
}

// Get 获取连接
func (p *ConnPool) Get(ctx context.Context) (*PooledConn, error) {
	if atomic.LoadInt32(&p.closed) == 1 {
		return nil, ErrPoolClosed
	}

	// 先尝试从空闲池获取
	select {
	case conn := <-p.conns:
		if p.isConnValid(conn) {
			conn.usedAt = time.Now()
			return conn, nil
		}
		// 连接无效，关闭并创建新的
		conn.ForceClose()
		atomic.AddInt32(&p.numOpen, -1)
	default:
	}

	// 检查是否可以创建新连接
	if atomic.LoadInt32(&p.numOpen) < int32(p.config.MaxSize) {
		conn, err := p.createConn()
		if err != nil {
			return nil, err
		}
		return conn, nil
	}

	// 等待空闲连接
	timer := time.NewTimer(p.config.WaitTimeout)
	defer timer.Stop()

	select {
	case conn := <-p.conns:
		if p.isConnValid(conn) {
			conn.usedAt = time.Now()
			return conn, nil
		}
		conn.ForceClose()
		atomic.AddInt32(&p.numOpen, -1)
		// 重新尝试创建
		return p.createConn()
	case <-timer.C:
		return nil, ErrPoolExhausted
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Put 归还连接
func (p *ConnPool) Put(conn *PooledConn) error {
	if atomic.LoadInt32(&p.closed) == 1 {
		conn.ForceClose()
		atomic.AddInt32(&p.numOpen, -1)
		return nil
	}

	// 检查连接是否有效
	if !p.isConnValid(conn) {
		conn.ForceClose()
		atomic.AddInt32(&p.numOpen, -1)
		return nil
	}

	conn.usedAt = time.Now()

	// 尝试放回空闲池
	select {
	case p.conns <- conn:
		return nil
	default:
		// 空闲池已满，关闭连接
		conn.ForceClose()
		atomic.AddInt32(&p.numOpen, -1)
		return nil
	}
}

// Close 关闭连接池
func (p *ConnPool) Close() {
	if !atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		return
	}

	close(p.conns)
	for conn := range p.conns {
		conn.ForceClose()
	}
}

// Stats 获取连接池状态
func (p *ConnPool) Stats() map[string]interface{} {
	return map[string]interface{}{
		"numOpen":     atomic.LoadInt32(&p.numOpen),
		"numIdle":     len(p.conns),
		"maxSize":     p.config.MaxSize,
		"maxIdleSize": p.config.MaxIdleSize,
		"closed":      atomic.LoadInt32(&p.closed) == 1,
	}
}

// createConn 创建新连接
func (p *ConnPool) createConn() (*PooledConn, error) {
	target := net.JoinHostPort(p.config.TargetAddr, fmt.Sprintf("%d", p.config.TargetPort))
	conn, err := net.DialTimeout("tcp", target, p.config.ConnectTimeout)
	if err != nil {
		return nil, err
	}

	// TCP 优化
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	atomic.AddInt32(&p.numOpen, 1)

	return &PooledConn{
		conn:      conn,
		pool:      p,
		createdAt: time.Now(),
		usedAt:    time.Now(),
	}, nil
}

// isConnValid 检查连接是否有效
func (p *ConnPool) isConnValid(conn *PooledConn) bool {
	now := time.Now()

	// 检查最大存活时间
	if p.config.MaxLifetime > 0 && now.Sub(conn.createdAt) > p.config.MaxLifetime {
		return false
	}

	// 检查空闲超时
	if p.config.IdleTimeout > 0 && now.Sub(conn.usedAt) > p.config.IdleTimeout {
		return false
	}

	// 快速检查连接是否可用 (设置极短的读超时)
	conn.conn.SetReadDeadline(time.Now().Add(time.Microsecond))
	buf := make([]byte, 1)
	_, err := conn.conn.Read(buf)
	conn.conn.SetReadDeadline(time.Time{}) // 清除超时

	// 如果返回 EOF 则连接已关闭
	if err != nil && err.Error() == "EOF" {
		return false
	}

	return true
}

// cleanupLoop 后台清理过期连接
func (p *ConnPool) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		if atomic.LoadInt32(&p.closed) == 1 {
			return
		}

		<-ticker.C

		// 检查并清理过期连接
		toCheck := len(p.conns)
		for i := 0; i < toCheck; i++ {
			select {
			case conn := <-p.conns:
				if p.isConnValid(conn) {
					select {
					case p.conns <- conn:
					default:
						conn.ForceClose()
						atomic.AddInt32(&p.numOpen, -1)
					}
				} else {
					conn.ForceClose()
					atomic.AddInt32(&p.numOpen, -1)
				}
			default:
				return
			}
		}
	}
}

// ==================== 连接池管理器 ====================

// PoolManager 连接池管理器
type PoolManager struct {
	pools sync.Map // map[ruleId]*ConnPool
}

var poolManager = &PoolManager{}

// GetPool 获取或创建连接池
func GetPool(ruleId int, targetAddr string, targetPort int) (*ConnPool, error) {
	// 先尝试获取已存在的池
	if poolInterface, ok := poolManager.pools.Load(ruleId); ok {
		return poolInterface.(*ConnPool), nil
	}

	// 创建新的连接池
	config := DefaultPoolConfig(targetAddr, targetPort)
	pool, err := NewConnPool(config)
	if err != nil {
		return nil, err
	}

	// 存储（处理并发创建）
	actual, loaded := poolManager.pools.LoadOrStore(ruleId, pool)
	if loaded {
		// 已有其他 goroutine 创建了，关闭我们新建的
		pool.Close()
		return actual.(*ConnPool), nil
	}

	return pool, nil
}

// ClosePool 关闭连接池
func ClosePool(ruleId int) {
	if poolInterface, ok := poolManager.pools.LoadAndDelete(ruleId); ok {
		poolInterface.(*ConnPool).Close()
	}
}

// CloseAllPools 关闭所有连接池
func CloseAllPools() {
	poolManager.pools.Range(func(key, value interface{}) bool {
		value.(*ConnPool).Close()
		poolManager.pools.Delete(key)
		return true
	})
}
