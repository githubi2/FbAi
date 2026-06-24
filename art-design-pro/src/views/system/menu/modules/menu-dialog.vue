<template>
  <ElDialog
    :title="dialogTitle"
    :model-value="visible"
    @update:model-value="handleCancel"
    width="860px"
    align-center
    class="menu-dialog"
    @closed="handleClosed"
  >
    <ArtForm
      ref="formRef"
      v-model="form"
      :items="formItems"
      :rules="rules"
      :span="width > 640 ? 12 : 24"
      :gutter="20"
      label-width="100px"
      :show-reset="false"
      :show-submit="false"
    >
      <template #menuType>
        <ElRadioGroup v-model="form.menuType" :disabled="disableMenuType">
          <ElRadioButton value="directory" label="directory">目录</ElRadioButton>
          <ElRadioButton value="menu" label="menu">菜单</ElRadioButton>
          <ElRadioButton value="button" label="button">按钮</ElRadioButton>
        </ElRadioGroup>
      </template>
      <template v-if="form.menuType !== 'button'" #parentId>
        <ElTreeSelect
          v-model="form.parentId"
          :data="menuTreeOptions"
          :props="{ label: 'label', value: 'id', children: 'children' } as any"
          placeholder="无（顶级菜单）"
          clearable
          check-strictly
          filterable
          style="width: 100%"
        />
      </template>
    </ArtForm>

    <template #footer>
      <span class="dialog-footer">
        <ElButton @click="handleCancel">取 消</ElButton>
        <ElButton type="primary" @click="handleSubmit">确 定</ElButton>
      </span>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormRules } from 'element-plus'
  import { ElIcon, ElTooltip } from 'element-plus'
  import { QuestionFilled } from '@element-plus/icons-vue'
  import type { AppRouteRecord } from '@/types/router'
  import type { FormItem } from '@/components/core/forms/art-form/index.vue'
  import ArtForm from '@/components/core/forms/art-form/index.vue'
  import { useWindowSize } from '@vueuse/core'

  const { width } = useWindowSize()

  /**
   * 创建带 tooltip 的表单标签
   * @param label 标签文本
   * @param tooltip 提示文本
   * @returns 渲染函数
   */
  const createLabelTooltip = (label: string, tooltip: string) => {
    return () =>
      h('span', { class: 'flex items-center' }, [
        h('span', label),
        h(
          ElTooltip,
          {
            content: tooltip,
            placement: 'top'
          },
          () => h(ElIcon, { class: 'ml-0.5 cursor-help' }, () => h(QuestionFilled))
        )
      ])
  }

  interface MenuFormData {
    id: number
    parentId: number
    title: string
    name: string
    path: string
    component: string
    icon: string
    isEnable: boolean
    sort: number
    isMenu: boolean
    keepAlive: boolean
    isHide: boolean
    isHideTab: boolean
    link: string
    isIframe: boolean
    showBadge: boolean
    showTextBadge: string
    fixedTab: boolean
    activePath: string
    roles: string[]
    isFullPage: boolean
    authName: string
    authLabel: string
    authIcon: string
    authSort: number
  }

  interface Props {
    visible: boolean
    editData?: AppRouteRecord | any
    type?: 'menu' | 'button' | 'directory'
    lockType?: boolean
    menuTree?: AppRouteRecord[]
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'submit', data: MenuFormData): void
  }

  const props = withDefaults(defineProps<Props>(), {
    visible: false,
    type: 'menu',
    lockType: false,
    menuTree: () => []
  })

  const emit = defineEmits<Emits>()

  const formRef = ref()
  const isEdit = ref(false)

  const form = reactive<MenuFormData & { menuType: 'menu' | 'button' | 'directory' }>({
    menuType: 'menu',
    id: 0,
    parentId: 0,
    title: '',
    name: '',
    path: '',
    component: '',
    icon: '',
    isEnable: true,
    sort: 1,
    isMenu: true,
    keepAlive: true,
    isHide: false,
    isHideTab: false,
    link: '',
    isIframe: false,
    showBadge: false,
    showTextBadge: '',
    fixedTab: false,
    activePath: '',
    roles: [],
    isFullPage: false,
    authName: '',
    authLabel: '',
    authIcon: '',
    authSort: 1
  })

  const rules = reactive<FormRules>({
    title: [
      { required: true, message: '请输入菜单名称', trigger: 'blur' },
      { min: 2, max: 64, message: '长度在 2 到 64 个字符', trigger: 'blur' }
    ],
    name: [
      { required: true, message: '请输入路由标识', trigger: 'blur' },
      { min: 2, max: 64, message: '长度在 2 到 64 个字符', trigger: 'blur' },
      {
        pattern: /^[a-zA-Z][a-zA-Z0-9_]*$/,
        message: '必须以字母开头，只允许字母/数字/下划线',
        trigger: 'blur'
      }
    ],
    path: [
      { required: true, message: '请输入路由地址', trigger: 'blur' },
      { max: 128, message: '长度不能超过 128 个字符', trigger: 'blur' }
    ],
    component: [{ max: 128, message: '长度不能超过 128 个字符', trigger: 'blur' }],
    icon: [{ max: 64, message: '长度不能超过 64 个字符', trigger: 'blur' }],
    sort: [{ type: 'number', min: 0, message: '排序必须 >= 0', trigger: 'blur' }],
    link: [{ type: 'url', message: '请输入有效的 URL 地址', trigger: 'blur' }],
    authName: [{ required: true, message: '请输入权限名称', trigger: 'blur' }],
    authLabel: [{ required: true, message: '请输入权限标识', trigger: 'blur' }]
  })

  /**
   * 将菜单树转为 ElTreeSelect 的选项格式
   */
  const menuTreeOptions = computed(() => {
    const transform = (nodes: any[], depth: number = 0): any[] => {
      if (!Array.isArray(nodes)) return []
      return nodes
        .filter((n: any) => {
          // 排除按钮类型，且编辑时排除自身及其子节点
          const isButton = n.meta?.menuType === 'button'
          if (isButton) return false
          if (isEdit.value && form.id && n.id === form.id) return false
          return true
        })
        .map((n: any) => ({
          id: n.id,
          label: (depth > 0 ? '├ ' : '') + (n.meta?.title || n.name || ''),
          children: n.children?.length ? transform(n.children, depth + 1) : undefined,
          disabled: isEdit.value && form.id && n.id === form.id
        }))
    }
    return transform(props.menuTree || [])
  })

  /**
   * 表单项配置
   */
  const formItems = computed<FormItem[]>(() => {
    const baseItems: FormItem[] = [{ label: '菜单类型', key: 'menuType', span: 24 }]

    // Switch 组件的 span：小屏幕 12，大屏幕 6
    const switchSpan = width.value < 640 ? 12 : 6

    // 上级菜单选择器（目录和菜单类型都显示）
    const parentMenuItem: FormItem = {
      label: '上级菜单',
      key: 'parentId',
      span: 12
    }

    if (form.menuType === 'directory') {
      return [
        ...baseItems,
        parentMenuItem,
        {
          label: '目录名称',
          key: 'title',
          type: 'input',
          props: { placeholder: '目录显示名称（支持中文）' }
        },
        {
          label: '路由标识',
          key: 'name',
          type: 'input',
          props: { placeholder: '路由 name（英文字母开头）' }
        },
        {
          label: createLabelTooltip('路由地址', '以 / 开头的绝对路径（如 /system、/dashboard）'),
          key: 'path',
          type: 'input',
          props: { placeholder: '如：/system' }
        },
        {
          label: createLabelTooltip('组件路径', '目录通常填写 /index/index'),
          key: 'component',
          type: 'input',
          props: { placeholder: '如：/index/index' }
        },
        { label: '图标', key: 'icon', type: 'input', props: { placeholder: '如：ri:folder-line' } },
        {
          label: '菜单排序',
          key: 'sort',
          type: 'number',
          props: { min: 1, controlsPosition: 'right', style: { width: '100%' } }
        },
        { label: '是否启用', key: 'isEnable', type: 'switch', span: switchSpan },
        { label: '隐藏菜单', key: 'isHide', type: 'switch', span: switchSpan }
      ]
    }

    if (form.menuType === 'menu') {
      return [
        ...baseItems,
        parentMenuItem,
        {
          label: '菜单名称',
          key: 'title',
          type: 'input',
          props: { placeholder: '菜单显示名称（支持中文）' }
        },
        {
          label: '路由标识',
          key: 'name',
          type: 'input',
          props: { placeholder: '路由 name（英文字母开头）' }
        },
        {
          label: createLabelTooltip(
            '路由地址',
            '一级菜单：以 / 开头的绝对路径（如 /dashboard）\n二级及以下：相对路径（如 console、user）'
          ),
          key: 'path',
          type: 'input',
          props: { placeholder: '如：/dashboard 或 console' }
        },
        {
          label: createLabelTooltip(
            '组件路径',
            '一级父级菜单：填写 /index/index\n具体页面：填写组件路径（如 /system/user）\n目录菜单：留空'
          ),
          key: 'component',
          type: 'input',
          props: { placeholder: '如：/system/user 或留空' }
        },
        { label: '图标', key: 'icon', type: 'input', props: { placeholder: '如：ri:user-line' } },
        {
          label: createLabelTooltip(
            '角色权限',
            '仅用于前端权限模式：配置角色标识（如 R_SUPER、R_ADMIN）\n后端权限模式：无需配置'
          ),
          key: 'roles',
          type: 'inputtag',
          props: { placeholder: '输入角色标识后按回车，如：R_SUPER' }
        },
        {
          label: '菜单排序',
          key: 'sort',
          type: 'number',
          props: { min: 1, controlsPosition: 'right', style: { width: '100%' } }
        },
        {
          label: '外部链接',
          key: 'link',
          type: 'input',
          props: { placeholder: '如：https://www.example.com' }
        },
        {
          label: '文本徽章',
          key: 'showTextBadge',
          type: 'input',
          props: { placeholder: '如：New、Hot' }
        },
        {
          label: createLabelTooltip(
            '激活路径',
            '用于详情页等隐藏菜单，指定高亮显示的父级菜单路径\n例如：用户详情页高亮显示"用户管理"菜单'
          ),
          key: 'activePath',
          type: 'input',
          props: { placeholder: '如：/system/user' }
        },
        { label: '是否启用', key: 'isEnable', type: 'switch', span: switchSpan },
        { label: '页面缓存', key: 'keepAlive', type: 'switch', span: switchSpan },
        { label: '隐藏菜单', key: 'isHide', type: 'switch', span: switchSpan },
        { label: '是否内嵌', key: 'isIframe', type: 'switch', span: switchSpan },
        { label: '显示徽章', key: 'showBadge', type: 'switch', span: switchSpan },
        { label: '固定标签', key: 'fixedTab', type: 'switch', span: switchSpan },
        { label: '标签隐藏', key: 'isHideTab', type: 'switch', span: switchSpan },
        { label: '全屏页面', key: 'isFullPage', type: 'switch', span: switchSpan }
      ]
    } else {
      return [
        ...baseItems,
        {
          label: '权限名称',
          key: 'authName',
          type: 'input',
          props: { placeholder: '如：新增、编辑、删除' }
        },
        {
          label: '权限标识',
          key: 'authLabel',
          type: 'input',
          props: { placeholder: '如：add、edit、delete' }
        },
        {
          label: '权限排序',
          key: 'authSort',
          type: 'number',
          props: { min: 1, controlsPosition: 'right', style: { width: '100%' } }
        }
      ]
    }
  })

  const dialogTitle = computed(() => {
    const typeMap: Record<string, string> = { directory: '目录', menu: '菜单', button: '按钮' }
    const type = typeMap[form.menuType] || '菜单'
    return isEdit.value ? `编辑${type}` : `新建${type}`
  })

  /**
   * 是否禁用菜单类型切换
   */
  const disableMenuType = computed(() => {
    if (props.lockType) return true
    return false
  })

  /**
   * 重置表单数据
   */
  const resetForm = (): void => {
    formRef.value?.reset()
    form.menuType = 'menu'
  }

  /**
   * 加载表单数据（编辑模式）
   */
  const loadFormData = (): void => {
    if (!props.editData) return

    isEdit.value = true

    // 根据后端实际 menuType 设置表单类型（编辑时）
    const backendType = props.editData.meta?.menuType || props.editData.menuType
    if (backendType && ['directory', 'menu', 'button'].includes(backendType)) {
      form.menuType = backendType
    }

    if (form.menuType === 'directory' || form.menuType === 'menu') {
      const row = props.editData
      form.id = row.id || 0
      form.parentId = row.meta?.parentId || row.parentId || 0
      form.title = row.meta?.title || ''
      form.name = row.name || ''
      form.path = row.path || ''
      form.component = row.component || ''
      form.icon = row.meta?.icon || ''
      form.sort = row.meta?.sort || 1
      form.isMenu = row.meta?.isMenu ?? true
      form.keepAlive = row.meta?.keepAlive ?? false
      form.isHide = row.meta?.isHide ?? false
      form.isHideTab = row.meta?.isHideTab ?? false
      form.isEnable = row.meta?.isEnable ?? true
      form.link = row.meta?.link || ''
      form.isIframe = row.meta?.isIframe ?? false
      form.showBadge = row.meta?.showBadge ?? false
      form.showTextBadge = row.meta?.showTextBadge || ''
      form.fixedTab = row.meta?.fixedTab ?? false
      form.activePath = row.meta?.activePath || ''
      form.roles = row.meta?.roles || []
      form.isFullPage = row.meta?.isFullPage ?? false
    } else {
      const row = props.editData
      form.authName = row.title || ''
      form.authLabel = row.authMark || ''
      form.authIcon = row.icon || ''
      form.authSort = row.sort || 1
    }
  }

  /**
   * 提交表单
   */
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    try {
      await formRef.value.validate()
      emit('submit', { ...form })
      ElMessage.success(`${isEdit.value ? '编辑' : '新增'}成功`)
      handleCancel()
    } catch {
      ElMessage.error('表单校验失败，请检查输入')
    }
  }

  /**
   * 取消操作
   */
  const handleCancel = (): void => {
    emit('update:visible', false)
  }

  /**
   * 对话框关闭后的回调
   */
  const handleClosed = (): void => {
    resetForm()
    isEdit.value = false
  }

  /**
   * 监听对话框显示状态
   */
  watch(
    () => props.visible,
    (newVal) => {
      if (newVal) {
        form.menuType = props.type
        nextTick(() => {
          if (props.editData) {
            loadFormData()
          }
        })
      }
    }
  )

  /**
   * 监听菜单类型变化
   */
  watch(
    () => props.type,
    (newType) => {
      if (props.visible) {
        form.menuType = newType
      }
    }
  )
</script>
