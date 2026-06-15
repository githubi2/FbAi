import request from '@/utils/http'

/**
 * 获取租户列表
 */
export function fetchGetTenantList() {
  return request.get<Api.Tenant.TenantListItem[]>({
    url: '/api/v1/tenants'
  })
}

/**
 * 获取租户详情
 */
export function fetchGetTenantById(id: number) {
  return request.get<Api.Tenant.TenantListItem>({
    url: `/api/v1/tenants/${id}`
  })
}

/**
 * 创建租户
 */
export function fetchCreateTenant(params: Api.Tenant.CreateTenantParams) {
  return request.post<Api.Tenant.CreateTenantResponse>({
    url: '/api/v1/tenants',
    params
  })
}

/**
 * 更新租户
 */
export function fetchUpdateTenant(params: Api.Tenant.UpdateTenantParams) {
  return request.put<Api.Tenant.TenantListItem>({
    url: `/api/v1/tenants/${params.id}`,
    params
  })
}

/**
 * 删除租户
 */
export function fetchDeleteTenant(id: number) {
  return request.del({
    url: `/api/v1/tenants/${id}`
  })
}

/**
 * 切换租户上下文
 */
export function fetchSwitchTenant(tenantId: number) {
  return request.post<Api.Tenant.TenantContext>({
    url: '/api/v1/tenants/switch',
    params: { tenantId }
  })
}

/**
 * 获取当前租户上下文
 */
export function fetchCurrentTenant() {
  return request.get<Api.Tenant.TenantContext>({
    url: '/api/v1/tenants/current'
  })
}
