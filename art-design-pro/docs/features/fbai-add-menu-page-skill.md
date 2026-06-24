# Skill: fbai-add-menu-page — 2026-06-16

## 概述

`fbai-add-menu-page` 是一个标准化技能，封装了在 FbAi art-design-pro 项目中**新增菜单+页面**的完整 12 步工作流。每次需要添加侧边栏菜单和对应页面时，加载此技能即可自动按规范执行。

## 作用

- 统一菜单创建的**所有环节**：前端页面、路由、i18n、MenuProcessor 映射、数据库插入、角色权限更新、后端 fallback、构建验证
- **消除遗漏**：不再出现"页面写了但忘记注册路由"、"菜单插了但没更新 role menu_ids"、"i18n key 没加导致显示英文"等问题
- **遵守项目规范**：自动应用 ArtTable + useTable、ArtButtonTable、plain ripple 按钮、Element Plus 组件优先等规则
- **附带 mock 数据模板**：后端 API 未就绪时，可先用硬编码数据快速出页面

## 使用方法

### 基本调用

向 AI 助手说：

> 加载 fbai-add-menu-page，创建一个「XX管理」菜单

AI 会自动：

1. 查询当前 DB 菜单状态（确定新 ID、sort_order）
2. 询问/推断菜单结构（独立的单页面 vs 父子菜单）
3. 依次执行全部 12 个步骤
4. 每一步完成即验证，最后跑 `pnpm build` + API 测试

### 手动参考

也可以让 AI 只执行特定步骤，例如：

> 用 fbai-add-menu-page 的 Step 8，把菜单插入数据库

### 技能文件位置

```
D:\hermes\profiles\fbai\skills\fbai\fbai-add-menu-page\
├── SKILL.md                          # 主流程文档（12 步）
└── references\
    └── mock-data-pattern.md          # Mock 数据页面完整模板
```

### 依赖

使用前需先加载（AI 自动处理）：

| 依赖技能 | 用途 |
|---------|------|
| `fbai-context` | 项目架构、DB 表结构、API 路由 |
| `fbai-standards` | 编码规范、表格/页面模式、所有 Rule |

## 12 步流程速览

```
Step 1  → 查询 DB 现状（菜单 ID / sort_order / role menu_ids）
Step 2  → 设计菜单结构（父子菜单 或 独立页面）
Step 3  → 创建 Vue 页面 (views/xxx/index.vue)
Step 4  → 创建路由模块 (router/modules/xxx.ts)
Step 5  → 注册路由 (router/modules/index.ts)
Step 6  → 添加 i18n (zh.json + en.json)
Step 7  → MenuProcessor 映射 (I18N_MAP + ICON_MAP)
Step 8  → 插入 DB (menus 表)
Step 9  → 更新角色权限 (role.menu_ids)
Step 10 → 更新后端 fallback (menu_service.go)
Step 11 → 验证 (API 菜单树 + pnpm build)
Step 12 → 写开发文档 (docs/dev/)
```

## 覆盖的 Pitfalls（8 个）

| # | Pitfall | 技能中的对策 |
|---|---------|-------------|
| 1 | `menus_id_seq` 序列不同步 | Step 8 先 `SELECT setval(...)` |
| 2 | DB `name` ≠ 路由 `name` | Step 2 明确命名约定，Step 7 强调 key 必须匹配 |
| 3 | 父菜单 component 非 `/index/index` | Step 2 Pattern A 硬性规定 |
| 4 | 子路径以 `/` 开头 | Step 2 命名约定 + Step 4 示例 |
| 5 | role menu_ids 不含父 ID | Step 9 明确要求两者都加 |
| 6 | i18n key 非 menus.xxx 格式 | Step 6 提供正确格式 |
| 7 | icon 缺 `ri:` 前缀 | Step 2 命名约定 + Step 7 示例 |
| 8 | 构建报错（路径拼写错误） | Step 11 强制 `pnpm build` 验证 |

## 版本历史

| 版本 | 日期 | 变更 |
|------|------|------|
| 1.0.0 | 2026-06-16 | 初始版本，基于「广告账户管理」菜单创建经验提炼 |
