<!-- 像素管理页面 -->
<template>
  <div class="pixel-manage-page art-full-height">
    <!-- 搜索筛选栏 -->
    <ElCard class="mb-4" shadow="never">
      <ElForm :inline="true" :model="searchForm" class="search-form">
        <ElFormItem :label="$t('menus.pixelManage.searchKeyword')">
          <ElInput
            v-model="searchForm.keyword"
            :placeholder="$t('menus.pixelManage.searchKeyword')"
            clearable
            @clear="handleSearch"
            @keyup.enter="handleSearch"
          />
        </ElFormItem>
        <ElFormItem :label="$t('menus.pixelManage.columns.active')">
          <ElSelect
            v-model="searchForm.active"
            :placeholder="$t('menus.pixelManage.activePlaceholder')"
            clearable
            @change="handleSearch"
          >
            <ElOption :label="$t('menus.pixelManage.activeYes')" :value="1" />
            <ElOption :label="$t('menus.pixelManage.activeNo')" :value="0" />
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
            <ElButton v-ripple @click="showCreateDialog">{{
              $t('menus.pixelManage.createPixel')
            }}</ElButton>
            <ElButton v-ripple @click="refreshData">{{
              $t('menus.pixelManage.refreshPixels')
            }}</ElButton>
            <ElTag v-if="isRefreshing" type="warning" effect="plain">
              {{ $t('menus.pixelManage.refreshing') }}
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

      <ElEmpty
        v-if="!loading && data.length === 0"
        :description="$t('menus.pixelManage.noPixels')"
      />
    </ElCard>

    <!-- 创建像素弹窗 -->
    <ElDialog
      v-model="createDialogVisible"
      :title="$t('menus.pixelManage.createPixel')"
      width="480px"
      destroy-on-close
    >
      <ElForm label-width="100px">
        <ElFormItem :label="$t('menus.pixelManage.selectAdAccount')" required>
          <ElSelect
            v-model="createForm.adAccountId"
            :placeholder="$t('menus.pixelManage.selectAdAccount')"
            :loading="adAccountsLoading"
            filterable
            style="width: 100%"
          >
            <ElOption
              v-for="acc in adAccountOptions"
              :key="acc.id"
              :label="`${acc.name} (${acc.id})`"
              :value="acc.id"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="$t('menus.pixelManage.pixelName')" required>
          <ElInput
            v-model="createForm.name"
            maxlength="256"
            :placeholder="$t('menus.pixelManage.pixelNamePlaceholder')"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="createDialogVisible = false">{{ $t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="createSaving" @click="handleCreatePixel">
          {{ $t('common.confirm') }}
        </ElButton>
      </template>
    </ElDialog>

    <!-- 备注编辑弹窗 -->
    <ElDialog
      v-model="remarkDialogVisible"
      :title="$t('menus.pixelManage.editRemark')"
      width="480px"
      destroy-on-close
    >
      <ElInput
        v-model="remarkForm.remark"
        type="textarea"
        :rows="3"
        maxlength="255"
        show-word-limit
        :placeholder="$t('menus.pixelManage.remarkPlaceholder')"
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
  import { buildPixelColumns } from './columns'
  import type { FbPixelItem } from '@/api/facebook'
  import {
    fetchFbPixels,
    fetchFbRefreshPixels,
    fetchFbUpdatePixelRemark,
    fetchFbCreatePixel,
    fetchFbAdAccountsDetail,
    fetchRefreshStatus
  } from '@/api/facebook'

  defineOptions({ name: 'PixelManageList' })

  const { t } = useI18n()

  // ==================== 刷新状态轮询 ====================
  const isRefreshing = ref(false)
  let pollTimer: ReturnType<typeof setInterval> | null = null

  const startPolling = () => {
    if (pollTimer) return
    pollTimer = setInterval(async () => {
      try {
        const status = await fetchRefreshStatus('pixels')
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
      await fetchFbRefreshPixels()
      const status = await fetchRefreshStatus('pixels')
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
    active: null as number | null
  })

  // ==================== useTable — 客户端分页+筛选 ====================
  const fetchPixels = async (params: any) => {
    const current = params?.current || 1
    const size = params?.size || 20

    try {
      const result = await fetchFbPixels()
      let pixels = result.list || []

      const keyword = (params?.keyword || '').toLowerCase().trim()
      const activeFilter = params?.active

      if (keyword) {
        pixels = pixels.filter(
          (p: FbPixelItem) =>
            (p.name || '').toLowerCase().includes(keyword) ||
            (p.pixelId || '').toLowerCase().includes(keyword) ||
            (p.fbOwnerName || '').toLowerCase().includes(keyword)
        )
      }

      if (activeFilter != null && activeFilter !== '') {
        pixels = pixels.filter((p: FbPixelItem) =>
          Number(activeFilter) === 1 ? !!p.lastFiredTime && !p.isUnavailable : !p.lastFiredTime
        )
      }

      const total = pixels.length
      const start = (current - 1) * size
      const list = pixels.slice(start, start + size)

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
      apiFn: fetchPixels,
      apiParams: { current: 1, size: 20 },
      columnsFactory: () =>
        buildPixelColumns({
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
      const res = await fetchPixels({
        current: pagination.current,
        size: pagination.size,
        keyword: searchForm.keyword,
        active: searchForm.active
      })
      data.value = res.list as any
    } catch {
      // 静默失败，保持现有数据
    }
  }

  // ==================== 创建像素 ====================
  const createDialogVisible = ref(false)
  const createSaving = ref(false)
  const adAccountsLoading = ref(false)
  const adAccountOptions = ref<Array<{ id: string; name: string }>>([])
  const createForm = reactive({ adAccountId: '', name: '' })

  const showCreateDialog = async () => {
    createForm.adAccountId = ''
    createForm.name = ''
    createDialogVisible.value = true
    adAccountsLoading.value = true
    try {
      const res = await fetchFbAdAccountsDetail()
      adAccountOptions.value = (res.accounts || []).map((a: any) => ({
        id: a.id || `act_${a.accountId}`,
        name: a.name || a.accountId
      }))
    } catch {
      // 错误提示由请求拦截器统一处理
    } finally {
      adAccountsLoading.value = false
    }
  }

  const handleCreatePixel = async () => {
    if (!createForm.adAccountId) {
      ElMessage.warning(t('menus.pixelManage.selectAdAccount'))
      return
    }
    if (!createForm.name.trim()) {
      ElMessage.warning(t('menus.pixelManage.pixelNamePlaceholder'))
      return
    }
    createSaving.value = true
    try {
      await fetchFbCreatePixel(createForm.adAccountId, createForm.name.trim())
      ElMessage.success(t('menus.pixelManage.createSuccess'))
      createDialogVisible.value = false
      // 创建成功后触发后台刷新并轮询
      await fetchFbRefreshPixels()
      isRefreshing.value = true
      startPolling()
    } catch {
      // 错误提示由请求拦截器统一处理
    } finally {
      createSaving.value = false
    }
  }

  // ==================== 备注编辑 ====================
  const remarkDialogVisible = ref(false)
  const remarkSaving = ref(false)
  const editingRemarkRow = ref<FbPixelItem | null>(null)
  const remarkForm = reactive({ remark: '' })

  const showEditRemark = (row: FbPixelItem) => {
    editingRemarkRow.value = row
    remarkForm.remark = row.remark || ''
    remarkDialogVisible.value = true
  }

  const handleSaveRemark = async () => {
    if (!editingRemarkRow.value) return
    remarkSaving.value = true
    const newRemark = remarkForm.remark.trim()
    const targetId = editingRemarkRow.value.pixelId
    try {
      await fetchFbUpdatePixelRemark(targetId, newRemark)
      ElMessage.success(t('menus.pixelManage.remarkUpdateSuccess'))
      remarkDialogVisible.value = false
      editingRemarkRow.value.remark = newRemark
      const row = (data.value as FbPixelItem[]).find((r) => r.pixelId === targetId)
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
      active: searchForm.active,
      current: 1,
      size: 20
    } as any)
  }

  const handleReset = () => {
    searchForm.keyword = ''
    searchForm.active = null
    replaceSearchParams({ keyword: '', active: null, current: 1, size: 20 } as any)
  }
</script>

<style lang="scss" scoped>
  .pixel-manage-page {
    .search-form {
      margin-bottom: -18px;
    }
  }
</style>
