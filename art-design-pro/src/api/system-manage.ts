import request from '@/utils/http'
import type { AppRouteRecord } from '@/types/router'

// ==================== 用户管理 API ====================

/** 获取用户列表 */
export function fetchGetUserList(params: Api.SystemManage.UserSearchParams) {
  return request.get<Api.SystemManage.UserList>({
    url: '/api/user/list',
    params
  })
}

/** 获取用户详情 */
export function fetchGetUserById(id: number) {
  return request.get<Api.SystemManage.UserListItem>({
    url: `/api/v1/users/${id}`
  })
}

/** 创建用户 */
export function fetchCreateUser(data: Api.SystemManage.CreateUserParams) {
  return request.post<Api.SystemManage.UserListItem>({
    url: '/api/v1/users',
    params: data
  })
}

/** 更新用户 */
export function fetchUpdateUser(id: number, data: Api.SystemManage.UpdateUserParams) {
  return request.put<Api.SystemManage.UserListItem>({
    url: `/api/v1/users/${id}`,
    params: data
  })
}

/** 删除用户 */
export function fetchDeleteUser(id: number) {
  return request.del<void>({
    url: `/api/v1/users/${id}`
  })
}

// ==================== 角色管理 API ====================

/** 获取角色列表 */
export function fetchGetRoleList(params?: Api.SystemManage.RoleSearchParams) {
  return request.get<Api.SystemManage.RoleList>({
    url: '/api/role/list',
    params
  })
}

/** 获取角色详情 */
export function fetchGetRoleById(id: number) {
  return request.get<Api.SystemManage.RoleListItem>({
    url: `/api/v1/roles/${id}`
  })
}

/** 创建角色 */
export function fetchCreateRole(data: Api.SystemManage.CreateRoleParams) {
  return request.post<Api.SystemManage.RoleListItem>({
    url: '/api/v1/roles',
    params: data
  })
}

/** 更新角色 */
export function fetchUpdateRole(id: number, data: Api.SystemManage.UpdateRoleParams) {
  return request.put<Api.SystemManage.RoleListItem>({
    url: `/api/v1/roles/${id}`,
    params: data
  })
}

/** 删除角色 */
export function fetchDeleteRole(id: number) {
  return request.del<void>({
    url: `/api/v1/roles/${id}`
  })
}

/** 获取角色菜单权限 */
export function fetchGetRoleMenus(id: number) {
  return request.get<{ allMenus: any[]; roleMenus: number[] }>({
    url: `/api/v1/roles/${id}/menus`
  })
}

// ==================== 菜单管理 API ====================

/** 获取菜单树 */
export function fetchGetMenuTree() {
  return request.get<AppRouteRecord[]>({
    url: '/api/v3/system/menus/simple'
  })
}

/** 获取菜单列表（平铺） */
export function fetchGetMenuList() {
  return request.get<any[]>({
    url: '/api/v1/menus'
  })
}

/** 获取菜单详情 */
export function fetchGetMenuById(id: number) {
  return request.get<any>({
    url: `/api/v1/menus/${id}`
  })
}

/** 创建菜单 */
export function fetchCreateMenu(data: any) {
  return request.post<any>({
    url: '/api/v1/menus',
    params: data
  })
}

/** 更新菜单 */
export function fetchUpdateMenu(id: number, data: any) {
  return request.put<any>({
    url: `/api/v1/menus/${id}`,
    params: data
  })
}

/** 删除菜单 */
export function fetchDeleteMenu(id: number) {
  return request.del<void>({
    url: `/api/v1/menus/${id}`
  })
}
