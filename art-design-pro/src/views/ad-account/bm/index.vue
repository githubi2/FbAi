<!-- BM 列表页面 — 展示所有已授权FB账号下的 Business Manager 列表 -->
<!-- 数据策略：GET 缓存直出（毫秒级）→ 后台显式刷新 → 轮询状态 → 静默更新 -->
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
            <ElTag v-if="isRefreshing" type="warning" size="small">后台刷新中…</ElTag>
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

    <!-- 备注编辑弹窗 -->
    <ElDialog v-model="remarkDialogVisible" :title="$t('menus.adAccount.editRemark')" width="420px">
      <ElInput
        v-model="remarkForm.remark"
        type="textarea"
        :rows="3"
        maxlength="255"
        show-word-limit
        :placeholder="$t('menus.adAccount.remarkPlaceholder')"
      />
      <template #footer>
        <ElButton @click="remarkDialogVisible = false">{{ $t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="remarkSaving" @click="handleSaveRemark">
          {{ $t('common.confirm') }}
        </ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { h, ref, reactive, onMounted, onUnmounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useTable } from '@/hooks/core/useTable'
  import { ElTag, ElEmpty, ElTooltip, ElButton, ElIcon, ElMessage } from 'element-plus'
  import { Edit } from '@element-plus/icons-vue'
  import type { FbBmItem } from '@/api/facebook'
  import {
    fetchFbBmList,
    fetchFbRefreshBmList,
    fetchFbUpdateBmRemark,
    fetchRefreshStatus
  } from '@/api/facebook'

  defineOptions({ name: 'AdAccountBmList' })

  const { t } = useI18n()

  // ==================== 刷新状态 ====================
  const isRefreshing = ref(false)
  let pollTimer: ReturnType<typeof setInterval> | null = null

  // 开始轮询刷新状态
  const startPolling = () => {
    if (pollTimer) return
    pollTimer = setInterval(async () => {
      try {
        const status = await fetchRefreshStatus('bm')
        if (!status.isRunning) {
          isRefreshing.value = false
          stopPolling()
          // 刷新完成，静默更新数据（不触发表格 loading，避免二次转圈）
          silentReload()
        }
      } catch {
        // 忽略轮询错误
      }
    }, 2000)
  }

  // 停止轮询
  const stopPolling = () => {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  // 静默重载声明（具体定义在 useTable 之后，轮询回调运行时已完成初始化）
  let silentReload: () => Promise<void> = async () => {}

  // 页面加载流程：
  // ① useTable immediate 自动请求列表 → 后端缓存直出 → 表格立即显示
  // ② 显式触发后台刷新（5 分钟冷却期内为 no-op）
  // ③ 若后台正在刷新 → 轮询状态 → 完成后 silentReload 静默替换数据
  onMounted(async () => {
    try {
      await fetchFbRefreshBmList()
      const status = await fetchRefreshStatus('bm')
      if (status.isRunning) {
        isRefreshing.value = true
        startPolling()
      }
    } catch {
      // 后台刷新失败不影响数据展示
    }
  })

  onUnmounted(() => {
    stopPolling()
  })

  // ==================== 搜索筛选 ====================
  const searchForm = reactive({
    keyword: ''
  })

  // ==================== 格式化 ====================
  const formatDate = (val: string) => {
    if (!val) return '—'
    const d = new Date(val)
    if (isNaN(d.getTime())) return val
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  }

  // 所有者角色映射（business_users.role）
  const formatOwnerRole = (role: string) => {
    if (role === 'ADMIN') return t('menus.adAccount.roleAdmin')
    if (role === 'EMPLOYEE') return t('menus.adAccount.roleEmployee')
    return role || '—'
  }

  // 认证状态映射（verification_status）
  const getVerificationConfig = (
    status: string
  ): { label: string; type: 'success' | 'info' | 'warning' } => {
    switch (status) {
      case 'verified':
        return { label: t('menus.adAccount.verified'), type: 'success' }
      case 'not_verified':
        return { label: t('menus.adAccount.notVerified'), type: 'info' }
      default:
        return { label: status || t('menus.adAccount.verifying'), type: 'warning' }
    }
  }

  // ==================== useTable — 客户端分页+筛选 ====================
  const fetchBmList = async (params: any) => {
    const current = params?.current || 1
    const size = params?.size || 20

    try {
      const result = await fetchFbBmList()
      let list = result.list || []

      // 客户端筛选
      const keyword = (params?.keyword || '').toLowerCase().trim()
      if (keyword) {
        list = list.filter(
          (item: FbBmItem) =>
            (item.name || '').toLowerCase().includes(keyword) ||
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
        { type: 'index', width: 55, label: '#' },
        {
          prop: 'statusLabel',
          label: t('menus.adAccount.columns.status'),
          width: 85,
          formatter: (row: FbBmItem) =>
            h(ElTag, { type: 'success', size: 'small' }, () => row.statusLabel || '—')
        },
        {
          prop: 'name',
          label: t('menus.adAccount.columns.bmName'),
          minWidth: 200,
          formatter: (row: FbBmItem) =>
            h(ElTooltip, { content: `ID: ${row.bmId}`, placement: 'top' }, () =>
              h('span', row.name || '—')
            )
        },
        {
          prop: 'remark',
          label: t('menus.adAccount.columns.remark'),
          minWidth: 140,
          formatter: (row: FbBmItem) =>
            h('div', { style: 'display:flex;align-items:center;gap:6px' }, [
              h('span', { style: row.remark ? '' : 'color:#999' }, row.remark || '—'),
              h(
                ElIcon,
                {
                  style: 'cursor:pointer;color:#409eff;font-size:14px',
                  onClick: () => showEditRemark(row)
                },
                () => h(Edit)
              )
            ])
        },
        {
          prop: 'ownerRole',
          label: t('menus.adAccount.columns.ownerRole'),
          minWidth: 110,
          formatter: (row: FbBmItem) => {
            if (!row.ownerRole) return '—'
            return h(
              'span',
              { style: { color: '#409eff', fontWeight: 500 } },
              formatOwnerRole(row.ownerRole)
            )
          }
        },
        {
          prop: 'verificationStatus',
          label: t('menus.adAccount.columns.verificationStatus'),
          width: 100,
          formatter: (row: FbBmItem) => {
            const config = getVerificationConfig(row.verificationStatus)
            return h(ElTag, { type: config.type, size: 'small' }, () => config.label)
          }
        },
        {
          prop: 'createdTime',
          label: t('menus.adAccount.columns.createdTime'),
          minWidth: 170,
          formatter: (row: FbBmItem) => formatDate(row.createdTime)
        },
        {
          prop: 'adminCount',
          label: t('menus.adAccount.columns.admin'),
          width: 90,
          formatter: (row: FbBmItem) => {
            // adminCount = business_users 真实总数（含灰号等不可枚举用户）
            const names = (row.adminNames || []).join('、')
            let tip = `${t('menus.adAccount.activeAdmins')} ${row.adminCount || 0}`
            if (names) tip += `（${names}）`
            if (row.pendingAdminCount > 0)
              tip += `\n${t('menus.adAccount.pendingAdmins')} ${row.pendingAdminCount}`
            return h(ElTooltip, { content: tip, placement: 'top' }, () =>
              h(ElTag, { type: 'primary', size: 'small' }, () => String(row.adminCount || 0))
            )
          }
        },
        {
          prop: 'partnerCount',
          label: t('menus.adAccount.columns.partner'),
          width: 95,
          formatter: (row: FbBmItem) => String(row.partnerCount || 0)
        },
        {
          prop: 'adAccountCount',
          label: t('menus.adAccount.columns.adAccounts'),
          width: 95,
          formatter: (row: FbBmItem) => String(row.adAccountCount || 0)
        }
      ]
    }
  })

  // 静默重载：后台刷新完成后直接更新数据，不触发表格 loading（避免二次转圈）
  silentReload = async () => {
    try {
      const res = await fetchBmList({
        current: pagination.current,
        size: pagination.size,
        keyword: searchForm.keyword
      })
      data.value = res.list as any
    } catch {
      // 静默失败，保持现有数据
    }
  }

  // ==================== 备注编辑 ====================
  const remarkDialogVisible = ref(false)
  const remarkSaving = ref(false)
  const editingRemarkRow = ref<FbBmItem | null>(null)
  const remarkForm = reactive({ remark: '' })

  const showEditRemark = (row: FbBmItem) => {
    editingRemarkRow.value = row
    remarkForm.remark = row.remark || ''
    remarkDialogVisible.value = true
  }

  const handleSaveRemark = async () => {
    if (!editingRemarkRow.value) return
    remarkSaving.value = true
    const newRemark = remarkForm.remark.trim()
    const targetId = editingRemarkRow.value.bmId
    try {
      await fetchFbUpdateBmRemark(targetId, newRemark)
      ElMessage.success(t('menus.adAccount.remarkUpdateSuccess'))
      remarkDialogVisible.value = false
      // 立即更新当前表格中对应行，无需刷新页面
      editingRemarkRow.value.remark = newRemark
      const row = (data.value as FbBmItem[]).find((r) => r.bmId === targetId)
      if (row) row.remark = newRemark
      // 后台静默同步一次完整数据
      silentReload()
    } catch {
      // 错误提示由请求拦截器统一处理
    } finally {
      remarkSaving.value = false
    }
  }

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
