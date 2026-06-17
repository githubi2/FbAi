# 广告账户管理页面 — 2026-06-17

## 概述

实现广告账户管理页面，展示当前用户所有已授权FB账号下的广告账户详细信息。

## 需求

列表展示逻辑：展示当前用户授权的FB账号的所有广告户

列表展示列：
| 列名 | 字段 | 说明 |
|------|------|------|
| 广告账户ID | `accountId` | 显示广告账户ID |
| 所属账户 | `fbOwnerName` | 这个广告账户属于哪个FB账号（显示名称）|
| 所属BM | `businessName` | 这个账号属于哪个BM |
| 状态 | `statusLabel` | 活跃/已禁用/未结算/待审核/非活跃/已关闭 |
| 平台 | `platform` | Facebook |
| 总消耗金额 | `amountSpent` | 累计消耗 |
| 限额 | `spendCap` | 广告账户限额（0=无限制）|
| 下笔扣款额度 | `balance` | 账户余额 |
| 管理员 | `adminName` | 主管理员名称 |
| 隐藏管理员 | `hiddenAdmins` | 显示隐藏管理员个数 |
| 创建时间 | `createdTime` | FB创建时间 |

## 修改文件

### Backend

| 文件 | 变更 | 原因 |
|------|------|------|
| `models/fb.go` | 新增 `FbAdAccountDetail`、`FbAdAccountDetailListResponse` 结构体 | 承载广告账户详细信息（含消耗/限额/管理员等字段）|
| `services/fb_service.go` | 新增 `GetAdAccountsDetail()`、`parseAdAccountDetail()`、`getAccountStatusLabel()`、`toFloat64()` 方法 | 遍历所有已授权FB token，逐一调用FB Graph API获取详细广告账户信息 |
| `handlers/fb_handler.go` | 新增 `AdAccountsDetail()` handler | HTTP 请求处理 |
| `routes/router.go` | 新增路由 `GET /api/v1/fb/ad-accounts/detail` | API 端点注册 |

### Frontend

| 文件 | 变更 | 原因 |
|------|------|------|
| `src/api/facebook.ts` | 新增 `FbAdAccountDetail`、`FbAdAccountDetailListResponse` 类型 + `fetchFbAdAccountsDetail()` 函数 | 前端 API 层 |
| `src/views/ad-account/manage/index.vue` | **重写**：替换原有 mock 数据占位页为真实数据展示页 | 展示所有已授权FB账号下的广告账户 |
| `src/locales/langs/zh.json` | 新增 16 个 i18n key（列名 + 空状态 + 无限制） | 国际化支持 |
| `src/locales/langs/en.json` | 新增 16 个 i18n key | 国际化支持 |

## 关键技术决策

### 1. 后端: 遍历所有 token 逐一调用 FB API
- 从 `fb_tokens` 表查询所有 status=1 的 token
- 对每个 token 调用 `/me/adaccounts?fields=id,account_id,name,account_status,currency,amount_spent,spend_cap,balance,business{name},owner,users{name,role},created_time`
- 合并所有广告账户到一个列表返回

### 2. 前端: 使用 ArtTable + useTable 标准模式
- 遵循项目规范（Rule 7.6）
- 使用 `ArtTableHeader` + `ArtTable` + `useTable()` hook
- API 返回扁平数组，用 wrapper 转为分页格式

### 3. 广告账户状态映射
| FB Status | Label |
|-----------|-------|
| 1 | 活跃 (green) |
| 2 | 已禁用 (red) |
| 3 | 未结算 (warning) |
| 7 | 待审核 (info) |
| 9 | 非活跃 (info) |
| 100 | 待关闭 (warning) |
| 101 | 已关闭 (red) |

### 4. 管理员逻辑
- 调用 FB API `users{name,role}` 获取账户用户列表
- `role=1001` (Admin) 的第一个用户作为主管理员显示
- 其余用户计入 `hiddenAdmins` 计数

## API 端点

```
GET /api/v1/fb/ad-accounts/detail
Authorization: Bearer <token>

Response:
{
  "code": 200,
  "msg": "success",
  "data": {
    "accounts": [
      {
        "id": "act_123456789",
        "accountId": "123456789",
        "name": "My Ad Account",
        "fbOwnerName": "Fshsh Ushsh",
        "fbOwnerId": "122196014900875069",
        "businessName": "My Business Manager",
        "accountStatus": 1,
        "statusLabel": "活跃",
        "platform": "Facebook",
        "amountSpent": 12500.50,
        "currency": "USD",
        "spendCap": 50000.00,
        "balance": 1234.56,
        "adminName": "John Doe",
        "hiddenAdmins": 2,
        "createdTime": "2025-06-15T10:30:00+0000"
      }
    ],
    "total": 1
  }
}
```

## 验证

- ✅ Backend build 成功 (`go build -o server.exe`)
- ✅ Backend API 端点返回 200 + 正确 JSON 结构
- ✅ Frontend build 无新增 TypeScript 错误
- ✅ Menu tree 包含 AdAccountManage (id=14)
- ✅ i18n keys 完整 (zh.json + en.json)
- ⚠️ 需要已授权的 FB token 才能看到实际数据（当前无授权）

## 待办

- 广告账户数据实际展示需要用户先通过 FB 页面授权连接 Facebook
- 可考虑添加缓存机制避免每次请求都调用 FB API
- 可添加筛选/搜索功能
