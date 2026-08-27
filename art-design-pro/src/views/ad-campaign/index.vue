<!-- 广告投放监控页 — 只读：广告系列/广告组/广告 三标签 + 近7天统计 -->
<!-- 数据策略：实时拉取 FB API；默认聚合全部授权广告账户，筛选支持多选账户 -->
<template>
  <div class="ad-campaign-page art-full-height">
    <!-- 搜索筛选栏 -->
    <ElCard class="mb-4" shadow="never">
      <ElForm :inline="true" :model="searchForm" class="search-form">
        <ElFormItem :label="$t('menus.adCampaign.selectAccount')">
          <ElSelect
            v-model="searchForm.accountIds"
            :placeholder="$t('menus.adCampaign.selectAccountPlaceholder')"
            style="width: 360px"
            multiple
            filterable
            collapse-tags
            clearable
          >
            <ElOption
              v-for="acc in accounts"
              :key="acc.id"
              :label="`${acc.name}${acc.businessName ? '（' + acc.businessName + '）' : ''}`"
              :value="acc.id"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem>
          <ElButton @click="handleSearch">{{ $t('table.searchBar.search') }}</ElButton>
          <ElButton @click="handleReset">{{ $t('table.searchBar.reset') }}</ElButton>
        </ElFormItem>
        <ElFormItem v-if="searchForm.accountIds.length">
          <ElTag type="info">
            {{ $t('menus.adCampaign.selectedAccounts') }}: {{ searchForm.accountIds.length }}
          </ElTag>
        </ElFormItem>
      </ElForm>
    </ElCard>

    <ElCard class="art-table-card">
      <ElTabs v-model="activeTab" class="ad-tabs">
        <ElTabPane :label="$t('menus.adCampaign.campaignTab')" :name="'campaign'">
          <ArtTableHeader
            v-model:columns="campaignColumnChecks"
            :loading="campaignLoading"
            @refresh="handleSearch"
          />
          <ArtTable
            :loading="campaignLoading"
            :data="campaignData"
            :columns="campaignColumns"
            :pagination="campaignPagination"
            :empty-text="$t('menus.adCampaign.noCampaigns')"
            @pagination:size-change="campaignSizeChange"
            @pagination:current-change="campaignCurrentChange"
          />
        </ElTabPane>

        <!-- 广告组 -->
        <ElTabPane :label="$t('menus.adCampaign.adSetTab')" :name="'adset'">
          <ArtTableHeader
            v-model:columns="adsetColumnChecks"
            :loading="adsetLoading"
            @refresh="handleAdSetRefresh"
          />
          <ArtTable
            :loading="adsetLoading"
            :data="adsetData"
            :columns="adsetColumns"
            :pagination="adsetPagination"
            :empty-text="$t('menus.adCampaign.noAdSets')"
            @pagination:size-change="adsetSizeChange"
            @pagination:current-change="adsetCurrentChange"
          />
        </ElTabPane>

        <!-- 广告 -->
        <ElTabPane :label="$t('menus.adCampaign.adTab')" :name="'ad'">
          <ArtTableHeader
            v-model:columns="adColumnChecks"
            :loading="adLoading"
            @refresh="handleAdRefresh"
          />
          <ArtTable
            :loading="adLoading"
            :data="adData"
            :columns="adColumns"
            :pagination="adPagination"
            :empty-text="$t('menus.adCampaign.noAds')"
            @pagination:size-change="adSizeChange"
            @pagination:current-change="adCurrentChange"
          />
        </ElTabPane>
      </ElTabs>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, watch, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import {
    ElCard,
    ElForm,
    ElFormItem,
    ElSelect,
    ElOption,
    ElButton,
    ElTag,
    ElTabs,
    ElTabPane
  } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchFbCampaigns,
    fetchFbAdSetsByAccount,
    fetchFbAdsByAccount,
    fetchFbAdAccountsDetail,
    type FbCampaign,
    type FbAdAccountDetail
  } from '@/api/facebook'
  import { buildCampaignColumns, buildAdSetColumns, buildAdColumns } from './columns'

  defineOptions({ name: 'AdCampaignMonitor' })

  const { t } = useI18n()

  // ==================== 广告账户下拉（多选） ====================
  const accounts = ref<FbAdAccountDetail[]>([])
  const searchForm = reactive({ accountIds: [] as string[] })

  onMounted(async () => {
    try {
      const res = await fetchFbAdAccountsDetail()
      accounts.value = res.accounts || []
    } catch {
      // 忽略
    }
  })

  // ==================== 数据加载（账户级聚合 + 客户端分页） ====================
  const fetchCampaignPaged = async (params: any) => {
    const current = params?.current || 1
    const size = params?.size || 20
    try {
      const res = await fetchFbCampaigns(params?.accountIds || [])
      const list = res.list || []
      const all = list.sort((a: FbCampaign, b: FbCampaign) =>
        (b.updatedTime || '').localeCompare(a.updatedTime || '')
      )
      return {
        list: all.slice((current - 1) * size, current * size),
        total: all.length,
        page: current,
        size
      }
    } catch {
      return { list: [], total: 0, page: 1, size: 20 }
    }
  }

  // 广告组/广告：账户级聚合（一次调用返回所选账户/全部账户数据，含归属列）
  const fetchAdSetPaged = async (params: any) => {
    const current = params?.current || 1
    const size = params?.size || 20
    try {
      const res = await fetchFbAdSetsByAccount(params?.accountIds || [])
      const list = res.list || []
      return {
        list: list.slice((current - 1) * size, current * size),
        total: list.length,
        page: current,
        size
      }
    } catch {
      return { list: [], total: 0, page: 1, size: 20 }
    }
  }

  const fetchAdPaged = async (params: any) => {
    const current = params?.current || 1
    const size = params?.size || 20
    try {
      const res = await fetchFbAdsByAccount(params?.accountIds || [])
      const list = res.list || []
      return {
        list: list.slice((current - 1) * size, current * size),
        total: list.length,
        page: current,
        size
      }
    } catch {
      return { list: [], total: 0, page: 1, size: 20 }
    }
  }

  // ==================== useTable × 3 ====================
  const {
    columns: campaignColumns,
    data: campaignData,
    loading: campaignLoading,
    pagination: campaignPagination,
    columnChecks: campaignColumnChecks,
    replaceSearchParams: campaignReplace,
    getData: campaignGetData,
    handleSizeChange: campaignSizeChange,
    handleCurrentChange: campaignCurrentChange
  } = useTable({
    core: {
      apiFn: fetchCampaignPaged,
      apiParams: { current: 1, size: 20, accountIds: [] },
      columnsFactory: () => buildCampaignColumns({ t, isAccountDisabled, onViewAdSets })
    }
  })

  const {
    columns: adsetColumns,
    data: adsetData,
    loading: adsetLoading,
    pagination: adsetPagination,
    columnChecks: adsetColumnChecks,
    replaceSearchParams: adsetReplace,
    getData: adsetGetData,
    handleSizeChange: adsetSizeChange,
    handleCurrentChange: adsetCurrentChange
  } = useTable({
    core: {
      apiFn: fetchAdSetPaged,
      apiParams: { current: 1, size: 20, accountIds: [] },
      columnsFactory: () => buildAdSetColumns({ t, isAccountDisabled }),
      immediate: false
    }
  })

  const {
    columns: adColumns,
    data: adData,
    loading: adLoading,
    pagination: adPagination,
    columnChecks: adColumnChecks,
    replaceSearchParams: adReplace,
    getData: adGetData,
    handleSizeChange: adSizeChange,
    handleCurrentChange: adCurrentChange
  } = useTable({
    core: {
      apiFn: fetchAdPaged,
      apiParams: { current: 1, size: 20, accountIds: [] },
      columnsFactory: () => buildAdColumns(t, isAccountDisabled),
      immediate: false
    }
  })

  // ==================== 交互 ====================
  const activeTab = ref<'campaign' | 'adset' | 'ad'>('campaign')

  const handleSearch = () => {
    campaignReplace({ accountIds: searchForm.accountIds, current: 1, size: 20 } as any)
    campaignGetData()
  }

  const handleReset = () => {
    searchForm.accountIds = []
    campaignReplace({ accountIds: [], current: 1, size: 20 } as any)
    campaignGetData()
  }

  // 函数声明（提升）：columnsFactory 在 setup 时立即执行仅读取引用，实际访问的 ref 在点击时已初始化
  function onViewAdSets() {
    activeTab.value = 'adset'
  }

  // 各 tab 表头刷新（广告组/广告为账户级全量，与 watch 逻辑一致）
  const handleAdSetRefresh = () => {
    adsetReplace({ accountIds: searchForm.accountIds, current: 1, size: 20 } as any)
    adsetGetData()
  }

  const handleAdRefresh = () => {
    adReplace({ accountIds: searchForm.accountIds, current: 1, size: 20 } as any)
    adGetData()
  }

  // 按行判定账户是否被封禁（account_status != 1 或 disable_reason > 0）
  // FB API 不返回"账户已停用"状态——按 FB 后台口径由账户状态推断，campaign/adset/ad 全部显示该状态
  function isAccountDisabled(row: any) {
    const acc = accounts.value.find((a) => a.id === row.accountId)
    return !!acc && (acc.accountStatus !== 1 || acc.disableReason > 0)
  }

  // 标签切换：广告组/广告为账户级全量（一次拉取所选/全部账户，含归属列）
  watch(activeTab, (tab) => {
    const accountIds = searchForm.accountIds
    if (tab === 'adset') {
      adsetReplace({ accountIds, current: 1, size: 20 } as any)
      adsetGetData()
    }
    if (tab === 'ad') {
      adReplace({ accountIds, current: 1, size: 20 } as any)
      adGetData()
    }
  })
</script>

<style lang="scss" scoped>
  .ad-campaign-page {
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

  .ad-tabs {
    height: 100%;
    display: flex;
    flex-direction: column;

    :deep(.el-tabs__header) {
      margin-top: 0;
      margin-bottom: 12px;
      flex-shrink: 0;
    }

    :deep(.el-tabs__content) {
      flex: 1;
      min-height: 0;
    }

    :deep(.el-tab-pane) {
      height: 100%;
    }

    // 分页与卡片底保留间隙：margin 不被 useTableHeight 测量吸收（padding 会被吸收导致视觉无变化）
    :deep(.custom-pagination) {
      margin-bottom: 20px;
    }
  }
</style>
