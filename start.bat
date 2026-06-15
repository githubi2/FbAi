@echo off
chcp 65001 >nul
title FbAi - Start

echo.
echo ============================================
echo   FbAi - art-design one-click start
echo   Backend :  http://localhost:9090
echo   Frontend:  http://localhost:3006
echo   Account :  admin / admin123
echo ============================================
echo.

echo [1/2] Starting backend (Go/Gin) ...
start "Backend :9090" cmd /c "cd /d %~dp0art-design-server && set GOPROXY=https://goproxy.cn,direct && go run main.go"

echo [2/2] Starting frontend (Vue3) ...
start "Frontend :3006" cmd /c "cd /d %~dp0art-design-pro && pnpm dev"

echo.
echo Both services started in separate windows.
echo Close this window or press any key to exit.
pause >nul
