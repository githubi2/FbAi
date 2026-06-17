# 2026-06-17 — Facebook 企业版登录迁移（scope → config_id）

## 问题
FB OAuth 授权报错："It looks like this app isn't available"

## 根因
App `1049101794354495` (AIADS) 是企业版 (Business) 应用，只配置了市场营销 API 用例，没有 Facebook 登录产品。
企业版应用的 OAuth 使用 **`config_id`** 而非 `scope` 参数。

## 修改内容

### 1. 创建新应用和配置
- 新 App ID: `1627308648332623`
- 新 App Secret: `5fb7267b782afdd13f8b7fd8e91a9214`
- 在企业版 Facebook 登录 → 配置中创建配置
- config_id: `2223437138400500`
- 配置了所有业务权限（ads_read, ads_management, business_management 等）

### 2. `art-design-server/.env`
- 更新 `FB_APP_ID=1627308648332623`
- 更新 `FB_APP_SECRET`
- 新增 `FB_CONFIG_ID=2223437138400500`

### 3. `art-design-server/services/fb_service.go`
- `FbService` 结构体新增 `configID` 字段
- `init()` 中读取 `FB_CONFIG_ID` 环境变量
- `GetAuthURL()`: OAuth URL 从 `scope=public_profile` 改为 `config_id=<CONFIG_ID>&override_default_response_type=true`
- `ExchangeCodeForToken()`: 更新 Scopes 为业务权限

### 4. `art-design-server/.env.example`
- 新增 `FB_CONFIG_ID` 配置项说明

## OAuth URL 格式变化
```
旧: .../oauth/authorize?client_id=X&redirect_uri=Y&scope=public_profile&state=Z&response_type=code
新: .../oauth/authorize?client_id=X&redirect_uri=Y&config_id=C&response_type=code&override_default_response_type=true&state=Z
```

## 验证
- 后端重建成功
- `GET /api/v1/fb/auth-url` 返回正确格式的 OAuth URL（含 config_id）
- Facebook 返回 302 重定向到授权页（而非 "app isn't available" 错误）
