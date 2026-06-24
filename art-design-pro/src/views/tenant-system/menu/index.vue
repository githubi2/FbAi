<!-- 租户菜单管理页面（只读查看） -->
<template>
  <div class="tenant-menu-page art-full-height">
    <ElCard class="art-table-card">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-base font-medium">菜单权限查看</span>
          <ElTag type="info">只读</ElTag>
        </div>
      </template>

      <ElTable
        :data="menuTreeData"
        row-key="id"
        border
        stripe
        style="width: 100%"
        default-expand-all
      >
        <ElTableColumn prop="title" label="菜单名称" minWidth="200">
          <template #default="{ row }">
            <div :style="{ paddingLeft: row._level * 24 + 'px' }" class="flex items-center gap-2">
              <ArtSvgIcon v-if="row.icon" :icon="formatIcon(row.icon)" class="text-base" />
              <span>{{ row.title }}</span>
              <ElTag v-if="row.menuType === 'directory'" size="small" type="info">目录</ElTag>
              <ElTag v-else-if="row.menuType === 'menu'" size="small">菜单</ElTag>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="name" label="路由名称" minWidth="150" />
        <ElTableColumn prop="path" label="路由路径" minWidth="180">
          <template #default="{ row }">
            <ElTag size="small" type="primary">{{ row.path }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="component" label="组件路径" minWidth="200">
          <template #default="{ row }">
            <code v-if="row.component" class="text-xs bg-gray-100 px-1 rounded">{{
              row.component
            }}</code>
            <span v-else class="text-gray-400">—</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="sortOrder" label="排序" width="80" align="center" />
      </ElTable>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { onMounted } from 'vue'
  import { fetchGetMenuList } from '@/api/system-manage'

  defineOptions({ name: 'TenantMenu' })

  interface FlatMenu {
    id: number
    title: string
    name: string
    path: string
    component: string
    icon: string
    menuType: string
    sortOrder: number
    _level: number
  }

  const menuTreeData = ref<FlatMenu[]>([])

  const formatIcon = (icon: string) => {
    if (!icon) return ''
    return icon.startsWith('ri:') ? icon : `ri:${icon}`
  }

  const flattenTree = (nodes: any[], level: number): FlatMenu[] => {
    const result: FlatMenu[] = []
    for (const node of nodes) {
      result.push({
        id: node.id,
        title: node.title || node.name || '',
        name: node.name || '',
        path: node.path || '',
        component: node.component || '',
        icon: node.icon || '',
        menuType: node.menuType || 'menu',
        sortOrder: node.sortOrder || 0,
        _level: level
      })
      if (node.children?.length) {
        result.push(...flattenTree(node.children, level + 1))
      }
    }
    return result
  }

  onMounted(async () => {
    try {
      const menus = await fetchGetMenuList()
      if (Array.isArray(menus)) {
        menuTreeData.value = flattenTree(menus, 0)
      }
    } catch {
      console.error('加载菜单列表失败')
    }
  })
</script>
