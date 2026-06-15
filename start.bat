@echo off
chcp 65001 >nul
title = FbAi — 一键启动前后端

echo.
echo ╔══════════════════════════════════════════════════╗
echo ║     FbAi — art-design 前后端一键启动              ║
echo ╚══════════════════════════════════════════════════╝
echo.
echo   [1/2] 启动后端服务 (Go/Gin)  →  http://localhost:9090
echo   [2/2] 启动前端服务 (Vue3)   →  http://localhost:3006
echo.

REM ============================================================
REM 1. 启动后端 (art-design-server)
REM ============================================================
echo   ▶ 正在启动后端服务...
start "art-design-server — Go/Gin Backend :9090" cmd /c ^
  "cd /d %~dp0art-design-server && ^
   set GOPROXY=https://goproxy.cn,direct && ^
   echo ===== art-design-server :9090 ===== && ^
   go run main.go && ^
   pause"

REM 等待后端先启动（避免前端请求失败），给 2 秒缓冲
timeout /t 2 /nobreak >nul

REM ============================================================
REM 2. 启动前端 (art-design-pro)
REM ============================================================
echo   ▶ 正在启动前端服务...
start "art-design-pro — Vue3 Frontend :3006" cmd /c ^
  "cd /d %~dp0art-design-pro && ^
   echo ===== art-design-pro :3006 ===== && ^
   pnpm dev && ^
   pause"

echo.
echo ╔══════════════════════════════════════════════════╗
echo ║  启动完成！                                       ║
echo ║  后端: http://localhost:9090/api/v1/ping          ║
echo ║  前端: http://localhost:3006                      ║
echo ║  默认账号: admin / admin123                       ║
echo ╚══════════════════════════════════════════════════╝
echo.
pause
