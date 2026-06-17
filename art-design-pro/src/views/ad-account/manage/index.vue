<!-- 广告账户管理页面 — 展示所有已授权FB账号下的广告账户详细列表 -->
<template>
  <div class="ad-account-manage-page art-full-height">
    <ElCard class="art-table-card">
      <!-- 表格头部 -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton @click="refreshData" v-ripple>{{ $t('menus.adAccount.refreshAccounts') }}</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- 表格 -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
      />

      <!-- 空状态 -->
      <ElEmpty
        v-if="!loading && data.length === 0"
        :description="$t('menus.adAccount.noAdAccounts')"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
import { h, computed } from 'vue'
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

const formatCurrency = (val: number, currency: string) => {
  const formatted = val.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  return `${currency || 'USD'} ${formatted}`
}

// 状态标签配置
const getStatusConfig = (status: number): { type: 'success' | 'warning' | 'danger' | 'info' } => {
  switch (status) {
    case 1: return { type: 'success' }
    case 2: return { type: 'danger' }
    case 3: return { type: 'warning' }
    case 7: return { type: 'info' }
    case 9: return { type: 'info' }
    case 100: return { type: 'warning' }
    case 101: return { type: 'danger' }
    default: return { type: 'info' }
  }
}

// ==================== useTable ====================
const fetchDetail = async (_params: any) => {
  try {
    const result = await fetchFbAdAccountsDetail()
    return { list: result.accounts || [], total: result.total || 0, page: 1, size: result.total || 0 }
  } catch {
    return { list: [], total: 0, page: 1, size: 0 }
  }
}

const {
  columns,
  columnChecks,
  data,
  loading,
  refreshData
} = useTable({
  core: {
    apiFn: fetchDetail,
    apiParams: { current: 1, size: 200 },
    columnsFactory: () => [
      { type: 'index', width: 58, label: '#' },
      {
        prop: 'accountId',
        label: t('menus.adAccount.columns.adAccountId'),
        minWidth: 180,
        formatter: (row: FbAdAccountDetail) => row.accountId || row.id || '—'
      },
      {
        prop: 'name',
        label: t('menus.adAccount.columns.adAccountName'),
        minWidth: 200,
        formatter: (row: FbAdAccountDetail) => row.name || '—'
      },
      {
        prop: 'fbOwnerName',
        label: t('menus.adAccount.columns.ownerAccount'),
        minWidth: 150,
        formatter: (row: FbAdAccountDetail) => {
          const name = row.fbOwnerName || row.fbOwnerId || '—'
          return h(ElTooltip, { content: `${t('menus.adAccount.columns.ownerAccount')}: ${row.fbOwnerId}`, placement: 'top' }, () =>
            h('span', name)
          )
        }
      },
      {
        prop: 'businessName',
        label: t('menus.adAccount.columns.bmName'),
        minWidth: 160,
        formatter: (row: FbAdAccountDetail) => row.businessName || '—'
      },
      {
        prop: 'accountStatus',
        label: t('menus.adAccount.columns.status'),
        width: 90,
        formatter: (row: FbAdAccountDetail) => {
          const config = getStatusConfig(row.accountStatus)
          return h(ElTag, { type: config.type, size: 'small' }, () => row.statusLabel || '—')
        }
      },
      {
        prop: 'platform',
        label: t('menus.adAccount.columns.platform'),
        width: 100,
        formatter: (row: FbAdAccountDetail) => {
          return h(ElTag, { type: 'primary', size: 'small' }, () => row.platform || 'Facebook')
        }
      },
      {
        prop: 'currency',
        label: t('menus.adAccount.columns.currency'),
        width: 80,
        formatter: (row: FbAdAccountDetail) => row.currency || '—'
      },
      {
        prop: 'timezoneName',
        label: t('menus.adAccount.columns.country'),
        minWidth: 140,
        formatter: (row: FbAdAccountDetail) => row.timezoneName || '—'
      },
      {
        prop: 'timezoneOffset',
        label: t('menus.adAccount.columns.timezone'),
        width: 90,
        formatter: (row: FbAdAccountDetail) => {
          if (row.timezoneOffset === 0 && !row.timezoneName) return '—'
          const sign = row.timezoneOffset >= 0 ? '+' : ''
          return `UTC${sign}${row.timezoneOffset}`
        }
      },
      {
        prop: 'amountSpent',
        label: t('menus.adAccount.columns.totalSpend'),
        minWidth: 150,
        formatter: (row: FbAdAccountDetail) => formatCurrency(row.amountSpent, row.currency)
      },
      {
        prop: 'spendCap',
        label: t('menus.adAccount.columns.spendCap'),
        minWidth: 140,
        formatter: (row: FbAdAccountDetail) => {
          if (row.spendCap === 0) {
            return h('span', { style: { color: '#999' } }, t('menus.adAccount.unlimited'))
          }
          return formatCurrency(row.spendCap, row.currency)
        }
      },
      {
        prop: 'balance',
        label: t('menus.adAccount.columns.nextBillAmount'),
        minWidth: 140,
        formatter: (row: FbAdAccountDetail) => formatCurrency(row.balance, row.currency)
      },
      {
        prop: 'adminName',
        label: t('menus.adAccount.columns.admin'),
        minWidth: 140,
        formatter: (row: FbAdAccountDetail) => row.adminName || '—'
      },
      {
        prop: 'hiddenAdmins',
        label: t('menus.adAccount.columns.hiddenAdmin'),
        width: 110,
        formatter: (row: FbAdAccountDetail) => {
          if (row.hiddenAdmins === 0) return '0'
          return h('span', { style: { color: '#909399' } }, row.hiddenAdmins.toString())
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
</script>

<style lang="scss" scoped>
.ad-account-manage-page {
  padding: 0;
}
</style>
