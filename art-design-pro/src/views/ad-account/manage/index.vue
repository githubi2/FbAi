<!-- 广告账户管理页面 -->
<template>
  <div class="ad-account-manage-page art-full-height">
    <!-- 搜索筛选栏 -->
    <ElCard class="mb-4" shadow="never">
      <ElForm :inline="true" :model="searchForm" class="search-form">
        <ElFormItem :label="$t('menus.adAccount.searchKeyword')">
          <ElInput
            v-model="searchForm.keyword"
            :placeholder="$t('menus.adAccount.searchKeyword')"
            clearable
            @clear="handleSearch"
            @keyup.enter="handleSearch"
          />
        </ElFormItem>
        <ElFormItem :label="$t('menus.adAccount.filterStatus')">
          <ElSelect
            v-model="searchForm.status"
            :placeholder="$t('menus.adAccount.statusPlaceholder')"
            clearable
            @change="handleSearch"
          >
            <ElOption :label="$t('menus.adAccount.statusActive')" :value="1" />
            <ElOption :label="$t('menus.adAccount.statusDisabled')" :value="2" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="$t('menus.adAccount.filterAccountType')">
          <ElSelect
            v-model="searchForm.accountType"
            :placeholder="$t('menus.adAccount.filterAccountType')"
            clearable
            @change="handleSearch"
          >
            <ElOption label="企业" value="企业" />
            <ElOption label="个人" value="个人" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem>
          <ElButton @click="handleSearch">{{ $t('table.searchBar.search') }}</ElButton>
          <ElButton @click="handleReset">{{ $t('table.searchBar.reset') }}</ElButton>
        </ElFormItem>
      </ElForm>

      <!-- 批量操作按钮组 -->
      <div class="batch-actions">
        <ElButton v-ripple @click="handleBatchAction('addAuth')">{{
          $t('menus.adAccount.addAuth')
        }}</ElButton>
        <ElButton v-ripple @click="handleBatchAction('deleteAuth')">{{
          $t('menus.adAccount.deleteAuth')
        }}</ElButton>
        <ElButton v-ripple @click="handleBatchAction('addToBM')">{{
          $t('menus.adAccount.addToBM')
        }}</ElButton>
        <ElButton v-ripple @click="handleBatchAction('setLimit')">{{
          $t('menus.adAccount.setLimit')
        }}</ElButton>
        <ElButton v-ripple @click="handleBatchAction('resetLimit')">{{
          $t('menus.adAccount.resetLimit')
        }}</ElButton>
        <ElButton v-ripple @click="handleBatchAction('hideAdmin')">{{
          $t('menus.adAccount.hideAdmin')
        }}</ElButton>
        <ElButton v-ripple @click="handleBatchAction('accountPush')">{{
          $t('menus.adAccount.accountPush')
        }}</ElButton>
      </div>
    </ElCard>

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton @click="refreshData" v-ripple>{{
              $t('menus.adAccount.refreshAccounts')
            }}</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @selection-change="handleSelectionChange"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />

      <ElEmpty
        v-if="!loading && data.length === 0"
        :description="$t('menus.adAccount.noAdAccounts')"
      />
    </ElCard>

    <!-- 支付记录弹窗 -->
    <ElDialog
      v-model="paymentDialogVisible"
      :title="paymentDialogTitle"
      width="800px"
      destroy-on-close
    >
      <ArtTable :loading="paymentLoading" :data="paymentRecords" :columns="paymentColumns" />
      <ElEmpty v-if="!paymentLoading && paymentRecords.length === 0" description="暂无支付记录" />
    </ElDialog>

    <!-- 管理员详情弹窗 -->
    <ElDialog v-model="adminDialogVisible" :title="adminDialogTitle" width="550px" destroy-on-close>
      <div class="admin-dialog-content">
        <!-- 步骤1：选择要删除的管理员 -->
        <div class="admin-step">
          <div class="admin-step-head">
            <span class="admin-step-num">1</span>
            <span class="admin-step-label">
              {{ $t('menus.adAccount.adminDialogStep1') }}
            </span>
          </div>
          <div class="admin-step-body">
            <ElCheckboxGroup v-model="selectedAdmins" class="admin-checklist">
              <ElCheckbox
                v-for="name in curOtherAdminNames"
                :key="name"
                :label="name"
                :value="name"
              >
                {{ name }}
              </ElCheckbox>
            </ElCheckboxGroup>
            <ElEmpty
              v-if="curOtherAdminNames.length === 0"
              :description="$t('menus.adAccount.adminDialogNoOther')"
              :image-size="60"
            />
          </div>
        </div>
        <!-- 步骤2：执行时间间隔 -->
        <div class="admin-step">
          <div class="admin-step-head">
            <span class="admin-step-num">2</span>
            <ElCheckbox v-model="useDefaultInterval">
              {{ $t('menus.adAccount.adminDialogStep2') }}
            </ElCheckbox>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <ElButton @click="adminDialogVisible = false">{{ $t('common.cancel') }}</ElButton>
          <ElButton
            type="primary"
            :disabled="selectedAdmins.length === 0"
            @click="handleAdminDelete"
          >
            {{ $t('menus.adAccount.adminDialogConfirm') }}
          </ElButton>
        </div>
      </template>
    </ElDialog>

    <!-- 增加授权弹窗 -->
    <ElDialog
      v-model="addAuthDialogVisible"
      :title="$t('menus.adAccount.addAuthTitle')"
      width="540px"
      :close-on-click-modal="false"
      @close="stopAuthPolling"
    >
      <div class="add-auth-dialog-body">
        <p class="add-auth-dialog-tip">{{ $t('menus.adAccount.addAuthTip') }}</p>
        <!-- 短链接 -->
        <div class="short-link-box">
          <label class="link-label">{{ $t('menus.adAccount.shortLinkLabel') }}</label>
          <div class="auth-link-box">
            <ElInput
              v-model="addAuthShortUrl"
              readonly
              :placeholder="$t('menus.adAccount.generatingLink')"
            />
            <ElButton type="primary" class="ml-2" @click="copyAddAuthShortUrl">
              {{
                addAuthCopySuccess ? $t('menus.adAccount.copied') : $t('menus.adAccount.copyLink')
              }}
            </ElButton>
          </div>
        </div>
        <!-- 完整链接 -->
        <div v-if="addAuthFullUrl" class="full-link-box mt-3">
          <label class="link-label">{{ $t('menus.adAccount.fullLinkLabel') }}</label>
          <div class="auth-link-box">
            <ElInput v-model="addAuthFullUrl" readonly size="small" />
            <ElButton size="small" class="ml-2" @click="copyAddAuthFullUrl">
              {{ $t('menus.adAccount.copyLink') }}
            </ElButton>
          </div>
        </div>
        <div class="add-auth-dialog-actions mt-4">
          <ElButton @click="openAddAuthUrl">
            {{ $t('menus.adAccount.openInBrowser') }}
          </ElButton>
          <span class="text-gray-400 text-sm ml-2">
            {{ $t('menus.adAccount.orCopyToOther') }}
          </span>
        </div>
        <!-- 授权成功提示 -->
        <div v-if="addAuthSuccess" class="add-auth-success-bar mt-4">
          <ElAlert type="success" :closable="false" show-icon>
            <template #title>
              {{ $t('menus.adAccount.addAuthSuccess') }}
            </template>
          </ElAlert>
        </div>
        <!-- 轮询状态 -->
        <div v-if="isAddAuthPolling && !addAuthSuccess" class="polling-status mt-4">
          <ElAlert type="info" :closable="false" show-icon>
            <template #title>
              <ElIcon class="is-loading mr-1"><Loading /></ElIcon>
              {{ $t('menus.adAccount.waitingAuth') }}
            </template>
          </ElAlert>
        </div>
      </div>
      <template #footer>
        <ElButton @click="stopAuthPollingAndClose">{{ $t('menus.adAccount.cancelAuth') }}</ElButton>
        <ElButton type="primary" :disabled="!addAuthSuccess" @click="handleAddAuthConfirm">
          {{ $t('common.confirm') }}
        </ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { h, ref, reactive, onUnmounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useTable } from '@/hooks/core/useTable'
  import {
    ElTag,
    ElEmpty,
    ElTooltip,
    ElButton,
    ElDialog,
    ElMessage,
    ElCheckbox,
    ElCheckboxGroup,
    ElAlert,
    ElIcon,
    ElInput
  } from 'element-plus'
  import { Loading } from '@element-plus/icons-vue'
  import type { FbAdAccountDetail, FbPaymentRecord } from '@/api/facebook'
  import {
    fetchFbAdAccountsDetail,
    fetchFbPaymentHistory,
    fetchFbAuthUrl,
    fetchFbAccountList
  } from '@/api/facebook'

  defineOptions({ name: 'AdAccountManage' })

  const { t } = useI18n()

  // ==================== 多选 ====================
  const selectedRows = ref<FbAdAccountDetail[]>([])

  const handleSelectionChange = (selection: FbAdAccountDetail[]) => {
    selectedRows.value = selection
  }

  // ==================== 批量操作 ====================
  const handleBatchAction = (action: string) => {
    if (action === 'addAuth') {
      handleOpenAddAuth()
      return
    }
    if (selectedRows.value.length === 0) {
      ElMessage.warning(t('menus.adAccount.selectRowsFirst'))
      return
    }
    // TODO: 实现各批量操作功能
    console.log('Batch action:', action, selectedRows.value)
  }

  // ==================== 增加授权弹窗 ====================
  const addAuthDialogVisible = ref(false)
  const addAuthShortUrl = ref('')
  const addAuthFullUrl = ref('')
  const addAuthSuccess = ref(false)
  const addAuthCopySuccess = ref(false)
  const isAddAuthPolling = ref(false)
  const isAddAuthLoading = ref(false)
  let addAuthPollTimer: ReturnType<typeof setInterval> | null = null

  const startAddAuthPolling = () => {
    if (addAuthPollTimer) return
    isAddAuthPolling.value = true
    const initialCount = pagination.total
    addAuthPollTimer = setInterval(async () => {
      try {
        const result = await fetchFbAccountList()
        if (result.total > initialCount) {
          stopAuthPolling()
          addAuthSuccess.value = true
          ElMessage.success(t('menus.adAccount.addAuthSuccess'))
        }
      } catch {
        // 轮询失败继续
      }
    }, 3000)
  }

  const stopAuthPolling = () => {
    if (addAuthPollTimer) {
      clearInterval(addAuthPollTimer)
      addAuthPollTimer = null
    }
    isAddAuthPolling.value = false
  }

  const stopAuthPollingAndClose = () => {
    stopAuthPolling()
    addAuthDialogVisible.value = false
    addAuthSuccess.value = false
  }

  const handleAddAuthConfirm = () => {
    addAuthDialogVisible.value = false
    addAuthSuccess.value = false
    refreshData()
  }

  const handleOpenAddAuth = async () => {
    isAddAuthLoading.value = true
    addAuthSuccess.value = false
    addAuthCopySuccess.value = false
    try {
      const { authUrl: full, shortUrl: short } = await fetchFbAuthUrl()
      addAuthShortUrl.value = short
      addAuthFullUrl.value = full
      addAuthDialogVisible.value = true
      startAddAuthPolling()
    } catch {
      ElMessage.error(t('menus.adAccount.authUrlError'))
    } finally {
      isAddAuthLoading.value = false
    }
  }

  const copyAddAuthShortUrl = async () => {
    try {
      await navigator.clipboard.writeText(addAuthShortUrl.value)
      addAuthCopySuccess.value = true
      ElMessage.success(t('menus.adAccount.copySuccess'))
      setTimeout(() => {
        addAuthCopySuccess.value = false
      }, 2000)
    } catch {
      fallbackAddAuthCopy(addAuthShortUrl.value)
    }
  }

  const copyAddAuthFullUrl = async () => {
    try {
      await navigator.clipboard.writeText(addAuthFullUrl.value)
      ElMessage.success(t('menus.adAccount.copySuccess'))
    } catch {
      fallbackAddAuthCopy(addAuthFullUrl.value)
    }
  }

  const fallbackAddAuthCopy = (text: string) => {
    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
    addAuthCopySuccess.value = true
    ElMessage.success(t('menus.adAccount.copySuccess'))
    setTimeout(() => {
      addAuthCopySuccess.value = false
    }, 2000)
  }

  const openAddAuthUrl = () => {
    if (addAuthFullUrl.value) {
      window.open(addAuthFullUrl.value, '_blank')
    }
  }

  onUnmounted(() => {
    stopAuthPolling()
  })

  // ==================== 搜索筛选 ====================
  const searchForm = reactive({
    keyword: '',
    status: null as number | null,
    accountType: ''
  })

  // ==================== 格式化 ====================
  const formatDate = (val: string) => {
    if (!val) return '—'
    const d = new Date(val)
    if (isNaN(d.getTime())) return val
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  }

  const formatCurrency = (val: number, currency: string) => {
    const formatted = Math.abs(val).toLocaleString('en-US', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    })
    return `${currency || 'USD'} ${formatted}`
  }

  // 状态标签配置
  const getStatusConfig = (status: number): { type: 'success' | 'warning' | 'danger' | 'info' } => {
    switch (status) {
      case 1:
        return { type: 'success' }
      case 2:
        return { type: 'danger' }
      case 3:
        return { type: 'warning' }
      case 7:
        return { type: 'info' }
      case 9:
        return { type: 'info' }
      case 100:
        return { type: 'warning' }
      case 101:
        return { type: 'danger' }
      default:
        return { type: 'info' }
    }
  }

  // 支付方式格式化：提取卡类型+后四位
  const formatPaymentMethod = (source: string) => {
    if (!source) return '—'
    const match = source.match(/([A-Za-z]+)\s+\*+\s*(\d+)/)
    if (match) {
      return `${match[1]} ····${match[2]}`
    }
    return source
  }

  // ==================== useTable — 客户端分页+筛选 ====================
  const fetchDetail = async (params: any) => {
    const current = params?.current || 1
    const size = params?.size || 20

    try {
      const result = await fetchFbAdAccountsDetail()
      let accounts = result.accounts || []

      // 客户端筛选
      const keyword = (params?.keyword || '').toLowerCase().trim()
      const statusFilter = params?.status
      const accountTypeFilter = params?.accountType || ''

      if (keyword) {
        accounts = accounts.filter(
          (a: FbAdAccountDetail) =>
            (a.name || '').toLowerCase().includes(keyword) ||
            (a.accountId || a.id || '').toLowerCase().includes(keyword) ||
            (a.businessName || '').toLowerCase().includes(keyword)
        )
      }

      if (statusFilter != null && statusFilter !== '') {
        accounts = accounts.filter(
          (a: FbAdAccountDetail) => a.accountStatus === Number(statusFilter)
        )
      }

      if (accountTypeFilter) {
        if (accountTypeFilter === '企业') {
          accounts = accounts.filter((a: FbAdAccountDetail) => !!a.businessName)
        } else if (accountTypeFilter === '个人') {
          accounts = accounts.filter((a: FbAdAccountDetail) => !a.businessName)
        }
      }

      // 客户端分页
      const total = accounts.length
      const start = (current - 1) * size
      const list = accounts.slice(start, start + size)

      return { list, total, page: current, size }
    } catch {
      return { list: [], total: 0, page: 1, size: 20 }
    }
  }

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    replaceSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchDetail,
      apiParams: { current: 1, size: 20 },
      columnsFactory: () => [
        { type: 'selection', width: 55 },
        { type: 'index', width: 55, label: '#' },
        {
          prop: 'accountStatus',
          label: t('menus.adAccount.columns.status'),
          width: 85,
          formatter: (row: FbAdAccountDetail) => {
            const config = getStatusConfig(row.accountStatus)
            return h(ElTag, { type: config.type, size: 'small' }, () => row.statusLabel || '—')
          }
        },
        {
          prop: 'accountId',
          label: t('menus.adAccount.columns.adAccountId'),
          minWidth: 170,
          formatter: (row: FbAdAccountDetail) => {
            return h(
              ElTooltip,
              { content: `${row.name || '—'} (${row.accountId || row.id})`, placement: 'top' },
              () => h('span', row.accountId || row.id || '—')
            )
          }
        },
        {
          prop: 'accountType',
          label: t('menus.adAccount.columns.accountType'),
          width: 85,
          formatter: (row: FbAdAccountDetail) => {
            if (row.businessName) return h(ElTag, { type: 'primary', size: 'small' }, () => '企业')
            if (row.accountId || row.id)
              return h(ElTag, { type: 'warning', size: 'small' }, () => '个人')
            return h(ElTag, { type: 'info', size: 'small' }, () => '—')
          }
        },
        {
          prop: 'adminCount',
          label: t('menus.adAccount.columns.admin'),
          minWidth: 110,
          formatter: (row: FbAdAccountDetail) => {
            const total = row.hiddenAdmins + (row.adminName ? 1 : 0)
            return h(
              ElTag,
              {
                type: 'primary',
                size: 'small',
                style: { cursor: 'pointer' },
                onClick: () => showAdminDetail(row, 'admin')
              },
              () => String(total)
            )
          }
        },
        {
          prop: 'hiddenAdmins',
          label: t('menus.adAccount.columns.hiddenAdmin'),
          minWidth: 110,
          formatter: (row: FbAdAccountDetail) => {
            const count = row.hiddenAdmins || 0
            return h(
              ElTag,
              {
                type: 'primary',
                size: 'small',
                style: { cursor: 'pointer' },
                onClick: () => showAdminDetail(row, 'hidden')
              },
              () => String(count)
            )
          }
        },
        {
          prop: 'fundingSource',
          label: t('menus.adAccount.columns.paymentMethod'),
          minWidth: 150,
          formatter: (row: FbAdAccountDetail) => formatPaymentMethod(row.fundingSource)
        },
        {
          prop: 'currency',
          label: t('menus.adAccount.columns.currency'),
          width: 75,
          formatter: (row: FbAdAccountDetail) => row.currency || '—'
        },
        {
          prop: 'balance',
          label: t('menus.adAccount.columns.balance'),
          minWidth: 130,
          formatter: (row: FbAdAccountDetail) => formatCurrency(row.balance, row.currency)
        },
        {
          prop: 'spendCap',
          label: t('menus.adAccount.columns.dailyLimit'),
          minWidth: 130,
          formatter: (row: FbAdAccountDetail) => {
            if (row.spendCap === 0)
              return h('span', { style: { color: '#999' } }, t('menus.adAccount.unlimited'))
            return formatCurrency(row.spendCap, row.currency)
          }
        },
        {
          prop: 'amountSpent',
          label: t('menus.adAccount.columns.totalSpend'),
          minWidth: 140,
          formatter: (row: FbAdAccountDetail) => formatCurrency(row.amountSpent, row.currency)
        },
        {
          prop: 'spentAmount',
          label: t('menus.adAccount.columns.spentAmount'),
          minWidth: 140,
          formatter: (row: FbAdAccountDetail) => formatCurrency(row.amountSpent, row.currency)
        },
        {
          prop: 'disableReason',
          label: t('menus.adAccount.columns.lockReason'),
          minWidth: 130,
          formatter: (row: FbAdAccountDetail) => {
            if (row.disableReason === 0) return '—'
            if (row.disableReasonLabel)
              return h(ElTag, { type: 'danger', size: 'small' }, () => row.disableReasonLabel)
            return `状态码: ${row.disableReason}`
          }
        },
        {
          prop: 'timezoneName',
          label: t('menus.adAccount.columns.timezone'),
          minWidth: 150,
          formatter: (row: FbAdAccountDetail) => {
            let tz = row.timezoneName || '—'
            if (row.timezoneOffset != null && row.timezoneName) {
              const sign = row.timezoneOffset >= 0 ? '+' : ''
              tz += ` (UTC${sign}${row.timezoneOffset})`
            }
            return tz
          }
        },
        {
          prop: 'countryCode',
          label: t('menus.adAccount.columns.countryCode'),
          width: 95,
          formatter: (row: FbAdAccountDetail) => {
            if (row.countryCode) return row.countryCode
            return '—'
          }
        },
        {
          prop: 'bmName',
          label: t('menus.adAccount.columns.bmName'),
          minWidth: 150,
          formatter: (row: FbAdAccountDetail) => row.businessName || '—'
        },
        {
          prop: 'createdFromBm',
          label: t('menus.adAccount.columns.createdFromBm'),
          minWidth: 140,
          formatter: (row: FbAdAccountDetail) => row.businessName || '—'
        },
        {
          prop: 'remark',
          label: t('menus.adAccount.columns.remark'),
          minWidth: 120,
          formatter: () => '—'
        },
        {
          prop: 'paymentRecord',
          label: t('menus.adAccount.columns.paymentRecord'),
          width: 100,
          formatter: (row: FbAdAccountDetail) => {
            return h(ElButton, { size: 'small', onClick: () => showPaymentHistory(row) }, () =>
              t('menus.adAccount.viewPayments')
            )
          }
        },
        {
          prop: 'createdTime',
          label: t('menus.adAccount.columns.createdTime'),
          minWidth: 170,
          formatter: (row: FbAdAccountDetail) => formatDate(row.createdTime)
        }
      ]
    }
  })

  // ==================== 搜索/重置 ====================
  const handleSearch = () => {
    replaceSearchParams({
      keyword: searchForm.keyword,
      status: searchForm.status,
      accountType: searchForm.accountType,
      current: 1,
      size: 20
    } as any)
  }

  const handleReset = () => {
    searchForm.keyword = ''
    searchForm.status = null
    searchForm.accountType = ''
    replaceSearchParams({ keyword: '', status: null, accountType: '', current: 1, size: 20 } as any)
  }

  // ==================== 支付记录弹窗 ====================
  const paymentDialogVisible = ref(false)
  const paymentDialogTitle = ref('')
  const paymentLoading = ref(false)
  const paymentRecords = ref<FbPaymentRecord[]>([])
  const curPaymentAccount = ref<FbAdAccountDetail | null>(null)

  const formatPayDate = (val: string) => {
    if (!val) return '—'
    const d = new Date(val)
    if (isNaN(d.getTime())) return val
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
  }

  const paymentColumns = [
    { type: 'index' as const, width: 50, label: '#' },
    {
      prop: 'time',
      label: '日期',
      minWidth: 110,
      formatter: (row: FbPaymentRecord) => formatPayDate(row.time)
    },
    {
      prop: 'description',
      label: '描述',
      minWidth: 200,
      formatter: (row: FbPaymentRecord) => row.description || '—'
    },
    {
      prop: 'amount',
      label: '金额',
      minWidth: 130,
      formatter: (row: FbPaymentRecord) => formatCurrency(row.amount, row.currency)
    },
    {
      prop: 'status',
      label: '状态',
      width: 90,
      formatter: (row: FbPaymentRecord) => {
        const type = row.status === 'success' ? ('success' as const) : ('info' as const)
        return h(ElTag, { type, size: 'small' }, () => row.status || '—')
      }
    },
    {
      prop: 'billingStart',
      label: '账单周期',
      minWidth: 200,
      formatter: (row: FbPaymentRecord) => {
        const s = formatPayDate(row.billingStart)
        const e = formatPayDate(row.billingEnd)
        if (s === '—' && e === '—') return '—'
        return `${s} ~ ${e}`
      }
    },
    {
      prop: 'paymentMethod',
      label: '支付方式',
      minWidth: 120,
      formatter: (row: FbPaymentRecord) => row.paymentMethod || '—'
    }
  ]

  const showPaymentHistory = async (row: FbAdAccountDetail) => {
    curPaymentAccount.value = row
    paymentDialogTitle.value = `${row.name || row.accountId || '—'} — 支付记录`
    paymentDialogVisible.value = true
    paymentLoading.value = true
    paymentRecords.value = []
    try {
      const result = await fetchFbPaymentHistory(row.id)
      paymentRecords.value = result.records || []
    } catch (e: any) {
      paymentRecords.value = []
      ElMessage.warning(e?.data?.msg || '支付记录暂不可用（需要 Facebook ads_read 权限）')
    } finally {
      paymentLoading.value = false
    }
  }
  // ==================== 管理员详情弹窗 ====================
  const adminDialogVisible = ref(false)
  const adminDialogTitle = ref('')
  const curAdminAccount = ref<FbAdAccountDetail | null>(null)
  const curAdminType = ref<'admin' | 'hidden'>('admin')
  const curOtherAdminNames = ref<string[]>([])
  const selectedAdmins = ref<string[]>([])
  const useDefaultInterval = ref(true)

  const showAdminDetail = (row: FbAdAccountDetail, type: 'admin' | 'hidden') => {
    // 标签值为 0 时直接提示，不弹窗
    if (type === 'hidden') {
      const count = row.hiddenAdmins || 0
      if (count === 0) {
        ElMessage.info(t('menus.adAccount.noHiddenAdmin'))
        return
      }
    } else {
      const total = row.hiddenAdmins + (row.adminName ? 1 : 0)
      if (total === 0) {
        ElMessage.info(t('menus.adAccount.noAdmin'))
        return
      }
    }
    curAdminAccount.value = row
    curAdminType.value = type
    curOtherAdminNames.value = row.otherAdminNames || []
    selectedAdmins.value = []
    useDefaultInterval.value = true
    adminDialogTitle.value = `${row.name || row.accountId || '—'} — ${t('menus.adAccount.adminDialogTitle')}`
    adminDialogVisible.value = true
  }

  const handleAdminDelete = () => {
    // TODO: 调用后端 API 删除选中的管理员
    ElMessage.success(
      t('menus.adAccount.adminDeleteSuccess', { count: selectedAdmins.value.length })
    )
    adminDialogVisible.value = false
  }
