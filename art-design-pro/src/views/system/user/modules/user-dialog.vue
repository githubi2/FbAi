<template>
  <ElDialog
    v-model="dialogVisible"
    :title="dialogType === 'add' ? '添加用户' : '编辑用户'"
    width="30%"
    align-center
  >
    <ElForm ref="formRef" :model="formData" :rules="rules" label-width="80px">
      <ElFormItem label="用户名" prop="userName">
        <ElInput v-model="formData.userName" placeholder="请输入用户名" />
      </ElFormItem>
      <ElFormItem v-if="dialogType === 'add'" label="密码" prop="password">
        <ElInput v-model="formData.password" type="password" placeholder="请输入密码" />
      </ElFormItem>
      <ElFormItem label="角色" prop="roleId">
        <ElSelect v-model="formData.roleId" placeholder="请选择角色">
          <ElOption
            v-for="role in roleList"
            :key="role.roleId"
            :label="role.roleName"
            :value="role.roleId"
          />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="性别" prop="userGender">
        <ElSelect v-model="formData.userGender">
          <ElOption label="男" value="男" />
          <ElOption label="女" value="女" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="邮箱" prop="userEmail">
        <ElInput v-model="formData.userEmail" placeholder="请输入邮箱" />
      </ElFormItem>
      <ElFormItem label="状态" prop="status">
        <ElSwitch v-model="formData.status" active-text="启用" inactive-text="禁用" />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <div class="dialog-footer">
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSubmit">提交</ElButton>
      </div>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchGetRoleList } from '@/api/system-manage'

  interface Props {
    visible: boolean
    type: string
    userData?: any
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'submit', formData: any): void
  }

  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value)
  })

  const dialogType = computed(() => props.type)

  const formRef = ref<FormInstance>()

  // 角色列表
  const roleList = ref<Api.SystemManage.RoleListItem[]>([])

  const formData = reactive({
    userName: '',
    password: '',
    roleId: undefined as number | undefined,
    userGender: '男',
    userEmail: '',
    status: true
  })

  const rules: FormRules = {
    userName: [
      { required: true, message: '请输入用户名', trigger: 'blur' },
      { min: 2, max: 20, message: '长度在 2 到 20 个字符', trigger: 'blur' }
    ],
    password: [
      { required: true, message: '请输入密码', trigger: 'blur' },
      { min: 6, max: 32, message: '长度在 6 到 32 个字符', trigger: 'blur' }
    ],
    roleId: [
      { required: true, message: '请选择角色', trigger: 'change' }
    ]
  }

  // 加载角色列表（标准化后端 id → roleId）
  const loadRoles = async () => {
    try {
      const res = await fetchGetRoleList()
      let rawList: any[] = []
      if (Array.isArray(res)) {
        rawList = res
      } else if ((res as any)?.records) {
        rawList = (res as any).records
      } else if ((res as any)?.list) {
        rawList = (res as any).list
      }
      // 标准化字段：后端返回 id，前端类型用 roleId
      roleList.value = rawList.map((item: any) => ({
        ...item,
        roleId: item.id ?? item.roleId
      }))
    } catch (e) {
      console.error('加载角色列表失败:', e)
    }
  }

  const initFormData = () => {
    const isEdit = props.type === 'edit' && props.userData
    const row = props.userData

    Object.assign(formData, {
      userName: isEdit && row ? row.userName || '' : '',
      password: '',
      roleId: isEdit && row ? (row.roleId || row.role_id) : undefined,
      userGender: isEdit && row ? row.userGender || '男' : '男',
      userEmail: isEdit && row ? row.userEmail || row.email || '' : '',
      status: isEdit && row ? (row.status === 1 || row.status === '1' || row.status === true) : true
    })
  }

  watch(
    () => [props.visible, props.type, props.userData],
    ([visible]) => {
      if (visible) {
        loadRoles()
        initFormData()
        nextTick(() => {
          formRef.value?.clearValidate()
        })
      }
    },
    { immediate: true }
  )

  const handleSubmit = async () => {
    if (!formRef.value) return

    await formRef.value.validate((valid) => {
      if (valid) {
        emit('submit', { ...formData })
      }
    })
  }
</script>
