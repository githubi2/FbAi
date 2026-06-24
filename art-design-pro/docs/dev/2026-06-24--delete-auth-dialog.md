# 删除广告账号权限弹窗 — 2026-06-24

## Modified
| File | Change | Reason |
|------|--------|--------|
| `src/views/ad-account/manage/index.vue` | 导入 DeleteAuthDialog，添加 deleteAuthDialogVisible，接入 deleteAuth 批量操作 | 接入删除授权弹窗 |
| `src/api/facebook.ts` | 新增 FbRemoveUserParams 接口 + fetchRemoveAdAccountUser 函数 | 删除权限 API |
| `src/locales/langs/zh.json` | 新增 menus.deleteAuth.* 17 个键 | 中文国际化 |
| `src/locales/langs/en.json` | 新增 menus.deleteAuth.* 17 个键 | 英文国际化 |

## Added
| File | Purpose |
|------|---------|
| `src/views/ad-account/manage/components/DeleteAuthDialog.vue` | 删除广告账号权限弹窗组件 |

## Why
广告账户管理页面的"删除授权"按钮之前只是占位，需要实现对应的弹窗功能。弹窗包含：
- 操作/结果标签页切换
- 5种删除模式下拉选择（删除它们的权限/除了它们删除所有/除了自己删除所有/删除自己/删除BM）
- 输入Facebook UID或主页地址的文本框
- 删除当前FB账号权限的选项
- 系统默认执行时间间隔选项
- 全宽确认按钮（带锁图标）

后端 API 端点: `POST /api/v1/fb/ad-accounts/remove-user`
