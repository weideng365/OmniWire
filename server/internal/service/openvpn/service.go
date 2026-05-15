package openvpn

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/bcrypt"

	"crypto/ecdsa"
	"crypto/x509"
)

var (
	ovpnProcess *exec.Cmd
	ovpnMutex   sync.Mutex
	caCert      *x509.Certificate
	caKey       *ecdsa.PrivateKey
	serverCert  []byte
	serverKey   []byte
)

const (
	ovpnDir    = "./data/openvpn"
	configFile = "server.conf"
	authScript = "auth.sh"
)

type StatusInfo struct {
	Running     bool
	Protocol    string
	Port        int
	ClientCount int
	RxBytes     int64
	TxBytes     int64
}

type ConfigInfo struct {
	Protocol    string
	Port        int
	Endpoint    string
	Subnet      string
	DNS         string
	AutoStart   bool
	RouteMode   string
	SplitRoutes string
}

type UserInfo struct {
	Id          int
	Username    string
	Enabled     int
	Online      int
	IP          string
	ConnectedAt string
	CreatedAt   string
	RxBytes     int64
	TxBytes     int64
}

func Status(ctx context.Context) (*StatusInfo, error) {
	ovpnMutex.Lock()
	running := ovpnProcess != nil && ovpnProcess.Process != nil
	ovpnMutex.Unlock()

	config, _ := GetConfig(ctx)
	var clientCount int
	var rxBytes, txBytes int64
	if running {
		for _, s := range parseStatusLog() {
			clientCount++
			rxBytes += s.RxBytes
			txBytes += s.TxBytes
		}
	}
	return &StatusInfo{Running: running, Protocol: config.Protocol, Port: config.Port, ClientCount: clientCount, RxBytes: rxBytes, TxBytes: txBytes}, nil
}

func Start(ctx context.Context) error {
	ovpnMutex.Lock()
	defer ovpnMutex.Unlock()

	if ovpnProcess != nil {
		return fmt.Errorf("服务已在运行")
	}
	if _, err := exec.LookPath("openvpn"); err != nil {
		return fmt.Errorf("未找到 openvpn 命令，请在 Linux 环境部署或安装 OpenVPN")
	}
	if err := ensureConfig(ctx); err != nil {
		return err
	}
	absDir, err := filepath.Abs(ovpnDir)
	if err != nil {
		return fmt.Errorf("解析配置路径失败: %v", err)
	}
	cmd := exec.Command("openvpn", "--config", filepath.Join(absDir, configFile))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动失败: %v", err)
	}
	ovpnProcess = cmd
	g.Log().Info(ctx, "[OpenVPN] 服务已启动")

	// 后台监控进程退出，防止僵尸进程和状态不同步
	go func() {
		cmd.Wait()
		ovpnMutex.Lock()
		if ovpnProcess == cmd {
			ovpnProcess = nil
			g.DB().Model("openvpn_user").Data(g.Map{"online": 0}).Update()
			g.Log().Warning(ctx, "[OpenVPN] 进程已意外退出")
		}
		ovpnMutex.Unlock()
	}()

	return nil
}

func Stop(ctx context.Context) error {
	ovpnMutex.Lock()
	defer ovpnMutex.Unlock()

	if ovpnProcess == nil {
		return fmt.Errorf("服务未运行")
	}
	if err := ovpnProcess.Process.Kill(); err != nil {
		return err
	}
	// 不再调用 Wait()，由 Start() 中的后台 goroutine 负责收割进程
	ovpnProcess = nil
	g.DB().Model("openvpn_user").Data(g.Map{"online": 0}).Update()
	g.Log().Info(ctx, "[OpenVPN] 服务已停止")
	return nil
}

func Restart(ctx context.Context) error {
	if err := Stop(ctx); err != nil && !strings.Contains(err.Error(), "未运行") {
		return err
	}
	return Start(ctx)
}

