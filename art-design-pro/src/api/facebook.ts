import request from '@/utils/http'

// ==================== Facebook 授权 API（多账号改造） ====================

/** Facebook 连接状态（保留用于轮询检测） */
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

// ==================== 多账号改造 — 新增类型 ====================

/** FB 账号列表项 */
export interface FbAccount {
  id: number
  fbUserId: string
  fbUserName: string
  label: string
  scopes: string[]
  expiresAt: string
  createdAt: string
  daysUntilExpiry: number
  hasAdPerm: boolean
  accountStatus: string // "正常" | "已过期" | "异常"
  bmCount: number
  personalAdCount: number
  bmAdCount: number
  dataError: string // 数据拉取失败时的错误信息
}

/** FB 账号列表响应 */
export interface FbAccountListResponse {
  accounts: FbAccount[]
  total: number
}

// ==================== API 函数 ====================

/** 获取 Facebook OAuth 授权链接 */
export function fetchFbAuthUrl() {
  return request.get<{ authUrl: string; shortUrl: string }>({
    url: '/api/v1/fb/auth-url'
  })
}

/** 获取 Facebook 连接状态（保留向后兼容） */
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

/** 断开 Facebook 连接（保留向后兼容 — 不传 id 断开全部） */
export function fetchFbDisconnect() {
  return request.del<void>({
    url: '/api/v1/fb/disconnect'
  })
}

// ==================== 多账号改造 — 新增 API ====================

/** 获取已授权的 FB 账号列表 */
export function fetchFbAccountList() {
  return request.get<FbAccountListResponse>({
    url: '/api/v1/fb/accounts',
    showErrorMessage: true
  })
}

/** 断开指定 FB 账号 */
export function fetchFbDisconnectAccount(id: number) {
  return request.del<void>({
    url: `/api/v1/fb/accounts/${id}`
  })
}

/** 更新 FB 账号备注 */
export function fetchFbUpdateLabel(id: number, label: string) {
  return request.put<void>({
    url: `/api/v1/fb/accounts/${id}/label`,
    params: { label }
  })
}

/** 刷新 FB 账号的 BM 和广告账户统计 */
export function fetchFbRefreshStats(id: number) {
  return request.post<void>({
    url: `/api/v1/fb/accounts/${id}/refresh`
  })
}

// ==================== 广告账户管理 ====================

/** 广告账户详细信息（管理页面用） */
export interface FbAdAccountDetail {
  id: string
  accountId: string
  name: string
  fbOwnerName: string
  fbOwnerId: string
  businessName: string
  ownerBusinessId: string
  accountStatus: number
  statusLabel: string
  platform: string
  amountSpent: number
  currency: string
  spendCap: number
  balance: number
  dailySpendLimit: number
  adminName: string
  hiddenAdmins: number
  timezoneName: string
  timezoneOffset: number
  countryCode: string
  isPersonal: number
  fundingSource: string
  disableReason: number
  disableReasonLabel: string
  nextBillDate: string
  createdTime: string
}

/** 广告账户详细列表响应 */
export interface FbAdAccountDetailListResponse {
  accounts: FbAdAccountDetail[]
  total: number
}

/** 获取所有已授权FB账号下的广告账户详细信息 */
export function fetchFbAdAccountsDetail() {
  return request.get<FbAdAccountDetailListResponse>({
    url: '/api/v1/fb/ad-accounts/detail',
    showErrorMessage: true
  })
}

// ==================== 支付记录 ====================

/** 支付记录 */
export interface FbPaymentRecord {
  id: string
  accountId: string
  time: string
  description: string
  amount: number
  currency: string
  billingStart: string
  billingEnd: string
  status: string
  paymentMethod: string
}

/** 支付记录列表响应 */
export interface FbPaymentListResponse {
  records: FbPaymentRecord[]
  total: number
}

/** 获取广告账户的支付记录 */
export function fetchFbPaymentHistory(adAccountId: string) {
  return request.get<FbPaymentListResponse>({
    url: `/api/v1/fb/ad-accounts/${adAccountId}/payments`,
    showErrorMessage: false
  })
}
