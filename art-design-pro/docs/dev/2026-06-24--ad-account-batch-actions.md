# 广告账户管理 — 多选框 + 批量操作按钮组 — 2026-06-24

## Modified

| File | Change | Reason |
|------|--------|--------|
| `src/views/ad-account/manage/index.vue` | ① 表格 columnsFactory 新增 `{ type: 'selection', width: 55 }` 列；② 搜索卡片内新增 7 个批量操作按钮；③ 新增 `selectedRows` ref + `handleSelectionChange` + `handleBatchAction`；④ `ArtTable` 新增 `@selection-change` 事件绑定；⑤ 新增 `.batch-actions` SCSS 样式 | 用户需求：列表加多选框 + 搜索区域加批量操作按钮组 |
| `src/locales/langs/zh.json` | 新增 8 个 i18n key：`addAuth`, `deleteAuth`, `addToBM`, `setLimit`, `resetLimit`, `hideAdmin`, `accountPush`, `selectRowsFirst` | 按钮和提示文本国际化 |
| `src/locales/langs/en.json` | 新增 8 个 i18n key 对应英文翻译 | 英文国际化同步 |

## Added

| Button | i18n Key | 中文 | English |
|--------|----------|------|---------|
| 增加授权 | `menus.adAccount.addAuth` | 增加授权 | Add Auth |
| 删除授权 | `menus.adAccount.deleteAuth` | 删除授权 | Delete Auth |
| 添加到BM | `menus.adAccount.addToBM` | 添加到BM | Add to BM |
| 设置限额 | `menus.adAccount.setLimit` | 设置限额 | Set Limit |
| 重置限额 | `menus.adAccount.resetLimit` | 重置限额 | Reset Limit |
| 隐藏管理员 | `menus.adAccount.hideAdmin` | 隐藏管理员 | Hide Admin |
| 账号推送 | `menus.adAccount.accountPush` | 账号推送 | Push Account |

## Why

用户需要在广告账户管理列表中批量操作广告账户。先实现 UI（多选框 + 按钮组），功能后续迭代实现。

按钮点击时会先检查是否有选中行，无选中时弹出提示"请先选择广告账户"。

## Design Decisions

- 按钮放在搜索卡片内（`ElCard`），搜索表单下方，用分隔线区分
- 按钮样式遵循 Rule 7.7：`<ElButton v-ripple>` 纯文字按钮，无 type
- 多选列使用 `{ type: 'selection' }`，与系统用户管理页面模式一致
- 暂不实现按钮功能，点击仅打印 console.log + 选中校验
