<!-- 广告账户管理页面 -->
<template>
  <div class="ad-account-manage-page art-full-height">
    <ElCard class="art-table-card">
      <!-- 表格头部 -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton @click="showDialog('add')" v-ripple>{{ $t('menus.adAccount.addAccount') }}</ElButton>
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
    </ElCard>

    <!-- 新增/编辑弹窗 -->
    <ElDialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? $t('menus.adAccount.addAccount') : $t('menus.adAccount.editAccount')"
      width="520px"
      :close-on-click-modal="false"
    >
      <ElForm ref="formRef" :model="formData" :rules="formRules" label-width="100px">
        <ElFormItem :label="$t('menus.adAccount.accountName')" prop="accountName">
          <ElInput v-model="formData.accountName" :placeholder="$t('menus.adAccount.accountNamePlaceholder')" />
        </ElFormItem>
        <ElFormItem :label="$t('menus.adAccount.accountId')" prop="accountId">
          <ElInput v-model="formData.accountId" :placeholder="$t('menus.adAccount.accountIdPlaceholder')" />
        </ElFormItem>
        <ElFormItem :label="$t('menus.adAccount.platform')" prop="platform">
          <ElSelect v-model="formData.platform" :placeholder="$t('menus.adAccount.platformPlaceholder')">
            <ElOption label="Facebook" value="facebook" />
            <ElOption label="Google Ads" value="google" />
            <ElOption label="TikTok" value="tiktok" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="$t('menus.adAccount.status')" prop="status">
          <ElSelect v-model="formData.status" :placeholder="$t('menus.adAccount.statusPlaceholder')">
            <ElOption :label="$t('menus.adAccount.statusActive')" :value="1" />
            <ElOption :label="$t('menus.adAccount.statusDisabled')" :value="0" />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <div class="dialog-footer">
          <ElButton @click="dialogVisible = false">{{ $t('common.cancel') }}</ElButton>
          <ElButton type="primary" @click="handleSubmit">{{ $t('common.confirm') }}</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
import { useTable } from '@/hooks/core/useTable'
import { ElTag, ElMessageBox, ElMessage } from 'element-plus'
import type { FormRules, FormInstance } from 'element-plus'
import { DialogType } from '@/types'

defineOptions({ name: 'AdAccountManage' })

// 广告账户数据类型
interface AdAccountItem {
  id: number
  accountName: string
  accountId: string
  platform: string
  status: number
  spend: number
  createTime: string
}

// Mock 数据
const mockData: AdAccountItem[] = [
  { id: 1, accountName: '主广告账户', accountId: 'act_123456789', platform: 'facebook', status: 1, spend: 12500.50, createTime: '2026-05-15T10:30:00Z' },
  { id: 2, accountName: '推广账户A', accountId: 'act_987654321', platform: 'google', status: 1, spend: 8320.00, createTime: '2026-05-20T14:20:00Z' },
  { id: 3, accountName: '测试账户', accountId: 'act_555555555', platform: 'facebook', status: 0, spend: 1200.75, createTime: '2026-06-01T09:00:00Z' }
]

// Mock API 适配器
const fetchPaged = async (_params: any) => {
  return { list: mockData, total: mockData.length, page: 1, size: mockData.length }
}

// 弹窗相关
const dialogType = ref<DialogType>('add')
const dialogVisible = ref(false)
const formRef = ref<FormInstance>()

const formData = reactive({
  accountName: '',
  accountId: '',
  platform: 'facebook',
  status: 1
})

const formRules: FormRules = {
  accountName: [
    { required: true, message: '请输入账户名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  accountId: [
    { required: true, message: '请输入账户ID', trigger: 'blur' }
  ],
  platform: [
    { required: true, message: '请选择平台', trigger: 'change' }
  ],
  status: [
    { required: true, message: '请选择状态', trigger: 'change' }
  ]
}

// 平台标签映射
const platformConfig: Record<string, { type: 'primary' | 'success' | 'warning' | 'info' | 'danger'; text: string }> = {
  facebook: { type: 'primary', text: 'Facebook' },
  google: { type: 'success', text: 'Google Ads' },
  tiktok: { type: 'warning', text: 'TikTok' }
}

// 状态配置
const getStatusConfig = (status: number) => {
  return status === 1
    ? { type: 'success' as const, text: '启用' }
    : { type: 'danger' as const, text: '禁用' }
}

// useTable 配置
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
    apiFn: fetchPaged,
    apiParams: { current: 1, size: 20 },
    columnsFactory: () => [
      { type: 'index', label: '#', width: 60 },
      { prop: 'accountName', label: '账户名称', minWidth: 160 },
      { prop: 'accountId', label: '账户ID', minWidth: 160 },
      {
        prop: 'platform',
        label: '平台',
        width: 120,
        formatter: (row: AdAccountItem) => {
          const config = platformConfig[row.platform] || { type: 'info', text: row.platform }
          return h(ElTag, { type: config.type, size: 'small' }, () => config.text)
        }
      },
      { prop: 'spend', label: '消耗金额', width: 120, formatter: (row: AdAccountItem) => `$${row.spend.toLocaleString()}` },
      {
        prop: 'status',
        label: '状态',
        width: 80,
        formatter: (row: AdAccountItem) => {
          const config = getStatusConfig(row.status)
          return h(ElTag, { type: config.type, size: 'small' }, () => config.text)
        }
      },
      {
        prop: 'createTime',
        label: '创建时间',
        minWidth: 170,
        formatter: (row: AdAccountItem) => {
          const d = new Date(row.createTime)
          const pad = (n: number) => String(n).padStart(2, '0')
          return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
        }
      },
      {
        label: '操作',
        width: 120,
        fixed: 'right',
        formatter: (row: AdAccountItem) => h('div', [
          h(ArtButtonTable, { type: 'edit', onClick: () => showDialog('edit', row) }),
          h(ArtButtonTable, { type: 'delete', onClick: () => handleDelete(row) })
        ])
      }
    ]
  }
})

// 显示弹窗
function showDialog(type: DialogType, row?: AdAccountItem) {
  dialogType.value = type
  if (type === 'edit' && row) {
    formData.accountName = row.accountName
    formData.accountId = row.accountId
    formData.platform = row.platform
    formData.status = row.status
  } else {
    formData.accountName = ''
    formData.accountId = ''
    formData.platform = 'facebook'
    formData.status = 1
  }
  dialogVisible.value = true
}

// 提交表单
async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate((valid) => {
    if (!valid) return
    const action = dialogType.value === 'add' ? '新增' : '编辑'
    ElMessage.success(`${action}成功`)
    dialogVisible.value = false
  })
}

// 删除
function handleDelete(row: AdAccountItem) {
  ElMessageBox.confirm(`确定删除账户「${row.accountName}」吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.success('删除成功')
  })
}
</script>

<style lang="scss" scoped>
.ad-account-manage-page {
  padding: 0;
}
</style>
