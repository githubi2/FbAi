<!-- 编辑列面板（FB 广告管理工具列设置同款交互，样式对齐本项目） -->
<!-- 分组勾选显示/隐藏列，实时生效；与 useTable 的 columnChecks 双向绑定 -->
<template>
  <ElDialog
    :model-value="modelValue"
    :title="$t('menus.adCampaign.columnSetting')"
    width="620px"
    append-to-body
    destroy-on-close
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
  >
    <div class="column-setting">
      <div v-for="group in groupedColumns" :key="group.key" class="group-card">
        <div class="group-header">
          <span class="group-title">{{ $t(group.labelKey) }}</span>
          <span class="group-count">
            {{ $t('menus.adCampaign.selectedItems', { count: group.checkedCount }) }}
          </span>
        </div>
        <div class="group-body">
          <ElCheckbox
            v-for="col in group.columns"
            :key="col.prop"
            :model-value="getVisibility(col)"
            :disabled="col.disabled"
            class="group-checkbox"
            @update:model-value="(v: boolean | string | number) => toggle(col, !!v)"
          >
            {{ col.label }}
          </ElCheckbox>
        </div>
      </div>
    </div>
    <template #footer>
      <ElButton @click="handleReset">{{ $t('table.form.reset') }}</ElButton>
      <ElButton type="primary" @click="emit('update:modelValue', false)">
        {{ $t('common.confirm') }}
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import { ElDialog, ElCheckbox, ElButton } from 'element-plus'
  import { COLUMN_GROUPS } from '../columns'
  import type { ColumnOption } from '@/types/component'

  defineOptions({ name: 'ColumnSettingPanel' })

  const props = defineProps<{
    modelValue: boolean
    /** columnChecks（useTable 返回，元素含 checked/visible/group/label） */
    columns: ColumnOption<any>[]
  }>()

  const emit = defineEmits<{ (e: 'update:modelValue', v: boolean): void }>()

  const getVisibility = (col: ColumnOption<any>) => {
    if (col.visible !== undefined) return col.visible
    return col.checked ?? true
  }

  const toggle = (col: ColumnOption<any>, v: boolean) => {
    col.checked = v
    col.visible = v
  }

  const groupedColumns = computed(() =>
    COLUMN_GROUPS.map((g) => {
      const columns = (props.columns || []).filter((c) => (c.group || 'basic') === g.key)
      return {
        key: g.key,
        labelKey: g.labelKey,
        columns,
        checkedCount: columns.filter((c) => getVisibility(c)).length
      }
    }).filter((g) => g.columns.length)
  )

  const handleReset = () => {
    ;(props.columns || []).forEach((c) => {
      c.checked = true
      c.visible = true
    })
  }
</script>

<style lang="scss" scoped>
  .column-setting {
    display: flex;
    flex-direction: column;
    gap: 12px;
    max-height: 480px;
    overflow-y: auto;
  }

  .group-card {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
    overflow: hidden;
  }

  .group-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 14px;
    background: var(--el-fill-color-light);
    border-bottom: 1px solid var(--el-border-color-lighter);

    .group-title {
      font-weight: 600;
      font-size: 13px;
    }

    .group-count {
      font-size: 12px;
      color: var(--el-text-color-secondary);
    }
  }

  .group-body {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 4px 16px;
    padding: 12px 14px;
  }

  .group-checkbox {
    height: 28px;
  }
</style>
