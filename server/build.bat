@echo off
chcp 65001 >nul
REM OmniWire 构建脚本 (Windows)

echo ==========================================
echo   OmniWire 构建脚本
echo ==========================================

REM 进入项目根目录
cd /d "%~dp0"

REM 1. 构建前端
echo.
echo [1/3] 构建前端...
cd ..\web
call npm install
call npm run build
cd ..\server

REM 2. 打包静态资源到 Go 代码
echo.
echo [2/3] 打包静态资源...
where gf >nul 2>nul
if %ERRORLEVEL% EQU 0 (
    gf pack resource/public internal/packed/packed.go -n packed
    echo   静态资源已打包到 internal/packed/packed.go
) else (
    echo   警告: 未安装 gf 命令行工具，跳过资源打包
    echo   安装方法: go install github.com/gogf/gf/cmd/gf/v2@latest
)

REM 3. 编译 Go 程序
echo.
echo [3/3] 编译 Go 程序...

if "%1"=="linux" (
    echo   目标平台: Linux ^(amd64^)
    set CGO_ENABLED=1
    set GOOS=linux
    set GOARCH=amd64
    go build -o omniwire main.go
    echo   输出文件: omniwire
) else (
    echo   目标平台: Windows ^(amd64^)
    set CGO_ENABLED=1
    go build -o omniwire.exe main.go
    echo   输出文件: omniwire.exe
)

echo.
echo ==========================================
echo   构建完成!
echo ==========================================
echo.
echo 运行方式:
echo   omniwire.exe
echo.

pause
