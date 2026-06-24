# 删除广告账号权限 — 后端实现 — 2026-06-24

## Modified (Backend)
| File | Change | Reason |
|------|--------|--------|
| `art-design-server/models/fb.go` | 新增 `FbRemoveUserRequest` 结构体 | 删除权限请求模型 |
| `art-design-server/services/fb_service.go` | 新增 `fbDelete`、`RemoveAdAccountUser`、`removeUserFromAdAccount`、`listAdAccountUsers`、`getCurrentFbUserID` | FB Graph API DELETE 调用 + 5种删除模式逻辑 |
| `art-design-server/handlers/fb_handler.go` | 新增 `RemoveUser` handler | POST /api/v1/fb/ad-accounts/remove-user |
| `art-design-server/routes/router.go` | 注册 `/ad-accounts/remove-user` 路由 | 路由映射 |

## Modified (Rule)
| File | Change | Reason |
|------|--------|--------|
| `art-design-pro/AGENTS.md` | 新增 Rule 44: 每次改完都要跑测试 | 项目规范 |

## API
- **端点**: `POST /api/v1/fb/ad-accounts/remove-user`
- **认证**: Bearer token required
- **请求体**: `{ adAccountIds: string[], uids: string[], mode: string, deleteCurrent: boolean }`
- **模式**: deleteTheirs / deleteExceptTheirs / deleteExceptSelf / deleteSelf / deleteBM
- **响应**: 复用 `FbAssignUserResponse` 格式 (results/total/success/failed)

## 测试结果
- ✅ 有效请求 → 200 + FB API 调用（测试账户返回预期错误）
- ✅ 缺少必填字段 → 400 参数校验错误
- ✅ 无效模式 → 500 明确错误信息
- ✅ 无认证 → 401
