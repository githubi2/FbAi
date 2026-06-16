<!-- Facebook 广告账户管理页面 -->
<template>
  <div class="ad-account-page art-full-height">
    <!-- Facebook 连接状态面板 -->
    <ElCard class="mb-4" shadow="never">
      <div class="fb-connect-panel">
        <div class="fb-connect-info">
          <ElSpace :size="12" alignment="center">
            <ElTag :type="connectionStatus.connected ? 'success' : 'info'" size="large">
              {{ connectionStatus.connected ? $t('menus.adAccount.connected') : $t('menus.adAccount.notConnected') }}
            </ElTag>
            <template v-if="connectionStatus.connected">
              <span class="text-gray-600">
                {{ $t('menus.adAccount.fbUser') }}: {{ connectionStatus.fbUserName }}
              </span>
              <span class="text-gray-400">|</span>
              <span class="text-gray-600">
                {{ $t('menus.adAccount.expiresAt') }}: {{ formatDate(connectionStatus.expiresAt) }}
              </span>
            </template>
          </ElSpace>
        </div>
        <div class="fb-connect-actions">
          <ElSpace :size="8" wrap>
            <ElButton v-if="!connectionStatus.connected" type="primary" :loading="isConnecting" @click="handleConnectFb">
              {{ $t('menus.adAccount.connectFb') }}
            </ElButton>
            <ElButton v-if="connectionStatus.connected" @click="handleRefreshAccounts" v-ripple>
              {{ $t('menus.adAccount.refreshAccounts') }}
            </ElButton>
            <ElButton v-if="connectionStatus.connected" type="danger" plain @click="handleDisconnect">
              {{ $t('menus.adAccount.disconnectFb') }}
            </ElButton>
          </ElSpace>
        </div>
      </div>
    </ElCard>

    <!-- 广告账户表格 -->
    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton v-if="connectionStatus.connected" @click="handleAdd" v-ripple>{{ $t('menus.adAccount.refreshAccounts') }}</ElButton>
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

      <!-- 未连接时的空状态 -->
      <ElEmpty v-if="!connectionStatus.connected && !loading" :description="$t('menus.adAccount.connectFirst')" />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
