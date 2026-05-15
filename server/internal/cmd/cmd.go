package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/golang-jwt/jwt/v5"

	"omniwire/internal/controller/forward"
	"omniwire/internal/controller/openvpn"
	"omniwire/internal/controller/port"
	"omniwire/internal/controller/system"
	"omniwire/internal/controller/wireguard"
	"omniwire/internal/packed"
	forwardService "omniwire/internal/service/forward"
	openvpnService "omniwire/internal/service/openvpn"
	wireguardService "omniwire/internal/service/wireguard"
)

const (
	AppName    = "OmniWire"
	AppVersion = "1.0.0"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "OmniWire - WireGuard Server & Port Forward Management",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 打印启动横幅
			printBanner()

			s := g.Server()

			// 设置静态文件服务：优先 embed FS，回退本地目录
			staticPath := "resource/public"
			var embedFS fs.FS
			var embedErr error
			if packed.Enabled {
				embedFS, embedErr = fs.Sub(packed.FS, "public")
			} else {
				embedErr = fs.ErrNotExist
			}
			if embedErr == nil {
				g.Log().Info(ctx, "[静态资源] 已加载内嵌资源")
			} else if gfile.Exists(staticPath) {
				s.SetServerRoot(staticPath)
				g.Log().Info(ctx, "[静态资源] 已加载本地目录: "+staticPath)
			} else {
				g.Log().Warning(ctx, "[静态资源] 未找到静态资源，Web界面不可用")
			}

			// 初始化数据库
			if err := InitDatabase(ctx); err != nil {
				g.Log().Errorf(ctx, "[数据库] 初始化失败: %v", err)
			}

			// 初始化端口转发规则（自动启动已启用的规则）
			forwardService.InitForwardRules(ctx)

			// 初始化 WireGuard 服务（根据配置自动启动）
			wireguardService.InitWireGuard(ctx)

			// 初始化 OpenVPN 服务（根据配置自动启动）
			openvpnService.InitOpenVPN(ctx)

			// API 路由组
			s.Group("/api/v1", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)

				// JWT 鉴权中间件（白名单放行）
				group.Middleware(func(r *ghttp.Request) {
					path := r.URL.Path
					// 白名单：登录、健康检查无需鉴权
					if path == "/api/v1/system/login" || path == "/api/v1/system/health" {
						r.Middleware.Next()
						return
					}
					// OpenVPN 回调（auth/connect/disconnect）由 OpenVPN 守护进程在本机执行 up/down 脚本调用，
					// 仅允许 loopback 源地址访问，防止外网伪造身份认证。
					if path == "/api/v1/openvpn/auth" || path == "/api/v1/openvpn/connect" || path == "/api/v1/openvpn/disconnect" {
						if isLoopbackRemote(r) {
							r.Middleware.Next()
							return
						}
						r.Response.WriteStatus(403)
						r.Response.WriteJsonExit(g.Map{"code": 403, "message": "OpenVPN 回调仅允许本机访问"})
						return
					}

					auth := r.GetHeader("Authorization")
					if !strings.HasPrefix(auth, "Bearer ") {
						r.Response.WriteStatus(401)
						r.Response.WriteJsonExit(g.Map{"code": 401, "message": "未授权"})
						return
					}

					tokenStr := auth[7:]
					secret := g.Cfg().MustGet(r.Context(), "security.jwtSecret", "omniwire-secret-key-change-in-production").String()
					token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
						return []byte(secret), nil
					})
					if err != nil || !token.Valid {
						r.Response.WriteStatus(401)
						r.Response.WriteJsonExit(g.Map{"code": 401, "message": "令牌无效或已过期"})
						return
					}

					if claims, ok := token.Claims.(jwt.MapClaims); ok {
						r.SetCtxVar("username", claims["username"])
					}
					r.Middleware.Next()
				})

				// 系统管理接口
				group.Group("/system", func(group *ghttp.RouterGroup) {
					group.Bind(system.NewV1())
				})

				// WireGuard 管理接口
				group.Group("/wireguard", func(group *ghttp.RouterGroup) {
					group.Bind(wireguard.NewV1())
				})

				// 端口转发管理接口
				group.Group("/forward", func(group *ghttp.RouterGroup) {
					group.Bind(forward.NewV1())
				})

				// 端口管理接口
				group.Group("/port", func(group *ghttp.RouterGroup) {
					group.Bind(port.NewV1())
				})

				// OpenVPN 管理接口
				group.Group("/openvpn", func(group *ghttp.RouterGroup) {
					group.Bind(openvpn.NewV1())
				})
			})

			// SPA fallback：embed FS 服务静态文件，其余返回 index.html
			if embedErr == nil {
				s.BindHandler("/*", func(r *ghttp.Request) {
					if strings.HasPrefix(r.URL.Path, "/api") {
						r.Middleware.Next()
						return
					}
					path := strings.TrimPrefix(r.URL.Path, "/")
					if path == "" {
						path = "index.html"
					}
					data, err := fs.ReadFile(embedFS, path)
					if err != nil {
						// SPA fallback to index.html
						data, err = fs.ReadFile(embedFS, "index.html")
						if err != nil {
							r.Response.WriteStatus(404)
							return
						}
						path = "index.html"
					}
					r.Response.Header().Set("Content-Type", detectMimeType(path, data))
					r.Response.Write(data)
				})
			}

			// 打印配置信息
			printConfig(ctx)

			// 打印API路由信息
			printRoutes()

			// 启动服务
			g.Log().Info(ctx, "========================================")
			g.Log().Infof(ctx, "[服务启动] 监听地址: %s", g.Cfg().MustGet(ctx, "server.address", ":8110").String())
			g.Log().Info(ctx, "========================================")

			s.Run()
			return nil
		},
	}
)