func GetConfig(ctx context.Context) (*ConfigInfo, error) {
	one, err := g.DB().Model("openvpn_config").Where("id", 1).One()
	if err != nil || one.IsEmpty() {
		return &ConfigInfo{Protocol: "udp", Port: 1194, Subnet: "10.8.0.0/24", DNS: "223.5.5.5", RouteMode: "split"}, nil
	}
	routeMode := one["route_mode"].String()
	if routeMode == "" {
		routeMode = "split"
	}
	return &ConfigInfo{
		Protocol:    one["protocol"].String(),
		Port:        one["port"].Int(),
		Endpoint:    one["endpoint"].String(),
		Subnet:      one["subnet"].String(),
		DNS:         one["dns"].String(),
		AutoStart:   one["auto_start"].Int() == 1,
		RouteMode:   routeMode,
		SplitRoutes: one["split_routes"].String(),
	}, nil
}

func UpdateConfig(ctx context.Context, input *ConfigInfo) error {
	data := g.Map{
		"protocol": input.Protocol, "port": input.Port, "endpoint": input.Endpoint,
		"subnet": input.Subnet, "dns": input.DNS, "auto_start": boolToInt(input.AutoStart),
		"route_mode": input.RouteMode, "split_routes": input.SplitRoutes,
	}
	count, _ := g.DB().Model("openvpn_config").Where("id", 1).Count()
	if count == 0 {
		data["id"] = 1
		_, err := g.DB().Model("openvpn_config").Insert(data)
		return err
	}
	_, err := g.DB().Model("openvpn_config").Where("id", 1).Update(data)
	return err
}

func GetUsers(ctx context.Context) ([]*UserInfo, error) {
	list, err := g.DB().Model("openvpn_user").OrderAsc("id").All()
	if err != nil {
		return nil, err
	}
	online := parseStatusLog()
	users := make([]*UserInfo, 0, len(list))
	for _, row := range list {
		username := row["username"].String()
		ip := row["ip"].String()
		connectedAt := row["connected_at"].String()
		isOnline := 0
		var rxBytes, txBytes int64
		if stat, ok := online[username]; ok {
			isOnline = 1
			ip = stat.IP
			rxBytes = stat.RxBytes
			txBytes = stat.TxBytes
		}
		users = append(users, &UserInfo{
			Id: row["id"].Int(), Username: username,
			Enabled: row["enabled"].Int(), Online: isOnline,
			IP: ip, ConnectedAt: connectedAt,
			CreatedAt: row["created_at"].String(),
			RxBytes:   rxBytes, TxBytes: txBytes,
		})
	}
	return users, nil
}

func CreateUser(ctx context.Context, username, password string) (*UserInfo, error) {
	count, _ := g.DB().Model("openvpn_user").Where("username", username).Count()
	if count > 0 {
		return nil, fmt.Errorf("用户名已存在")
	}
	if err := ensurePKI(ctx); err != nil {
		return nil, err
	}
	bundle, err := generateCert(username, caCert, caKey, false)
	if err != nil {
		return nil, err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	config, _ := GetConfig(ctx)
	staticIP, _ := assignNextIP(ctx, config.Subnet)
	res, err := g.DB().Model("openvpn_user").Insert(g.Map{
		"username": username, "password": string(hashed),
		"cert": string(bundle.CertPEM), "key": string(bundle.KeyPEM),
		"enabled": 1, "static_ip": staticIP,
	})
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &UserInfo{Id: int(id), Username: username, Enabled: 1}, nil
}

func UpdateUser(ctx context.Context, id int, password string, enabled bool) error {
	data := g.Map{"enabled": boolToInt(enabled)}
	if password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		data["password"] = string(hashed)
	}
	_, err := g.DB().Model("openvpn_user").Where("id", id).Update(data)
	return err
}

func DeleteUser(ctx context.Context, id int) error {
	_, err := g.DB().Model("openvpn_user").Where("id", id).Delete()
	return err
}

