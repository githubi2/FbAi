<!-- 广告投放监控页 — 只读：广告系列/广告组/广告 三标签 + 近7天统计 -->
<!-- 数据策略：实时拉取 FB API（按广告账户筛选，数据量小） -->
<template>
  <div class="ad-campaign-page art-full-height">
    <!-- 搜索筛选栏 -->
    <ElCard class="mb-4" shadow="never">
      <ElForm :inline="true" :model="searchForm" class="search-form">
        <ElFormItem :label="$t('menus.adCampaign.selectAccount')">
          <ElSelect
            v-model="searchForm.accountId"
            :placeholder="$t('menus.adCampaign.selectAccountPlaceholder')"
            style="width: 280px"
            filterable
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
        <ElFormItem v-if="currentAccountName">
          <ElTag type="info"
            >{{ $t('menus.adCampaign.accountName') }}: {{ currentAccountName }}</ElTag
          >
        </ElFormItem>
      </ElForm>
    </ElCard>

    <ElCard class="art-table-card">
      <ArtTableHeader :loading="campaignLoading" @refresh="handleSearch" />
      <ElTabs v-model="activeTab" class="ad-tabs">
        <!-- 广告系列 -->
        <ElTabPane :label="$t('menus.adCampaign.campaignTab')" :name="'campaign'">
          <ArtTable
            :loading="campaignLoading"
            :data="campaignData"
            :columns="campaignColumns"
            :pagination="campaignPagination"
            @pagination:size-change="campaignSizeChange"
            @pagination:current-change="campaignCurrentChange"
          />
          <ElEmpty
            v-if="!campaignLoading && campaignData.length === 0"
            :description="$t('menus.adCampaign.noCampaigns')"
          />
        </ElTabPane>

        <!-- 广告组 -->
        <ElTabPane :label="$t('menus.adCampaign.adSetTab')" :name="'adset'">
          <ArtTable
            :loading="adsetLoading"
            :data="adsetData"
            :columns="adsetColumns"
            :pagination="adsetPagination"
            @pagination:size-change="adsetSizeChange"
            @pagination:current-change="adsetCurrentChange"
          />
          <ElEmpty
            v-if="!adsetLoading && adsetData.length === 0"
            :description="$t('menus.adCampaign.noAdSets')"
          />
        </ElTabPane>

        <!-- 广告 -->
        <ElTabPane :label="$t('menus.adCampaign.adTab')" :name="'ad'">
          <ArtTable
            :loading="adLoading"
            :data="adData"
            :columns="adColumns"
            :pagination="adPagination"
            @pagination:size-change="adSizeChange"
            @pagination:current-change="adCurrentChange"
          />
          <ElEmpty
            v-if="!adLoading && adData.length === 0"
            :description="$t('menus.adCampaign.noAds')"
          />
        </ElTabPane>
      </ElTabs>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, watch, onMounted } from 'vue'
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
    ElTabPane,
    ElEmpty
  } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchFbCampaigns,
    fetchFbAdSets,
    fetchFbAds,
    fetchFbAdAccountsDetail,
    type FbCampaign,
    type FbAdSet,
    type FbAdAccountDetail
  } from '@/api/facebook'
  import { buildCampaignColumns, buildAdSetColumns, buildAdColumns } from './columns'

  defineOptions({ name: 'AdCampaignMonitor' })

  const { t } = useI18n()

  // ==================== 广告账户下拉 ====================
  const accounts = ref<FbAdAccountDetail[]>([])
  const searchForm = reactive({ accountId: '' })

  const currentAccountName = computed(() => {
    const acc = accounts.value.find((a) => a.id === searchForm.accountId)
    return acc ? acc.name : ''
  })

  onMounted(async () => {
    try {
      const res = await fetchFbAdAccountsDetail()
      accounts.value = res.accounts || []
    } catch {
      // 忽略
    }
  })

  // ==================== 数据加载（客户端分页） ====================
  const fetchCampaignPaged = async (params: any) => {
    const accountId = params?.accountId || ''
    if (!accountId) return { list: [], total: 0, page: 1, size: 20 }
    const current = params?.current || 1
    const size = params?.size || 20
    try {
      const res = await fetchFbCampaigns(accountId)
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

  const fetchAdSetPaged = async (params: any) => {
    const { accountId = '', campaignId = '' } = params || {}
    if (!accountId || !campaignId) return { list: [], total: 0, page: 1, size: 20 }
    const current = params?.current || 1
    const size = params?.size || 20
    try {
      const res = await fetchFbAdSets(campaignId, accountId)
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
    const { accountId = '', adsetId = '' } = params || {}
    if (!accountId || !adsetId) return { list: [], total: 0, page: 1, size: 20 }
    const current = params?.current || 1
    const size = params?.size || 20
    try {
      const res = await fetchFbAds(adsetId, accountId)
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
    replaceSearchParams: campaignReplace,
    getData: campaignGetData,
    handleSizeChange: campaignSizeChange,
    handleCurrentChange: campaignCurrentChange
  } = useTable({
    core: {
      apiFn: fetchCampaignPaged,
      apiParams: { current: 1, size: 20, accountId: '' },
      columnsFactory: () => buildCampaignColumns({ t, isAccountDisabled, onViewAdSets })
    }
  })

  const {
    columns: adsetColumns,
    data: adsetData,
    loading: adsetLoading,
    pagination: adsetPagination,
    replaceSearchParams: adsetReplace,
    getData: adsetGetData,
    handleSizeChange: adsetSizeChange,
    handleCurrentChange: adsetCurrentChange
  } = useTable({
    core: {
      apiFn: fetchAdSetPaged,
      apiParams: { current: 1, size: 20, accountId: '', campaignId: '' },
      columnsFactory: () => buildAdSetColumns({ t, isAccountDisabled, onViewAds })
    }
  })

  const {
    columns: adColumns,
    data: adData,
    loading: adLoading,
    pagination: adPagination,
    replaceSearchParams: adReplace,
    getData: adGetData,
    handleSizeChange: adSizeChange,
    handleCurrentChange: adCurrentChange
  } = useTable({
    core: {
      apiFn: fetchAdPaged,
      apiParams: { current: 1, size: 20, accountId: '', adsetId: '' },
      columnsFactory: () => buildAdColumns(t, isAccountDisabled)
    }
  })

  // ==================== 交互 ====================
  const activeTab = ref<'campaign' | 'adset' | 'ad'>('campaign')
  const activeCampaignId = ref('')
  const activeAdsetId = ref('')

  const handleSearch = () => {
    campaignReplace({ accountId: searchForm.accountId, current: 1, size: 20 } as any)
    campaignGetData()
  }

  const handleReset = () => {
    searchForm.accountId = ''
    campaignReplace({ accountId: '', current: 1, size: 20 } as any)
    campaignGetData()
  }

  // 函数声明（提升）：columnsFactory 在 setup 时立即执行仅读取引用，实际访问的 ref 在点击时已初始化
  function onViewAdSets(row: FbCampaign) {
    activeCampaignId.value = row.id
    activeTab.value = 'adset'
  }

  function onViewAds(row: FbAdSet) {
    activeAdsetId.value = row.id
    activeTab.value = 'ad'
  }

  // 当前所选广告账户是否被封禁（account_status != 1 或 disable_reason > 0）
  // FB API 不返回"账户已停用"状态——按 FB 后台口径由账户状态推断，campaign/adset/ad 全部显示该状态
  function isAccountDisabled() {
    const acc = accounts.value.find((a) => a.id === searchForm.accountId)
    return !!acc && (acc.accountStatus !== 1 || acc.disableReason > 0)
  }

  // 标签切换：自动以第一行作为下钻上下文加载（无需先点操作按钮）
  watch(activeTab, (tab) => {
    const accountId = searchForm.accountId
    if (tab === 'adset') {
      if (!activeCampaignId.value && (campaignData.value as any[]).length) {
        activeCampaignId.value = (campaignData.value as any[])[0].id
      }
      if (activeCampaignId.value && accountId) {
        adsetReplace({ accountId, campaignId: activeCampaignId.value, current: 1, size: 20 } as any)
        adsetGetData()
      }
    }
    if (tab === 'ad') {
      if (!activeAdsetId.value && (adsetData.value as any[]).length) {
        activeAdsetId.value = (adsetData.value as any[])[0].id
      }
      if (activeAdsetId.value && accountId) {
        adReplace({ accountId, adsetId: activeAdsetId.value, current: 1, size: 20 } as any)
        adGetData()
      }
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
    :deep(.el-tabs__header) {
      margin-bottom: 12px;
    }
  }
</style>
