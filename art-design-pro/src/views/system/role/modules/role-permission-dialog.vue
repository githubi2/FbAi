<template>
  <ElDialog
    v-model="visible"
    title="菜单权限"
    width="520px"
    align-center
    class="el-dialog-border"
    @close="handleClose"
  >
    <ElScrollbar height="70vh">
      <ElTree
        ref="treeRef"
        :data="processedMenuList"
        show-checkbox
        node-key="name"
        :default-expand-all="isExpandAll"
        :props="defaultProps"
        @check="handleTreeCheck"
      >
        <template #default="{ data }">
          <div style="display: flex; align-items: center">
            <span v-if="data.isAuth">
              {{ data.label }}
            </span>
            <span v-else>{{ defaultProps.label(data) }}</span>
          </div>
        </template>
      </ElTree>
    </ElScrollbar>
    <template #footer>
      <ElButton @click="outputSelectedData" style="margin-left: 8px">获取选中数据</ElButton>

      <ElButton @click="toggleExpandAll">{{ isExpandAll ? '全部收起' : '全部展开' }}</ElButton>
      <ElButton @click="toggleSelectAll" style="margin-left: 8px">{{
        isSelectAll ? '取消全选' : '全部选择'
      }}</ElButton>
      <ElButton type="primary" @click="savePermission">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { useMenuStore } from '@/store/modules/menu'
  import { formatMenuTitle } from '@/utils/router'
  import { fetchGetRoleMenus, fetchUpdateRole } from '@/api/system-manage'

  type RoleListItem = Api.SystemManage.RoleListItem

  interface Props {
    modelValue: boolean
    roleData?: RoleListItem
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    roleData: undefined
  })

  const emit = defineEmits<Emits>()

  const { menuList } = storeToRefs(useMenuStore())
  const treeRef = ref()
  const isExpandAll = ref(true)
  const isSelectAll = ref(false)

  /**
   * 弹窗显示状态双向绑定
   */
  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  /**
   * 菜单节点类型
   */
  interface MenuNode {
    id?: string | number
    name?: string
    label?: string
    meta?: {
      title?: string
      authList?: Array<{
        authMark: string
        title: string
        checked?: boolean
      }>
    }
    children?: MenuNode[]
    [key: string]: any
  }

  /**
   * 处理菜单数据，将 authList 转换为树形子节点
   * 递归处理菜单树，将权限列表展开为可选择的子节点
   */
  const processedMenuList = computed(() => {
    const processNode = (node: MenuNode): MenuNode => {
      const processed = { ...node }

      // 如果有 authList，将其转换为子节点
      if (node.meta?.authList?.length) {
        const authNodes = node.meta.authList.map((auth) => ({
          id: `${node.id}_${auth.authMark}`,
          name: `${node.name}_${auth.authMark}`,
          label: auth.title,
          authMark: auth.authMark,
          isAuth: true,
          checked: auth.checked || false
        }))

        processed.children = processed.children ? [...processed.children, ...authNodes] : authNodes
      }

      // 递归处理子节点
      if (processed.children) {
        processed.children = processed.children.map(processNode)
      }

      return processed
    }

    return (menuList.value as any[]).map(processNode)
  })

  /**
   * 树形组件配置
   */
  const defaultProps = {
    children: 'children',
    label: (data: any) => formatMenuTitle(data.meta?.title) || data.label || ''
  }

  /**
   * 后端菜单数据（id → name 映射，用于桥接前端路由 name 和后端菜单 ID）
   */
  const idToNameMap = ref<Map<number, string>>(new Map())
  const nameToIdMap = ref<Map<string, number>>(new Map())

  /**
   * 监听弹窗打开，从后端加载角色的菜单权限数据
   */
  watch(
    () => props.modelValue,
    async (newVal) => {
      if (newVal && props.roleData?.roleId) {
        try {
          const res = await fetchGetRoleMenus(props.roleData.roleId)
          const roleMenus: number[] = res.roleMenus || []
          const allMenus: any[] = (res as any).allMenus || []

          // 构建 id↔name 双向映射（前端路由用 name 作为 node-key，后端用 id 存储权限）
          const id2Name = new Map<number, string>()
          const name2Id = new Map<string, number>()
          allMenus.forEach((menu: any) => {
            if (menu.id != null && menu.name) {
              id2Name.set(Number(menu.id), menu.name)
              name2Id.set(menu.name, Number(menu.id))
            }
          })
          idToNameMap.value = id2Name
          nameToIdMap.value = name2Id

          // 将后端的 menuId 转换为前端路由的 name
          const checkedNames: string[] = roleMenus
            .map((id) => id2Name.get(id)!)
            .filter(Boolean)

          await nextTick()
          treeRef.value?.setCheckedKeys(checkedNames)
        } catch (error) {
          console.error('加载角色菜单权限失败:', error)
          ElMessage.error('加载菜单权限失败')
        }
      }
    }
  )

  /**
   * 关闭弹窗并清空选中状态
   */
  const handleClose = () => {
    visible.value = false
    treeRef.value?.setCheckedKeys([])
    isSelectAll.value = false
  }

  /**
   * 保存权限配置
   */
  const savePermission = async () => {
    const tree = treeRef.value
    if (!tree) return

    if (!props.roleData?.roleId) {
      ElMessage.error('角色数据异常')
      return
    }

    const checkedKeys: string[] = tree.getCheckedKeys()
    // 将前端路由 name 转换为后端菜单 ID
    // auth 子节点的 key 格式为 "MenuName_authMark"，过滤掉只保留纯菜单 name
    const menuIds: number[] = checkedKeys
      .map((name) => nameToIdMap.value.get(name))
      .filter((id): id is number => id != null)

    try {
      await fetchUpdateRole(props.roleData.roleId, {
        roleName: props.roleData.roleName,
        roleCode: props.roleData.roleCode,
        description: props.roleData.description,
        status: props.roleData.enabled ? 1 : 0,
        menuIds
      })
      ElMessage.success('权限保存成功')
      emit('success')
      handleClose()
    } catch (error) {
      console.error('保存权限失败:', error)
      ElMessage.error('保存权限失败')
    }
  }

  /**
   * 切换全部展开/收起状态
   */
  const toggleExpandAll = () => {
    const tree = treeRef.value
    if (!tree) return

    const nodes = tree.store.nodesMap
    // 这里保留 any，因为 Element Plus 的内部节点类型较复杂
    Object.values(nodes).forEach((node: any) => {
      node.expanded = !isExpandAll.value
    })

    isExpandAll.value = !isExpandAll.value
  }

  /**
   * 切换全选/取消全选状态
   */
  const toggleSelectAll = () => {
    const tree = treeRef.value
    if (!tree) return

    if (!isSelectAll.value) {
      const allKeys = getAllNodeKeys(processedMenuList.value)
      tree.setCheckedKeys(allKeys)
    } else {
      tree.setCheckedKeys([])
    }

    isSelectAll.value = !isSelectAll.value
  }

  /**
   * 递归获取所有节点的 key (name)
   * @param nodes 节点列表
   * @returns 所有节点的 key 数组
   */
  const getAllNodeKeys = (nodes: MenuNode[]): string[] => {
    const keys: string[] = []
    const traverse = (nodeList: MenuNode[]): void => {
      nodeList.forEach((node) => {
        if (node.name) keys.push(node.name)
        if (node.children?.length) traverse(node.children)
      })
    }
    traverse(nodes)
    return keys
  }

  /**
   * 处理树节点选中状态变化
   * 同步更新全选按钮状态
   */
  const handleTreeCheck = () => {
    const tree = treeRef.value
    if (!tree) return

    const checkedKeys = tree.getCheckedKeys()
    const allKeys = getAllNodeKeys(processedMenuList.value)

    isSelectAll.value = checkedKeys.length === allKeys.length && allKeys.length > 0
  }

  /**
   * 输出选中的权限数据到控制台
   * 用于调试和查看当前选中的权限配置
   */
  const outputSelectedData = () => {
    const tree = treeRef.value
    if (!tree) return

    const selectedData = {
      checkedKeys: tree.getCheckedKeys(),
      halfCheckedKeys: tree.getHalfCheckedKeys(),
      checkedNodes: tree.getCheckedNodes(),
      halfCheckedNodes: tree.getHalfCheckedNodes(),
      totalChecked: tree.getCheckedKeys().length,
      totalHalfChecked: tree.getHalfCheckedKeys().length
    }

    console.log('=== 选中的权限数据 ===', selectedData)
    ElMessage.success(`已输出选中数据到控制台，共选中 ${selectedData.totalChecked} 个节点`)
  }
</script>
