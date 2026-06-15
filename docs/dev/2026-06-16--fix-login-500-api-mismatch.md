# 修复登录后 500 错误 — 前后端 API 对齐

**日期**: 2026-06-16
**问题**: 登录成功后跳转 500 错误页

## 根因

登录成功后，前端路由守卫 `beforeEach.ts` → `handleDynamicRoutes()` 调用 `fetchGetUserInfo()` 请求 `GET /api/user/info`，但 Go 后端没有该端点，返回 404 导致路由初始化失败 → 跳转 Exception500。

此外，后端返回的用户数据格式与前端 `Api.Auth.UserInfo` 接口不匹配（缺少 `roles` 数组、`buttons` 数组、字段名不对应）。

## 修改内容

### 1. `models/user.go` — 新增 UserInfoResponse 结构体
```go
type UserInfoResponse struct {
    Buttons  []string `json:"buttons"`
    Roles    []string `json:"roles"`
    UserID   uint     `json:"userId"`
    UserName string   `json:"userName"`
    Email    string   `json:"email"`
    Avatar   string   `json:"avatar,omitempty"`
}
```

### 2. `handlers/auth_handler.go` — 新增 GetUserInfoHandler
- 从 JWT 中间件注入的 context 获取 userID
- 查询用户表获取用户信息
- 通过 `role_name` 匹配内存 RoleService 获取 `role_code`（如 `"超级管理员"` → `"R_SUPER"`）
- 兜底：如果 role_name 本身就是 code，直接使用
- 从角色关联菜单中提取 `authMark` 作为按钮权限（Type=3）
- 返回标准 UserInfoResponse 格式

### 3. `routes/router.go` — 注册新路由
```go
r.GET("/api/user/info", middleware.AuthRequired(), handlers.DefaultAuthHandler.GetUserInfoHandler)
```
注：该路由不在 `/api/v1` 组下，直接匹配前端 `fetchGetUserInfo()` 的路径。

## 验证

```bash
# 登录获取 token
curl -s -X POST http://localhost:9090/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"userName":"admin","password":"admin123"}'

# 获取用户信息（用上面返回的 token）
curl -s http://localhost:9090/api/user/info \
  -H "Authorization: Bearer <token>"
```

返回：
```json
{
  "code": 200,
  "data": {
    "buttons": [],
    "roles": ["R_SUPER"],
    "userId": 1,
    "userName": "admin",
    "email": "admin@art-design.com"
  }
}
```

## 待办（其他 API 不匹配）

前端 `src/api/system-manage.ts` 中还有以下 API 与后端路径不匹配（仅在 backend 权限模式时使用，当前为 frontend 模式不受影响）：

| 前端调用 | 后端实际 | 
|---------|---------|
| `GET /api/user/list` | `GET /api/v1/users` |
| `GET /api/role/list` | `GET /api/v1/roles` |
| `GET /api/v3/system/menus/simple` | `GET /api/v1/menus` |
