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

        <!-- 输入目标账号 -->
        <div class="auth-field">
          <div class="auth-field-label">
            <span class="required-star">*</span>
            {{ $t('menus.addAuth.step1Label') }}
          </div>
          <ElInput
            v-model="uidInput"
            type="textarea"
            :rows="5"
            :placeholder="$t('menus.addAuth.step1Placeholder')"
            class="uid-textarea"
          />
        </div>

        <!-- 检测好友关系 -->
        <div class="auth-field">
          <div class="auth-field-label">{{ $t('menus.addAuth.step2Label') }}</div>
          <ElButton type="success" class="detect-btn" :loading="detecting" @click="handleDetect">
            {{ detectedOnce ? $t('menus.addAuth.reDetect') : $t('menus.addAuth.step2Button') }}
          </ElButton>
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
        <!-- 检测结果：一个账号一个方块 -->
        <template v-if="detectResult">
          <div class="result-section-title">{{ $t('menus.addAuth.lookupResult') }}</div>
          <div class="result-cards">
            <div v-for="user in detectResult" :key="user.uid" class="result-card">
              <div class="result-row">
                <span class="result-label">{{ $t('menus.addAuth.cardStatus') }}</span>
                <ElTag :type="user.isFriend ? 'success' : 'danger'" size="small">
                  {{
                    user.isFriend ? $t('menus.addAuth.assignOk') : $t('menus.addAuth.assignFail')
                  }}
                </ElTag>
              </div>
              <div class="result-row">
                <span class="result-label">
                  {{
                    user.isFriend
                      ? $t('menus.addAuth.cardSuccessAccount')
                      : $t('menus.addAuth.cardFailAccount')
                  }}
                </span>
                <span class="result-value link">{{ user.rawInput }}</span>
              </div>
              <div class="result-row">
                <span class="result-label">{{ $t('menus.addAuth.cardCurrentAccount') }}</span>
                <span class="result-value link">{{ fbProfileUrl(user.uid) }}</span>
              </div>
              <div class="result-row">
                <span class="result-label">{{ $t('menus.addAuth.cardMessage') }}</span>
                <span class="result-value">
                  {{ user.isFriend ? $t('menus.addAuth.friendYes') : $t('menus.addAuth.friendNo') }}
                </span>
              </div>
            </div>
          </div>
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
          <div class="result-cards">
            <div
              v-for="(item, idx) in assignResult.results"
              :key="`${item.adAccountId}-${idx}`"
              class="result-card"
            >
              <div class="result-row">
                <span class="result-label">{{ $t('menus.addAuth.assignStatus') }}</span>
                <ElTag :type="item.success ? 'success' : 'danger'" size="small">
                  {{ item.success ? $t('menus.addAuth.assignOk') : $t('menus.addAuth.assignFail') }}
                </ElTag>
              </div>
              <div class="result-row">
                <span class="result-label">{{ $t('menus.addAuth.cardAccount') }}</span>
                <span class="result-value link">{{ item.adAccountId }}</span>
              </div>
              <div class="result-row">
                <span class="result-label">{{ $t('menus.addAuth.cardMessage') }}</span>
                <span class="result-value">{{ item.message || '-' }}</span>
              </div>
            </div>
          </div>
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

  // 检测结果（rawInput 为用户输入的原始行，用于结果展示）
  interface DetectResultItem extends FbLookupUserResult {
    rawInput: string
  }
  const detectResult = ref<DetectResultItem[] | null>(null)

  // 是否已完成过一次检测（控制按钮文案：开始检测 → 重新检测）
  const detectedOnce = ref(false)

  // 授权结果
  const assignResult = ref<FbAssignUserResponse | null>(null)

  // authType → FB role 映射
  const roleMap: Record<string, 'ADMIN' | 'ADVERTISER' | 'ANALYST'> = {
    authorizeAdmin: 'ADMIN',
    authorizeAdManager: 'ADVERTISER',
    authorizeAdAnalyst: 'ANALYST'
  }

  // 解析输入行：提取 UID，并保留原始输入用于结果展示
  const parseUIDPairs = (): { raw: string; uid: string }[] => {
    return uidInput.value
      .split('\n')
      .map((l) => l.trim())
      .filter(Boolean)
      .map((line) => {
        // 从 URL 中提取 UID
        const urlMatch = line.match(/facebook\.com\/profile\.php\?id=(\d+)/)
        if (urlMatch) return { raw: line, uid: urlMatch[1] }
        // 如果是纯数字 UID，直接返回
        if (/^\d+$/.test(line)) return { raw: line, uid: line }
        // 其他 URL 格式，提取路径部分
        const pathMatch = line.match(/facebook\.com\/([^/?]+)/)
        if (pathMatch) return { raw: line, uid: pathMatch[1] }
        return { raw: line, uid: line }
      })
  }

  // 解析输入的 UID 列表
  const parseUIDs = (): string[] => parseUIDPairs().map((p) => p.uid)

  // 生成 Facebook 主页地址
  const fbProfileUrl = (uid: string): string => {
    if (/^\d+$/.test(uid)) return `https://www.facebook.com/profile.php?id=${uid}`
    return `https://www.facebook.com/${uid}`
  }

  // ==================== 检测好友关系 ====================
  const handleDetect = async () => {
    if (!uidInput.value.trim()) {
      ElMessage.warning('请先输入Facebook账号UID或主页地址')
      return
    }

    detecting.value = true

    try {
      const pairs = parseUIDPairs()
      const result = await fetchLookupFbUsers(pairs.map((p) => p.uid))
      // 把原始输入行关联回检测结果（优先按 UID 匹配，退化为按顺序）
      detectResult.value = result.users.map((u, idx) => ({
        ...u,
        rawInput: pairs.find((p) => p.uid === u.uid)?.raw ?? pairs[idx]?.raw ?? u.uid
      }))
      detectedOnce.value = true
      // 等结果返回后再跳转到结果标签页
      activeTab.value = 'result'
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

      /* 结果方块卡片 */
      .result-cards {
        display: flex;
        flex-direction: column;
        gap: 12px;

        .result-card {
          padding: 12px 16px;
          border: 1px solid var(--el-border-color-lighter);
          border-radius: 6px;

          .result-row {
            display: flex;
            align-items: center;
            gap: 12px;
            padding: 6px 0;

            .result-label {
              flex-shrink: 0;
              width: 64px;
              font-size: 13px;
              color: var(--el-text-color-secondary);
            }

            .result-value {
              font-size: 13px;
              color: var(--el-text-color-primary);
              word-break: break-all;

              &.link {
                color: var(--el-color-primary);
              }
            }
          }
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
