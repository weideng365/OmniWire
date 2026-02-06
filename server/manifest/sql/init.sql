-- ==========================================================================
-- OmniWire - 数据库初始化脚本 (SQLite)
-- ==========================================================================

-- 用户表
CREATE TABLE IF NOT EXISTS user (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- WireGuard 客户端表
CREATE TABLE IF NOT EXISTS wireguard_peer (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    public_key VARCHAR(100) NOT NULL UNIQUE,
    private_key VARCHAR(100) NOT NULL,
    allowed_ips VARCHAR(100) NOT NULL,
    enabled INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 端口转发规则表
CREATE TABLE IF NOT EXISTS forward_rule (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    protocol VARCHAR(10) NOT NULL DEFAULT 'tcp',
    listen_port INTEGER NOT NULL,
    target_addr VARCHAR(255) NOT NULL,
    target_port INTEGER NOT NULL,
    enabled INTEGER DEFAULT 1,
    max_conn INTEGER DEFAULT 1000,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 操作日志表
CREATE TABLE IF NOT EXISTS operation_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    action VARCHAR(50) NOT NULL,
    target VARCHAR(100),
    detail TEXT,
    ip VARCHAR(50),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_wireguard_peer_enabled ON wireguard_peer(enabled);
CREATE INDEX IF NOT EXISTS idx_forward_rule_enabled ON forward_rule(enabled);
CREATE INDEX IF NOT EXISTS idx_operation_log_created_at ON operation_log(created_at);

-- 插入默认管理员用户 (密码: admin123)
INSERT OR IGNORE INTO user (username, password, role) 
VALUES ('admin', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z5EHsM8lE9lBOsl7iKTVKIUi', 'admin');
