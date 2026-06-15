# 用户管理列表列简化 — 2026-06-16

## 需求

将用户管理列表改为：用户名、账号、角色、状态（启用状态）、创建时间，去除邮箱、性别、手机号。

## Modified

| File | Change | Reason |
|------|--------|--------|
| `src/views/system/user/index.vue` | 重构列表列定义；移除 userInfo 联合列（头像+用户名+邮箱）、userGender、userPhone；新增 nickName（用户名）、userName（账号）、roleName（角色）独立列；状态从四态（在线/离线/异常/注销）改为启停二态（启用/禁用）；移除 ACCOUNT_TABLE_DATA 和 ElImage 引用；简化 searchForm；更新 dataTransformer 和 handleDialogSubmit | 按需求精简列 |
| `src/views/system/user/modules/user-search.vue` | 移除手机号、邮箱、性别搜索字段；状态选项从四态改为启停二态(1/0)；搜索标签从"用户名"改为"账号" | 搜索栏对齐新列定义 |
| `src/views/system/user/modules/user-dialog.vue` | 移除性别下拉、邮箱输入框；简化 formData 和 initFormData | 弹窗表单对齐新字段 |

## Added

无新增文件。

## Why

客户要求简化用户管理页面，只保留核心字段：
- **用户名** = `nickName`（显示名称）
- **账号** = `userName`（登录账号）
- **角色** = `roleName`（角色标签）
- **状态** = 启用/禁用（二态开关）
- **创建时间** = `createTime`

非核心字段（邮箱、性别、手机号）已从列表、搜索栏、弹窗中完全移除。

## Verification

- `pnpm build` 通过（vue-tsc + vite build，0 错误）
- 后端 DSN: `postgres://fbai:***@127.0.0.1:5432/fbai?sslmode=disable`
- 构建产物: `dist/`，gzip 启用
