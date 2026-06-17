# 管理员弹窗 — 零值检查 + 步骤式弹窗 UI — 2026-06-18

## 需求
1. 管理员/隐藏管理员标签值为 0 时，不弹窗，直接提示
2. 弹窗改为步骤式 UI：步骤1 选择要删除的管理员 / 步骤2 执行间隔 checkbox / 确认按钮

## Modified

### Frontend
| File | Change | Reason |
|------|--------|--------|
| `src/views/ad-account/manage/index.vue` | `showAdminDetail()` 零值判断；弹窗 UI 改为步骤式（ElCheckboxGroup + ElCheckbox）；新增 `handleAdminDelete` | 业务需求 |
| `src/api/facebook.ts` | `FbAdAccountDetail` 新增 `otherAdminNames: string[]` | 接收后端返回的其他管理员列表 |
| `src/locales/langs/zh.json` | 修改 `noAdmin` → "没有管理员"；新增 `noHiddenAdmin`, `adminDialogStep1`, `adminDialogStep2`, `adminDialogConfirm`, `adminDialogNoOther`, `adminDeleteSuccess` | 中文文案 |
| `src/locales/langs/en.json` | 新增 `noHiddenAdmin`, `adminDialogStep1`-`adminDeleteSuccess` 英文对应 | 英文文案 |

### Backend
| File | Change | Reason |
|------|--------|--------|
| `models/fb.go` | `FbAdAccountDetail` 新增 `OtherAdminNames []string` | 返回其他管理员名称列表 |
| `services/fb_service.go` | `parseAdAccountDetail()` 收集所有非首位管理员名称 | 前端弹窗需要显示可删除的管理员列表 |

## Logic

```
showAdminDetail(row, type):
  ├── type='hidden' → count = hiddenAdmins || 0
  │   ├── count===0 → ElMessage.info("没有隐藏管理员"), return
  │   └── count>0  → 打开弹窗
  └── type='admin' → total = hiddenAdmins + (adminName ? 1 : 0)
      ├── total===0 → ElMessage.info("没有管理员"), return
      └── total>0  → 打开弹窗

弹窗内容:
  ├── otherAdminNames.length===0 → ElEmpty("没有其他管理员可删除")
  └── otherAdminNames.length>0
      ├── 步骤1: ElCheckboxGroup — 选择要删除的管理员
      ├── 步骤2: ElCheckbox — "系统默认执行时间间隔 4s" (默认勾选)
      └── 确认按钮: type="primary", :disabled="selectedAdmins.length===0"
```

## Why
- 避免空白弹窗，零值时直接提示用户
- 步骤式 UI 清晰引导用户操作流程
- 后端新增 `otherAdminNames` 字段复用已有 FB API 数据（`users{name}`），无需额外 API 调用
