<!-- 添加到BM弹窗 -->
<template>
  <ElDialog
    :model-value="modelValue"
    :title="$t('menus.adAccount.addToBM')"
    width="640px"
    destroy-on-close
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <div class="add-to-bm-dialog">
      <!-- 标签页：操作 / 结果 -->
      <div class="bm-tabs">
        <div
          class="bm-tab"
          :class="{ active: activeTab === 'action' }"
          @click="activeTab = 'action'"
        >
          {{ $t('menus.adAccount.addAuthTabAction') }}
        </div>
        <div
          class="bm-tab"
          :class="{ active: activeTab === 'result' }"
          @click="activeTab = 'result'"
        >
          {{ $t('menus.adAccount.addAuthTabResult') }}
        </div>
      </div>

      <!-- 操作面板 -->
      <div v-show="activeTab === 'action'" class="bm-action-panel">
        <!-- 已选广告账户 -->
        <div v-if="selectedAdAccounts.length > 0" class="selected-accounts-info">
          <div class="info-label">{{ $t('menus.addAuth.selectedAccounts') }}</div>
          <div class="account-tags">
            <ElTag
              v-for="acc in selectedAdAccounts"
              :key="acc.id"
              size="small"
              type="info"
              class="account-tag"
            >
              {{ acc.accountId }} - {{ acc.name || acc.businessName }}
            </ElTag>
          </div>
        </div>
        <ElAlert
          v-else
          :title="$t('menus.addAuth.noAccountsSelected')"
          type="warning"
          :closable="false"
          show-icon
          class="no-accounts-alert"
        />

        <!-- 操作类型选择 -->
        <div class="bm-select-group">
          <ElSelect
            v-model="bmActionType"
            class="bm-select"
            :placeholder="$t('menus.addToBm.selectActionType')"
          >
            <ElOption :label="$t('menus.addToBm.claimByBmOption')" value="claimByBm" />
            <ElOption :label="$t('menus.addToBm.shareToBmOption')" value="shareToBm" />
          </ElSelect>
        </div>

        <!-- Step 1: 填写BM ID -->
        <div class="bm-step">
          <div class="bm-step-header">
            <span class="step-badge">1</span>
            <span class="step-label">
              <span class="required-star">*</span>
              {{ $t('menus.addToBm.step1Label') }}
            </span>
          </div>
          <ElInput
            v-model="bmIdInput"
            type="textarea"
            :rows="4"
            :placeholder="$t('menus.addToBm.step1Placeholder')"
            class="bm-id-textarea"
          />
        </div>

        <!-- Step 2: 系统默认执行时间间隔 -->
        <div class="bm-step">
          <div class="bm-step-header">
            <span class="step-badge">2</span>
            <ElCheckbox v-model="useDefaultInterval">
              {{ $t('menus.addToBm.step2Label') }}
            </ElCheckbox>
          </div>
        </div>
      </div>

      <!-- 结果面板 -->
      <div v-show="activeTab === 'result'" class="bm-result-panel">
        <template v-if="submitResult">
          <div class="result-section-title">{{ $t('menus.addToBm.resultTitle') }}</div>
          <div class="assign-summary">
            <ElTag type="success" size="small"
              >{{ $t('menus.addAuth.assignSuccess') }}: {{ submitResult.success }}</ElTag
            >
            <ElTag type="danger" size="small"
              >{{ $t('menus.addAuth.assignFailed') }}: {{ submitResult.failed }}</ElTag
            >
            <ElTag type="info" size="small"
              >{{ $t('menus.addAuth.assignTotal') }}: {{ submitResult.total }}</ElTag
            >
          </div>
          <ElTable :data="submitResult.results" border size="small">
            <ElTableColumn prop="adAccountId" label="BM ID" min-width="160" />
            <ElTableColumn prop="success" :label="$t('menus.addAuth.assignStatus')" width="100">
              <template #default="{ row }">
                <ElTag :type="row.success ? 'success' : 'danger'" size="small">
                  {{ row.success ? $t('menus.addAuth.assignOk') : $t('menus.addAuth.assignFail') }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn
              prop="message"
              :label="$t('menus.addAuth.assignMessage')"
              min-width="200"
            />
          </ElTable>
        </template>

        <ElEmpty v-if="!submitResult" :description="$t('menus.addAuth.noResultYet')" />
      </div>
    </div>

    <template #footer>
      <ElButton
        type="primary"
        class="confirm-btn"
        :loading="submitting"
        :disabled="selectedAdAccounts.length === 0"
        @click="handleConfirm"
      >
        <ElIcon class="confirm-icon"><Lock /></ElIcon>
        {{ $t('menus.addAuth.confirm') }}
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ref, computed } from 'vue'
  import { Lock } from '@element-plus/icons-vue'
  import { ElMessage } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import type { FbAdAccountDetail } from '@/api/facebook'

  defineOptions({ name: 'AddToBmDialog' })

  const { t } = useI18n()

  const props = defineProps<{
    modelValue: boolean
    selectedAdAccounts?: FbAdAccountDetail[]
  }>()

  defineEmits<{
    'update:modelValue': [value: boolean]
  }>()

  // 已选广告账户（从父组件传入）
  const selectedAdAccounts = computed(() => props.selectedAdAccounts ?? [])

  // ==================== 状态 ====================
  const activeTab = ref<'action' | 'result'>('action')
  const bmActionType = ref('claimByBm')
  const bmIdInput = ref('')
  const useDefaultInterval = ref(true)
  const submitting = ref(false)

  // 提交结果
  const submitResult = ref<{
    results: { adAccountId: string; success: boolean; message: string }[]
    total: number
    success: number
    failed: number
  } | null>(null)

  // 解析输入的BM ID列表
  const parseBmIds = (): string[] => {
    return bmIdInput.value
      .split('\n')
      .map((l) => l.trim())
      .filter(Boolean)
  }

  // ==================== 确认提交 ====================
  const handleConfirm = async () => {
    if (selectedAdAccounts.value.length === 0) {
      ElMessage.warning(t('menus.adAccount.selectRowsFirst'))
      return
    }
    if (!bmIdInput.value.trim()) {
      ElMessage.warning(t('menus.addToBm.enterBmId'))
      return
    }

    submitting.value = true
    activeTab.value = 'result'

    try {
      const bmIds = parseBmIds()
      const adAccountIds = selectedAdAccounts.value.map((acc) => acc.id)

      // TODO: 调用后端API - 添加到BM
      // const result = await fetchAddToBm({
      //   adAccountIds,
      //   bmIds,
      //   action: bmActionType.value,
      //   useDefaultInterval: useDefaultInterval.value
      // })

      // 模拟结果
      const results: { adAccountId: string; success: boolean; message: string }[] = []
      let successCount = 0
      let failedCount = 0

      for (const bmId of bmIds) {
        for (const adAccountId of adAccountIds) {
          const success = Math.random() > 0.3 // 模拟成功率
          results.push({
            adAccountId: `${adAccountId} → ${bmId}`,
            success,
            message: success ? '操作成功' : '操作失败，请检查权限'
          })
          if (success) successCount++
          else failedCount++
        }
      }

      submitResult.value = {
        results,
        total: results.length,
        success: successCount,
        failed: failedCount
      }

      if (failedCount === 0) {
        ElMessage.success(t('menus.addToBm.successMsg', { count: successCount }))
      } else if (successCount > 0) {
        ElMessage.warning(
          t('menus.addToBm.partialMsg', { success: successCount, failed: failedCount })
        )
      } else {
        ElMessage.error(t('menus.addToBm.failMsg'))
      }
    } catch {
      ElMessage.error(t('menus.addToBm.failMsg'))
      submitResult.value = null
    } finally {
      submitting.value = false
    }
  }
