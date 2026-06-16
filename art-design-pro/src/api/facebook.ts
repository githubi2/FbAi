import request from '@/utils/http'

// ==================== Facebook 授权 API ====================

/** Facebook 连接状态 */
export interface FbConnectionStatus {
  connected: boolean
  fbUserId: string
  fbUserName: string
  expiresAt: string
  selectedAdAccountId: string
  scopes: string[]
}

/** Facebook 广告账户 */
export interface FbAdAccount {
  id: string
  accountId: string
  name: string
  accountStatus: number
  currency: string
  businessName: string
}

/** Facebook Business Manager */
export interface FbBusinessManager {
  id: string
  name: string
}

/** 广告账户列表响应 */
export interface FbAdAccountListResponse {
  adAccounts: FbAdAccount[]
  businesses: FbBusinessManager[]
}

/** 获取 Facebook OAuth 授权链接 */
export function fetchFbAuthUrl() {
  return request.get<{ authUrl: string }>({
    url: '/api/v1/fb/auth-url'
  })
}

/** 获取 Facebook 连接状态 */
export function fetchFbConnectionStatus() {
  return request.get<FbConnectionStatus>({
    url: '/api/v1/fb/status'
  })
}

/** 获取广告账户列表 */
export function fetchFbAdAccounts() {
  return request.get<FbAdAccountListResponse>({
    url: '/api/v1/fb/ad-accounts',
    showErrorMessage: true
  })
}

/** 断开 Facebook 连接 */
export function fetchFbDisconnect() {
  return request.del<void>({
    url: '/api/v1/fb/disconnect'
  })
}
