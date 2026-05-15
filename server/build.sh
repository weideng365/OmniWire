#!/bin/bash
# OmniWire 构建脚本 (Linux/macOS)

set -e

echo "=========================================="
echo "  OmniWire 构建脚本"
echo "=========================================="

# 进入项目根目录
cd "$(dirname "$0")"

# 1. 构建前端
echo ""
echo "[1/3] 构建前端..."
cd ../web
npm install
npm run build
cd ../server

# 2. 同步前端产物到 embed 目录
echo ""
echo "[2/3] 同步静态资源到 embed 目录..."
rm -rf internal/packed/public
mkdir -p internal/packed/public
cp -r resource/public/. internal/packed/public/
echo "  静态资源已同步到 internal/packed/public"

# 3. 编译 Go 程序（启用 embed build tag）
echo ""
echo "[3/3] 编译 Go 程序..."

# 检测目标平台
if [ "$1" == "linux" ]; then
    echo "  目标平台: Linux (amd64)"
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags embed -o omniwire main.go
    echo "  输出文件: omniwire"
elif [ "$1" == "windows" ]; then
    echo "  目标平台: Windows (amd64)"
    CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -tags embed -o omniwire.exe main.go
    echo "  输出文件: omniwire.exe"
else
    echo "  目标平台: 当前系统"
    go build -tags embed -o omniwire main.go
    echo "  输出文件: omniwire"
fi

echo ""
echo "=========================================="
echo "  构建完成!"
echo "=========================================="
echo ""
echo "运行方式:"
echo "  ./omniwire"
echo ""
