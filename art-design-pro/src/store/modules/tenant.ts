/**
 * 租户状态管理模块
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { fetchCurrentTenant, fetchSwitchTenant } from '@/api/tenant'

export const useTenantStore = defineStore(
  'tenant',
  () => {
    /** 当前租户ID (null = 全局视角) */
    const tenantId = ref<number | null>(null)

    /** 当前租户名称 */
    const tenantName = ref<string>('')

    /** 租户列表 */
    const tenantList = ref<Api.Tenant.TenantListItem[]>([])

    /** 是否全局视角 */
    const isGlobalView = computed(() => tenantId.value === null)

    /** 获取当前租户上下文 */
    async function loadCurrentTenant() {
      try {
        const data = await fetchCurrentTenant()
        tenantId.value = data.tenantId
        tenantName.value = data.tenantName || '全局视角'
      } catch {
        // ignore
      }
    }

    /** 切换租户 */
    async function switchTenant(id: number) {
      try {
        const data = await fetchSwitchTenant(id)
        tenantId.value = data.tenantId
        tenantName.value = data.tenantName
      } catch {
        // ignore
      }
    }

    /** 重置状态 */
    function reset() {
      tenantId.value = null
      tenantName.value = ''
      tenantList.value = []
    }

    return {
      tenantId,
      tenantName,
      tenantList,
      isGlobalView,
      loadCurrentTenant,
      switchTenant,
      reset
    }
  },
  {
    persist: {
      key: 'tenant-state',
      storage: localStorage
    }
  }
)
