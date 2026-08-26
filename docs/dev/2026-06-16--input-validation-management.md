# Input Validation — User/Role/Menu/Tenant Management — 2026-06-16

## Summary

Added comprehensive input validation to all 4 management modules (user, role, menu, tenant) on BOTH frontend AND backend, per Rule 18 (前后端双重校验).

## Modified

### Backend — Request Struct Binding Tags

| File | Change | Reason |
|------|--------|--------|
| `models/user.go` | `CreateUserRequest.Status`: added `binding:"oneof=0 1"` | Validate status enum |
| `models/user.go` | `CreateUserRequest.RoleID`: added `binding:"required,gt=0"` | Role must be specified |
| `models/user.go` | `UpdateUserRequest.Status`: added `binding:"oneof=0 1"` | Validate status enum |
| `models/user.go` | `UpdateUserRequest.RoleID`: added `binding:"required,gt=0"` | Role must be specified |
| `models/user.go` | `UpdateUserRequest.Password`: added `binding:"omitempty,min=6,max=32"` | Password length when provided |
| `models/role.go` | `CreateRoleRequest.Status`: added `binding:"oneof=0 1"` | Validate status enum |
| `models/role.go` | `UpdateRoleRequest.Status`: added `binding:"oneof=0 1"` | Validate status enum |
| `models/menu.go` | `CreateMenuRequest.MenuType`: added `binding:"omitempty,oneof=directory menu button"` | Validate menu type enum |
| `models/menu.go` | `CreateMenuRequest.Status`: added `binding:"oneof=0 1"` | Validate status enum |
| `models/menu.go` | `CreateMenuRequest.SortOrder`: added `binding:"gte=0"` | Sort must be non-negative |
| `models/menu.go` | `UpdateMenuRequest.MenuType`: added `binding:"omitempty,oneof=directory menu button"` | Validate menu type enum |
| `models/menu.go` | `UpdateMenuRequest.Status`: added `binding:"oneof=0 1"` | Validate status enum |
| `models/menu.go` | `UpdateMenuRequest.SortOrder`: added `binding:"gte=0"` | Sort must be non-negative |
| `models/tenant.go` | `CreateTenantRequest.ContactEmail`: added `binding:"omitempty,email,max=128"` | Validate email format when provided |
| `models/tenant.go` | `UpdateTenantRequest.ContactEmail`: added `binding:"omitempty,email,max=128"` | Validate email format when provided |
| `models/tenant.go` | `UpdateTenantRequest.Status`: added `binding:"oneof=0 1"` | Validate status enum |

### Frontend — Form Validation Rules

| File | Change | Reason |
|------|--------|--------|
| `views/system/user/modules/user-dialog.vue` | Added `userName` pattern: `/^[a-zA-Z\u4e00-\u9fa5][a-zA-Z0-9\u4e00-\u9fa5_-]*$/` | Prevent special chars at start, allow CJK |
| `views/tenant-system/user/modules/user-dialog.vue` | Same as above | Synchronize tenant user dialog |
| `views/system/role/modules/role-edit-dialog.vue` | Added `roleCode` pattern: `/^[a-zA-Z][a-zA-Z0-9_]*$/` | Ensure code starts with letter |
| `views/tenant-system/role/modules/role-edit-dialog.vue` | Same as above | Synchronize tenant role dialog |
| `views/system/menu/modules/menu-dialog.vue` | Enhanced all rules: name pattern, max lengths for path/label/component/icon, sort min, link URL validation | Comprehensive menu form validation |
| `views/system/tenant/modules/TenantForm.vue` | Enhanced: name min/max, code pattern (`/^[a-z][a-z0-9_]*$/`), contactPhone pattern, contactEmail email type, adminUserName pattern, adminNickName max | Comprehensive tenant form validation |

## Why

Per Rule 18 — all user input must be validated on BOTH frontend AND backend:
- **Frontend**: instant UX feedback, prevents obviously invalid submissions
- **Backend**: security baseline, Gin's `ShouldBindJSON` auto-rejects invalid requests with 400

## Verification

```bash
# Backend build
cd art-design-server && GOPROXY=https://goproxy.cn,direct go build -o server.exe ./main.go
# → exit 0

# Frontend type check
cd art-design-pro && npx vue-tsc --noEmit
# → exit 0

# Frontend lint
cd art-design-pro && pnpm lint
# → pre-existing errors only (TenantSwitcher.vue), no new errors
```
