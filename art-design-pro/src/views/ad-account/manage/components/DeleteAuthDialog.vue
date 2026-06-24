<!-- 删除广告账号权限弹窗 -->
<template>
  <ElDialog
    :model-value="modelValue"
    :title="$t('menus.deleteAuth.dialogTitle')"
    width="640px"
    destroy-on-close
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <div class="delete-auth-dialog">
      <!-- 标签页：操作 / 结果 -->
      <div class="auth-tabs">
        <div
          class="auth-tab"
          :class="{ active: activeTab === 'action' }"
          @click="activeTab = 'action'"
        >
          {{ $t('menus.adAccount.addAuthTabAction') }}
        </div>
        <div
          class="auth-tab"
          :class="{ active: activeTab === 'result' }"
          @click="activeTab = 'result'"
        >
          {{ $t('menus.adAccount.addAuthTabResult') }}
        </div>
      </div>

      <!-- 操作面板 -->
      <div v-show="activeTab === 'action'" class="auth-action-panel">
        <!-- 删除模式选择 -->
        <div class="auth-select-group">
          <ElSelect v-model="deleteMode" class="auth-select">
            <ElOption
              v-for="opt in deleteModeOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </ElSelect>
        </div>

        <!-- 输入目标账号 -->
        <div class="auth-field">
          <div class="auth-field-label">
            <span class="required-star">*</span>
            {{ $t('menus.deleteAuth.step1Label') }}
          </div>
          <ElInput
            v-model="uidInput"
            type="textarea"
            :rows="5"
            :placeholder="$t('menus.deleteAuth.step1Placeholder')"
            class="uid-textarea"
          />
        </div>

        <!-- 删除当前FB账号权限 -->
        <div class="auth-field">
          <ElCheckbox v-model="deleteCurrentAccount">
            {{ $t('menus.deleteAuth.step2Label') }}
          </ElCheckbox>
        </div>

        <!-- 系统默认执行时间间隔 -->
        <div class="auth-field">
          <ElCheckbox v-model="useDefaultInterval">
            {{ $t('menus.addAuth.step3Label') }}
          </ElCheckbox>
        </div>
      </div>

      <!-- 结果面板 -->
      <div v-show="activeTab === 'result'" class="auth-result-panel">
        <template v-if="deleteResult">
          <div class="result-section-title">{{ $t('menus.deleteAuth.resultTitle') }}</div>
          <div class="assign-summary">
            <ElTag type="success" size="small">
              {{ $t('menus.addAuth.assignSuccess') }}: {{ deleteResult.success }}
            </ElTag>
            <ElTag type="danger" size="small">
              {{ $t('menus.addAuth.assignFailed') }}: {{ deleteResult.failed }}
            </ElTag>
            <ElTag type="info" size="small">
              {{ $t('menus.addAuth.assignTotal') }}: {{ deleteResult.total }}
            </ElTag>
          </div>
          <ElTable :data="deleteResult.results" border size="small">
            <ElTableColumn prop="adAccountId" label="Ad Account ID" min-width="160" />
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

        <ElEmpty v-if="!deleteResult" :description="$t('menus.addAuth.noResultYet')" />
      </div>
    </div>

    <template #footer>
      <ElButton type="primary" class="confirm-btn" :loading="submitting" @click="handleConfirm">
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
  import {
    fetchRemoveAdAccountUser,
    type FbAdAccountDetail,
    type FbAssignUserResponse
  } from '@/api/facebook'

  defineOptions({ name: 'DeleteAuthDialog' })

  const props = defineProps<{
    modelValue: boolean
    selectedAdAccounts?: FbAdAccountDetail[]
  }>()

  defineEmits<{
    'update:modelValue': [value: boolean]
  }>()

  const { t } = useI18n()

  // 已选广告账户（从父组件传入）
  const selectedAdAccounts = computed(() => props.selectedAdAccounts ?? [])

  // ==================== 状态 ====================
  const activeTab = ref<'action' | 'result'>('action')
  const deleteMode = ref('deleteTheirs')
  const uidInput = ref('')
  const deleteCurrentAccount = ref(false)
  const useDefaultInterval = ref(true)
  const submitting = ref(false)

  // 删除结果
  const deleteResult = ref<FbAssignUserResponse | null>(null)

  // 删除模式选项
  const deleteModeOptions = computed(() => [
    { label: t('menus.deleteAuth.modeDeleteTheirs'), value: 'deleteTheirs' },
    { label: t('menus.deleteAuth.modeDeleteExceptTheirs'), value: 'deleteExceptTheirs' },
    { label: t('menus.deleteAuth.modeDeleteExceptSelf'), value: 'deleteExceptSelf' },
    { label: t('menus.deleteAuth.modeDeleteSelf'), value: 'deleteSelf' },
    { label: t('menus.deleteAuth.modeDeleteBM'), value: 'deleteBM' }
  ])

  // 解析输入的 UID 列表
  const parseUIDs = (): string[] => {
    return uidInput.value
      .split('\n')
      .map((l) => l.trim())
      .filter(Boolean)
      .map((line) => {
        // 从 URL 中提取 UID
        const urlMatch = line.match(/facebook\.com\/profile\.php\?id=(\d+)/)
        if (urlMatch) return urlMatch[1]
        // 如果是纯数字 UID，直接返回
        if (/^\d+$/.test(line)) return line
        // 其他 URL 格式，提取路径部分
        const pathMatch = line.match(/facebook\.com\/([^/?]+)/)
        if (pathMatch) return pathMatch[1]
        return line
      })
  }

  // ==================== 确认提交 ====================
  const handleConfirm = async () => {
    // 校验：至少需要一种输入方式
    const hasUIDs = uidInput.value.trim().length > 0
    if (!hasUIDs && !deleteCurrentAccount.value) {
      ElMessage.warning(t('menus.deleteAuth.inputRequired'))
      return
    }
    if (selectedAdAccounts.value.length === 0) {
      ElMessage.warning(t('menus.adAccount.selectRowsFirst'))
      return
    }

    submitting.value = true
    activeTab.value = 'result'

    try {
      const uids = hasUIDs ? parseUIDs() : []
      const result = await fetchRemoveAdAccountUser({
        adAccountIds: selectedAdAccounts.value.map((acc) => acc.id),
        uids,
        mode: deleteMode.value,
        deleteCurrent: deleteCurrentAccount.value
      })

      deleteResult.value = result

      if (result.failed === 0) {
        ElMessage.success(t('menus.deleteAuth.deleteSuccess', { count: result.success }))
      } else if (result.success > 0) {
        ElMessage.warning(
          t('menus.deleteAuth.deletePartial', { success: result.success, failed: result.failed })
        )
      } else {
        ElMessage.error(t('menus.deleteAuth.deleteFailed'))
      }
    } catch {
      ElMessage.error(t('menus.deleteAuth.deleteFailed'))
      deleteResult.value = null
    } finally {
      submitting.value = false
    }
  }
</script>

<style lang="scss" scoped>
  .delete-auth-dialog {
    /* 标签页 */
    .auth-tabs {
      display: flex;
      gap: 0;
      margin-bottom: 20px;
      border-bottom: 1px solid var(--el-border-color-lighter);

      .auth-tab {
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

    /* 删除模式选择 */
    .auth-select-group {
      margin-bottom: 20px;

      .auth-select {
        width: 100%;
      }
    }

    /* 表单字段 */
    .auth-field {
      margin-bottom: 16px;

      &:last-child {
        margin-bottom: 0;
      }
    }

    .auth-field-label {
      font-size: 14px;
      font-weight: 500;
      color: var(--el-text-color-primary);
      margin-bottom: 8px;

      .required-star {
        color: var(--el-color-danger);
        margin-right: 4px;
      }
    }
    }

    /* UID输入框 */
    .uid-textarea {
      :deep(.el-textarea__inner) {
        background-color: var(--el-fill-color-lighter);
        font-size: 13px;
        line-height: 1.6;
      }
    }

    /* 结果面板 */
    .auth-result-panel {
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
