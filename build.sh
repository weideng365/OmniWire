#!/bin/bash
# OmniWire 一键构建脚本 (Linux/macOS)

set -e

echo "=========================================="
echo "  OmniWire 一键构建脚本"
echo "=========================================="

# 进入项目根目录
cd "$(dirname "$0")"
PROJECT_ROOT=$(pwd)

# 解析参数
TARGET=${1:-"current"}  # current, linux, windows, all

# 1. 构建前端
echo ""
echo "[1/3] 构建前端..."
cd "$PROJECT_ROOT/web"
npm install --silent
npm run build
echo "  前端构建完成"

# 2. 打包静态资源
echo ""
echo "[2/3] 打包静态资源..."
cd "$PROJECT_ROOT/server"
echo "y" | gf pack resource/public,manifest/config internal/packed/packed.go -n packed -k
echo "  静态资源已打包"

# 3. 编译 Go 程序
echo ""
echo "[3/3] 编译 Go 程序..."
cd "$PROJECT_ROOT/server"

# 创建输出目录
mkdir -p "$PROJECT_ROOT/dist"

build_linux() {
    echo "  编译 Linux (amd64)..."
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "$PROJECT_ROOT/dist/omniwire-linux-amd64" main.go
    echo "  输出: dist/omniwire-linux-amd64"
}

build_windows() {
    echo "  编译 Windows (amd64)..."
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "$PROJECT_ROOT/dist/omniwire-windows-amd64.exe" main.go
    echo "  输出: dist/omniwire-windows-amd64.exe"
}

build_current() {
    echo "  编译当前平台..."
    go build -ldflags="-s -w" -o "$PROJECT_ROOT/dist/omniwire" main.go
    echo "  输出: dist/omniwire"
}

case "$TARGET" in
    linux)
        build_linux
        ;;
    windows)
        build_windows
        ;;
    all)
        build_linux
        build_windows
        ;;
    *)
        build_current
        ;;
esac

echo ""
echo "=========================================="
echo "  构建完成!"
echo "=========================================="
ls -lh "$PROJECT_ROOT/dist/"
echo ""
echo "运行: ./dist/omniwire"
echo ""
