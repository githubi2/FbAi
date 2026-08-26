<!-- 广告账户管理页面 -->
<template>
  <div class="ad-account-manage-page art-full-height">
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
        <ElFormItem :label="$t('menus.adAccount.filterStatus')">
          <ElSelect
            v-model="searchForm.status"
            :placeholder="$t('menus.adAccount.statusPlaceholder')"
            clearable
            @change="handleSearch"
          >
            <ElOption :label="$t('menus.adAccount.statusActive')" :value="1" />
            <ElOption :label="$t('menus.adAccount.statusDisabled')" :value="2" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="$t('menus.adAccount.filterAccountType')">
          <ElSelect
            v-model="searchForm.accountType"
            :placeholder="$t('menus.adAccount.filterAccountType')"
            clearable
            @change="handleSearch"
          >
            <ElOption label="企业" value="企业" />
            <ElOption label="个人" value="个人" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem>
          <ElButton @click="handleSearch">{{ $t('table.searchBar.search') }}</ElButton>
          <ElButton @click="handleReset">{{ $t('table.searchBar.reset') }}</ElButton>
        </ElFormItem>
      </ElForm>

      <!-- 批量操作按钮组 -->
      <div class="batch-actions">
        <ElButton v-ripple @click="handleBatchAction('addAuth')">{{
          $t('menus.adAccount.addAuth')
        }}</ElButton>
        <ElButton v-ripple @click="handleBatchAction('deleteAuth')">{{
          $t('menus.adAccount.deleteAuth')
        }}</ElButton>
        <ElButton v-ripple @click="handleBatchAction('addToBM')">{{
          $t('menus.adAccount.addToBM')
        }}</ElButton>
        <ElButton v-ripple @click="handleBatchAction('hideAdmin')">{{
          $t('menus.adAccount.hideAdmin')
        }}</ElButton>
        <ElButton v-ripple @click="handleBatchAction('accountPush')">{{
          $t('menus.adAccount.accountPush')
        }}</ElButton>
      </div>
    </ElCard>

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton @click="refreshData" v-ripple>{{
              $t('menus.adAccount.refreshAccounts')
            }}</ElButton>
            <ElTag v-if="isRefreshing" type="warning" effect="plain">
              {{ $t('menus.adAccount.refreshing') || '数据更新中...' }}
            </ElTag>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @selection-change="handleSelectionChange"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />

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
      <ArtTable :loading="paymentLoading" :data="pagedPaymentRecords" :columns="paymentColumns" />
      <ElEmpty v-if="!paymentLoading && paymentRecords.length === 0" description="暂无支付记录" />
      <div v-if="paymentRecords.length > 0" class="payment-pagination">
        <ElPagination
          v-model:current-page="paymentPage"
          v-model:page-size="paymentPageSize"
          :total="paymentRecords.length"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          background
        />
      </div>
    </ElDialog>

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

    <!-- 管理员详情弹窗 -->
    <ElDialog v-model="adminDialogVisible" :title="adminDialogTitle" width="550px" destroy-on-close>
      <div class="admin-dialog-content">
        <!-- 步骤1：选择要删除的管理员 -->
        <div class="admin-step">
          <div class="admin-step-head">
            <span class="admin-step-num">1</span>
            <span class="admin-step-label">
              {{ $t('menus.adAccount.adminDialogStep1') }}
            </span>
          </div>
          <div class="admin-step-body">
            <ElCheckboxGroup v-model="selectedAdmins" class="admin-checklist">
              <ElCheckbox
                v-for="name in curOtherAdminNames"
                :key="name"
                :label="name"
                :value="name"
              >
                {{ name }}
              </ElCheckbox>
            </ElCheckboxGroup>
            <ElEmpty
              v-if="curOtherAdminNames.length === 0"
              :description="$t('menus.adAccount.adminDialogNoOther')"
              :image-size="60"
            />
          </div>
        </div>
        <!-- 步骤2：执行时间间隔 -->
        <div class="admin-step">
          <div class="admin-step-head">
            <span class="admin-step-num">2</span>
            <ElCheckbox v-model="useDefaultInterval">
              {{ $t('menus.adAccount.adminDialogStep2') }}
            </ElCheckbox>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <ElButton @click="adminDialogVisible = false">{{ $t('common.cancel') }}</ElButton>
          <ElButton
            type="primary"
            :disabled="selectedAdmins.length === 0"
            @click="handleAdminDelete"
          >
            {{ $t('menus.adAccount.adminDialogConfirm') }}
          </ElButton>
        </div>
      </template>
    </ElDialog>

    <!-- 增加授权弹窗 -->
    <AddAuthDialog v-model="addAuthDialogVisible" :selected-ad-accounts="selectedRows" />

    <!-- 删除授权弹窗 -->
    <DeleteAuthDialog v-model="deleteAuthDialogVisible" :selected-ad-accounts="selectedRows" />

    <!-- 添加到BM弹窗 -->
    <AddToBmDialog v-model="addToBmDialogVisible" :selected-ad-accounts="selectedRows" />
  </div>
