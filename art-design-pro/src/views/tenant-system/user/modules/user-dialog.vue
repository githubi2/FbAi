<template>
  <ElDialog
    v-model="dialogVisible"
    :title="dialogType === 'add' ? '添加用户' : '编辑用户'"
    width="36%"
    align-center
  >
    <ElForm ref="formRef" :model="formData" :rules="rules" label-width="90px">
      <ElFormItem label="用户名" prop="userName">
        <ElInput v-model="formData.userName" placeholder="请输入用户名" />
      </ElFormItem>

      <!-- 新增时的密码 -->
      <ElFormItem v-if="dialogType === 'add'" label="密码" prop="password">
        <ElInput
          v-model="formData.password"
          :type="showPassword ? 'text' : 'password'"
          placeholder="请输入密码"
        >
          <template #suffix>
            <ElButton link type="primary" size="small" @click="generatePassword('password')">
              生成
            </ElButton>
            <ElButton link size="small" @click="showPassword = !showPassword">
              <ArtSvgIcon :icon="showPassword ? 'ri:eye-off-line' : 'ri:eye-line'" />
            </ElButton>
          </template>
        </ElInput>
      </ElFormItem>

      <!-- 编辑时的修改密码 -->
      <template v-if="dialogType === 'edit'">
        <ElFormItem label="新密码" prop="newPassword">
          <ElInput
            v-model="formData.newPassword"
            :type="showNewPassword ? 'text' : 'password'"
            placeholder="留空则不修改密码"
          >
            <template #suffix>
              <ElButton link type="primary" size="small" @click="generatePassword('newPassword')">
                生成
              </ElButton>
              <ElButton link size="small" @click="showNewPassword = !showNewPassword">
                <ArtSvgIcon :icon="showNewPassword ? 'ri:eye-off-line' : 'ri:eye-line'" />
              </ElButton>
            </template>
          </ElInput>
        </ElFormItem>
        <ElFormItem v-if="formData.newPassword" label="确认密码" prop="confirmPassword">
          <ElInput
            v-model="formData.confirmPassword"
            :type="showConfirmPassword ? 'text' : 'password'"
            placeholder="请再次输入新密码"
          >
            <template #suffix>
              <ElButton link size="small" @click="showConfirmPassword = !showConfirmPassword">
                <ArtSvgIcon :icon="showConfirmPassword ? 'ri:eye-off-line' : 'ri:eye-line'" />
              </ElButton>
            </template>
          </ElInput>
        </ElFormItem>
      </template>

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
  const roleList = ref<Api.SystemManage.RoleListItem[]>([])

  const showPassword = ref(false)
  const showNewPassword = ref(false)
  const showConfirmPassword = ref(false)

  const formData = reactive({
    userName: '',
    password: '',
    newPassword: '',
    confirmPassword: '',
    roleId: undefined as number | undefined,
    status: true
  })

  const validateConfirmPassword = (_rule: any, value: string, callback: any) => {
    if (value && value !== formData.newPassword) {
      callback(new Error('两次输入的密码不一致'))
    } else {
      callback()
    }
  }

  const rules: FormRules = {
    userName: [
      { required: true, message: '请输入用户名', trigger: 'blur' },
      { min: 2, max: 20, message: '长度在 2 到 20 个字符', trigger: 'blur' }
    ],
    password: [
      { required: true, message: '请输入密码', trigger: 'blur' },
      { min: 6, max: 32, message: '长度在 6 到 32 个字符', trigger: 'blur' }
    ],
    newPassword: [{ min: 6, max: 32, message: '长度在 6 到 32 个字符', trigger: 'blur' }],
    confirmPassword: [{ validator: validateConfirmPassword, trigger: 'blur' }],
    roleId: [{ required: true, message: '请选择角色', trigger: 'change' }]
  }

  const generatePassword = (field: 'password' | 'newPassword') => {
    const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789!@#$'
    let pwd = ''
    for (let i = 0; i < 12; i++) {
      pwd += chars[Math.floor(Math.random() * chars.length)]
    }
    formData[field] = pwd
    if (field === 'newPassword') {
      formData.confirmPassword = pwd
    }
  }

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
      newPassword: '',
      confirmPassword: '',
      roleId: isEdit && row ? (row.roleId || row.role_id) : undefined,
      status: isEdit && row ? (row.status === 1 || row.status === '1' || row.status === true) : true
    })
    showPassword.value = false
    showNewPassword.value = false
    showConfirmPassword.value = false
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
        const data = { ...formData }
        if (props.type === 'edit' && data.newPassword) {
          data.password = data.newPassword
        }
        emit('submit', data)
      }
    })
  }
</script>
