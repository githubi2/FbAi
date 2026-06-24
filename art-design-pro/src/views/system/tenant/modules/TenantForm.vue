<!-- 租户创建/编辑弹窗 -->
<template>
  <ElDialog
    :model-value="visible"
    :title="type === 'add' ? '新增租户' : '编辑租户'"
    width="580px"
    :close-on-click-modal="false"
    @update:model-value="$emit('update:visible', $event)"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="100px">
      <ElFormItem label="租户名称" prop="name">
        <ElInput v-model="form.name" placeholder="请输入租户名称" maxlength="128" />
      </ElFormItem>

      <ElFormItem v-if="type === 'add'" label="租户编码" prop="code">
        <ElInput v-model="form.code" placeholder="英文编码，如 company_x" maxlength="64" />
      </ElFormItem>

      <ElFormItem label="联系方式" prop="contactPhone">
        <ElInput v-model="form.contactPhone" placeholder="请输入联系电话" maxlength="20" />
      </ElFormItem>

      <ElFormItem v-if="type === 'edit'" label="状态" prop="status">
        <ElSwitch v-model="form.status" :active-value="1" :inactive-value="0" />
      </ElFormItem>

      <ElFormItem label="备注" prop="description">
        <ElInput
          v-model="form.description"
          type="textarea"
          :rows="3"
          placeholder="请输入备注信息"
          maxlength="256"
        />
      </ElFormItem>

      <!-- 创建租户时需填写管理员账号 -->
      <template v-if="type === 'add'">
        <ElDivider content-position="left">租户管理员账号</ElDivider>

        <ElFormItem label="管理员用户名" prop="adminUserName">
          <ElInput v-model="form.adminUserName" placeholder="请输入管理员用户名" maxlength="64" />
        </ElFormItem>

        <ElFormItem label="管理员密码" prop="adminPassword">
          <ElInput
            v-model="form.adminPassword"
            type="password"
            placeholder="至少6位密码"
            maxlength="32"
            show-password
          />
        </ElFormItem>

        <ElFormItem label="管理员昵称" prop="adminNickName">
          <ElInput v-model="form.adminNickName" placeholder="请输入管理员昵称" maxlength="64" />
        </ElFormItem>
      </template>
    </ElForm>

    <template #footer>
      <ElButton @click="$emit('update:visible', false)">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">确定</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ref, reactive, watch } from 'vue'
  import type { FormInstance, FormRules } from 'element-plus'
  import { DialogType } from '@/types'

  interface Props {
    visible: boolean
    type: DialogType
    tenantData?: Partial<Api.Tenant.TenantListItem>
  }

  interface Emits {
    (e: 'update:visible', val: boolean): void
    (e: 'submit', data: any): void
  }

  const props = withDefaults(defineProps<Props>(), {
    tenantData: () => ({})
  })
  const emit = defineEmits<Emits>()

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive<Record<string, any>>({
    name: '',
    code: '',
    contactPhone: '',
    description: '',
    status: 1,
    adminUserName: '',
    adminPassword: '',
    adminNickName: ''
  })

  const rules: FormRules = {
    name: [
      { required: true, message: '请输入租户名称', trigger: 'blur' },
      { min: 2, max: 128, message: '长度在 2 到 128 个字符', trigger: 'blur' }
    ],
    code: [
      { required: true, message: '请输入租户编码', trigger: 'blur' },
      { min: 2, max: 64, message: '长度在 2 到 64 个字符', trigger: 'blur' },
      {
        pattern: /^[a-z][a-z0-9_]*$/,
        message: '必须小写字母开头，只允许小写字母/数字/下划线',
        trigger: 'blur'
      }
    ],
    contactPhone: [
      { pattern: /^[\d\-+() ]{0,20}$/, message: '请输入有效的电话号码', trigger: 'blur' }
    ],
    contactEmail: [{ type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }],
    adminUserName: [
      { required: true, message: '请输入管理员用户名', trigger: 'blur' },
      { min: 2, max: 64, message: '长度在 2 到 64 个字符', trigger: 'blur' },
      {
        pattern: /^[a-zA-Z][a-zA-Z0-9_]*$/,
        message: '必须以字母开头，只允许字母/数字/下划线',
        trigger: 'blur'
      }
    ],
    adminPassword: [
      { required: true, message: '请输入管理员密码', trigger: 'blur' },
      { min: 6, max: 32, message: '密码长度 6-32 位', trigger: 'blur' }
    ],
    adminNickName: [{ max: 64, message: '长度不能超过 64 个字符', trigger: 'blur' }]
  }

  watch(
    () => props.visible,
    (val) => {
      if (val && props.type === 'edit' && props.tenantData) {
        form.name = props.tenantData.name || ''
        form.contactPhone = props.tenantData.contactPhone || ''
        form.description = props.tenantData.description || ''
        form.status = props.tenantData.status ?? 1
      } else if (val && props.type === 'add') {
        form.name = ''
        form.code = ''
        form.contactPhone = ''
        form.description = ''
        form.status = 1
        form.adminUserName = ''
        form.adminPassword = ''
        form.adminNickName = ''
      }
    }
  )

  async function handleSubmit() {
    if (!formRef.value) return
    try {
      await formRef.value.validate()
    } catch {
      return
    }
    submitting.value = true
    try {
      emit('submit', { ...form })
    } finally {
      submitting.value = false
    }
  }
</script>
