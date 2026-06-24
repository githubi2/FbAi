<!-- 租户角色管理页面 -->
<template>
  <div class="art-full-height">
    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton @click="showDialog('add')" v-ripple>新增角色</ElButton>
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
    </ElCard>

    <!-- 角色编辑弹窗 -->
    <TenantRoleEditDialog
      v-model="dialogVisible"
      :dialog-type="dialogType"
      :role-data="currentRoleData"
      @success="refreshData"
    />

    <!-- 菜单权限弹窗 -->
    <TenantRolePermissionDialog
      v-model="permissionDialog"
      :role-data="currentRoleData"
      @success="refreshData"
    />
  </div>
</template>

<script setup lang="ts">
  import { h, onMounted } from 'vue'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetRoleList, fetchDeleteRole, fetchGetAllMenus } from '@/api/system-manage'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import TenantRoleEditDialog from './modules/role-edit-dialog.vue'
  import TenantRolePermissionDialog from './modules/role-permission-dialog.vue'
  import { ElTag, ElButton, ElMessageBox, ElMessage } from 'element-plus'

  defineOptions({ name: 'TenantRole' })

  type RoleListItem = Api.SystemManage.RoleListItem

  const menuNameMap = ref<Map<number, string>>(new Map())

  onMounted(async () => {
    try {
      const menus = await fetchGetAllMenus()
      const map = new Map<number, string>()
      if (Array.isArray(menus)) {
        for (const m of menus) {
          map.set(m.id, m.title)
        }
      }
      menuNameMap.value = map
    } catch {
      // 静默失败
    }
  })

  const dialogVisible = ref(false)
  const permissionDialog = ref(false)
  const currentRoleData = ref<RoleListItem | undefined>(undefined)
  const dialogType = ref<'add' | 'edit'>('add')

  const formatDate = (val: string) => {
    if (!val) return '-'
    const d = new Date(val)
    if (isNaN(d.getTime())) return val
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
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
      apiFn: fetchGetRoleList,
      apiParams: {
        current: 1,
        size: 20
      },
      columnsFactory: () => [
        {
          prop: 'roleId',
          label: '角色ID',
          width: 100
        },
        {
          prop: 'roleName',
          label: '角色名称',
          minWidth: 120
        },
        {
          prop: 'roleCode',
          label: '角色编码',
          minWidth: 120
        },
        {
          prop: 'description',
          label: '角色描述',
          minWidth: 150,
          showOverflowTooltip: true
        },
        {
          prop: 'menuIds',
          label: '权限',
          minWidth: 200,
          formatter: (row) => {
            const ids: number[] = row.menuIds || []
            if (!ids.length) return h('span', { class: 'text-gray-400 text-xs' }, '暂无权限')
            const map = menuNameMap.value
            return h(
              'div',
              { class: 'flex flex-wrap gap-1' },
              ids.map((id: number) => {
                const name = map.get(id) || `#${id}`
                return h(ElTag, { size: 'small' }, () => name)
              })
            )
          }
        },
        {
          prop: 'enabled',
          label: '角色状态',
          width: 100,
          formatter: (row) => {
            const isEnabled = row.enabled
            return h(
              ElButton,
              {
                size: 'small',
                type: isEnabled ? 'success' : 'warning',
                plain: true,
                disabled: true
              },
              () => (isEnabled ? '启用' : '禁用')
            )
          }
        },
        {
          prop: 'createTime',
          label: '创建日期',
          width: 180,
          sortable: true,
          formatter: (row) => formatDate(row.createTime)
        },
        {
          prop: 'operation',
          label: '操作',
          width: 160,
          fixed: 'right',
          formatter: (row) => {
            return h('div', { class: 'flex items-center' }, [
              h(ArtButtonTable, {
                type: 'view',
                icon: 'ri:shield-keyhole-line',
                onClick: () => showPermissionDialog(row)
              }),
              h(ArtButtonTable, {
                type: 'edit',
                onClick: () => showDialog('edit', row)
              }),
              h(ArtButtonTable, {
                type: 'delete',
                onClick: () => deleteRole(row)
              })
            ])
          }
        }
      ]
    },
    transform: {
      dataTransformer: (records) => {
        if (!Array.isArray(records)) return []
        return records.map((item: any) => ({
          ...item,
          roleId: item.id || item.roleId,
          roleName: item.roleName || '',
          roleCode: item.roleCode || '',
          description: item.description || '',
          enabled: item.status === 1 || item.enabled === true,
          createTime: item.createTime || '',
          menuIds: item.menuIds || []
        }))
      }
    }
  })

  const showDialog = (type: 'add' | 'edit', row?: RoleListItem) => {
    dialogVisible.value = true
    dialogType.value = type
    currentRoleData.value = row
  }

  const showPermissionDialog = (row?: RoleListItem) => {
    permissionDialog.value = true
    currentRoleData.value = row
  }

  const deleteRole = async (row: RoleListItem) => {
    try {
      await ElMessageBox.confirm(`确定删除角色"${row.roleName}"吗？此操作不可恢复！`, '删除确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      await fetchDeleteRole(row.roleId)
      ElMessage.success('删除成功')
      refreshData()
    } catch (error: any) {
      if (error !== 'cancel') {
        ElMessage.error('删除失败')
      }
    }
  }
</script>
