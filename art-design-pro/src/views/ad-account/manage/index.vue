<!-- 广告账户管理页面 -->
<template>
  <div class="ad-account-manage-page art-full-height">
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

      <ArtTable :loading="loading" :data="data" :columns="columns" />

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
  </div>
</template>

<script setup lang="ts">
  import { h, ref } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useTable } from '@/hooks/core/useTable'
  import { ElTag, ElEmpty, ElTooltip, ElButton, ElDialog } from 'element-plus'
  import type { FbAdAccountDetail, FbPaymentRecord } from '@/api/facebook'
  import { fetchFbAdAccountsDetail, fetchFbPaymentHistory } from '@/api/facebook'

  defineOptions({ name: 'AdAccountManage' })

  const { t } = useI18n()

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
    // FB returns "Visa **** 1234" or "credit_card" type
    // Try to extract card type and last 4
    const match = source.match(/([A-Za-z]+)\s+\*+\s*(\d+)/)
    if (match) {
      return `${match[1]} ····${match[2]}`
    }
    return source
  }

  // ==================== useTable ====================
  const fetchDetail = async () => {
    try {
      const result = await fetchFbAdAccountsDetail()
      return {
        list: result.accounts || [],
        total: result.total || 0,
        page: 1,
        size: result.total || 0
      }
    } catch {
      return { list: [], total: 0, page: 1, size: 0 }
    }
  }

  const { columns, columnChecks, data, loading, refreshData } = useTable({
    core: {
      apiFn: fetchDetail,
      apiParams: { current: 1, size: 200 },
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
          width: 95,
          formatter: (row: FbAdAccountDetail) => {
            const total = row.hiddenAdmins + (row.adminName ? 1 : 0)
            return h(
              ElTooltip,
              { content: row.adminName || t('menus.adAccount.noAdmin'), placement: 'top' },
              () => h('span', String(total))
            )
          }
        },
        {
          prop: 'hiddenAdmins',
          label: t('menus.adAccount.columns.hiddenAdmin'),
          width: 105,
          formatter: (row: FbAdAccountDetail) => String(row.hiddenAdmins)
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
    } catch {
      paymentRecords.value = []
    } finally {
      paymentLoading.value = false
    }
  }
</script>

<style lang="scss" scoped>
  .ad-account-manage-page {
    padding: 0;
  }
</style>