func GetUserConfig(ctx context.Context, id int) (string, error) {
	row, err := g.DB().Model("openvpn_user").Where("id", id).One()
	if err != nil || row.IsEmpty() {
		return "", fmt.Errorf("用户不存在")
	}
	config, err := GetConfig(ctx)
	if err != nil {
		return "", err
	}
	caRow, err := g.DB().Model("openvpn_config").Where("id", 1).One()
	if err != nil || caRow.IsEmpty() {
		return "", fmt.Errorf("CA证书未初始化")
	}

	host := config.Endpoint
	if host == "" {
		host = "your-server-ip"
	}

	var routeBlock string
	if config.RouteMode == "split" {
		for _, r := range strings.Split(parseCIDRRoutes(config.SplitRoutes), "\n") {
			if r != "" {
				routeBlock += fmt.Sprintf("route %s\n", r)
			}
		}
	} else {
		routeBlock = "redirect-gateway def1 bypass-dhcp\n"
	}

	tmplStr := `client
dev tun
proto {{.Protocol}}
remote {{.Host}} {{.Port}}
resolv-retry infinite
nobind
persist-tun
auth-user-pass
auth SHA256
cipher AES-256-GCM
disable-dco
verb 3
{{.Routes}}<ca>
{{.CA}}</ca>
<cert>
{{.Cert}}</cert>
<key>
{{.Key}}</key>`

	var buf bytes.Buffer
	t := template.Must(template.New("ovpn").Parse(tmplStr))
	t.Execute(&buf, map[string]string{
		"Protocol": config.Protocol,
		"Host":     host,
		"Port":     fmt.Sprintf("%d", config.Port),
		"Routes":   routeBlock,
		"CA":       caRow["ca_cert"].String(),
		"Cert":     row["cert"].String(),
		"Key":      row["key"].String(),
	})
	return buf.String(), nil
}

func AuthUser(ctx context.Context, username, password string) bool {
	row, err := g.DB().Model("openvpn_user").Where("username", username).Where("enabled", 1).One()
	if err != nil || row.IsEmpty() {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(row["password"].String()), []byte(password)) == nil
}

func InitOpenVPN(ctx context.Context) {
	config, err := GetConfig(ctx)
	if err != nil || !config.AutoStart {
		return
	}
	if err := Start(ctx); err != nil {
		g.Log().Warningf(ctx, "[OpenVPN] 自动启动失败: %v", err)
	}
}

func ensurePKI(ctx context.Context) error {
	if caCert != nil {
		return nil
	}
	row, err := g.DB().Model("openvpn_config").Where("id", 1).One()
	if err == nil && !row.IsEmpty() && row["ca_cert"].String() != "" {
		caCert, caKey, err = parseCACert([]byte(row["ca_cert"].String()), []byte(row["ca_key"].String()))
		serverCert = []byte(row["server_cert"].String())
		serverKey = []byte(row["server_key"].String())
		return err
	}
	var caCertPEM, caKeyPEM []byte
	caCert, caKey, caCertPEM, caKeyPEM, err = generateCA()
	if err != nil {
		return err
	}
	bundle, err := generateCert("server", caCert, caKey, true)
	if err != nil {
		return err
	}
	serverCert = bundle.CertPEM
	serverKey = bundle.KeyPEM

	count, _ := g.DB().Model("openvpn_config").Where("id", 1).Count()
	if count == 0 {
		_, err = g.DB().Model("openvpn_config").Insert(g.Map{
			"id": 1, "protocol": "udp", "port": 1194, "subnet": "10.8.0.0/24", "dns": "223.5.5.5",
			"ca_cert": string(caCertPEM), "ca_key": string(caKeyPEM),
			"server_cert": string(serverCert), "server_key": string(serverKey),
		})
	} else {
		_, err = g.DB().Model("openvpn_config").Where("id", 1).Update(g.Map{
			"ca_cert": string(caCertPEM), "ca_key": string(caKeyPEM),
			"server_cert": string(serverCert), "server_key": string(serverKey),
		})
	}
	return err
}

