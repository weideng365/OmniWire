#!/bin/bash
# ==========================================================================
# OmniWire WireGuard 安装脚本
# ==========================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查 root 权限
if [[ $EUID -ne 0 ]]; then
   log_error "此脚本需要 root 权限运行"
   exit 1
fi

# 检测操作系统
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        VERSION=$VERSION_ID
    else
        log_error "无法检测操作系统"
        exit 1
    fi
}

# 安装 WireGuard
install_wireguard() {
    log_info "正在安装 WireGuard..."
    
    case $OS in
        ubuntu|debian)
            apt-get update
            apt-get install -y wireguard wireguard-tools
            ;;
        centos|rhel|rocky|almalinux)
            if [[ "$VERSION" == "7"* ]]; then
                yum install -y epel-release elrepo-release
                yum install -y yum-plugin-elrepo
                yum install -y kmod-wireguard wireguard-tools
            else
                dnf install -y epel-release
                dnf install -y wireguard-tools
            fi
            ;;
        fedora)
            dnf install -y wireguard-tools
            ;;
        arch|manjaro)
            pacman -S --noconfirm wireguard-tools
            ;;
        alpine)
            apk add wireguard-tools
            ;;
        *)
            log_error "不支持的操作系统: $OS"
            exit 1
            ;;
    esac
    
    log_info "WireGuard 安装完成"
}

# 加载内核模块
load_module() {
    log_info "加载 WireGuard 内核模块..."
    
    if ! lsmod | grep -q wireguard; then
        modprobe wireguard
    fi
    
    # 确保开机自动加载
    echo "wireguard" > /etc/modules-load.d/wireguard.conf
    
    log_info "内核模块加载完成"
}

# 启用 IP 转发
enable_ip_forward() {
    log_info "启用 IP 转发..."
    
    echo "net.ipv4.ip_forward = 1" > /etc/sysctl.d/99-wireguard.conf
    echo "net.ipv4.conf.all.src_valid_mark = 1" >> /etc/sysctl.d/99-wireguard.conf
    sysctl -p /etc/sysctl.d/99-wireguard.conf
    
    log_info "IP 转发已启用"
}

# 创建配置目录
create_config_dir() {
    log_info "创建配置目录..."
    
    mkdir -p /etc/wireguard
    chmod 700 /etc/wireguard
    
    log_info "配置目录已创建"
}

# 主函数
main() {
    log_info "开始安装 WireGuard..."
    
    detect_os
    log_info "检测到操作系统: $OS $VERSION"
    
    install_wireguard
    load_module
    enable_ip_forward
    create_config_dir
    
    # 验证安装
    if command -v wg &> /dev/null; then
        log_info "WireGuard 安装成功!"
        wg --version
    else
        log_error "WireGuard 安装失败"
        exit 1
    fi
}

main "$@"
