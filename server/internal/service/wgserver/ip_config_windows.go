//go:build windows

package wgserver

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

// configureIPWithLUID 使用 Windows LUID API 配置 IP 地址
func configureIPWithLUID(tunDevice tun.Device, address string) error {
	fmt.Printf("[DEBUG] configureIPWithLUID called with address: %s\n", address)

	// 解析 CIDR 地址
	prefix, err := netip.ParsePrefix(address)
	if err != nil {
		fmt.Printf("[DEBUG] ParsePrefix failed: %v\n", err)
		return fmt.Errorf("解析地址失败: %v", err)
	}
	fmt.Printf("[DEBUG] Parsed prefix: %s\n", prefix.String())

	// 如果 IP 等于网络地址（主机位全 0），自动修正为 .1
	addr := prefix.Addr()
	if addr.Is4() {
		ip4 := addr.As4()
		bits := prefix.Bits()
		shift := uint(32 - bits)
		ipNum := uint32(ip4[0])<<24 | uint32(ip4[1])<<16 | uint32(ip4[2])<<8 | uint32(ip4[3])
		mask := uint32(0xFFFFFFFF) << shift
		netNum := ipNum & mask
		if ipNum == netNum {
			hostNum := netNum | 1
			ip4[0] = byte(hostNum >> 24)
			ip4[1] = byte(hostNum >> 16)
			ip4[2] = byte(hostNum >> 8)
			ip4[3] = byte(hostNum)
			addr = netip.AddrFrom4(ip4)
			prefix = netip.PrefixFrom(addr, bits)
			fmt.Printf("[DEBUG] Auto-corrected to: %s\n", prefix.String())
			g.Log().Infof(context.Background(), "[WireGuard] 检测到网络地址，自动修正服务端 IP 为: %s", addr.String())
		}
	}

	// 获取 NativeTun 以访问 LUID
	fmt.Printf("[DEBUG] Getting NativeTun...\n")
	nativeTun, ok := tunDevice.(*tun.NativeTun)
	if !ok {
		fmt.Printf("[DEBUG] Failed to cast to NativeTun\n")
		return fmt.Errorf("无法获取 NativeTun 类型")
	}

	// 获取 LUID
	luid := winipcfg.LUID(nativeTun.LUID())
	fmt.Printf("[DEBUG] Got LUID: %d\n", luid)
	g.Log().Infof(context.Background(), "[WireGuard] 获取到 LUID: %d", luid)

	// 等待接口完全就绪
	fmt.Printf("[DEBUG] Waiting 2 seconds for interface...\n")
	g.Log().Infof(context.Background(), "[WireGuard] 等待接口初始化...")
	time.Sleep(2 * time.Second)

	// 重试机制
	var lastErr error
	for i := 0; i < 5; i++ {
		fmt.Printf("[DEBUG] Attempt %d/5: Setting IP %s\n", i+1, prefix.String())
		g.Log().Infof(context.Background(), "[WireGuard] 使用 LUID 配置 IP (%d/5): %s", i+1, prefix.String())

		// 先清除现有 IP 配置 (AF_INET = 2 for IPv4)
		flushErr := luid.FlushIPAddresses(winipcfg.AddressFamily(windows.AF_INET))
		if flushErr != nil {
			fmt.Printf("[DEBUG] FlushIPAddresses failed: %v\n", flushErr)
			g.Log().Warningf(context.Background(), "[WireGuard] FlushIPAddresses 失败: %v", flushErr)
		}
		time.Sleep(500 * time.Millisecond)

		// 设置新 IP
		err = luid.SetIPAddressesForFamily(winipcfg.AddressFamily(windows.AF_INET), []netip.Prefix{prefix})
		if err == nil {
			fmt.Printf("[DEBUG] IP configured successfully: %s\n", prefix.String())
			g.Log().Infof(context.Background(), "[WireGuard] IP 配置成功: %s", prefix.String())
			return nil
		}

		lastErr = err
		fmt.Printf("[DEBUG] SetIPAddressesForFamily failed: %v\n", err)
		g.Log().Warningf(context.Background(), "[WireGuard] SetIPAddressesForFamily 尝试 %d 失败: %v", i+1, err)
		time.Sleep(time.Duration(1+i) * time.Second)
	}

	fmt.Printf("[DEBUG] All attempts failed, last error: %v\n", lastErr)
	return fmt.Errorf("LUID IP 配置最终失败: %v", lastErr)
}