// isLoopbackRemote 判断请求源是否来自本机回环地址。
// 用于限制 OpenVPN up/down 脚本回调，仅允许 OpenVPN 守护进程在本机调用。
func isLoopbackRemote(r *ghttp.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// detectMimeType 按文件扩展名推断 Content-Type；未知类型回退到内容嗅探。
func detectMimeType(path string, data []byte) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".js", ".mjs":
		return "application/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	case ".ico":
		return "image/x-icon"
	case ".map":
		return "application/json; charset=utf-8"
	}
	if t := mime.TypeByExtension(ext); t != "" {
		return t
	}
	return http.DetectContentType(data)
}

// printBanner 打印启动横幅
func printBanner() {
	banner := `
  ___                  _ __        __ _            
 / _ \  _ __ ___   __ _ (_)\ \      / /(_) _ __  ___  
| | | || '_ ' _ \ / /' || | \ \ /\ / / | || '__|/ _ \ 
| |_| || | | | | || | | || |  \ V  V /  | || |  |  __/ 
 \___/ |_| |_| |_||_| |_||_|   \_/\_/   |_||_|   \___| 

  WireGuard Server & Port Forward Management System
`
	fmt.Println(banner)
	fmt.Printf("  Version: %s | Go: %s | OS: %s/%s\n\n", AppVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// printConfig 打印配置信息
func printConfig(ctx context.Context) {
	g.Log().Info(ctx, "")
	g.Log().Info(ctx, "================ 系统配置 ================")
	g.Log().Infof(ctx, "[WireGuard] 接口: %s", g.Cfg().MustGet(ctx, "wireguard.interface", "omniwire").String())
	g.Log().Infof(ctx, "[WireGuard] 端口: %d", g.Cfg().MustGet(ctx, "wireguard.listenPort", 51820).Int())
	g.Log().Infof(ctx, "[WireGuard] 子网: %s", g.Cfg().MustGet(ctx, "wireguard.addressRange", "10.66.66.0/24").String())
	g.Log().Infof(ctx, "[数据库] 类型: %s", g.Cfg().MustGet(ctx, "database.default.type", "sqlite").String())
}

// printRoutes 打印API路由信息
func printRoutes() {
	fmt.Println("")
	fmt.Println("================ API 路由 ================")
	fmt.Println("")
	fmt.Println("  系统管理:")
	fmt.Println("    POST /api/v1/system/login          - 用户登录")
	fmt.Println("    POST /api/v1/system/change-password - 修改密码")
	fmt.Println("    GET  /api/v1/system/info            - 获取系统信息")
	fmt.Println("    GET  /api/v1/system/dashboard       - 获取仪表盘数据")
	fmt.Println("    GET  /api/v1/system/health          - 健康检查")
	fmt.Println("")
	fmt.Println("  WireGuard 管理:")
	fmt.Println("    GET  /api/v1/wireguard/status  - 获取服务状态")
	fmt.Println("    POST /api/v1/wireguard/start   - 启动服务")
	fmt.Println("    POST /api/v1/wireguard/stop    - 停止服务")
	fmt.Println("    POST /api/v1/wireguard/restart - 重启服务")
	fmt.Println("    GET  /api/v1/wireguard/config  - 获取配置")
	fmt.Println("    PUT  /api/v1/wireguard/config  - 更新配置")
	fmt.Println("    GET  /api/v1/wireguard/peers   - 获取客户端列表")
	fmt.Println("    POST /api/v1/wireguard/peers   - 创建客户端")
	fmt.Println("")
	fmt.Println("  端口转发:")
	fmt.Println("    GET  /api/v1/forward           - 获取转发规则列表")
	fmt.Println("    POST /api/v1/forward           - 创建转发规则")
	fmt.Println("    PUT  /api/v1/forward/:id       - 更新转发规则")
	fmt.Println("    DEL  /api/v1/forward/:id       - 删除转发规则")
	fmt.Println("")
	fmt.Println("  端口管理:")
	fmt.Println("    POST /api/v1/port/scan         - 扫描端口")
	fmt.Println("    GET  /api/v1/port/check/:port  - 检查端口占用")
	fmt.Println("    GET  /api/v1/port/listen       - 获取监听端口")
	fmt.Println("")
	fmt.Println("  OpenVPN 管理:")
	fmt.Println("    GET  /api/v1/openvpn/status    - 获取服务状态")
	fmt.Println("    POST /api/v1/openvpn/start     - 启动服务")
	fmt.Println("    POST /api/v1/openvpn/stop      - 停止服务")
	fmt.Println("    POST /api/v1/openvpn/restart   - 重启服务")
	fmt.Println("    GET  /api/v1/openvpn/config    - 获取配置")
	fmt.Println("    PUT  /api/v1/openvpn/config    - 更新配置")
	fmt.Println("    GET  /api/v1/openvpn/users     - 获取用户列表")
	fmt.Println("    POST /api/v1/openvpn/users     - 创建用户")
	fmt.Println("")
}
