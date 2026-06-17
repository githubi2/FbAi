# 广告账户管理列表字段扩展 — 2026-06-17

## Modified

| File | Change | Reason |
|------|--------|--------|
| `art-design-server/models/fb.go` | `FbAdAccountDetail` struct: 新增 8 个字段 (`OwnerBusinessID`, `DailySpendLimit`, `CountryCode`, `IsPersonal`, `FundingSource`, `DisableReason`, `DisableReasonLabel`, `NextBillDate`) | 展示用户要求的完整广告账户列表字段 |
| `art-design-server/services/fb_service.go` | ① FB API fields 参数新增: `daily_spend_limit,funding_source_details,disable_reason,next_bill_date,business_country_code,is_personal`；② `parseAdAccountDetail` 新增对应字段解析逻辑；③ 新增 `getDisableReasonLabel()` 和 `deriveCountryCode()` 辅助函数 | 从 Facebook Graph API 拉取更多广告账户字段 |
| `art-design-pro/src/api/facebook.ts` | `FbAdAccountDetail` 接口: 新增 8 个字段 | TypeScript 类型同步后端模型 |
| `art-design-pro/src/views/ad-account/manage/index.vue` | 重写 `columnsFactory`: 从 16 列扩展到 28 列，按用户指定顺序排列 | 广告账户管理列表展示 28 个业务字段 |
| `art-design-pro/src/locales/langs/zh.json` | `menus.adAccount.columns` 新增 20 个 i18n key | 新列中文标签 |
| `art-design-pro/src/locales/langs/en.json` | `menus.adAccount.columns` 新增 20 个 i18n key | 新列英文标签 |

## Added

| File | Purpose |
|------|---------|
| `art-design-server/services/fb_service.go:deriveCountryCode()` | 从时区名称推断国家编码（30+ 城市映射） |
| `art-design-server/services/fb_service.go:getDisableReasonLabel()` | 广告账户禁用原因状态码→中文标签映射 |

## 新增字段详情

| 序号 | 列名 | 数据来源 | 说明 |
|------|------|----------|------|
| 1 | 序号 | `<ArtTable>` index | 保留 |
| 2 | 状态 | FB `account_status` | 保留（ElTag 彩色标签） |
| 3 | 账号 | FB `name` + `account_id` | Tooltip 显示完整 ID |
| 4 | 推送状态 | — | 暂无数据源，显示 "—" |
| 5 | 管理员 | FB `users{name}` [0] | 保留 |
| 6 | 隐藏管理员 | FB `users{name}` 计数-1 | 保留 |
| 7 | 账号类型 | FB `is_personal` | 1=个人, 0=企业 |
| 8 | 账单金额 | FB `balance` | 与余额同值展示 |
| 9 | 门槛 | — | 暂无数据源（FB `min_campaign_group_spend_cap` 已弃用） |
| 10 | 日限额 | FB `daily_spend_limit` | **新增**，0 显示"无限制" |
| 11 | 总花费 | FB `amount_spent` | 保留 |
| 12 | 花费限额 | FB `spend_cap` | 保留 |
| 13 | 已花费 | FB `amount_spent` | 与总花费同值 |
| 14 | 余额 | FB `balance` | 保留 |
| 15 | 备注 | — | 暂无本地存储，显示 "—" |
| 16 | 币种 | FB `currency` | 保留 |
| 17 | 账户类型 | `platform` (Facebook) | ElTag 显示平台名称 |
| 18 | 所有者角色 | FB `owner` → `ownerBusinessId` | **新增**，显示所有者 BM ID |
| 19 | 支付方法 | FB `funding_source_details.display_string` | **新增** |
| 20 | 账单期 | FB `next_bill_date` | **新增**，仅显示日期 |
| 21 | 锁定原因 | FB `disable_reason` → 中文标签 | **新增**，0 表示未锁定 |
| 22 | 创建日期 | FB `created_time` | 保留 |
| 23 | 时区 | FB `timezone_name` + `timezone_offset_hours_utc` | 合并显示 |
| 24 | 原始ID | FB `account_id` | 显示数字 ID |
| 25 | 创建自BM | FB `business{name}` | 与所属BM同值 |
| 26 | 所属BM | FB `business{name}` | 保留 |
| 27 | 国家编码 | FB `business_country_code` / 时区推断 | **新增** |
| 28 | 支付记录 | — | 暂无数据源，显示 "—" |

## Why

用户要求广告账户管理列表按完整业务字段展示 28 列。大部分字段可从 Facebook Graph API 直接获取（新增 6 个 API 字段请求）；部分字段（推送状态、备注、支付记录）暂无数据源，后续可扩展本地存储。
