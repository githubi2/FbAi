# 管理员/隐藏管理员标签零值处理 — 2026-06-18

## 需求
广告账户管理页中，"管理员"和"隐藏管理员"标签可点击弹窗。如果标签值为 0，不应弹窗，直接提示"没有管理员"或"没有隐藏管理员"。

## Modified
| File | Change | Reason |
|------|--------|--------|
| `src/views/ad-account/manage/index.vue` | `showAdminDetail()` 增加零值判断：值为 0 时 `ElMessage.info()` 提示并 `return`，不打开弹窗 | 业务需求 |
| `src/locales/langs/zh.json` | 修改 `noAdmin`: `"无管理员"` → `"没有管理员"`；新增 `noHiddenAdmin`: `"没有隐藏管理员"` | 提示文案 |
| `src/locales/langs/en.json` | 新增 `noHiddenAdmin`: `"No Hidden Admins"` | 英文对应 |

## Logic
```
管理员标签 (admin) → total = hiddenAdmins + (adminName ? 1 : 0)
  ├── total === 0 → ElMessage.info("没有管理员")
  └── total > 0  → 打开管理员详情弹窗

隐藏管理员标签 (hidden) → count = hiddenAdmins || 0
  ├── count === 0 → ElMessage.info("没有隐藏管理员")
  └── count > 0  → 打开隐藏管理员详情弹窗
```

## Why
避免空白弹窗，提升用户体验。当没有管理员/隐藏管理员时，直接告知用户而非弹出空内容弹窗。
