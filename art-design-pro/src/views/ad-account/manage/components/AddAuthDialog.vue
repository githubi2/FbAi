<!-- 增加广告账号权限弹窗 -->
<template>
  <ElDialog
    :model-value="modelValue"
    :title="$t('menus.adAccount.addAuthDialogTitle')"
    width="600px"
    destroy-on-close
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <div class="add-auth-dialog">
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
        <!-- 权限类型选择 -->
        <div class="auth-select-group">
          <ElSelect v-model="authType" class="auth-select" :placeholder="$t('menus.addAuth.selectAuthType')">
            <ElOption :label="$t('menus.addAuth.authorizeAdmin')" value="authorizeAdmin" />
            <ElOption :label="$t('menus.addAuth.authorizeAdManager')" value="authorizeAdManager" />
            <ElOption :label="$t('menus.addAuth.authorizeAdAnalyst')" value="authorizeAdAnalyst" />
          </ElSelect>
        </div>

        <!-- 步骤1：输入目标账号 -->
        <div class="auth-step">
          <div class="auth-step-head">
            <span class="auth-step-num">1</span>
            <span class="auth-step-label">
              <span class="required-star">*</span>
              {{ $t('menus.addAuth.step1Label') }}
            </span>
          </div>
          <div class="auth-step-body">
            <ElInput
              v-model="uidInput"
              type="textarea"
              :rows="5"
              :placeholder="$t('menus.addAuth.step1Placeholder')"
              class="uid-textarea"
            />
          </div>
        </div>

        <!-- 步骤2：检测好友关系 -->
        <div class="auth-step">
          <div class="auth-step-head">
            <span class="auth-step-num">2</span>
            <span class="auth-step-label">{{ $t('menus.addAuth.step2Label') }}</span>
          </div>
          <div class="auth-step-body">
            <ElButton
              type="success"
              class="detect-btn"
              :loading="detecting"
              @click="handleDetect"
            >
              {{ $t('menus.addAuth.step2Button') }}
            </ElButton>
          </div>
        </div>

        <!-- 步骤3：系统默认执行时间间隔 -->
        <div class="auth-step">
          <div class="auth-step-head">
            <span class="auth-step-num">3</span>
            <ElCheckbox v-model="useDefaultInterval" class="auth-step-label">
              {{ $t('menus.addAuth.step3Label') }}
            </ElCheckbox>
          </div>
        </div>
      </div>

      <!-- 结果面板 -->
      <div v-show="activeTab === 'result'" class="auth-result-panel">
        <ElEmpty v-if="!detectResult" :description="$t('menus.addAuth.noResultYet')" />
        <div v-else class="result-content">
          <ElTable :data="detectResult" border size="small">
            <ElTableColumn prop="uid" label="UID" min-width="160" />
            <ElTableColumn prop="status" :label="$t('menus.addAuth.friendStatus')" width="120">
              <template #default="{ row }">
                <ElTag :type="row.isFriend ? 'success' : 'danger'" size="small">
                  {{ row.isFriend ? $t('menus.addAuth.isFriend') : $t('menus.addAuth.notFriend') }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="name" :label="$t('menus.addAuth.accountName')" min-width="140" />
          </ElTable>
        </div>
      </div>
    </div>

    <template #footer>
      <ElButton
        type="primary"
        class="confirm-btn"
        :loading="submitting"
        @click="handleConfirm"
      >
        <ElIcon class="confirm-icon"><Lock /></ElIcon>
        {{ $t('menus.addAuth.confirm') }}
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ref } from 'vue'
  import { Lock } from '@element-plus/icons-vue'
  import { ElMessage } from 'element-plus'

  defineOptions({ name: 'AddAuthDialog' })

  defineProps<{
    modelValue: boolean
  }>()

  defineEmits<{
    'update:modelValue': [value: boolean]
  }>()

  // ==================== 状态 ====================
  const activeTab = ref<'action' | 'result'>('action')
  const authType = ref('authorizeAdmin')
  const uidInput = ref('')
  const useDefaultInterval = ref(true)
  const detecting = ref(false)
  const submitting = ref(false)

  // 检测结果
  interface DetectRow {
    uid: string
    isFriend: boolean
    name: string
  }
  const detectResult = ref<DetectRow[] | null>(null)

  // ==================== 检测好友关系 ====================
  const handleDetect = async () => {
    if (!uidInput.value.trim()) {
      ElMessage.warning('请先输入Facebook账号UID或主页地址')
      return
    }

    detecting.value = true
    activeTab.value = 'result'

    try {
      // 解析输入的UID列表（支持UID和URL混合输入，每行一个）
      const lines = uidInput.value
        .split('\n')
        .map((l) => l.trim())
        .filter(Boolean)

      // 模拟检测结果（实际应调用后端API）
      await new Promise((r) => setTimeout(r, 1500))

      detectResult.value = lines.map((line) => {
        // 从URL中提取UID或直接使用UID
        let uid = line
        const urlMatch = line.match(/facebook\.com\/profile\.php\?id=(\d+)/)
        const pathMatch = line.match(/facebook\.com\/([^/?]+)/)
        if (urlMatch) {
          uid = urlMatch[1]
        } else if (pathMatch && !/^\d+$/.test(pathMatch[1])) {
          uid = pathMatch[1]
        }

        return {
          uid,
          isFriend: Math.random() > 0.3, // 模拟
          name: '' // 实际由后端返回
        }
      })

      ElMessage.success('检测完成')
    } catch {
      ElMessage.error('检测失败，请重试')
      detectResult.value = null
    } finally {
      detecting.value = false
    }
  }

  // ==================== 确认提交 ====================
  const handleConfirm = async () => {
    if (!uidInput.value.trim()) {
      ElMessage.warning('请先输入Facebook账号UID或主页地址')
      return
    }

    submitting.value = true
    try {
      // TODO: 调用后端API执行授权操作
      await new Promise((r) => setTimeout(r, 1000))
      ElMessage.success('授权操作已提交')
    } catch {
      ElMessage.error('授权失败，请重试')
    } finally {
      submitting.value = false
    }
  }
