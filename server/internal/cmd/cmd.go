package cmd

import (
	"context"
	"fmt"
	"runtime"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"

	"omniwire/internal/controller/forward"
	"omniwire/internal/controller/port"
	"omniwire/internal/controller/system"
	"omniwire/internal/controller/wireguard"
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

			// 设置静态文件目录（仅在目录存在时设置）
			staticPath := "resource/public"
			if gfile.Exists(staticPath) {
				s.SetServerRoot(staticPath)
				g.Log().Info(ctx, "[静态资源] 已加载: "+staticPath)
			}

			// 初始化数据库
			if err := InitDatabase(ctx); err != nil {
				g.Log().Errorf(ctx, "[数据库] 初始化失败: %v", err)
			}

			// API 路由组
			s.Group("/api/v1", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)

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
			})

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
	g.Log().Infof(ctx, "[WireGuard] 接口: %s", g.Cfg().MustGet(ctx, "wireguard.interface", "wg0").String())
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
	fmt.Println("    GET  /api/v1/system/info       - 获取系统信息")
	fmt.Println("    GET  /api/v1/system/dashboard  - 获取仪表盘数据")
	fmt.Println("    GET  /api/v1/system/health     - 健康检查")
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
}
