# 移除登录页选择框 + 后端 PostgreSQL 登录接口 — 2026-06-16

## 概述
1. 移除前端登录页的演示账号选择下拉框（ElSelect）
2. 后端从内存 mock 数据迁移到 PostgreSQL 数据库
3. 登录接口改为查询 PostgreSQL 验证用户名密码

---

## Modified（修改）

| 文件 | 变更 | 原因 |
|------|------|------|
| `art-design-pro/src/views/auth/login/index.vue` | 移除 ElSelect 角色选择框及所有相关代码 | 登录不需要演示账号选择 |
| `art-design-pro/src/api/auth.ts` | 登录 URL：`/api/auth/login` → `/api/v1/auth/login` | 匹配后端 Gin 路由前缀 |
| `art-design-pro/.env.development` | VITE_API_PROXY_URL 指向本地后端 | 对接真实后端 |
| `art-design-server/config/config.go` | 数据库 DSN 从环境变量读取，默认 PostgreSQL | 安全配置管理 |
| `art-design-server/main.go` | 启动时加载 .env + 连接数据库 | 数据库初始化 |
| `art-design-server/middleware/auth.go` | ValidateUser 从内存 mockUsers → PostgreSQL 查询 | 真实用户验证 |
| `art-design-server/services/user_service.go` | 从内存 map → PostgreSQL (pgx/v5) | 持久化存储 |
| `art-design-server/go.mod / go.sum` | 新增 pgx/v5 + godotenv 依赖 | PostgreSQL 驱动 |
| `.gitignore` | 新增 .codegraph/ | 忽略 CodeGraph 索引 |

## Added（新增）

| 文件 | 说明 |
|------|------|
| `art-design-server/db/postgres.go` | pgx/v5 连接池管理模块 |
| `art-design-server/.env.example` | 环境变量配置模板 |

## Database（数据库）

| 项目 | 值 |
|------|-----|
| 数据库 | PostgreSQL 17 |
| 数据库名 | `fbai` |
| 用户 | `fbai` / `***` |
| 表 | `users`, `roles`, `menus` |
| 种子数据 | admin 用户 (admin / admin123)，3 个角色 |

## 验证

```bash
# 编译
GOPROXY=https://goproxy.cn,direct go build -o server.exe ./main.go

# 启动
./server.exe

# 测试
curl http://localhost:9090/api/v1/ping
# → {"code":200,"msg":"pong"}

curl -X POST http://localhost:9090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"userName":"admin","password":"admin123"}'
# → {"code":200,"msg":"success","data":{"token":"...","userInfo":{...}}}
```