</script>

<style lang="scss" scoped>
  .add-auth-dialog {
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

    /* 权限类型选择 */
    .auth-select-group {
      margin-bottom: 20px;

      .auth-select {
        width: 100%;
      }
    }

    /* 步骤 */
    .auth-step {
      position: relative;
      padding-left: 36px;
      margin-bottom: 20px;

      /* 连接线 */
      &::before {
        content: '';
        position: absolute;
        left: 11px;
        top: 28px;
        bottom: -20px;
        width: 2px;
        background-color: var(--el-border-color);
      }

      &:last-child::before {
        display: none;
      }

      &:last-child {
        margin-bottom: 0;
      }
    }

    .auth-step-head {
      display: flex;
      align-items: center;
      gap: 12px;
      min-height: 24px;
      margin-bottom: 10px;
    }

    .auth-step-num {
      position: absolute;
      left: 0;
      top: 0;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: 24px;
      height: 24px;
      border-radius: 50%;
      background-color: var(--el-fill-color-dark);
      color: var(--el-text-color-regular);
      font-size: 13px;
      font-weight: 600;
      flex-shrink: 0;
    }

    .auth-step-label {
      font-size: 14px;
      font-weight: 500;
      color: var(--el-text-color-primary);
      line-height: 24px;

      .required-star {
        color: var(--el-color-danger);
        margin-right: 4px;
      }
    }

    .auth-step-body {
      min-width: 0;
    }

    /* UID输入框 */
    .uid-textarea {
      :deep(.el-textarea__inner) {
        background-color: var(--el-fill-color-lighter);
        font-size: 13px;
        line-height: 1.6;
      }
    }

    /* 检测按钮 */
    .detect-btn {
      width: 100%;
      height: 40px;
      font-size: 14px;
    }

    /* 结果面板 */
    .auth-result-panel {
      min-height: 200px;
    }

    /* 确认按钮 */
    .confirm-btn {
      width: 100%;
      height: 42px;
      font-size: 15px;

      .confirm-icon {
        margin-right: 6px;
      }
    }
  }
</style>
