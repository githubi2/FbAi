<!-- 增加广告账号权限弹窗 -->
<template>
  <ElDialog
    :model-value="modelValue"
    :title="$t('menus.adAccount.addAuthDialogTitle')"
    width="640px"
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

        <!-- 权限类型选择 -->
        <div class="auth-select-group">
          <ElSelect
            v-model="authType"
            class="auth-select"
            :placeholder="$t('menus.addAuth.selectAuthType')"
          >
            <ElOption :label="$t('menus.addAuth.authorizeAdmin')" value="authorizeAdmin" />
            <ElOption :label="$t('menus.addAuth.authorizeAdManager')" value="authorizeAdManager" />
            <ElOption :label="$t('menus.addAuth.authorizeAdAnalyst')" value="authorizeAdAnalyst" />
          </ElSelect>
        </div>

        <!-- 步骤1：输入目标账号 -->
        <div class="auth-step">
          <div class="auth-step-indicator">
            <span class="auth-step-num">1</span>
            <div class="auth-step-line"></div>
          </div>
          <div class="auth-step-body">
            <div class="auth-step-head">
              <span class="auth-step-label">
                <span class="required-star">*</span>
                {{ $t('menus.addAuth.step1Label') }}
              </span>
            </div>
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
          <div class="auth-step-indicator">
            <span class="auth-step-num">2</span>
            <div class="auth-step-line"></div>
          </div>
          <div class="auth-step-body">
            <div class="auth-step-head">
              <span class="auth-step-label">{{ $t('menus.addAuth.step2Label') }}</span>
            </div>
            <ElButton type="success" class="detect-btn" :loading="detecting" @click="handleDetect">
              {{ $t('menus.addAuth.step2Button') }}
            </ElButton>
          </div>
        </div>

        <!-- 步骤3：系统默认执行时间间隔 -->
        <div class="auth-step">
          <div class="auth-step-indicator">
            <span class="auth-step-num">3</span>
          </div>
          <div class="auth-step-head">
            <ElCheckbox v-model="useDefaultInterval" class="auth-step-label">
              {{ $t('menus.addAuth.step3Label') }}
            </ElCheckbox>
          </div>
        </div>
      </div>

      <!-- 结果面板 -->
      <div v-show="activeTab === 'result'" class="auth-result-panel">
        <!-- 检测结果 -->
        <template v-if="detectResult">
          <div class="result-section-title">{{ $t('menus.addAuth.lookupResult') }}</div>
          <ElTable :data="detectResult" border size="small">
            <ElTableColumn prop="uid" label="UID" min-width="140" />
            <ElTableColumn prop="name" :label="$t('menus.addAuth.accountName')" min-width="120">
              <template #default="{ row }">
                <div class="user-info">
                  <img v-if="row.avatar" :src="row.avatar" class="user-avatar" alt="" />
                  <span>{{ row.name || '-' }}</span>
                </div>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="isFriend" :label="$t('menus.addAuth.friendStatus')" width="100">
              <template #default="{ row }">
                <ElTag :type="row.isFriend ? 'success' : 'danger'" size="small">
                  {{ row.isFriend ? $t('menus.addAuth.isFriend') : $t('menus.addAuth.notFriend') }}
                </ElTag>
              </template>
            </ElTableColumn>
          </ElTable>
        </template>

        <!-- 授权结果 -->
        <template v-if="assignResult">
          <div class="result-section-title">{{ $t('menus.addAuth.assignResult') }}</div>
          <div class="assign-summary">
            <ElTag type="success" size="small"
              >{{ $t('menus.addAuth.assignSuccess') }}: {{ assignResult.success }}</ElTag
            >
            <ElTag type="danger" size="small"
              >{{ $t('menus.addAuth.assignFailed') }}: {{ assignResult.failed }}</ElTag
            >
            <ElTag type="info" size="small"
              >{{ $t('menus.addAuth.assignTotal') }}: {{ assignResult.total }}</ElTag
            >
          </div>
          <ElTable :data="assignResult.results" border size="small">
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

        <ElEmpty
          v-if="!detectResult && !assignResult"
          :description="$t('menus.addAuth.noResultYet')"
        />
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
  import {
    fetchLookupFbUsers,
    fetchAssignAdAccountUser,
    type FbAdAccountDetail,
    type FbLookupUserResult,
    type FbAssignUserResponse
  } from '@/api/facebook'

  defineOptions({ name: 'AddAuthDialog' })

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
  const authType = ref('authorizeAdmin')
  const uidInput = ref('')
  const useDefaultInterval = ref(true)
  const detecting = ref(false)
  const submitting = ref(false)

  // 检测结果
  const detectResult = ref<FbLookupUserResult[] | null>(null)

  // 授权结果
  const assignResult = ref<FbAssignUserResponse | null>(null)

  // authType → FB role 映射
  const roleMap: Record<string, 'ADMIN' | 'ADVERTISER' | 'ANALYST'> = {
    authorizeAdmin: 'ADMIN',
    authorizeAdManager: 'ADVERTISER',
    authorizeAdAnalyst: 'ANALYST'
  }

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

  // ==================== 检测好友关系 ====================
  const handleDetect = async () => {
    if (!uidInput.value.trim()) {
      ElMessage.warning('请先输入Facebook账号UID或主页地址')
      return
    }

    detecting.value = true
    activeTab.value = 'result'

    try {
      const uids = parseUIDs()
      const result = await fetchLookupFbUsers(uids)
      detectResult.value = result.users
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
    if (selectedAdAccounts.value.length === 0) {
      ElMessage.warning('请先在表格中选择广告账户')
      return
    }
    if (!uidInput.value.trim()) {
      ElMessage.warning('请先输入Facebook账号UID或主页地址')
      return
    }

    submitting.value = true
    activeTab.value = 'result'

    try {
      const uids = parseUIDs()
      const role = roleMap[authType.value] || 'ADMIN'

      // 对每个 UID 执行授权
      let allResults: FbAssignUserResponse | null = null
      for (const uid of uids) {
        const result = await fetchAssignAdAccountUser({
          adAccountIds: selectedAdAccounts.value.map((acc) => acc.id),
          userId: uid,
          role
        })

        if (!allResults) {
          allResults = result
        } else {
          allResults.results.push(...result.results)
          allResults.success += result.success
          allResults.failed += result.failed
          allResults.total += result.total
        }
      }

      assignResult.value = allResults

      if (allResults && allResults.failed === 0) {
        ElMessage.success(`授权成功！共 ${allResults.success} 个账户`)
      } else if (allResults && allResults.success > 0) {
        ElMessage.warning(`部分授权成功：${allResults.success} 成功，${allResults.failed} 失败`)
      } else {
        ElMessage.error('授权失败，请检查权限或重试')
      }
    } catch {
      ElMessage.error('授权失败，请重试')
      assignResult.value = null
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

    /* 权限类型选择 */
    .auth-select-group {
      margin-bottom: 20px;

      .auth-select {
        width: 100%;
      }
    }

    /* 步骤 */
    .auth-step {
      display: flex;
      gap: 12px;
      margin-bottom: 16px;
    }

    .auth-step:last-child {
      margin-bottom: 0;
    }

    .auth-step-indicator {
      display: flex;
      flex-direction: column;
      align-items: center;
      flex-shrink: 0;
      width: 24px;
    }

    .auth-step-num {
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
      line-height: 1;
      flex-shrink: 0;
    }

    .auth-step-line {
      flex: 1;
      width: 2px;
      min-height: 12px;
      background-color: var(--el-border-color);
      margin-top: 4px;
    }

    .auth-step:last-child .auth-step-line {
      display: none;
    }

    .auth-step-body {
      flex: 1;
      min-width: 0;
      padding-bottom: 20px;
    }

    .auth-step:last-child .auth-step-body {
      padding-bottom: 0;
    }

    .auth-step-head {
      display: flex;
      align-items: center;
      gap: 12px;
      min-height: 24px;
      margin-bottom: 10px;

      :deep(.el-checkbox) {
        height: 24px;
        line-height: 24px;
      }
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
      padding-left: 36px;
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

      .result-section-title {
        font-size: 14px;
        font-weight: 600;
        color: var(--el-text-color-primary);
        margin-bottom: 12px;
        padding-bottom: 8px;
        border-bottom: 1px solid var(--el-border-color-lighter);
      }

      .user-info {
        display: flex;
        align-items: center;
        gap: 8px;

        .user-avatar {
          width: 24px;
          height: 24px;
          border-radius: 50%;
          object-fit: cover;
        }
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
