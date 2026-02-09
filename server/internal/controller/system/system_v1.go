// ==========================================================================
// OmniWire - 系统管理控制器
// ==========================================================================

package system

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"omniwire/api/v1/system"
	"omniwire/internal/service/forward"
	"omniwire/internal/service/wgserver"
)

// ControllerV1 系统管理控制器
type ControllerV1 struct{}

// NewV1 创建系统管理控制器实例
func NewV1() *ControllerV1 {
	return &ControllerV1{}
}

// Info 系统信息
func (c *ControllerV1) Info(ctx context.Context, req *system.InfoReq) (res *system.InfoRes, err error) {
	res = &system.InfoRes{
		Name:    "OmniWire",
		Version: "1.0.0",
		Status:  "running",
	}
	return
}

// Dashboard 仪表盘数据
func (c *ControllerV1) Dashboard(ctx context.Context, req *system.DashboardReq) (res *system.DashboardRes, err error) {
	res = &system.DashboardRes{}

	// WireGuard 状态
	if wgserver.GetServer() != nil && wgserver.GetServer().IsRunning() {
		res.WireguardStatus = "running"
	} else {
		res.WireguardStatus = "stopped"
	}

	// 从数据库获取真实数据
	res.WireguardPeers, _ = g.DB().Model("wireguard_peer").Count()
	res.ForwardRules, _ = g.DB().Model("forward_rule").Count()
	res.ActiveConnections = forward.GetTotalActiveConnections()

	return
}

// Login 用户登录
func (c *ControllerV1) Login(ctx context.Context, req *system.LoginReq) (res *system.LoginRes, err error) {
	// 查询用户密码
	password, err := g.DB().Model("user").Where("username", req.Username).Fields("password").Value()
	if err != nil || password.IsEmpty() {
		return nil, gerror.New("用户名或密码错误")
	}

	// 验证密码
	if bcrypt.CompareHashAndPassword([]byte(password.String()), []byte(req.Password)) != nil {
		return nil, gerror.New("用户名或密码错误")
	}

	// 生成 JWT
	secret := g.Cfg().MustGet(ctx, "security.jwtSecret", "omniwire-secret-key-change-in-production").String()
	expireStr := g.Cfg().MustGet(ctx, "security.tokenExpire", "24h").String()
	expire, _ := time.ParseDuration(expireStr)
	if expire == 0 {
		expire = 24 * time.Hour
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.Username,
		"exp":      time.Now().Add(expire).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, gerror.New("生成令牌失败")
	}

	res = &system.LoginRes{Token: tokenStr}
	return
}

// ChangePassword 修改密码
func (c *ControllerV1) ChangePassword(ctx context.Context, req *system.ChangePasswordReq) (res *system.ChangePasswordRes, err error) {
	username := g.RequestFromCtx(ctx).GetCtxVar("username").String()
	if username == "" {
		return nil, gerror.New("未授权")
	}

	// 查询当前密码
	currentHash, err := g.DB().Model("user").Where("username", username).Fields("password").Value()
	if err != nil || currentHash.IsEmpty() {
		return nil, gerror.New("用户不存在")
	}

	// 验证旧密码
	if bcrypt.CompareHashAndPassword([]byte(currentHash.String()), []byte(req.OldPassword)) != nil {
		return nil, gerror.New("旧密码错误")
	}

	// 哈希新密码并更新
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, gerror.New("密码加密失败")
	}
	_, err = g.DB().Model("user").Where("username", username).Update(g.Map{"password": string(newHash)})
	if err != nil {
		return nil, gerror.New("密码更新失败")
	}

	res = &system.ChangePasswordRes{}
	return
}

// Health 健康检查
func (c *ControllerV1) Health(ctx context.Context, req *system.HealthReq) (res *system.HealthRes, err error) {
	res = &system.HealthRes{
		Status: "healthy",
	}
	return
}
