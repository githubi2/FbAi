# FB管理表格添加分页筛选 — 2026-06-18

## Modified

| File | Change | Reason |
|------|--------|--------|
| `src/views/ad-account/index.vue` | 添加搜索栏（关键词+状态筛选）+ 客户端分页 | FB账号列表需要分页和搜索功能 |
| `src/views/ad-account/manage/index.vue` | 添加搜索栏（关键词+状态+账号类型筛选）+ 客户端分页 | 广告账户管理需要分页和搜索功能 |
| `src/views/ad-account/bm/index.vue` | 添加搜索栏（关键词筛选）+ 客户端分页 | BM管理需要分页和搜索功能 |
| `src/locales/langs/zh.json` | 添加搜索相关 i18n 键 | searchKeyword, filterStatus, filterAccountType, bmNoData |
| `src/locales/langs/en.json` | 添加搜索相关 i18n 键 | 对应英文翻译 |

## Added

N/A — 无新增文件，均为现有文件修改

## Why

三个 FB 管理表格（FB账号列表、广告账户管理、BM管理）之前没有分页和搜索筛选功能：
- 所有数据一次性展示，数据量大时页面过长
- 无法按关键词搜索或按状态/类型筛选

用户要求与用户管理列表保持一致的交互体验。

## Implementation Details

### 客户端分页模式
由于 FB 数据通过 Graph API 一次性获取（非后端分页 API），采用**客户端分页**策略：
1. `fetchXxx(params)` 包装函数接收 `current`/`size` + 筛选参数
2. 调用真实 API 获取全量数据
3. 在客户端应用筛选（关键词、状态、类型）
4. 对筛选结果进行切片分页
5. 返回 `{ list, total, page, size }` 格式

### 搜索栏设计
每个页面的搜索栏使用 `ElForm :inline="true"` + `ElCard` 包裹：
- **FB账号列表**: 关键词搜索（FB用户名/备注/FB UserID）+ 状态筛选（正常/已过期/异常）
- **广告账户管理**: 关键词搜索（账户名/ID/BM名）+ 状态筛选（启用/禁用）+ 账号类型（企业/个人）
- **BM管理**: 关键词搜索（BM名称/ID/所属账户）（当前使用模拟数据，待接入真实API）

### Pagination 集成
- 使用 `useTable` hook 自带的 `pagination`、`handleSizeChange`、`handleCurrentChange`
- `ArtTable` 绑定 `:pagination="pagination"` + 相应事件
- `replaceSearchParams` 用于触发搜索（重置到第1页）
