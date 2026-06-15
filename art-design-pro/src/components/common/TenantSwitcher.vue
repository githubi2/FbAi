<!-- 租户切换下拉（仅超级管理员可见） -->
<template>
  <ElDropdown
    v-if="isSuperAdmin"
    @command="handleSwitch"
    popper-class="tenant-dropdown"
    trigger="click"
  >
    <span class="tenant-switcher flex-c gap-1 px-2 py-1 text-xs c-p border border-g-300 rounded select-none">
      <ArtSvgIcon icon="ri:building-line" class="text-sm" />
      <span>{{ tenantStore.tenantName || '全局视角' }}</span>
      <ArtSvgIcon icon="ri:arrow-down-s-line" class="text-xs" />
    </span>
    <template #dropdown>
      <ElDropdownMenu>
        <ElDropdownItem command="0" :class="{ 'is-active': tenantStore.isGlobalView }">
          <ArtSvgIcon icon="ri:global-line" class="mr-1" />
          全局视角
        </ElDropdownItem>
        <ElDropdownItem
          v-for="t in tenantList"
          :key="t.id"
          :command="String(t.id)"
          :class="{ 'is-active': tenantStore.tenantId === t.id }"
        >
          <ArtSvgIcon icon="ri:building-line" class="mr-1" />
          {{ t.name }}
        </ElDropdownItem>
      </ElDropdownMenu>
    </template>
  </ElDropdown>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useTenantStore } from '@/store/modules/tenant'
import { fetchGetTenantList } from '@/api/tenant'
import { useUserStore } from '@/store/modules/user'

const tenantStore = useTenantStore()
const userStore = useUserStore()
const tenantList = ref<Api.Tenant.TenantListItem[]>([])

/** 是否超级管理员 */
const isSuperAdmin = computed(() => {
  const roles = userStore.info?.roles || []
  return roles.includes('R_SUPER')
})

async function loadTenants() {
  if (!isSuperAdmin.value) return
  try {
    tenantList.value = await fetchGetTenantList()
  } catch {
    // ignore
  }
}

async function handleSwitch(tenantId: string) {
  const id = Number(tenantId)
  await tenantStore.switchTenant(id)
  // 重新加载页面以刷新菜单和权限
  window.location.reload()
}

onMounted(() => {
  loadTenants()
  tenantStore.loadCurrentTenant()
})
</script>

<style scoped>
.tenant-switcher {
  transition: border-color 0.3s;
}
.tenant-switcher:hover {
  border-color: var(--el-color-primary);
}
.is-active {
  color: var(--el-color-primary);
  font-weight: bold;
}
</style>
