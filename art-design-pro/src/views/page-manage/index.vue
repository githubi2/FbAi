<!-- 主页管理页面 -->
<template>
  <div class="page-manage-page art-full-height">
    <!-- 搜索筛选栏 -->
    <ElCard class="mb-4" shadow="never">
      <ElForm :inline="true" :model="searchForm" class="search-form">
        <ElFormItem :label="$t('menus.pageManage.searchKeyword')">
          <ElInput
            v-model="searchForm.keyword"
            :placeholder="$t('menus.pageManage.searchKeyword')"
            clearable
            @clear="handleSearch"
            @keyup.enter="handleSearch"
          />
        </ElFormItem>
        <ElFormItem :label="$t('menus.pageManage.columns.publishStatus')">
          <ElSelect
            v-model="searchForm.isPublished"
            :placeholder="$t('menus.pageManage.statusPlaceholder')"
            clearable
            @change="handleSearch"
          >
            <ElOption :label="$t('menus.pageManage.published')" :value="1" />
            <ElOption :label="$t('menus.pageManage.unpublished')" :value="0" />
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
            <ElButton v-ripple @click="refreshData">{{
              $t('menus.pageManage.refreshPages')
            }}</ElButton>
            <ElTag v-if="isRefreshing" type="warning" effect="plain">
              {{ $t('menus.pageManage.refreshing') }}
            </ElTag>
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

      <ElEmpty v-if="!loading && data.length === 0" :description="$t('menus.pageManage.noPages')" />
    </ElCard>

    <!-- 备注编辑弹窗 -->
    <ElDialog
      v-model="remarkDialogVisible"
      :title="$t('menus.pageManage.editRemark')"
      width="480px"
      destroy-on-close
    >
      <ElInput
        v-model="remarkForm.remark"
        type="textarea"
        :rows="3"
        maxlength="255"
        show-word-limit
        :placeholder="$t('menus.pageManage.remarkPlaceholder')"
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
  import { ref, reactive, onMounted, onUnmounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useTable } from '@/hooks/core/useTable'
  import { ElMessage } from 'element-plus'
  import { buildPageColumns } from './columns'
  import type { FbPageItem } from '@/api/facebook'
  import {
    fetchFbPages,
    fetchFbRefreshPages,
    fetchFbUpdatePageRemark,
    fetchRefreshStatus
  } from '@/api/facebook'

  defineOptions({ name: 'PageManageList' })

  const { t } = useI18n()

  // ==================== 刷新状态轮询 ====================
  const isRefreshing = ref(false)
  let pollTimer: ReturnType<typeof setInterval> | null = null

  const startPolling = () => {
    if (pollTimer) return
    pollTimer = setInterval(async () => {
      try {
        const status = await fetchRefreshStatus('pages')
        if (!status.isRunning) {
          isRefreshing.value = false
          stopPolling()
          silentReload()
        }
      } catch {
        // 忽略轮询错误
      }
    }, 2000)
  }

  const stopPolling = () => {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  // 静默重载声明（具体定义在 useTable 之后）
  let silentReload: () => Promise<void> = async () => {}

  onMounted(async () => {
    try {
      await fetchFbRefreshPages()
      const status = await fetchRefreshStatus('pages')
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
    keyword: '',
    isPublished: null as number | null
  })

  // ==================== useTable — 客户端分页+筛选 ====================
  const fetchPages = async (params: any) => {
    const current = params?.current || 1
    const size = params?.size || 20

    try {
      const result = await fetchFbPages()
      let pages = result.list || []

      const keyword = (params?.keyword || '').toLowerCase().trim()
      const publishFilter = params?.isPublished

      if (keyword) {
        pages = pages.filter(
          (p: FbPageItem) =>
            (p.name || '').toLowerCase().includes(keyword) ||
            (p.pageId || '').toLowerCase().includes(keyword) ||
            (p.fbOwnerName || '').toLowerCase().includes(keyword)
        )
      }

      if (publishFilter != null && publishFilter !== '') {
        pages = pages.filter((p: FbPageItem) => p.isPublished === Number(publishFilter))
      }

      const total = pages.length
      const start = (current - 1) * size
      const list = pages.slice(start, start + size)

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
      apiFn: fetchPages,
      apiParams: { current: 1, size: 20 },
      columnsFactory: () =>
        buildPageColumns({
          t,
          // 惰性包装：columnsFactory 在 setup 阶段就会被立即调用，
          // 直接引用下方声明的 showEditRemark 会触发 TDZ 错误
          onEditRemark: (row) => showEditRemark(row)
        })
    }
  })

  // 静默重载：后台刷新完成后直接更新数据，不触发表格 loading
  silentReload = async () => {
    try {
      const res = await fetchPages({
        current: pagination.current,
        size: pagination.size,
        keyword: searchForm.keyword,
        isPublished: searchForm.isPublished
      })
      data.value = res.list as any
    } catch {
      // 静默失败，保持现有数据
    }
  }

  // ==================== 备注编辑 ====================
  const remarkDialogVisible = ref(false)
  const remarkSaving = ref(false)
  const editingRemarkRow = ref<FbPageItem | null>(null)
  const remarkForm = reactive({ remark: '' })

  const showEditRemark = (row: FbPageItem) => {
    editingRemarkRow.value = row
    remarkForm.remark = row.remark || ''
    remarkDialogVisible.value = true
  }

  const handleSaveRemark = async () => {
    if (!editingRemarkRow.value) return
    remarkSaving.value = true
    const newRemark = remarkForm.remark.trim()
    const targetId = editingRemarkRow.value.pageId
    try {
      await fetchFbUpdatePageRemark(targetId, newRemark)
      ElMessage.success(t('menus.pageManage.remarkUpdateSuccess'))
      remarkDialogVisible.value = false
      editingRemarkRow.value.remark = newRemark
      const row = (data.value as FbPageItem[]).find((r) => r.pageId === targetId)
      if (row) row.remark = newRemark
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
      isPublished: searchForm.isPublished,
      current: 1,
      size: 20
    } as any)
  }

  const handleReset = () => {
    searchForm.keyword = ''
    searchForm.isPublished = null
    replaceSearchParams({ keyword: '', isPublished: null, current: 1, size: 20 } as any)
  }
</script>

<style lang="scss" scoped>
  .page-manage-page {
    .search-form {
      margin-bottom: -18px;
    }
  }
</style>
