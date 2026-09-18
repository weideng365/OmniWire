package forward

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestUDPSessionCloseIdempotent 测试 UDP 会话关闭的幂等性与连接计数的精确性
func TestUDPSessionCloseIdempotent(t *testing.T) {
	fr := &ForwardRule{
		stats: &ForwardStats{},
	}

	// 初始状态：增加 1 个连接
	atomic.AddInt32(&fr.stats.CurrentConn, 1)

	session := &udpSession{
		lastActive: time.Now(),
	}
	key := "127.0.0.1:12345"
	fr.udpSessions.Store(key, session)

	// 并发模拟 Ticker 超时清理与 Goroutine defer 退出同时发生
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			session.Close(fr, key)
		}()
	}
	wg.Wait()

	// 校验 CurrentConn 必须精确为 0，不能为负数
	current := atomic.LoadInt32(&fr.stats.CurrentConn)
	if current != 0 {
		t.Fatalf("预期 CurrentConn 为 0，实际为 %d", current)
	}

	// 校验 session 已被删除
	if _, exists := fr.udpSessions.Load(key); exists {
		t.Fatalf("预期 udpSessions 已删除 key，但仍然存在")
	}
}
