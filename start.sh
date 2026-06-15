#!/usr/bin/env bash
# FbAi — 一键启动前后端 (bash / git-bash / MSYS2)
set -e

ROOT="$(cd "$(dirname "$0")" && pwd)"

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║     FbAi — art-design 前后端一键启动              ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""
echo "  [1/2] 后端服务 (Go/Gin)  →  http://localhost:9090"
echo "  [2/2] 前端服务 (Vue3)   →  http://localhost:3006"
echo ""

cleanup() {
    echo ""
    echo "[STOP] 正在关闭所有服务..."
    [ -n "$BACKEND_PID" ] && kill "$BACKEND_PID" 2>/dev/null
    [ -n "$FRONTEND_PID" ] && kill "$FRONTEND_PID" 2>/dev/null
    echo "[STOP] 已关闭"
    exit 0
}
trap cleanup INT TERM

# -------- 1. 启动后端 --------
echo "  ▶ 启动后端 (art-design-server :9090) ..."
cd "$ROOT/art-design-server"
export GOPROXY=https://goproxy.cn,direct
go run main.go &
BACKEND_PID=$!

# 等待后端就绪
echo "  ⏳ 等待后端就绪..."
for i in $(seq 1 15); do
    if curl -s http://localhost:9090/api/v1/ping > /dev/null 2>&1; then
        echo "  ✓ 后端已就绪"
        break
    fi
    sleep 1
done

# -------- 2. 启动前端 --------
echo "  ▶ 启动前端 (art-design-pro :3006) ..."
cd "$ROOT/art-design-pro"
pnpm dev &
FRONTEND_PID=$!

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║  启动完成！                                       ║"
echo "║  后端: http://localhost:9090/api/v1/ping          ║"
echo "║  前端: http://localhost:3006                      ║"
echo "║  默认账号: admin / admin123                       ║"
echo "║  Ctrl+C 关闭所有服务                              ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""

# 等待任意子进程退出
wait