</script>

<style lang="scss" scoped>
  .ad-account-manage-page {
    padding: 0;
  }

  .search-form {
    display: flex;
    align-items: flex-end;
    flex-wrap: wrap;
    gap: 8px;

    :deep(.el-form-item) {
      margin-bottom: 0;
    }
  }

  .batch-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid var(--el-border-color-lighter);
  }

  .admin-dialog-content {
    min-height: 120px;

    .admin-step {
      position: relative;
      padding-left: 36px;

      // 连接线：圆圈下方到下一步骤
      &::before {
        content: '';
        position: absolute;
        left: 11px;
        top: 28px;
        bottom: 0;
        width: 2px;
        background-color: var(--el-border-color);
      }

      &:last-child::before {
        display: none;
      }

      &:not(:last-child) {
        padding-bottom: 24px;
      }
    }

    .admin-step-head {
      display: flex;
      align-items: center;
      gap: 12px;
      min-height: 24px;
      margin-bottom: 8px;

      :deep(.el-checkbox) {
        height: 24px;
        line-height: 24px;
      }
    }

    .admin-step-num {
      position: absolute;
      left: 0;
      top: 0;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: 24px;
      height: 24px;
      border-radius: 50%;
      background-color: var(--el-color-primary-light-5);
      color: var(--el-color-primary);
      font-size: 13px;
      font-weight: 600;
      line-height: 1;
      flex-shrink: 0;
    }

    .admin-step-label {
      font-size: 14px;
      font-weight: 500;
      color: var(--el-text-color-primary);
      line-height: 24px;
    }

    .admin-step-body {
      min-width: 0;
    }

    .admin-checklist {
      display: flex;
      flex-direction: column;
      gap: 6px;
      padding-left: 4px;

      :deep(.el-checkbox) {
        margin-right: 0;
      }
    }

    .admin-interval-check {
      margin-top: 4px;
    }
  }

  .add-auth-dialog-body {
    .add-auth-dialog-tip {
      font-size: 14px;
      color: var(--el-text-color-regular);
      margin-bottom: 16px;
      line-height: 1.6;
    }

    .short-link-box,
    .full-link-box {
      .link-label {
        display: block;
        font-size: 13px;
        font-weight: 500;
        color: var(--el-text-color-secondary);
        margin-bottom: 6px;
      }

      .auth-link-box {
        display: flex;
        align-items: center;

        :deep(.el-input) {
          flex: 1;
        }
      }
    }

    .add-auth-dialog-actions {
      display: flex;
      align-items: center;
    }

    .add-auth-success-bar {
      :deep(.el-alert__title) {
        font-weight: 500;
      }
    }
  }
</style>