</script>

<style lang="scss" scoped>
  .add-to-bm-dialog {
    /* 标签页 */
    .bm-tabs {
      display: flex;
      gap: 0;
      margin-bottom: 20px;
      border-bottom: 1px solid var(--el-border-color-lighter);

      .bm-tab {
        padding: 8px 20px;
        font-size: 14px;
        color: var(--el-text-color-secondary);
        cursor: pointer;
        border-bottom: 2px solid transparent;
        transition: all 0.2s;

        &:hover {
          color: var(--el-color-primary);
        }

        &.active {
          color: var(--el-color-primary);
          border-bottom-color: var(--el-color-primary);
        }
      }
    }

    /* 已选广告账户 */
    .selected-accounts-info {
      margin-bottom: 16px;
      padding: 12px;
      background: var(--el-fill-color-lighter);
      border-radius: 6px;

      .info-label {
        font-size: 13px;
        color: var(--el-text-color-secondary);
        margin-bottom: 8px;
      }

      .account-tags {
        display: flex;
        flex-wrap: wrap;
        gap: 6px;

        .account-tag {
          font-size: 12px;
        }
      }
    }

    .no-accounts-alert {
      margin-bottom: 16px;
    }

    /* 操作类型选择 */
    .bm-select-group {
      margin-bottom: 20px;

      .bm-select {
        width: 100%;
      }
    }

    /* 步骤 */
    .bm-step {
      margin-bottom: 20px;

      &:last-child {
        margin-bottom: 0;
      }
    }

    .bm-step-header {
      display: flex;
      align-items: center;
      gap: 10px;
      margin-bottom: 10px;
    }

    .step-badge {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: 22px;
      height: 22px;
      border-radius: 50%;
      background-color: var(--el-color-primary);
      color: #fff;
      font-size: 13px;
      font-weight: 600;
      flex-shrink: 0;
    }

    .step-label {
      font-size: 14px;
      font-weight: 500;
      color: var(--el-text-color-primary);

      .required-star {
        color: var(--el-color-danger);
        margin-right: 4px;
      }
    }

    /* BM ID 输入框 */
    .bm-id-textarea {
      :deep(.el-textarea__inner) {
        background-color: var(--el-fill-color-lighter);
        font-size: 13px;
        line-height: 1.6;
      }
    }

    /* 结果面板 */
    .bm-result-panel {
      min-height: 200px;

      .result-section-title {
        font-size: 14px;
        font-weight: 600;
        color: var(--el-text-color-primary);
        margin-bottom: 12px;
        padding-bottom: 8px;
        border-bottom: 1px solid var(--el-border-color-lighter);
      }

      .assign-summary {
        display: flex;
        gap: 8px;
        margin-bottom: 12px;
      }
    }

    /* 确认按钮 */
    .confirm-btn {
      width: 100%;
      height: 42px;
      font-size: 15px;

      .confirm-icon {
        margin-right: 8px;
      }
    }
  }
</style>
