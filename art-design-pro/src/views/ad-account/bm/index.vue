<!-- BM 管理页面 — 展示所有已授权FB账号下的 Business Manager 列表 -->
<template>
  <div class="ad-account-bm-page art-full-height">
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
        :description="$t('menus.adAccount.bmNoData')"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { h, computed } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useTable } from '@/hooks/core/useTable'
  import { ElTag, ElEmpty, ElTooltip } from 'element-plus'

  defineOptions({ name: 'AdAccountBm' })

  const { t } = useI18n()

  // ==================== 数据类型 ====================
  interface BmItem {
    bmId: string
    bmName: string
    fbOwnerName: string
    fbOwnerId: string
    adAccountCount: number
    status: string
    createdTime: string
  }

  // ==================== 格式化 ====================
  const formatDate = (val: string) => {
    if (!val) return '—'
    const d = new Date(val)
    if (isNaN(d.getTime())) return val
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  }

  // ==================== useTable ====================
  const fetchBmList = async (_params: any) => {
    try {
      // TODO: 接入真实 API — GET /api/v1/fb/bm/list
      const res = await (await import('@/api/facebook')).fetchFbAccountList()
      // 从 token 的 bm_list 聚合 BM（暂时用 mock）
      const mockBmList: BmItem[] = [
        {
          bmId: '123456789',
          bmName: 'Demo Business Manager 1',
          fbOwnerName: 'admin',
          fbOwnerId: '100001234',
          adAccountCount: 3,
          status: '正常',
          createdTime: '2025-01-15T08:00:00Z'
        },
        {
          bmId: '987654321',
          bmName: 'Demo Business Manager 2',
          fbOwnerName: 'admin',
          fbOwnerId: '100001234',
          adAccountCount: 1,
          status: '正常',
          createdTime: '2025-03-20T10:30:00Z'
        }
      ]
      return { list: mockBmList, total: mockBmList.length, page: 1, size: mockBmList.length }
    } catch {
      return { list: [], total: 0, page: 1, size: 0 }
    }
  }

  const { columns, columnChecks, data, loading, refreshData } = useTable({
    core: {
      apiFn: fetchBmList,
      apiParams: { current: 1, size: 200 },
      columnsFactory: () => [
        { type: 'index', width: 58, label: '#' },
        {
          prop: 'bmId',
          label: t('menus.adAccount.columns.bmId'),
          minWidth: 160,
          formatter: (row: BmItem) => row.bmId || '—'
        },
        {
          prop: 'bmName',
          label: t('menus.adAccount.columns.bmName'),
          minWidth: 200,
          formatter: (row: BmItem) => row.bmName || '—'
        },
        {
          prop: 'fbOwnerName',
          label: t('menus.adAccount.columns.ownerAccount'),
          minWidth: 150,
          formatter: (row: BmItem) => {
            const name = row.fbOwnerName || row.fbOwnerId || '—'
            return h(
              ElTooltip,
              {
                content: `${t('menus.adAccount.columns.ownerAccount')}: ${row.fbOwnerId}`,
                placement: 'top'
              },
              () => h('span', name)
            )
          }
        },
        {
          prop: 'adAccountCount',
          label: t('menus.adAccount.columns.adAccounts'),
          width: 120,
          formatter: (row: BmItem) => row.adAccountCount.toString()
        },
        {
          prop: 'status',
          label: t('menus.adAccount.columns.status'),
          width: 90,
          formatter: (row: BmItem) => {
            return h(ElTag, { type: 'success', size: 'small' }, () => row.status)
          }
        },
        {
          prop: 'createdTime',
          label: t('menus.adAccount.columns.createdTime'),
          minWidth: 170,
          formatter: (row: BmItem) => formatDate(row.createdTime)
        }
      ]
    }
  })
</script>

<style lang="scss" scoped>
  .ad-account-bm-page {
    padding: 0;
  }
</style>
