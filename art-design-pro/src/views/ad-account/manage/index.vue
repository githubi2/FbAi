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
    <ElDialog v-model="adminDialogVisible" :title="adminDialogTitle" width="600px" destroy-on-close>
      <div class="admin-dialog-content">
        <!-- TODO: 内容待完善 -->
      </div>
      <template #footer>
        <div class="dialog-footer">
          <ElButton @click="adminDialogVisible = false">{{ $t('common.cancel') }}</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { h, ref, reactive } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useTable } from '@/hooks/core/useTable'
  import { ElTag, ElEmpty, ElTooltip, ElButton, ElDialog, ElMessage } from 'element-plus'
  import type { FbAdAccountDetail, FbPaymentRecord } from '@/api/facebook'
  import { fetchFbAdAccountsDetail, fetchFbPaymentHistory } from '@/api/facebook'

  defineOptions({ name: 'AdAccountManage' })

  const { t } = useI18n()

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

  const showAdminDetail = (row: FbAdAccountDetail, type: 'admin' | 'hidden') => {
    curAdminAccount.value = row
    curAdminType.value = type
    if (type === 'admin') {
      adminDialogTitle.value = `${row.name || row.accountId || '—'} — 管理员`
    } else {
      adminDialogTitle.value = `${row.name || row.accountId || '—'} — 隐藏管理员`
    }
    adminDialogVisible.value = true
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

  .admin-dialog-content {
    min-height: 120px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--el-text-color-secondary);
  }
</style>
