package wireguard

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestBuildPeerConfigPreservesCommaSeparatedAllowedIPs(t *testing.T) {
	serverConfig := &ConfigOutput{
		PublicKey:           "server-public-key",
		ListenPort:          51820,
		EndpointAddress:     "vpn.example.com",
		DNS:                 "223.5.5.5",
		MTU:                 1420,
		PersistentKeepalive: 25,
		ClientAllowedIPs:    "10.0.0.0/8,192.168.0.0/16, 172.16.0.0/12",
	}

	config := buildPeerConfig("peer-private-key", "10.66.66.2/32", serverConfig)

	expected := "AllowedIPs = 10.0.0.0/8,192.168.0.0/16, 172.16.0.0/12"
	if !strings.Contains(config, expected) {
		t.Fatalf("expected config to contain %q, got:\n%s", expected, config)
	}
}

func TestValidatePeerInput(t *testing.T) {
	// 1. 空公钥
	err := ValidatePeerInput("", "10.66.66.2/32")
	if err == nil {
		t.Fatalf("expected error for empty public key, got nil")
	}

	// 2. 长度不对的公钥 (非 32 字节)
	shortKey := base64.StdEncoding.EncodeToString([]byte("short-key"))
	err = ValidatePeerInput(shortKey, "10.66.66.2/32")
	if err == nil {
		t.Fatalf("expected error for invalid key length, got nil")
	}

	// 3. 非法 Base64
	err = ValidatePeerInput("invalid!!!base64", "10.66.66.2/32")
	if err == nil {
		t.Fatalf("expected error for non-base64 key, got nil")
	}

	// 4. 空 AllowedIPs
	validKey := base64.StdEncoding.EncodeToString(make([]byte, 32))
	err = ValidatePeerInput(validKey, "")
	if err == nil {
		t.Fatalf("expected error for empty allowed-ips, got nil")
	}

	// 5. 合法输入
	err = ValidatePeerInput(validKey, "10.66.66.2/32, 192.168.1.0/24")
	if err != nil {
		t.Fatalf("expected valid peer input to succeed, got: %v", err)
	}
}