func ensureConfig(ctx context.Context) error {
	if err := ensurePKI(ctx); err != nil {
		return err
	}
	absDir, err := filepath.Abs(ovpnDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(absDir, 0700); err != nil {
		return err
	}
	config, _ := GetConfig(ctx)
	row, _ := g.DB().Model("openvpn_config").Where("id", 1).One()

	os.WriteFile(filepath.Join(absDir, "ca.crt"), []byte(row["ca_cert"].String()), 0600)
	os.WriteFile(filepath.Join(absDir, "server.crt"), serverCert, 0600)
	os.WriteFile(filepath.Join(absDir, "server.key"), serverKey, 0600)

	// 回调 API 基地址：从主服务监听地址解析端口，避免 Docker(8080)/开发(8110) 不一致
	callbackBase := callbackBaseURL(ctx)

	// auth 脚本（via-file：$1 为临时文件，第一行用户名，第二行密码）
	// 使用 curl --data-urlencode 自动 URL 编码，避免用户名/密码包含 " \ $ 时
	// 拼接 JSON 出现注入或 JSON 格式破坏导致的拒绝服务。
	authScriptContent := "#!/bin/sh\n" +
		"USERNAME=$(sed -n '1p' \"$1\")\n" +
		"PASSWORD=$(sed -n '2p' \"$1\")\n" +
		"curl -sf -X POST " + callbackBase + "/api/v1/openvpn/auth" +
		" --data-urlencode \"username=$USERNAME\"" +
		" --data-urlencode \"password=$PASSWORD\" | grep -q '\"success\":true'\n"
	os.WriteFile(filepath.Join(absDir, authScript), []byte(authScriptContent), 0700)

	// client-connect 脚本
	connectScript := "#!/bin/sh\ncurl -sf -X POST " + callbackBase + "/api/v1/openvpn/connect" +
		" --data-urlencode \"username=$common_name\"" +
		" --data-urlencode \"ip=$ifconfig_pool_remote_ip\" || true\n"
	os.WriteFile(filepath.Join(absDir, "client-connect.sh"), []byte(connectScript), 0700)

	// client-disconnect 脚本
	disconnectScript := "#!/bin/sh\ncurl -sf -X POST " + callbackBase + "/api/v1/openvpn/disconnect" +
		" --data-urlencode \"username=$common_name\" || true\n"
	os.WriteFile(filepath.Join(absDir, "client-disconnect.sh"), []byte(disconnectScript), 0700)

	routeLines := "push \"redirect-gateway def1 bypass-dhcp\"\n"
	if config.RouteMode == "split" {
		routeLines = ""
		for _, r := range strings.Split(parseCIDRRoutes(config.SplitRoutes), "\n") {
			if r != "" {
				routeLines += fmt.Sprintf("push \"route %s\"\n", r)
			}
		}
	}

	// 生成 CCD 目录和每用户固定 IP 文件
	ccdDir := filepath.Join(absDir, "ccd")
	os.MkdirAll(ccdDir, 0700)
	_, ipNet, err := net.ParseCIDR(config.Subnet)
	if err != nil {
		return fmt.Errorf("解析子网失败: %v", err)
	}
	subnetMask := fmt.Sprintf("%d.%d.%d.%d", ipNet.Mask[0], ipNet.Mask[1], ipNet.Mask[2], ipNet.Mask[3])
	if users, err := g.DB().Model("openvpn_user").Fields("username,static_ip").All(); err == nil {
		for _, u := range users {
			if ip := u["static_ip"].String(); ip != "" {
				content := fmt.Sprintf("ifconfig-push %s %s\n", ip, subnetMask)
				os.WriteFile(filepath.Join(ccdDir, u["username"].String()), []byte(content), 0600)
			}
		}
	}

	network := ipNet.IP.String()
	serverConf := fmt.Sprintf("port %d\nproto %s\ndev tun\nca %s/ca.crt\ncert %s/server.crt\nkey %s/server.key\n"+
		"dh none\ntopology subnet\n"+
		"server %s %s\nclient-config-dir %s\n%spush \"dhcp-option DNS %s\"\n"+
		"keepalive 10 120\ncipher AES-256-GCM\nauth SHA256\npersist-key\npersist-tun\n"+
		"status %s/openvpn-status.log 5\nstatus-version 2\nverb 3\nscript-security 2\n"+
		"auth-user-pass-verify %s/%s via-file\nusername-as-common-name\n"+
		"client-connect %s/client-connect.sh\nclient-disconnect %s/client-disconnect.sh\n",
		config.Port, config.Protocol,
		absDir, absDir, absDir, network, subnetMask, ccdDir, routeLines, config.DNS,
		absDir, absDir, authScript,
		absDir, absDir,
	)
	return os.WriteFile(filepath.Join(absDir, configFile), []byte(serverConf), 0600)
}

// callbackBaseURL 根据主服务监听地址，构造 OpenVPN 脚本回调 OmniWire API 的基地址。
// 总是回环到 127.0.0.1，端口取自 server.address 配置（默认 8110）。
func callbackBaseURL(ctx context.Context) string {
	addr := g.Cfg().MustGet(ctx, "server.address", ":8110").String()
	port := "8110"
	if idx := strings.LastIndex(addr, ":"); idx >= 0 {
		if p, err := strconv.Atoi(addr[idx+1:]); err == nil && p > 0 {
			port = strconv.Itoa(p)
		}
	}
	return "http://127.0.0.1:" + port
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// assignNextIP 从 VPN 子网中分配下一个可用 IP（从 .2 开始，.1 为服务端）
func assignNextIP(ctx context.Context, subnet string) (string, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return "", err
	}
	rows, _ := g.DB().Model("openvpn_user").Fields("static_ip").All()
	used := map[string]bool{}
	for _, r := range rows {
		used[r["static_ip"].String()] = true
	}
	ip := ipNet.IP.To4()
	for i := 2; i < 254; i++ {
		candidate := fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], byte(i))
		if !used[candidate] {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("IP 地址池已耗尽")
}

func parseCIDRRoutes(routes string) string {
	var out string
	for _, cidr := range strings.Split(routes, ",") {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		mask := fmt.Sprintf("%d.%d.%d.%d", ipNet.Mask[0], ipNet.Mask[1], ipNet.Mask[2], ipNet.Mask[3])
		out += fmt.Sprintf("%s %s\n", ipNet.IP.String(), mask)
	}
	return out
}

type clientStat struct {
	IP      string
	RxBytes int64
	TxBytes int64
}

// parseStatusLog 解析 openvpn-status.log，返回 username->clientStat 映射
func parseStatusLog() map[string]*clientStat {
	absDir, _ := filepath.Abs(ovpnDir)
	data, err := os.ReadFile(filepath.Join(absDir, "openvpn-status.log"))
	if err != nil {
		return nil
	}
	result := map[string]*clientStat{}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "CLIENT_LIST,") {
			continue
		}
		parts := strings.Split(line, ",")
		// CLIENT_LIST,CommonName,RealAddr,VirtualAddr,VirtualIPv6,BytesReceived,BytesSent,...
		if len(parts) < 7 {
			continue
		}
		stat := &clientStat{IP: parts[3]}
		fmt.Sscanf(parts[5], "%d", &stat.RxBytes)
		fmt.Sscanf(parts[6], "%d", &stat.TxBytes)
		result[parts[1]] = stat
	}
	return result
}

func Connect(ctx context.Context, username, ip string) {
	g.DB().Model("openvpn_user").Where("username", username).Update(g.Map{
		"online": 1, "ip": ip, "connected_at": time.Now().Format("2006-01-02 15:04:05"),
	})
}

func Disconnect(ctx context.Context, username string) {
	g.DB().Model("openvpn_user").Where("username", username).Update(g.Map{"online": 0})
}
