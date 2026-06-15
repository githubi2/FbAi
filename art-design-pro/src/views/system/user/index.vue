<!-- 用户管理页面 -->
<template>
  <div class="user-page art-full-height">
    <!-- 搜索栏 -->
    <UserSearch v-model="searchForm" @search="handleSearch" @reset="resetSearchParams"></UserSearch>

    <ElCard class="art-table-card">
      <!-- 表格头部 -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton @click="showDialog('add')" v-ripple>新增用户</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- 表格 -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @selection-change="handleSelectionChange"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      >
      </ArtTable>

      <!-- 用户弹窗 -->
      <UserDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :user-data="currentUserData"
        @submit="handleDialogSubmit"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchGetUserList,
    fetchCreateUser,
    fetchUpdateUser,
    fetchDeleteUser
  } from '@/api/system-manage'
  import UserSearch from './modules/user-search.vue'
  import UserDialog from './modules/user-dialog.vue'
  import { ElTag, ElMessageBox, ElMessage } from 'element-plus'
  import { DialogType } from '@/types'

  defineOptions({ name: 'User' })

  type UserListItem = Api.SystemManage.UserListItem

  // 弹窗相关
  const dialogType = ref<DialogType>('add')
  const dialogVisible = ref(false)
  const currentUserData = ref<Partial<UserListItem>>({})

  // 选中行
  const selectedRows = ref<UserListItem[]>([])

  // 搜索表单
  const searchForm = ref({
    userName: undefined,
    status: undefined
  })

  // 用户状态配置（1:启用 0:禁用）
  const getUserStatusConfig = (status: number) => {
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
    getData,
    replaceSearchParams,
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchGetUserList,
      apiParams: {
        current: 1,
        size: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { type: 'selection' },
        { type: 'index', width: 55, label: '序号' },
        {
          prop: 'nickName',
          label: '用户名',
          minWidth: 120,
          formatter: (row) => row.nickName || row.userName || ''
        },
        {
          prop: 'userName',
          label: '账号',
          minWidth: 100,
          formatter: (row) => row.userName || ''
        },
        {
          prop: 'roleName',
          label: '角色',
          minWidth: 100,
          formatter: (row) => {
            const roleName = row.roleName || (row.userRoles?.[0]) || '—'
            return h(ElTag, { type: 'primary' }, () => roleName)
          }
        },
        {
          prop: 'status',
          label: '状态',
          minWidth: 90,
          formatter: (row) => {
            const statusNum = typeof row.status === 'number' ? row.status : parseInt(row.status) || 0
            const statusConfig = getUserStatusConfig(statusNum)
            return h(ElTag, { type: statusConfig.type }, () => statusConfig.text)
          }
        },
        {
          prop: 'createTime',
          label: '创建时间',
          minWidth: 160,
          sortable: true,
          formatter: (row) => {
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
          formatter: (row) =>
            h('div', [
              h(ArtButtonTable, {
                type: 'edit',
                onClick: () => showDialog('edit', row)
              }),
              h(ArtButtonTable, {
                type: 'delete',
                onClick: () => deleteUser(row)
              })
            ])
        }
      ]
    },
    transform: {
      dataTransformer: (records) => {
        if (!Array.isArray(records)) {
          return []
        }

        return records.map((item: any) => {
          return {
            ...item,
            // 后端返回字段映射
            userName: item.userName || item.username || '',
            nickName: item.nickName || '',
            roleName: item.roleName || (item.userRoles?.[0]) || '',
            roleId: item.roleId || item.role_id,
            status: Number(item.status),
            createTime: item.createTime || '',
            updateTime: item.updateTime || ''
          }
        })
      }
    }
  })

  const handleSearch = (params: Api.SystemManage.UserSearchParams) => {
    replaceSearchParams(params)
    getData()
  }

  const showDialog = (type: DialogType, row?: UserListItem): void => {
    console.log('打开弹窗:', { type, row })
    dialogType.value = type
    currentUserData.value = row || {}
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  /** 删除用户 */
  const deleteUser = async (row: UserListItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要注销该用户吗？`, '注销用户', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'error'
      })
      await fetchDeleteUser(row.id)
      ElMessage.success('注销成功')
      refreshData()
    } catch (error: any) {
      if (error !== 'cancel') {
        ElMessage.error('注销失败')
      }
    }
  }

  /** 处理弹窗提交事件 */
  const handleDialogSubmit = async (formData: any) => {
    try {
      if (dialogType.value === 'add') {
        await fetchCreateUser({
          userName: formData.userName,
          password: formData.password || '123456',
          status: formData.status ? 1 : 0,
          roleId: formData.roleId
        })
        ElMessage.success('添加成功')
      } else {
        const userId = currentUserData.value.id
        if (userId) {
          const updateData: any = {
            status: formData.status ? 1 : 0,
            roleId: formData.roleId
          }
          // 如果填了新密码，传给后端
          if (formData.password || formData.newPassword) {
            updateData.password = formData.password || formData.newPassword
          }
          await fetchUpdateUser(userId, updateData)
          ElMessage.success('更新成功')
        }
      }
      dialogVisible.value = false
      currentUserData.value = {}
      refreshData()
    } catch (error) {
      console.error('提交失败:', error)
      ElMessage.error('操作失败')
    }
  }

  const handleSelectionChange = (selection: UserListItem[]): void => {
    selectedRows.value = selection
  }
</script>
