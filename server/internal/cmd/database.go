package cmd

import (
	"context"
	"os"
	"path/filepath"

	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/bcrypt"

	"omniwire/internal/service/wgserver"
)

// InitDatabase 初始化数据库表
func InitDatabase(ctx context.Context) error {
	// 确保 data 目录存在
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		g.Log().Warningf(ctx, "[数据库] 创建数据目录失败: %v", err)
	} else {
		absPath, _ := filepath.Abs(dataDir)
		g.Log().Debugf(ctx, "[数据库] 数据目录: %s", absPath)
	}

	g.Log().Info(ctx, "[数据库] 正在初始化数据库表...")

	// 创建用户表
	_, err := g.DB().Exec(ctx, `
		CREATE TABLE IF NOT EXISTS user (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username VARCHAR(50) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(20) DEFAULT 'admin',
			status INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// 创建 WireGuard 客户端表
	_, err = g.DB().Exec(ctx, `
		CREATE TABLE IF NOT EXISTS wireguard_peer (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(100) NOT NULL,
			public_key VARCHAR(255) NOT NULL UNIQUE,
			private_key VARCHAR(255) NOT NULL,
			preshared_key VARCHAR(255),
			allowed_ips VARCHAR(255) NOT NULL,
			endpoint VARCHAR(255),
			persistent_keepalive INTEGER DEFAULT 25,
			enabled INTEGER DEFAULT 1,
			upload_limit INTEGER DEFAULT 0,
			download_limit INTEGER DEFAULT 0,
			total_upload INTEGER DEFAULT 0,
			total_download INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// 创建端口转发规则表
	_, err = g.DB().Exec(ctx, `
		CREATE TABLE IF NOT EXISTS forward_rule (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(100) NOT NULL,
			protocol VARCHAR(10) NOT NULL DEFAULT 'tcp',
			listen_port INTEGER NOT NULL,
			target_addr VARCHAR(255) NOT NULL,
			target_port INTEGER NOT NULL,
			max_conn INTEGER DEFAULT 1000,
			upload_limit INTEGER DEFAULT 0,
			download_limit INTEGER DEFAULT 0,
			total_upload INTEGER DEFAULT 0,
			total_download INTEGER DEFAULT 0,
			enabled INTEGER DEFAULT 1,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// 创建操作日志表
	_, err = g.DB().Exec(ctx, `
		CREATE TABLE IF NOT EXISTS operation_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			action VARCHAR(100) NOT NULL,
			target VARCHAR(100),
			detail TEXT,
			ip VARCHAR(50),
			user_agent TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// 创建 WireGuard 配置表
	_, err = g.DB().Exec(ctx, `
		CREATE TABLE IF NOT EXISTS wireguard_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			interface_name VARCHAR(20) DEFAULT 'omniwire',
			listen_port INTEGER DEFAULT 51820,
			private_key VARCHAR(255),
			public_key VARCHAR(255),
			address VARCHAR(100) DEFAULT '10.66.66.1/24',
			dns VARCHAR(255) DEFAULT '223.5.5.5',
			mtu INTEGER DEFAULT 1420,
			endpoint_address VARCHAR(100),
			eth_device VARCHAR(20) DEFAULT '',
			persistent_keepalive INTEGER DEFAULT 25,
			client_allowed_ips VARCHAR(255) DEFAULT '0.0.0.0/0, ::/0',
			proxy_address VARCHAR(100) DEFAULT ':50122',
			log_level VARCHAR(20) DEFAULT 'error',
			auto_start INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// 创建 WireGuard 连接日志表
	_, err = g.DB().Exec(ctx, `
		CREATE TABLE IF NOT EXISTS wireguard_connection_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			peer_id INTEGER,
			peer_name VARCHAR(100),
			public_key VARCHAR(255),
			event VARCHAR(20),
			endpoint VARCHAR(255),
			transfer_rx INTEGER DEFAULT 0,
			transfer_tx INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// 为已存在的 wireguard_config 表添加 auto_start 字段（兼容旧数据库）
	hasAutoStart, _ := g.DB().GetValue(ctx, `SELECT COUNT(*) FROM pragma_table_info('wireguard_config') WHERE name='auto_start'`)
	if hasAutoStart.Int() == 0 {
		_, _ = g.DB().Exec(ctx, `ALTER TABLE wireguard_config ADD COLUMN auto_start INTEGER DEFAULT 0`)
	}

	// 插入默认管理员（如果不存在）
	count, _ := g.DB().Model("user").Where("username", "admin").Count()
	if count == 0 {
		hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		_, err = g.DB().Exec(ctx, `
			INSERT INTO user (username, password, role) VALUES ('admin', ?, 'admin')
		`, string(hashedPwd))
		if err != nil {
			g.Log().Warning(ctx, "[数据库] 创建默认用户失败:", err)
		}
	}

	// 插入默认 WireGuard 配置（如果不存在）
	configCount, _ := g.DB().Model("wireguard_config").Where("id", 1).Count()
	if configCount == 0 {
		privateKey, publicKey, keyErr := wgserver.GenerateKeyPair()
		if keyErr != nil {
			g.Log().Warning(ctx, "[数据库] 生成WireGuard密钥失败:", keyErr)
			privateKey, publicKey = "", ""
		}
		_, err = g.DB().Exec(ctx, `
			INSERT INTO wireguard_config (id, interface_name, listen_port, private_key, public_key, address, dns, mtu, eth_device, persistent_keepalive, client_allowed_ips, proxy_address, log_level)
			VALUES (1, 'omniwire', 51820, ?, ?, '10.66.66.1/24', '223.5.5.5', 1420, '', 25, '0.0.0.0/0, ::/0', ':50122', 'error')
		`, privateKey, publicKey)
		if err != nil {
			g.Log().Warning(ctx, "[数据库] 创建默认WireGuard配置失败:", err)
		}
	}

	g.Log().Info(ctx, "[数据库] 数据库表初始化完成")
	return nil
}
