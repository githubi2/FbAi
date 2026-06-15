<!-- 租户管理页面 -->
<template>
  <div class="tenant-page art-full-height">
    <ElCard class="art-table-card">
      <!-- 表格头部 -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton @click="showDialog('add')" v-ripple>新增租户</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- 表格 -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />

      <!-- 租户弹窗 -->
      <TenantForm
        v-model:visible="dialogVisible"
        :type="dialogType"
        :tenant-data="currentTenant"
        @submit="handleDialogSubmit"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, h } from 'vue'
import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
import { useTable } from '@/hooks/core/useTable'
import {
  fetchGetTenantList,
  fetchCreateTenant,
  fetchUpdateTenant,
  fetchDeleteTenant
} from '@/api/tenant'
import TenantForm from './modules/TenantForm.vue'
import { ElTag, ElMessageBox, ElMessage } from 'element-plus'
import { DialogType } from '@/types'

defineOptions({ name: 'Tenant' })

type TenantListItem = Api.Tenant.TenantListItem

// 包装为分页格式（useTable 要求分页响应）
const fetchTenantListPaged = async (_params: any) => {
  const list = await fetchGetTenantList()
  return { list: list || [], total: (list || []).length, page: 1, size: (list || []).length || 20 }
}

// 弹窗相关
const dialogType = ref<DialogType>('add')
const dialogVisible = ref(false)
const currentTenant = ref<Partial<TenantListItem>>({})

// 租户状态配置
const getStatusConfig = (status: number) => {
  return status === 1
    ? { type: 'success' as const, text: '启用' }
    : { type: 'danger' as const, text: '禁用' }
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
    apiFn: fetchTenantListPaged,
    apiParams: {
      current: 1,
      size: 20
    },
    columnsFactory: () => [
      { type: 'index', width: 55, label: '序号' },
      {
        prop: 'name',
        label: '租户名称',
        minWidth: 150,
        formatter: (row: any) => row.name || '—'
      },
      {
        prop: 'code',
        label: '租户编码',
        minWidth: 120,
        formatter: (row: any) => row.code || '—'
      },
      {
        prop: 'status',
        label: '状态',
        minWidth: 90,
        formatter: (row: any) => {
          const config = getStatusConfig(Number(row.status))
          return h(ElTag, { type: config.type }, () => config.text)
        }
      },
      {
        prop: 'contactName',
        label: '联系人',
        minWidth: 100,
        formatter: (row: any) => row.contactName || '—'
      },
      {
        prop: 'contactPhone',
        label: '联系电话',
        minWidth: 120,
        formatter: (row: any) => row.contactPhone || '—'
      },
      {
        prop: 'createTime',
        label: '创建时间',
        minWidth: 160,
        sortable: true,
        formatter: (row: any) => {
          const val = row.createTime
          if (!val) return '—'
          const d = new Date(val)
          if (isNaN(d.getTime())) return val
          const pad = (n: number) => String(n).padStart(2, '0')
          return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
        }
      },
      {
        prop: 'operation',
        label: '操作',
        width: 120,
        fixed: 'right',
        formatter: (row: any) =>
          h('div', [
            h(ArtButtonTable, {
              type: 'edit',
              onClick: () => showDialog('edit', row)
            }),
            h(ArtButtonTable, {
              type: 'delete',
              onClick: () => deleteTenant(row)
            })
          ])
      }
    ]
  }
})

const showDialog = (type: DialogType, row?: TenantListItem): void => {
  dialogType.value = type
  currentTenant.value = row || {}
  nextTick(() => {
    dialogVisible.value = true
  })
}

const deleteTenant = async (row: TenantListItem): Promise<void> => {
  try {
    await ElMessageBox.confirm(
      `确定要删除租户「${row.name}」吗？此操作将级联删除该租户下所有数据，不可撤销。`,
      '删除租户',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'error'
      }
    )
    await fetchDeleteTenant(row.id)
    ElMessage.success('租户已删除')
    refreshData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleDialogSubmit = async (formData: any) => {
  try {
    if (dialogType.value === 'add') {
      await fetchCreateTenant({
        name: formData.name,
        code: formData.code,
        contactName: formData.contactName,
        contactPhone: formData.contactPhone,
        contactEmail: formData.contactEmail,
        description: formData.description,
        adminUserName: formData.adminUserName,
        adminPassword: formData.adminPassword,
        adminNickName: formData.adminNickName
      })
      ElMessage.success('租户创建成功')
    } else {
      const id = currentTenant.value.id
      if (id) {
        await fetchUpdateTenant({
          id,
          name: formData.name,
          contactName: formData.contactName,
          contactPhone: formData.contactPhone,
          contactEmail: formData.contactEmail,
          description: formData.description,
          status: formData.status
        })
        ElMessage.success('租户更新成功')
      }
    }
    dialogVisible.value = false
    currentTenant.value = {}
    refreshData()
  } catch {
    ElMessage.error('操作失败')
  }
}
</script>
