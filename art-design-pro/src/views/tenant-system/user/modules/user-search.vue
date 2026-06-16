<!-- 租户用户搜索栏 -->
<template>
  <ElCard class="search-card">
    <ElForm :model="localForm" inline>
      <ElFormItem label="用户名">
        <ElInput v-model="localForm.userName" placeholder="请输入用户名" clearable />
      </ElFormItem>
      <ElFormItem label="状态">
        <ElSelect v-model="localForm.status" placeholder="请选择" clearable style="width: 120px">
          <ElOption label="启用" value="1" />
          <ElOption label="禁用" value="0" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem>
        <ElSpace>
          <ElButton @click="handleSearch" v-ripple>查询</ElButton>
          <ElButton @click="handleReset" v-ripple>重置</ElButton>
        </ElSpace>
      </ElFormItem>
    </ElForm>
  </ElCard>
</template>

<script setup lang="ts">
  interface SearchForm {
    userName?: string
    status?: string
  }

  interface Props {
    modelValue: SearchForm
  }

  interface Emits {
    (e: 'update:modelValue', value: SearchForm): void
    (e: 'search', value: SearchForm): void
    (e: 'reset'): void
  }

  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  const localForm = reactive<SearchForm>({ ...props.modelValue })

  watch(
    () => props.modelValue,
    (val) => {
      Object.assign(localForm, val)
    }
  )

  const handleSearch = () => {
    const params = { ...localForm }
    emit('update:modelValue', params)
    emit('search', params)
  }

  const handleReset = () => {
    const empty: SearchForm = { userName: undefined, status: undefined }
    Object.assign(localForm, empty)
    emit('update:modelValue', empty)
    emit('reset')
  }
</script>

<style lang="scss" scoped>
  .search-card {
    margin-bottom: 12px;

    :deep(.el-card__body) {
      padding-bottom: 0;
    }
  }
</style>
