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

/** 触发后台刷新所有 FB 账号统计（5分钟冷却期内为 no-op，幂等） */
export function fetchFbRefreshAccounts() {
  return request.post<{ started: boolean }>({
    url: '/api/v1/fb/accounts/refresh-all'
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
  otherAdminNames: string[]
  timezoneName: string
  timezoneOffset: number
  countryCode: string
  isPersonal: number
  fundingSource: string
  disableReason: number
  disableReasonLabel: string
  nextBillDate: string
  createdTime: string
  /** 是否预付费账户（1=预付费, 0=后付费） */
  isPrepay: number
  /** 授权用户在该账户的角色（Admin/Advertiser/Analyst） */
  ownerRole: string
  /** 本地备注 */
  remark: string
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

/** 触发后台刷新广告账户详情（5分钟冷却期内为 no-op，幂等） */
export function fetchFbRefreshAdAccounts() {
  return request.post<{ started: boolean }>({
    url: '/api/v1/fb/ad-accounts/refresh-all'
  })
}

/** 更新广告账户本地备注 */
export function fetchFbUpdateAdAccountRemark(adAccountId: string, remark: string) {
  return request.put<{ remark: string }>({
    url: `/api/v1/fb/ad-accounts/${adAccountId}/remark`,
    data: { remark }
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

// ==================== 广告账户授权 ====================

/** 授权请求参数 */
export interface FbAssignUserParams {
  adAccountIds: string[]
  userId: string
  role: 'ADMIN' | 'ADVERTISER' | 'ANALYST'
}

/** 单个广告账户授权结果 */
export interface FbAssignUserResult {
  adAccountId: string
  success: boolean
  message: string
}

/** 授权响应 */
export interface FbAssignUserResponse {
  results: FbAssignUserResult[]
  total: number
  success: number
  failed: number
}

/** 将用户分配到广告账户 */
export function fetchAssignAdAccountUser(params: FbAssignUserParams) {
  return request.post<FbAssignUserResponse>({
    url: '/api/v1/fb/ad-accounts/assign-user',
    data: params,
    showErrorMessage: true
  })
}

/** 查找用户结果 */
export interface FbLookupUserResult {
  uid: string
  name: string
  isFriend: boolean
  avatar: string
}

/** 查找用户响应 */
export interface FbLookupUserResponse {
  users: FbLookupUserResult[]
}

/** 查找 Facebook 用户信息 */
export function fetchLookupFbUsers(uids: string[]) {
  return request.post<FbLookupUserResponse>({
    url: '/api/v1/fb/users/lookup',
    data: { uids },
    showErrorMessage: true
  })
}

// ==================== 删除广告账号权限 ====================

/** 删除权限请求参数 */
export interface FbRemoveUserParams {
  adAccountIds: string[]
  uids: string[]
  mode: string
  deleteCurrent: boolean
}

/** 删除权限响应 */
export function fetchRemoveAdAccountUser(params: FbRemoveUserParams) {
  return request.post<FbAssignUserResponse>({
    url: '/api/v1/fb/ad-accounts/remove-user',
    data: params,
    showErrorMessage: true
  })
}

// ==================== 刷新状态 ====================

/** 刷新状态响应 */
export interface FbRefreshStatusResponse {
  status: string // pending/running/completed/failed/none
  isRunning: boolean
  error: string
}

/** 获取刷新状态 */
export function fetchRefreshStatus(
  type: 'accounts' | 'ad_accounts' | 'bm' | 'pages' | 'all' = 'all'
) {
  return request.get<FbRefreshStatusResponse>({
    url: '/api/v1/fb/refresh-status',
    params: { type },
    showErrorMessage: false
  })
}

// ==================== BM 列表 ====================

/**
 * BM 列表项
 * 官方公开 API 不提供：BM类型、日限额、隐藏管理员、BM质量（前端显示 —）
 */
export interface FbBmItem {
  bmId: string
  name: string
  fbOwnerName: string
  fbOwnerId: string
  /** 状态：API 可达即为「正常」 */
  statusLabel: string
  /** 本地推送状态 */
  pushStatus: string
  /** 本地备注 */
  remark: string
  /** 授权用户在 BM 中的角色（ADMIN/EMPLOYEE） */
  ownerRole: string
  /** 认证状态（verified/not_verified/...） */
  verificationStatus: string
  /** 管理员总数 = 在职 + 邀请中 */
  adminCount: number
  /** 邀请中管理员数 */
  pendingAdminCount: number
  /** 在职管理员名单 */
  adminNames: string[]
  /** 合作伙伴数（owned_businesses + agencies） */
  partnerCount: number
  /** 广告账户数（owned + client） */
  adAccountCount: number
  createdTime: string
  lastRefreshAt: string | null
}

/** BM 列表响应 */
export interface FbBmListResponse {
  list: FbBmItem[]
  total: number
}

/** 获取 BM 列表（后端缓存直出，毫秒级） */
export function fetchFbBmList() {
  return request.get<FbBmListResponse>({
    url: '/api/v1/fb/bm-list',
    showErrorMessage: true
  })
}

/** 触发后台刷新 BM 列表（5分钟冷却期内为 no-op，幂等） */
export function fetchFbRefreshBmList() {
  return request.post<{ started: boolean }>({
    url: '/api/v1/fb/bm-list/refresh'
  })
}

/** 更新 BM 本地备注 */
export function fetchFbUpdateBmRemark(bmId: string, remark: string) {
  return request.put<{ remark: string }>({
    url: `/api/v1/fb/bm-list/${bmId}/remark`,
    data: { remark }
  })
}

// ==================== FB 公共主页 ====================

/**
 * 公共主页列表项
 * FB 官方 API 不提供：创建时间、创建渠道、主页状态、申诉时间、允许评论、
 * 隐藏不文明用语、广告权限、屏蔽词设置、黑名单列表、主页类型（前端显示 —）
 */
export interface FbPageItem {
  pageId: string
  name: string
  link: string
  fbOwnerName: string
  fbOwnerId: string
  /** 主页分类 */
  category: string
  /** 点赞数 */
  fanCount: number
  /** 粉丝数 */
  followersCount: number
  /** 发布状态 1=已发布 0=未发布 */
  isPublished: number
  /** 认证状态（verified/not_verified/...） */
  verificationStatus: string
  website: string
  phone: string
  email: string
  address: string
  /** 管理员名单 */
  adminNames: string[]
  /** 本地推送状态 */
  pushStatus: string
  /** 本地备注 */
  remark: string
  lastRefreshAt: string | null
}

/** 公共主页列表响应 */
export interface FbPageListResponse {
  list: FbPageItem[]
  total: number
}

/** 获取公共主页列表（后端缓存直出，毫秒级） */
export function fetchFbPages() {
  return request.get<FbPageListResponse>({
    url: '/api/v1/fb/pages',
    showErrorMessage: true
  })
}

/** 触发后台刷新主页列表（5分钟冷却期内为 no-op，幂等） */
export function fetchFbRefreshPages() {
  return request.post<{ started: boolean }>({
    url: '/api/v1/fb/pages/refresh-all'
  })
}

/** 更新主页本地备注 */
export function fetchFbUpdatePageRemark(pageId: string, remark: string) {
  return request.put<{ remark: string }>({
    url: `/api/v1/fb/pages/${pageId}/remark`,
    data: { remark }
  })
}
