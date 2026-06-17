<!-- Facebook 账号列表页面（多账号改造） -->
<template>
  <div class="ad-account-page art-full-height">
    <!-- 顶部状态栏 -->
    <ElCard class="mb-4" shadow="never">
      <div class="fb-connect-panel">
        <div class="fb-connect-info">
          <ElSpace :size="12" alignment="center">
            <span class="text-base text-gray-700">
              {{ $t('menus.adAccount.totalAccounts', { count: totalAccounts }) }}
            </span>
          </ElSpace>
        </div>
        <div class="fb-connect-actions">
          <ElButton type="primary" :loading="isConnecting" @click="handleConnectFb">
            {{ $t('menus.adAccount.connectFb') }}
          </ElButton>
        </div>
      </div>
    </ElCard>

    <!-- 账号表格 -->
    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton @click="handleConnectFb" v-ripple>
              {{ $t('menus.adAccount.connectFb') }}
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
      />

      <!-- 空状态 -->
      <ElEmpty
        v-if="!loading && totalAccounts === 0"
        :description="$t('menus.adAccount.noAccounts')"
      />
    </ElCard>

    <!-- 授权链接弹窗 -->
    <ElDialog
      v-model="showAuthDialog"
      :title="$t('menus.adAccount.authLinkTitle')"
      width="540px"
      :close-on-click-modal="false"
      @close="stopPolling"
    >
      <div class="auth-dialog-body">
        <p class="auth-dialog-tip">{{ $t('menus.adAccount.authLinkTip') }}</p>
        <!-- 短链接 -->
        <div class="short-link-box">
          <label class="link-label">{{ $t('menus.adAccount.shortLinkLabel') }}</label>
          <div class="auth-link-box">
            <ElInput v-model="shortUrl" readonly :placeholder="$t('menus.adAccount.generatingLink')" />
            <ElButton type="primary" class="ml-2" @click="copyShortUrl">
              {{ copySuccess ? $t('menus.adAccount.copied') : $t('menus.adAccount.copyLink') }}
            </ElButton>
          </div>
        </div>
        <!-- 完整链接 -->
        <div v-if="fullAuthUrl" class="full-link-box mt-3">
          <label class="link-label">{{ $t('menus.adAccount.fullLinkLabel') }}</label>
          <div class="auth-link-box">
            <ElInput v-model="fullAuthUrl" readonly size="small" />
            <ElButton size="small" class="ml-2" @click="copyFullUrl">
              {{ $t('menus.adAccount.copyLink') }}
            </ElButton>
          </div>
        </div>
        <div class="auth-dialog-actions mt-4">
          <ElButton @click="openAuthUrl">
            {{ $t('menus.adAccount.openInBrowser') }}
          </ElButton>
          <span class="text-gray-400 text-sm ml-2">
            {{ $t('menus.adAccount.orCopyToOther') }}
          </span>
        </div>
        <!-- 轮询状态 -->
        <div v-if="isPolling" class="polling-status mt-4">
          <ElAlert type="info" :closable="false" show-icon>
            <template #title>
              <ElIcon class="is-loading mr-1"><Loading /></ElIcon>
              {{ $t('menus.adAccount.waitingAuth') }}
            </template>
          </ElAlert>
        </div>
      </div>
      <template #footer>
        <ElButton @click="stopPollingAndClose">{{ $t('menus.adAccount.cancelAuth') }}</ElButton>
      </template>
    </ElDialog>

    <!-- 编辑备注弹窗 -->
    <ElDialog
      v-model="showLabelDialog"
      :title="$t('menus.adAccount.editLabel')"
      width="400px"
    >
      <ElForm :model="labelForm" label-position="top">
        <ElFormItem :label="$t('menus.adAccount.editLabel')">
          <ElInput
            v-model="labelForm.label"
            maxlength="64"
            show-word-limit
            :placeholder="$t('menus.adAccount.labelPlaceholder')"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="showLabelDialog = false">{{ $t('common.cancel') }}</ElButton>
        <ElButton type="primary" @click="handleSaveLabel">{{ $t('common.confirm') }}</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
import { h, ref, reactive, onMounted, onUnmounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useTable } from '@/hooks/core/useTable'
import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
import { ElTag, ElMessage, ElMessageBox, ElEmpty } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import {
  fetchFbAuthUrl,
  fetchFbAccountList,
  fetchFbDisconnectAccount,
  fetchFbUpdateLabel,
  fetchFbRefreshStats,
  type FbAccount
} from '@/api/facebook'