import { h, ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
import { useTable } from '@/hooks/core/useTable'
import { ElTag, ElMessage, ElMessageBox, ElEmpty } from 'element-plus'
import {
  fetchFbAuthUrl,
  fetchFbConnectionStatus,
  fetchFbAdAccounts,
  fetchFbDisconnect,
  type FbConnectionStatus,
  type FbAdAccount
} from '@/api/facebook'

defineOptions({ name: 'AdAccount' })

const { t } = useI18n()

// ==================== 连接状态 ====================
const connectionStatus = reactive<FbConnectionStatus>({
  connected: false,
  fbUserId: '',
  fbUserName: '',
  expiresAt: '',
  selectedAdAccountId: '',
  scopes: []
})
const isConnecting = ref(false)

// ==================== 状态映射 ====================
const getStatusConfig = (status: number) => {
  switch (status) {
    case 1:
      return { type: 'success' as const, text: '启用' }
    case 2:
      return { type: 'warning' as const, text: '禁用' }
    case 3:
      return { type: 'danger' as const, text: '未结算' }
    case 7:
      return { type: 'info' as const, text: '待审核' }
    case 9:
      return { type: 'warning' as const, text: '不活跃' }
    case 101:
      return { type: 'danger' as const, text: '已关闭' }
    default:
      return { type: 'info' as const, text: '未知' }
  }
}

const formatDate = (val: string) => {
  if (!val) return '—'
  const d = new Date(val)
  if (isNaN(d.getTime())) return val
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// ==================== useTable ====================
const fetchAdAccounts = async (_params: any) => {
  if (!connectionStatus.connected) {
    return { list: [], total: 0, page: 1, size: 0 }
  }
  try {
    const result = await fetchFbAdAccounts()
    const accounts = result.adAccounts || []
    const list = accounts.map((acc: FbAdAccount) => ({
      ...acc,
      id: acc.id,
      accountId: acc.accountId || acc.id,
      name: acc.name,
      accountStatus: acc.accountStatus,
      currency: acc.currency,
      businessName: acc.businessName || ''
    }))
    return { list, total: list.length, page: 1, size: list.length }
  } catch {
    return { list: [], total: 0, page: 1, size: 0 }
  }
}

const {
  columns,
  columnChecks,
  data,
  loading,
  pagination,
  handleSizeChange,
  handleCurrentChange,
  refreshData
} = useTable({
  core: {
    apiFn: fetchAdAccounts,
    apiParams: { current: 1, size: 50 },
    columnsFactory: () => [
      { type: 'index', width: 60, label: '序号' },
      {
        prop: 'name',
        label: '账户名称',
        minWidth: 200,
        formatter: (row: any) => row.name || '—'
      },
      {
        prop: 'accountId',
        label: '账户 ID',
        minWidth: 180,
        formatter: (row: any) => {
          const id = row.accountId || row.id || ''
          return id.replace('act_', '')
        }
      },
      {
        prop: 'businessName',
        label: 'BM 名称',
        minWidth: 160,
        formatter: (row: any) => row.businessName || '—'
      },
      {
        prop: 'accountStatus',
        label: '状态',
        width: 90,
        formatter: (row: any) => {
          const config = getStatusConfig(Number(row.accountStatus))
          return h(ElTag, { type: config.type, size: 'small' }, () => config.text)
        }
      },
      {
        prop: 'currency',
        label: '币种',
        width: 80,
        formatter: (row: any) => row.currency || '—'
      }
    ]
  }
})

// ==================== 操作方法 ====================
const checkConnectionStatus = async () => {
  try {
    const status = await fetchFbConnectionStatus()
    Object.assign(connectionStatus, status)
  } catch {
    connectionStatus.connected = false
  }
}

const handleConnectFb = async () => {
  isConnecting.value = true
  try {
    const { authUrl } = await fetchFbAuthUrl()
    window.open(authUrl, '_blank')
    ElMessage.success('请在打开的页面中完成 Facebook 授权，完成后刷新页面')
  } catch {
    ElMessage.error('获取授权链接失败，请检查 Facebook 应用配置')
  } finally {
    isConnecting.value = false
  }
}

const handleRefreshAccounts = async () => {
  await refreshData()
  ElMessage.success('账户列表已刷新')
}

const handleDisconnect = async () => {
  try {
    await ElMessageBox.confirm(
      t('menus.adAccount.confirmDisconnect'),
      t('common.warning'),
      { type: 'warning' }
    )
    await fetchFbDisconnect()
    connectionStatus.connected = false
    ElMessage.success(t('menus.adAccount.disconnectSuccess'))
    await refreshData()
  } catch {
    // 用户取消
  }
}

const handleAdd = () => {
  if (!connectionStatus.connected) {
    ElMessage.warning(t('menus.adAccount.connectFirst'))
    return
  }
  handleRefreshAccounts()
}

// ==================== 生命周期 ====================
onMounted(async () => {
  await checkConnectionStatus()
  if (connectionStatus.connected) {
    await refreshData()
  }
})

// 监听 URL 参数，处理 OAuth 回调成功
const urlParams = new URLSearchParams(window.location.hash.split('?')[1] || '')
if (urlParams.get('fb_connected') === 'success') {
  ElMessage.success('Facebook 授权成功！')
  // 清除 URL 参数
  window.history.replaceState({}, '', window.location.pathname + window.location.hash.split('?')[0])
  checkConnectionStatus().then(() => {
    if (connectionStatus.connected) {
      refreshData()
    }
  })
}
</script>

<style lang="scss" scoped>
.fb-connect-panel {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}

.fb-connect-info {
  flex: 1;
  min-width: 240px;
}

.fb-connect-actions {
  flex-shrink: 0;
}
</style>
