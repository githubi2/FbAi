<!-- 租户管理页面 -->
<template>
  <div class="tenant-page art-full-height">
    <ElCard class="art-table-card">
      <ArtTableHeader :loading="loading" @refresh="loadData">
        <template #left>
          <ElSpace wrap>
            <ElButton type="primary" @click="showCreateDialog" v-ripple>新增租户</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ElTable :data="tableData" v-loading="loading" stripe border>
        <ElTableColumn prop="id" label="ID" width="60" />
        <ElTableColumn prop="name" label="租户名称" min-width="150" />
        <ElTableColumn prop="code" label="租户编码" min-width="120" />
        <ElTableColumn label="状态" width="80">
          <template #default="{ row }">
            <ElTag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="contactName" label="联系人" min-width="100" />
        <ElTableColumn prop="contactPhone" label="联系电话" min-width="120" />
        <ElTableColumn prop="createTime" label="创建时间" min-width="160">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <ArtButtonTable type="edit" @click="showEditDialog(row)" />
            <ArtButtonTable type="delete" @click="handleDelete(row)" />
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>

    <!-- 租户弹窗 -->
    <TenantForm
      v-model:visible="dialogVisible"
      :type="dialogType"
      :tenant-data="currentTenant"
      @submit="handleDialogSubmit"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
import {
  fetchGetTenantList,
  fetchCreateTenant,
  fetchUpdateTenant,
  fetchDeleteTenant
} from '@/api/tenant'
import TenantForm from './modules/TenantForm.vue'
import { DialogType } from '@/types'

defineOptions({ name: 'Tenant' })

const loading = ref(false)
const tableData = ref<Api.Tenant.TenantListItem[]>([])
const dialogVisible = ref(false)
const dialogType = ref<DialogType>('add')
const currentTenant = ref<Partial<Api.Tenant.TenantListItem>>({})

function formatTime(val: string) {
  if (!val) return '-'
  const d = new Date(val)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

async function loadData() {
  loading.value = true
  try {
    tableData.value = await fetchGetTenantList()
  } finally {
    loading.value = false
  }
}

function showCreateDialog() {
  dialogType.value = 'add'
  currentTenant.value = {}
  dialogVisible.value = true
}

function showEditDialog(row: Api.Tenant.TenantListItem) {
  dialogType.value = 'edit'
  currentTenant.value = { ...row }
  dialogVisible.value = true
}

async function handleDelete(row: Api.Tenant.TenantListItem) {
  try {
    await ElMessageBox.confirm(`确定要删除租户「${row.name}」吗？此操作不可撤销，将级联删除该租户下所有数据。`, '警告', {
      type: 'warning',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    await fetchDeleteTenant(row.id)
    ElMessage.success('租户已删除')
    loadData()
  } catch {
    ElMessage.error('删除失败')
  }
}

async function handleDialogSubmit(formData: any) {
  try {
    if (dialogType.value === 'add') {
      await fetchCreateTenant(formData)
      ElMessage.success('租户创建成功')
    } else {
      await fetchUpdateTenant({ id: currentTenant.value.id!, ...formData })
      ElMessage.success('租户更新成功')
    }
    dialogVisible.value = false
    loadData()
  } catch {
    ElMessage.error('操作失败')
  }
}

onMounted(() => {
  loadData()
})
</script>