defineOptions({ name: 'AdAccount' })

const { t } = useI18n()

// ==================== 授权链接弹窗 ====================
const showAuthDialog = ref(false)
const shortUrl = ref('')
const fullAuthUrl = ref('')
const isPolling = ref(false)
const isConnecting = ref(false)
const copySuccess = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null

// ==================== 编辑备注 ====================
const showLabelDialog = ref(false)
const editingAccount = ref<FbAccount | null>(null)
const labelForm = reactive({ label: '' })

// ==================== 格式化 ====================
const formatDate = (val: string) => {
  if (!val) return '—'
  const d = new Date(val)
  if (isNaN(d.getTime())) return val
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// ==================== useTable ====================
const fetchAccounts = async (_params: any) => {
  try {
    const result = await fetchFbAccountList()
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
    apiFn: fetchAccounts,
    apiParams: { current: 1, size: 50 },
    columnsFactory: () => [
      { type: 'index', width: 60, label: t('menus.adAccount.columns.index') },
      {
        prop: 'accountName',
        label: t('menus.adAccount.columns.accountName'),
        minWidth: 160,
        formatter: (row: FbAccount) => {
          const name = row.label ? `${row.label} (${row.fbUserName})` : row.fbUserName
          return name || '—'
        }
      },
      {
        prop: 'fbUserId',
        label: t('menus.adAccount.columns.accountId'),
        minWidth: 160,
        formatter: (row: FbAccount) => row.fbUserId || '—'
      },
      {
        prop: 'accountStatus',
        label: t('menus.adAccount.columns.status'),
        minWidth: 90,
        formatter: (row: FbAccount) => {
          const isNormal = row.accountStatus === '正常'
          return h(ElTag, {
            type: isNormal ? 'success' : 'danger',
            size: 'small'
          }, () => isNormal ? t('menus.adAccount.status.normal') : t('menus.adAccount.status.expired'))
        }
      },
      {
        prop: 'hasAdPerm',
        label: t('menus.adAccount.columns.adPerm'),
        minWidth: 100,
        formatter: (row: FbAccount) => {
          return h(ElTag, {
            type: row.hasAdPerm ? 'success' : 'info',
            size: 'small'
          }, () => row.hasAdPerm ? t('menus.adAccount.adPerm.granted') : t('menus.adAccount.adPerm.none'))
        }
      },
      {
        prop: 'bmCount',
        label: t('menus.adAccount.columns.bm'),
        minWidth: 70,
        formatter: (row: FbAccount) => row.bmCount.toString()
      },
      {
        prop: 'adAccounts',
        label: t('menus.adAccount.columns.adAccounts'),
        minWidth: 160,
        formatter: (row: FbAccount) => {
          return `BM: ${row.bmAdCount}，个人: ${row.personalAdCount}`
        }
      },
      {
        prop: 'validity',
        label: t('menus.adAccount.columns.validity'),
        minWidth: 130,
        formatter: (row: FbAccount) => {
          const days = row.daysUntilExpiry
          if (days < 0) {
            return h(ElTag, { type: 'danger', size: 'small' }, () => t('menus.adAccount.expiredBadge'))
          }
          let color = '#67c23a'
          if (days <= 7) color = '#f56c6c'
          else if (days <= 30) color = '#e6a23c'
          return h('span', { style: { color, fontWeight: '600' } }, t('menus.adAccount.daysLeft', { days }))
        }
      },
      {
        prop: 'createdAt',
        label: t('menus.adAccount.columns.authTime'),
        minWidth: 150,
        formatter: (row: FbAccount) => formatDate(row.createdAt)
      },
      {
        label: t('menus.adAccount.columns.actions'),
        width: 200,
        fixed: 'right',
        formatter: (row: FbAccount) =>
          h('div', [
            h(ArtButtonTable, {
              type: 'edit',
              onClick: () => showEditLabel(row)
            }),
            h(ArtButtonTable, {
              type: 'view',
              icon: 'ri:refresh-line',
              onClick: () => handleRefreshStats(row)
            }),
            h(ArtButtonTable, {
              type: 'delete',
              onClick: () => handleDisconnect(row)
            })
          ])
      }
    ]
  }
})

const totalAccounts = computed(() => data.value.length)