</template>

<script setup lang="ts">
  import { h, ref, reactive, computed, onMounted, onUnmounted } from 'vue'
  import AddAuthDialog from './components/AddAuthDialog.vue'
  import DeleteAuthDialog from './components/DeleteAuthDialog.vue'
  import AddToBmDialog from './components/AddToBmDialog.vue'
  import { useI18n } from 'vue-i18n'
  import { useTable } from '@/hooks/core/useTable'
  import {
    ElTag,
    ElEmpty,
    ElTooltip,
    ElButton,
    ElDialog,
    ElMessage,
    ElCheckbox,
    ElCheckboxGroup,
    ElIcon,
    ElPagination
  } from 'element-plus'
  import { Edit } from '@element-plus/icons-vue'
  import type { FbAdAccountDetail, FbPaymentRecord } from '@/api/facebook'
  import {
    fetchFbAdAccountsDetail,
    fetchFbPaymentHistory,
    fetchFbRefreshAdAccounts,
    fetchFbUpdateAdAccountRemark,
    fetchRefreshStatus
  } from '@/api/facebook'

  defineOptions({ name: 'AdAccountManage' })

  const { t } = useI18n()

  // ==================== 刷新状态 ====================
  const isRefreshing = ref(false)
  let pollTimer: ReturnType<typeof setInterval> | null = null

  // 开始轮询刷新状态
  const startPolling = () => {
    if (pollTimer) return
    pollTimer = setInterval(async () => {
      try {
        const status = await fetchRefreshStatus('ad_accounts')
        if (!status.isRunning) {
          isRefreshing.value = false
          stopPolling()
          // 刷新完成，静默更新数据（不触发表格 loading，避免二次转圈）
          silentReload()
        }
      } catch {
        // 忽略轮询错误
      }
    }, 2000) // 每2秒检查一次
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

  // 页面加载流程（单向、职责清晰）：
  // ① useTable immediate 自动请求详情 → 后端缓存直出 → 表格立即显示
  // ② 显式触发后台刷新（5 分钟冷却期内为 no-op）
  // ③ 若后台正在刷新 → 轮询状态 → 完成后 silentReload 静默替换数据
  onMounted(async () => {
    try {
      await fetchFbRefreshAdAccounts()
      const status = await fetchRefreshStatus('ad_accounts')
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

  // ==================== 多选 ====================
  const selectedRows = ref<FbAdAccountDetail[]>([])

  const handleSelectionChange = (selection: FbAdAccountDetail[]) => {
    selectedRows.value = selection
  }

  // ==================== 批量操作（占位） ====================
  const handleBatchAction = (action: string) => {
    if (selectedRows.value.length === 0) {
      ElMessage.warning(t('menus.adAccount.selectRowsFirst'))
      return
    }
    // 实现各批量操作功能
    switch (action) {
      case 'addAuth':
        addAuthDialogVisible.value = true
        break
      case 'deleteAuth':
        deleteAuthDialogVisible.value = true
        break
      case 'addToBM':
        addToBmDialogVisible.value = true
        break
      default:
        console.log('Batch action:', action, selectedRows.value)
        break
    }
  }

  // ==================== 增加授权弹窗 ====================
  const addAuthDialogVisible = ref(false)
  const deleteAuthDialogVisible = ref(false)
  const addToBmDialogVisible = ref(false)

  // ==================== 搜索筛选 ====================
  const searchForm = reactive({
    keyword: '',
    status: null as number | null,
    accountType: ''
  })

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

  // ==================== useTable — 客户端分页+筛选 ====================
  const fetchDetail = async (params: any) => {
    const current = params?.current || 1
    const size = params?.size || 20

    try {
      const result = await fetchFbAdAccountsDetail()
      let accounts = result.accounts || []

      // 客户端筛选
      const keyword = (params?.keyword || '').toLowerCase().trim()
      const statusFilter = params?.status
      const accountTypeFilter = params?.accountType || ''

      if (keyword) {
        accounts = accounts.filter(
          (a: FbAdAccountDetail) =>
            (a.name || '').toLowerCase().includes(keyword) ||
            (a.accountId || a.id || '').toLowerCase().includes(keyword) ||
            (a.businessName || '').toLowerCase().includes(keyword)
        )
      }

      if (statusFilter != null && statusFilter !== '') {
        accounts = accounts.filter(
          (a: FbAdAccountDetail) => a.accountStatus === Number(statusFilter)
        )
      }

      if (accountTypeFilter) {
        if (accountTypeFilter === '企业') {
          accounts = accounts.filter((a: FbAdAccountDetail) => !!a.businessName)
        } else if (accountTypeFilter === '个人') {
          accounts = accounts.filter((a: FbAdAccountDetail) => !a.businessName)
        }
      }

      // 客户端分页
      const total = accounts.length
      const start = (current - 1) * size
      const list = accounts.slice(start, start + size)

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
      apiFn: fetchDetail,
      apiParams: { current: 1, size: 20 },
      columnsFactory: () => [
        { type: 'selection', width: 55 },
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
          prop: 'remark',
          label: t('menus.adAccount.columns.remark'),
          minWidth: 140,
          formatter: (row: FbAdAccountDetail) =>
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
          prop: 'accountType',
          label: t('menus.adAccount.columns.accountCategory'),
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
          minWidth: 110,
          formatter: (row: FbAdAccountDetail) => {
            const total = row.hiddenAdmins + (row.adminName ? 1 : 0)
            return h(
              ElTag,
              {
                type: 'primary',
                size: 'small',
                style: { cursor: 'pointer' },
                onClick: () => showAdminDetail(row, 'admin')
              },
              () => String(total)
            )
          }
        },
        {
          prop: 'hiddenAdmins',
          label: t('menus.adAccount.columns.hiddenAdmin'),
          minWidth: 110,
          formatter: (row: FbAdAccountDetail) => {
            const count = row.hiddenAdmins || 0
            return h(
              ElTag,
              {
                type: 'primary',
                size: 'small',
                style: { cursor: 'pointer' },
                onClick: () => showAdminDetail(row, 'hidden')
              },
              () => String(count)
            )
          }
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
          label: t('menus.adAccount.columns.spendCap'),
          minWidth: 120,
          formatter: (row: FbAdAccountDetail) => {
            if (!row.spendCap)
              return h('span', { style: { color: '#999' } }, t('menus.adAccount.unlimited'))
            return formatCurrency(row.spendCap, row.currency)
          }
        },
        {
          prop: 'spentAmount',
          label: t('menus.adAccount.columns.spentAmount'),
          minWidth: 120,
          formatter: (row: FbAdAccountDetail) => formatCurrency(row.amountSpent, row.currency)
        },
        {
          prop: 'isPrepay',
          label: t('menus.adAccount.columns.accountType'),
          minWidth: 100,
          formatter: (row: FbAdAccountDetail) =>
            row.isPrepay ? t('menus.adAccount.prepay') : t('menus.adAccount.postpay')
        },
        {
          prop: 'ownerRole',
          label: t('menus.adAccount.columns.ownerRole'),
          minWidth: 110,
          formatter: (row: FbAdAccountDetail) => {
            if (!row.ownerRole) return '—'
            return h('span', { style: { color: '#409eff', fontWeight: 500 } }, row.ownerRole)
          }
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

  // 静默重载：后台刷新完成后直接更新数据，不触发表格 loading（避免二次转圈）
  silentReload = async () => {
    try {
      const res = await fetchDetail({
        current: pagination.current,
        size: pagination.size,
        keyword: searchForm.keyword,
        status: searchForm.status,
        accountType: searchForm.accountType
      })
      data.value = res.list as any
    } catch {
      // 静默失败，保持现有数据
    }
  }

  // ==================== 备注编辑 ====================
  const remarkDialogVisible = ref(false)
  const remarkSaving = ref(false)
  const editingRemarkRow = ref<FbAdAccountDetail | null>(null)
  const remarkForm = reactive({ remark: '' })

  const showEditRemark = (row: FbAdAccountDetail) => {
    editingRemarkRow.value = row
    remarkForm.remark = row.remark || ''
    remarkDialogVisible.value = true
  }

  const handleSaveRemark = async () => {
    if (!editingRemarkRow.value) return
    remarkSaving.value = true
    const newRemark = remarkForm.remark.trim()
    const targetId = editingRemarkRow.value.id
    try {
      await fetchFbUpdateAdAccountRemark(targetId, newRemark)
      ElMessage.success(t('menus.adAccount.remarkUpdateSuccess'))
      remarkDialogVisible.value = false
      // 立即更新当前表格中对应行，无需刷新页面
      editingRemarkRow.value.remark = newRemark
      const row = (data.value as FbAdAccountDetail[]).find((r) => r.id === targetId)
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
      status: searchForm.status,
      accountType: searchForm.accountType,
      current: 1,
      size: 20
    } as any)
  }

  const handleReset = () => {
    searchForm.keyword = ''
    searchForm.status = null
    searchForm.accountType = ''
    replaceSearchParams({ keyword: '', status: null, accountType: '', current: 1, size: 20 } as any)
  }

  // ==================== 支付记录弹窗 ====================
  const paymentDialogVisible = ref(false)
  const paymentDialogTitle = ref('')
  const paymentLoading = ref(false)
  const paymentRecords = ref<FbPaymentRecord[]>([])
  const curPaymentAccount = ref<FbAdAccountDetail | null>(null)
  // 分页（客户端）
  const paymentPage = ref(1)
  const paymentPageSize = ref(10)
  const pagedPaymentRecords = computed(() => {
    const start = (paymentPage.value - 1) * paymentPageSize.value
    return paymentRecords.value.slice(start, start + paymentPageSize.value)
  })

  const formatPayDateTime = (val: string) => {
    if (!val) return '—'
    const d = new Date(val)
    if (isNaN(d.getTime())) return val
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
  }

  const paymentColumns = [
    {
      prop: 'id',
      label: '交易编号',
      minWidth: 260,
      formatter: (row: FbPaymentRecord) => row.id || '—'
    },
    {
      prop: 'time',
      label: '日期',
      minWidth: 150,
      formatter: (row: FbPaymentRecord) => formatPayDateTime(row.time)
    },
    {
      prop: 'amount',
      label: '金额',
      minWidth: 120,
      formatter: (row: FbPaymentRecord) => formatCurrency(row.amount, row.currency)
    },
    {
      prop: 'paymentMethod',
      label: '支付方式',
      minWidth: 130,
      formatter: (row: FbPaymentRecord) => row.paymentMethod || '—'
    },
    {
      prop: 'status',
      label: '支付状态',
      width: 100,
      formatter: (row: FbPaymentRecord) => {
        const paid = ['已支付', '成功']
        const danger = ['失败', '退单']
        const warning = ['处理中']
        const type = paid.includes(row.status)
          ? ('success' as const)
          : danger.includes(row.status)
            ? ('danger' as const)
            : warning.includes(row.status)
              ? ('warning' as const)
              : ('info' as const)
        return h(ElTag, { type, size: 'small' }, () => row.status || '—')
      }
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
      paymentPage.value = 1
    } catch (e: any) {
      paymentRecords.value = []
      ElMessage.warning(e?.data?.msg || '支付记录暂不可用（需要 Facebook ads_read 权限）')
    } finally {
      paymentLoading.value = false
    }
  }
  // ==================== 管理员详情弹窗 ====================
  const adminDialogVisible = ref(false)
  const adminDialogTitle = ref('')
  const curAdminAccount = ref<FbAdAccountDetail | null>(null)
  const curAdminType = ref<'admin' | 'hidden'>('admin')
  const curOtherAdminNames = ref<string[]>([])
  const selectedAdmins = ref<string[]>([])
  const useDefaultInterval = ref(true)

  const showAdminDetail = (row: FbAdAccountDetail, type: 'admin' | 'hidden') => {
    // 标签值为 0 时直接提示，不弹窗
    if (type === 'hidden') {
      const count = row.hiddenAdmins || 0
      if (count === 0) {
        ElMessage.info(t('menus.adAccount.noHiddenAdmin'))
        return
      }
    } else {
      const total = row.hiddenAdmins + (row.adminName ? 1 : 0)
      if (total === 0) {
        ElMessage.info(t('menus.adAccount.noAdmin'))
        return
      }
    }
    curAdminAccount.value = row
    curAdminType.value = type
    curOtherAdminNames.value = row.otherAdminNames || []
    selectedAdmins.value = []
    useDefaultInterval.value = true
    adminDialogTitle.value = `${row.name || row.accountId || '—'} — ${t('menus.adAccount.adminDialogTitle')}`
    adminDialogVisible.value = true
  }

  const handleAdminDelete = () => {
    // TODO: 调用后端 API 删除选中的管理员
    ElMessage.success(
      t('menus.adAccount.adminDeleteSuccess', { count: selectedAdmins.value.length })
    )
    adminDialogVisible.value = false
  }
</script>

<style lang="scss" scoped>
  .ad-account-manage-page {
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

  .batch-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid var(--el-border-color-lighter);
  }

  .admin-dialog-content {
    min-height: 120px;

    .admin-step {
      position: relative;
      padding-left: 36px;

      // 连接线：圆圈下方到下一步骤
      &::before {
        content: '';
        position: absolute;
        left: 11px;
        top: 28px;
        bottom: 0;
        width: 2px;
        background-color: var(--el-border-color);
      }

      &:last-child::before {
        display: none;
      }

      &:not(:last-child) {
        padding-bottom: 24px;
      }
    }

    .admin-step-head {
      display: flex;
      align-items: center;
      gap: 12px;
      min-height: 24px;
      margin-bottom: 8px;

      :deep(.el-checkbox) {
        height: 24px;
        line-height: 24px;
      }
    }

    .admin-step-num {
      position: absolute;
      left: 0;
      top: 0;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: 24px;
      height: 24px;
      border-radius: 50%;
      background-color: var(--el-color-primary-light-5);
      color: var(--el-color-primary);
      font-size: 13px;
      font-weight: 600;
      line-height: 1;
      flex-shrink: 0;
    }

    .admin-step-label {
      font-size: 14px;
      font-weight: 500;
      color: var(--el-text-color-primary);
      line-height: 24px;
    }

    .admin-step-body {
      min-width: 0;
    }

    .admin-checklist {
      display: flex;
      flex-direction: column;
      gap: 6px;
      padding-left: 4px;

      :deep(.el-checkbox) {
        margin-right: 0;
      }
    }

    .admin-interval-check {
      margin-top: 4px;
    }
  }

  .payment-pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 12px;
  }
</style>
