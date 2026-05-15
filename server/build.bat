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

REM 2. 同步前端产物到 embed 目录
echo.
echo [2/3] 同步静态资源到 embed 目录...
if exist "internal\packed\public" rmdir /S /Q "internal\packed\public"
mkdir "internal\packed\public"
xcopy /E /I /Y "resource\public\*" "internal\packed\public\" >nul
echo   静态资源已同步到 internal\packed\public

REM 3. 编译 Go 程序（启用 embed build tag）
echo.
echo [3/3] 编译 Go 程序...

if "%1"=="linux" (
    echo   目标平台: Linux ^(amd64^)
    set CGO_ENABLED=1
    set GOOS=linux
    set GOARCH=amd64
    go build -tags embed -o omniwire main.go
    echo   输出文件: omniwire
) else (
    echo   目标平台: Windows ^(amd64^)
    set CGO_ENABLED=1
    go build -tags embed -o omniwire.exe main.go
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
