@echo off
chcp 65001 >nul
title FbAi - Production Start (Fast)

echo.
echo ============================================
echo   FbAi - production mode (fast loading)
echo   Backend :  http://localhost:9090
echo   Frontend:  http://localhost:3006
echo   Account :  admin / admin123
echo ============================================
echo.

echo [1/3] Starting backend (Go/Gin) ...
start "Backend :9090" cmd /c "cd /d %~dp0art-design-server && set GOPROXY=https://goproxy.cn,direct && go run main.go"

echo [2/3] Checking frontend build ...
cd /d %~dp0art-design-pro
if not exist dist\index.html (
    echo     dist not found, building... ^(first run takes about 40s^)
    call pnpm exec vite build
) else (
    echo     dist exists, skip build.
    echo     TIP: after changing frontend code, delete dist folder and rerun to rebuild.
)

echo [3/3] Starting frontend (production preview) ...
start "Frontend :3006" cmd /c "cd /d %~dp0art-design-pro && pnpm exec vite preview --port 3006 --strictPort"

echo.
echo Started. Open http://localhost:3006
echo NOTE: production mode has NO hot-reload. For development use start.bat instead.
pause >nul