// ==================== 操作方法 ====================
// 开启授权轮询
const startPolling = () => {
  if (pollTimer) return
  isPolling.value = true
  const initialCount = totalAccounts.value
  pollTimer = setInterval(async () => {
    try {
      const result = await fetchFbAccountList()
      if (result.total > initialCount) {
        stopPolling()
        showAuthDialog.value = false
        ElMessage.success(t('menus.adAccount.authSuccess'))
        await refreshData()
      }
    } catch {
      // 轮询失败继续
    }
  }, 3000)
}

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  isPolling.value = false
}

const stopPollingAndClose = () => {
  stopPolling()
  showAuthDialog.value = false
}

// 点击"连接 Facebook"
const handleConnectFb = async () => {
  isConnecting.value = true
  try {
    const { authUrl: full, shortUrl: short } = await fetchFbAuthUrl()
    shortUrl.value = short
    fullAuthUrl.value = full
    copySuccess.value = false
    showAuthDialog.value = true
    startPolling()
  } catch {
    ElMessage.error(t('menus.adAccount.authUrlError'))
  } finally {
    isConnecting.value = false
  }
}

// 复制链接
const copyShortUrl = async () => {
  try {
    await navigator.clipboard.writeText(shortUrl.value)
    copySuccess.value = true
    ElMessage.success(t('menus.adAccount.copySuccess'))
    setTimeout(() => { copySuccess.value = false }, 2000)
  } catch {
    fallbackCopy(shortUrl.value)
  }
}

const copyFullUrl = async () => {
  try {
    await navigator.clipboard.writeText(fullAuthUrl.value)
    ElMessage.success(t('menus.adAccount.copySuccess'))
  } catch {
    fallbackCopy(fullAuthUrl.value)
  }
}

const fallbackCopy = (text: string) => {
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)
  textarea.select()
  document.execCommand('copy')
  document.body.removeChild(textarea)
  copySuccess.value = true
  ElMessage.success(t('menus.adAccount.copySuccess'))
  setTimeout(() => { copySuccess.value = false }, 2000)
}

const openAuthUrl = () => {
  window.open(fullAuthUrl.value, '_blank')
}

// 编辑备注
const showEditLabel = (row: FbAccount) => {
  editingAccount.value = row
  labelForm.label = row.label || ''
  showLabelDialog.value = true
}

const handleSaveLabel = async () => {
  if (!editingAccount.value) return
  try {
    await fetchFbUpdateLabel(editingAccount.value.id, labelForm.label)
    ElMessage.success(t('menus.adAccount.labelUpdateSuccess'))
    showLabelDialog.value = false
    await refreshData()
  } catch {
    ElMessage.error('操作失败')
  }
}

// 刷新统计
const handleRefreshStats = async (row: FbAccount) => {
  try {
    await fetchFbRefreshStats(row.id)
    ElMessage.success(t('menus.adAccount.refreshStatsSuccess'))
    await refreshData()
  } catch {
    ElMessage.error('刷新失败')
  }
}

// 断开连接
const handleDisconnect = async (row: FbAccount) => {
  try {
    await ElMessageBox.confirm(
      t('menus.adAccount.confirmDisconnect'),
      '提示',
      { type: 'warning' }
    )
    await fetchFbDisconnectAccount(row.id)
    ElMessage.success(t('menus.adAccount.disconnectSuccess'))
    await refreshData()
  } catch {
    // 用户取消
  }
}

// ==================== 生命周期 ====================
onMounted(async () => {
  await refreshData()
})

onUnmounted(() => {
  stopPolling()
})
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
  min-width: 200px;
}

.fb-connect-actions {
  flex-shrink: 0;
}

.auth-dialog-body {
  .auth-dialog-tip {
    font-size: 14px;
    color: #666;
    margin-bottom: 16px;
    line-height: 1.6;
  }

  .short-link-box,
  .full-link-box {
    .link-label {
      display: block;
      font-size: 13px;
      font-weight: 500;
      color: #303133;
      margin-bottom: 6px;
    }
  }

  .auth-link-box {
    display: flex;
    align-items: center;
    gap: 8px;

    .el-input {
      flex: 1;
    }
  }

  .auth-dialog-actions {
    display: flex;
    align-items: center;
  }

  .polling-status {
    :deep(.el-alert__title) {
      display: flex;
      align-items: center;
    }
  }
}
</style>
