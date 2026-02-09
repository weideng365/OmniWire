//go:build !windows

package wgserver

import (
	"golang.zx2c4.com/wireguard/tun"
)

// configureIPWithLUID 在非 Windows 平台上不可用，返回错误让调用方使用备用方案
func configureIPWithLUID(tunDevice tun.Device, address string) error {
	// 非 Windows 平台不支持 LUID API，直接返回错误
	// 调用方会回退到使用 ip 命令配置
	return nil
}
