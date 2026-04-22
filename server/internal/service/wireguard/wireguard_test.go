package wireguard

import (
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

