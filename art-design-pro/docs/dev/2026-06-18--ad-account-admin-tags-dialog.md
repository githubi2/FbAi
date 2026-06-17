# 广告账户管理 — 管理员/隐藏管理员列改为标签+弹窗 — 2026-06-18

## Modified

| File | Change | Reason |
|------|--------|--------|
| `src/views/ad-account/manage/index.vue` | 管理员列：数字+tooltip → ElTag（点击打开弹窗） | 用户要求改用标签展示 |
| `src/views/ad-account/manage/index.vue` | 隐藏管理员列：纯数字 → ElTag（点击打开弹窗） | 用户要求改用标签展示 |
| `src/views/ad-account/manage/index.vue` | 新增管理员详情 ElDialog 组件 | 点击标签弹出详情弹窗 |

## Added

| Component / State | Purpose |
|-------------------|---------|
| `<ElDialog>` admin dialog | 管理员/隐藏管理员详情弹窗（内容待完善） |
| `adminDialogVisible` ref | 弹窗显示状态 |
| `adminDialogTitle` ref | 弹窗标题（区分"管理员"/"隐藏管理员"） |
| `curAdminAccount` ref | 当前点击行的数据 |
| `curAdminType` ref | 区分点击的是管理员还是隐藏管理员 |
| `showAdminDetail()` function | 点击标签时打开弹窗 |

## Details

### 管理员列
- 有 `adminName` → 显示为 `primary` 类型 ElTag，标签文字为管理员名称
- 无 `adminName` → 显示为 `info` 类型 ElTag，标签文字为 "0"
- 点击 → 弹窗标题：`{账户名} — 管理员`

### 隐藏管理员列
- `hiddenAdmins > 0` → 显示为 `warning` 类型 ElTag
- `hiddenAdmins === 0` → 显示为 `info` 类型 ElTag
- 点击 → 弹窗标题：`{账户名} — 隐藏管理员`

### 弹窗
- 内容区域当前为空白（TODO 占位）
- 底部"取消"按钮关闭弹窗
- 支持 X 按钮关闭
