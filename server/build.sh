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

# 2. 打包静态资源到 Go 代码
echo ""
echo "[2/3] 打包静态资源..."
if command -v gf &> /dev/null; then
    gf pack resource/public internal/packed/packed.go -n packed
    echo "  静态资源已打包到 internal/packed/packed.go"
else
    echo "  警告: 未安装 gf 命令行工具，跳过资源打包"
    echo "  安装方法: go install github.com/gogf/gf/cmd/gf/v2@latest"
fi

# 3. 编译 Go 程序
echo ""
echo "[3/3] 编译 Go 程序..."

# 检测目标平台
if [ "$1" == "linux" ]; then
    echo "  目标平台: Linux (amd64)"
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o omniwire main.go
    echo "  输出文件: omniwire"
elif [ "$1" == "windows" ]; then
    echo "  目标平台: Windows (amd64)"
    CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o omniwire.exe main.go
    echo "  输出文件: omniwire.exe"
else
    echo "  目标平台: 当前系统"
    go build -o omniwire main.go
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
