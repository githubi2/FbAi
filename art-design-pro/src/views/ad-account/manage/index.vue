<!-- 广告账户管理页面 — 展示所有已授权FB账号下的广告账户详细列表 -->
<template>
  <div class="ad-account-manage-page art-full-height">
    <ElCard class="art-table-card">
      <!-- 表格头部 -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton @click="refreshData" v-ripple>{{
              $t('menus.adAccount.refreshAccounts')
            }}</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- 表格 -->
      <ArtTable :loading="loading" :data="data" :columns="columns" />

      <!-- 空状态 -->
      <ElEmpty
        v-if="!loading && data.length === 0"
        :description="$t('menus.adAccount.noAdAccounts')"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { h } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useTable } from '@/hooks/core/useTable'
  import { ElTag, ElEmpty, ElTooltip } from 'element-plus'
  import type { FbAdAccountDetail } from '@/api/facebook'
  import { fetchFbAdAccountsDetail } from '@/api/facebook'

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

  const formatDateOnly = (val: string) => {
    if (!val) return '—'
    const d = new Date(val)
    if (isNaN(d.getTime())) return val
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
  }

  const formatCurrency = (val: number, currency: string) => {
    const formatted = val.toLocaleString('en-US', {
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
        // 1. 序号
        { type: 'index', width: 58, label: '#' },
        // 2. 状态
        {
          prop: 'accountStatus',
          label: t('menus.adAccount.columns.status'),
          width: 90,
          formatter: (row: FbAdAccountDetail) => {
            const config = getStatusConfig(row.accountStatus)
            return h(ElTag, { type: config.type, size: 'small' }, () => row.statusLabel || '—')
          }
        },
        // 3. 账号
        {
          prop: 'name',
          label: t('menus.adAccount.columns.adAccountName'),
          minWidth: 180,
          formatter: (row: FbAdAccountDetail) => {
            return h(
              ElTooltip,
              {
                content: `ID: ${row.accountId || row.id}`,
                placement: 'top'
              },
              () => h('span', row.name || '—')
            )
          }
        },
        // 4. 推送状态
        {
          prop: 'pushStatus',
          label: t('menus.adAccount.columns.pushStatus'),
          width: 90,
          formatter: () => '—'
        },
        // 5. 管理员
        {
          prop: 'adminName',
          label: t('menus.adAccount.columns.admin'),
          minWidth: 130,
          formatter: (row: FbAdAccountDetail) => row.adminName || '—'
        },
        // 6. 隐藏管理员
        {
          prop: 'hiddenAdmins',
          label: t('menus.adAccount.columns.hiddenAdmin'),
          width: 110,
          formatter: (row: FbAdAccountDetail) => row.hiddenAdmins.toString()
        },
        // 7. 账号类型（个人/企业）
        {
          prop: 'accountType',
          label: t('menus.adAccount.columns.accountType'),
          width: 90,
          formatter: (row: FbAdAccountDetail) => {
            if (row.isPersonal === 1) return '个人'
            if (row.isPersonal === 0) return '企业'
            return '—'
          }
        },
        // 8. 账单金额
        {
          prop: 'billAmount',
          label: t('menus.adAccount.columns.billAmount'),
          minWidth: 130,
          formatter: (row: FbAdAccountDetail) => formatCurrency(row.balance, row.currency)
        },
        // 9. 门槛
        {
          prop: 'threshold',
          label: t('menus.adAccount.columns.threshold'),
          minWidth: 110,
          formatter: () => '—'
        },
        // 10. 日限额
        {
          prop: 'dailySpendLimit',
          label: t('menus.adAccount.columns.dailyLimit'),
          minWidth: 130,
          formatter: (row: FbAdAccountDetail) => {
            if (row.dailySpendLimit === 0) {
              return h('span', { style: { color: '#999' } }, t('menus.adAccount.unlimited'))
            }
            return formatCurrency(row.dailySpendLimit, row.currency)
          }
        },
        // 11. 总花费
        {
          prop: 'totalSpend',
          label: t('menus.adAccount.columns.totalSpend'),
          minWidth: 140,
          formatter: (row: FbAdAccountDetail) => formatCurrency(row.amountSpent, row.currency)
        },
        // 12. 花费限额
        {
          prop: 'spendCap',
          label: t('menus.adAccount.columns.spendCap'),
          minWidth: 130,
          formatter: (row: FbAdAccountDetail) => {
            if (row.spendCap === 0) {
              return h('span', { style: { color: '#999' } }, t('menus.adAccount.unlimited'))
            }
            return formatCurrency(row.spendCap, row.currency)
          }
        },
        // 13. 已花费
        {
          prop: 'spentAmount',
          label: t('menus.adAccount.columns.spentAmount'),
          minWidth: 140,
          formatter: (row: FbAdAccountDetail) => formatCurrency(row.amountSpent, row.currency)
        },
        // 14. 余额
        {
          prop: 'balance',
          label: t('menus.adAccount.columns.balance'),
          minWidth: 130,
          formatter: (row: FbAdAccountDetail) => formatCurrency(row.balance, row.currency)
        },
        // 15. 备注
        {
          prop: 'remark',
          label: t('menus.adAccount.columns.remark'),
          minWidth: 120,
          formatter: () => '—'
        },
        // 16. 币种
        {
          prop: 'currency',
          label: t('menus.adAccount.columns.currency'),
          width: 80,
          formatter: (row: FbAdAccountDetail) => row.currency || '—'
        },
        // 17. 账户类型
        {
          prop: 'accountType2',
          label: t('menus.adAccount.columns.accountType'),
          width: 90,
          formatter: (row: FbAdAccountDetail) => {
            return h(ElTag, { type: 'primary', size: 'small' }, () => row.platform || 'Facebook')
          }
        },
        // 18. 所有者角色
        {
          prop: 'ownerBusinessId',
          label: t('menus.adAccount.columns.ownerRole'),
          minWidth: 140,
          formatter: (row: FbAdAccountDetail) => row.ownerBusinessId || '—'
        },
        // 19. 支付方法
        {
          prop: 'fundingSource',
          label: t('menus.adAccount.columns.paymentMethod'),
          minWidth: 150,
          formatter: (row: FbAdAccountDetail) => row.fundingSource || '—'
        },
        // 20. 账单期
        {
          prop: 'nextBillDate',
          label: t('menus.adAccount.columns.billPeriod'),
          minWidth: 120,
          formatter: (row: FbAdAccountDetail) => formatDateOnly(row.nextBillDate)
        },
        // 21. 锁定原因
        {
          prop: 'disableReason',
          label: t('menus.adAccount.columns.lockReason'),
          minWidth: 130,
          formatter: (row: FbAdAccountDetail) => {
            if (row.disableReason === 0) return '—'
            return row.disableReasonLabel || `状态码: ${row.disableReason}`
          }
        },
        // 22. 创建日期
        {
          prop: 'createdTime',
          label: t('menus.adAccount.columns.createdTime'),
          minWidth: 170,
          formatter: (row: FbAdAccountDetail) => formatDate(row.createdTime)
        },
        // 23. 时区
        {
          prop: 'timezoneName',
          label: t('menus.adAccount.columns.timezone'),
          minWidth: 140,
          formatter: (row: FbAdAccountDetail) => {
            let tz = row.timezoneName || '—'
            if (row.timezoneOffset != null && row.timezoneName) {
              const sign = row.timezoneOffset >= 0 ? '+' : ''
              tz += ` (UTC${sign}${row.timezoneOffset})`
            }
            return tz
          }
        },
        // 24. 原始ID
        {
          prop: 'originalId',
          label: t('menus.adAccount.columns.originalId'),
          minWidth: 160,
          formatter: (row: FbAdAccountDetail) => row.accountId || row.id || '—'
        },
        // 25. 创建自BM
        {
          prop: 'createdFromBm',
          label: t('menus.adAccount.columns.createdFromBm'),
          minWidth: 150,
          formatter: (row: FbAdAccountDetail) => row.businessName || '—'
        },
        // 26. 所属BM
        {
          prop: 'bmName',
          label: t('menus.adAccount.columns.bmName'),
          minWidth: 150,
          formatter: (row: FbAdAccountDetail) => row.businessName || '—'
        },
        // 27. 国家编码
        {
          prop: 'countryCode',
          label: t('menus.adAccount.columns.countryCode'),
          width: 100,
          formatter: (row: FbAdAccountDetail) => {
            if (row.countryCode) return row.countryCode
            // fallback: extract from timezone
            if (row.timezoneName) {
              const parts = row.timezoneName.split('/')
              return parts.length === 2 ? parts[0] : row.timezoneName
            }
            return '—'
          }
        },
        // 28. 支付记录
        {
          prop: 'paymentRecord',
          label: t('menus.adAccount.columns.paymentRecord'),
          minWidth: 120,
          formatter: () => '—'
        }
      ]
    }
  })
</script>

<style lang="scss" scoped>
  .ad-account-manage-page {
    padding: 0;
  }
</style>
