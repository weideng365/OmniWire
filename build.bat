@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

REM OmniWire 一键构建脚本 (Windows)

echo ==========================================
echo   OmniWire 一键构建脚本
echo ==========================================

REM 进入项目根目录
cd /d "%~dp0"
set PROJECT_ROOT=%cd%

REM 解析参数
set TARGET=%1
if "%TARGET%"=="" set TARGET=current

REM 1. 构建前端
echo.
echo [1/3] 构建前端...
cd "%PROJECT_ROOT%\web"
call npm install --silent
call npm run build
echo   前端构建完成

REM 2. 同步前端产物到 embed 目录
echo.
echo [2/3] 同步静态资源到 embed 目录...
cd "%PROJECT_ROOT%\server"
if exist "internal\packed\public" rmdir /S /Q "internal\packed\public"
mkdir "internal\packed\public"
xcopy /E /I /Y "resource\public\*" "internal\packed\public\" >nul
echo   静态资源已同步到 internal\packed\public

REM 3. 编译 Go 程序（启用 embed build tag）
echo.
echo [3/3] 编译 Go 程序...
cd "%PROJECT_ROOT%\server"

REM 创建输出目录
if not exist "%PROJECT_ROOT%\dist" mkdir "%PROJECT_ROOT%\dist"

if "%TARGET%"=="linux" (
    echo   编译 Linux ^(amd64^)...
    set CGO_ENABLED=0
    set GOOS=linux
    set GOARCH=amd64
    go build -tags embed -ldflags="-s -w" -o "%PROJECT_ROOT%\dist\omniwire-linux-amd64" main.go
    echo   输出: dist\omniwire-linux-amd64
) else if "%TARGET%"=="windows" (
    echo   编译 Windows ^(amd64^)...
    set CGO_ENABLED=0
    set GOOS=windows
    set GOARCH=amd64
    go build -tags embed -ldflags="-s -w" -o "%PROJECT_ROOT%\dist\omniwire-windows-amd64.exe" main.go
    echo   输出: dist\omniwire-windows-amd64.exe
) else if "%TARGET%"=="all" (
    echo   编译 Linux ^(amd64^)...
    set CGO_ENABLED=0
    set GOOS=linux
    set GOARCH=amd64
    go build -tags embed -ldflags="-s -w" -o "%PROJECT_ROOT%\dist\omniwire-linux-amd64" main.go
    echo   输出: dist\omniwire-linux-amd64

    echo   编译 Windows ^(amd64^)...
    set CGO_ENABLED=0
    set GOOS=windows
    set GOARCH=amd64
    go build -tags embed -ldflags="-s -w" -o "%PROJECT_ROOT%\dist\omniwire-windows-amd64.exe" main.go
    echo   输出: dist\omniwire-windows-amd64.exe
) else (
    echo   编译当前平台...
    go build -tags embed -ldflags="-s -w" -o "%PROJECT_ROOT%\dist\omniwire.exe" main.go
    echo   输出: dist\omniwire.exe
)

echo.
echo ==========================================
echo   构建完成!
echo ==========================================
echo.
dir "%PROJECT_ROOT%\dist"
echo.
echo 运行: dist\omniwire.exe
echo.

endlocal
