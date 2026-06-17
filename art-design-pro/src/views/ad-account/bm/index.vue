<!-- BM 管理页面 — 展示所有已授权FB账号下的 Business Manager 列表 -->
<template>
  <div class="ad-account-bm-page art-full-height">
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
        <ElFormItem>
          <ElButton @click="handleSearch">{{ $t('table.searchBar.search') }}</ElButton>
          <ElButton @click="handleReset">{{ $t('table.searchBar.reset') }}</ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>

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
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />

      <!-- 空状态 -->
      <ElEmpty v-if="!loading && data.length === 0" :description="$t('menus.adAccount.bmNoData')" />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { h, reactive } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useTable } from '@/hooks/core/useTable'
  import { ElTag, ElEmpty, ElTooltip } from 'element-plus'

  defineOptions({ name: 'AdAccountBm' })

  const { t } = useI18n()

  // ==================== 搜索筛选 ====================
  const searchForm = reactive({
    keyword: ''
  })

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

  // ==================== useTable — 客户端分页+筛选 ====================
  // TODO: 接入真实 API — GET /api/v1/fb/bm/list
  const fetchBmList = async (params: any) => {
    const current = params?.current || 1
    const size = params?.size || 20

    try {
      // 模拟数据（后续替换为真实 API 调用）
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

      let list = mockBmList

      // 客户端筛选
      const keyword = (params?.keyword || '').toLowerCase().trim()
      if (keyword) {
        list = list.filter(
          (item) =>
            (item.bmName || '').toLowerCase().includes(keyword) ||
            (item.bmId || '').toLowerCase().includes(keyword) ||
            (item.fbOwnerName || '').toLowerCase().includes(keyword)
        )
      }

      // 客户端分页
      const total = list.length
      const start = (current - 1) * size
      const paged = list.slice(start, start + size)

      return { list: paged, total, page: current, size }
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
      apiFn: fetchBmList,
      apiParams: { current: 1, size: 20 },
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

  // ==================== 搜索/重置 ====================
  const handleSearch = () => {
    replaceSearchParams({
      keyword: searchForm.keyword,
      current: 1,
      size: 20
    } as any)
  }

  const handleReset = () => {
    searchForm.keyword = ''
    replaceSearchParams({ keyword: '', current: 1, size: 20 } as any)
  }
</script>

<style lang="scss" scoped>
  .ad-account-bm-page {
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
</style>
