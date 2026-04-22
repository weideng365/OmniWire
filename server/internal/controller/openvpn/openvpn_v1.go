package openvpn

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"omniwire/api/v1/openvpn"
	svc "omniwire/internal/service/openvpn"
)

type ControllerV1 struct{}

func NewV1() *ControllerV1 { return &ControllerV1{} }

func (c *ControllerV1) Status(ctx context.Context, req *openvpn.StatusReq) (res *openvpn.StatusRes, err error) {
	info, err := svc.Status(ctx)
	if err != nil {
		return nil, err
	}
	return &openvpn.StatusRes{
		Running:     info.Running,
		Protocol:    info.Protocol,
		Port:        info.Port,
		ClientCount: info.ClientCount,
		RxBytes:     info.RxBytes,
		TxBytes:     info.TxBytes,
	}, nil
}

func (c *ControllerV1) Start(ctx context.Context, req *openvpn.StartReq) (res *openvpn.StartRes, err error) {
	if err = svc.Start(ctx); err != nil {
		return nil, err
	}
	g.Log().Info(ctx, "OpenVPN 服务已启动")
	return &openvpn.StartRes{Success: true}, nil
}

func (c *ControllerV1) Stop(ctx context.Context, req *openvpn.StopReq) (res *openvpn.StopRes, err error) {
	if err = svc.Stop(ctx); err != nil {
		return nil, err
	}
	g.Log().Info(ctx, "OpenVPN 服务已停止")
	return &openvpn.StopRes{Success: true}, nil
}

func (c *ControllerV1) Restart(ctx context.Context, req *openvpn.RestartReq) (res *openvpn.RestartRes, err error) {
	if err = svc.Restart(ctx); err != nil {
		return nil, err
	}
	return &openvpn.RestartRes{Success: true}, nil
}

func (c *ControllerV1) Config(ctx context.Context, req *openvpn.ConfigReq) (res *openvpn.ConfigRes, err error) {
	cfg, err := svc.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &openvpn.ConfigRes{
		Protocol:    cfg.Protocol,
		Port:        cfg.Port,
		Endpoint:    cfg.Endpoint,
		Subnet:      cfg.Subnet,
		DNS:         cfg.DNS,
		AutoStart:   cfg.AutoStart,
		RouteMode:   cfg.RouteMode,
		SplitRoutes: cfg.SplitRoutes,
	}, nil
}

func (c *ControllerV1) UpdateConfig(ctx context.Context, req *openvpn.UpdateConfigReq) (res *openvpn.UpdateConfigRes, err error) {
	if err = svc.UpdateConfig(ctx, &svc.ConfigInfo{
		Protocol:    req.Protocol,
		Port:        req.Port,
		Endpoint:    req.Endpoint,
		Subnet:      req.Subnet,
		DNS:         req.DNS,
		AutoStart:   req.AutoStart,
		RouteMode:   req.RouteMode,
		SplitRoutes: req.SplitRoutes,
	}); err != nil {
		return nil, err
	}
	return &openvpn.UpdateConfigRes{Success: true}, nil
}

func (c *ControllerV1) UserList(ctx context.Context, req *openvpn.UserListReq) (res *openvpn.UserListRes, err error) {
	users, err := svc.GetUsers(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*openvpn.UserInfo, 0, len(users))
	for _, u := range users {
		list = append(list, &openvpn.UserInfo{
			Id:          u.Id,
			Username:    u.Username,
			Enabled:     u.Enabled == 1,
			Online:      u.Online == 1,
			IP:          u.IP,
			ConnectedAt: u.ConnectedAt,
			CreatedAt:   u.CreatedAt,
			RxBytes:     u.RxBytes,
			TxBytes:     u.TxBytes,
		})
	}
	return &openvpn.UserListRes{Users: list}, nil
}

func (c *ControllerV1) UserCreate(ctx context.Context, req *openvpn.UserCreateReq) (res *openvpn.UserCreateRes, err error) {
	u, err := svc.CreateUser(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return &openvpn.UserCreateRes{User: &openvpn.UserInfo{
		Id: u.Id, Username: u.Username, Enabled: true,
	}}, nil
}

func (c *ControllerV1) UserUpdate(ctx context.Context, req *openvpn.UserUpdateReq) (res *openvpn.UserUpdateRes, err error) {
	if err = svc.UpdateUser(ctx, req.Id, req.Password, req.Enabled); err != nil {
		return nil, err
	}
	return &openvpn.UserUpdateRes{Success: true}, nil
}

func (c *ControllerV1) UserDelete(ctx context.Context, req *openvpn.UserDeleteReq) (res *openvpn.UserDeleteRes, err error) {
	if err = svc.DeleteUser(ctx, req.Id); err != nil {
		return nil, err
	}
	return &openvpn.UserDeleteRes{Success: true}, nil
}

func (c *ControllerV1) UserConfig(ctx context.Context, req *openvpn.UserConfigReq) (res *openvpn.UserConfigRes, err error) {
	cfg, err := svc.GetUserConfig(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &openvpn.UserConfigRes{Config: cfg}, nil
}

func (c *ControllerV1) Auth(ctx context.Context, req *openvpn.AuthReq) (res *openvpn.AuthRes, err error) {
	return &openvpn.AuthRes{Success: svc.AuthUser(ctx, req.Username, req.Password)}, nil
}

func (c *ControllerV1) Connect(ctx context.Context, req *openvpn.ConnectReq) (res *openvpn.ConnectRes, err error) {
	svc.Connect(ctx, req.Username, req.IP)
	return &openvpn.ConnectRes{}, nil
}

func (c *ControllerV1) Disconnect(ctx context.Context, req *openvpn.DisconnectReq) (res *openvpn.DisconnectRes, err error) {
	svc.Disconnect(ctx, req.Username)
	return &openvpn.DisconnectRes{}, nil
}
