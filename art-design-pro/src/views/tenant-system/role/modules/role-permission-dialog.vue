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
        :check-strictly="true"
        @check="handleTreeCheck"
      >
        <template #default="{ data }">
          <div style="display: flex; align-items: center">
            <span v-if="data.isAuth">{{ data.label }}</span>
            <span v-else>{{ defaultProps.label(data) }}</span>
          </div>
        </template>
      </ElTree>
    </ElScrollbar>
    <template #footer>
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

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

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

  const processedMenuList = computed(() => {
    const processNode = (node: MenuNode): MenuNode => {
      const processed = { ...node }
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
      if (processed.children) {
        processed.children = processed.children.map(processNode)
      }
      return processed
    }
    return (menuList.value as any[]).map(processNode)
  })

  const defaultProps = {
    children: 'children',
    label: (data: any) => formatMenuTitle(data.meta?.title) || data.label || ''
  }

  const idToNameMap = ref<Map<number, string>>(new Map())
  const nameToIdMap = ref<Map<string, number>>(new Map())

  watch(
    () => props.modelValue,
    async (newVal) => {
      if (newVal && props.roleData?.roleId) {
        try {
          const res = await fetchGetRoleMenus(props.roleData.roleId)
          const roleMenus: number[] = res.roleMenus || []
          const allMenus: any[] = (res as any).allMenus || []

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

  const handleClose = () => {
    visible.value = false
    treeRef.value?.setCheckedKeys([])
    isSelectAll.value = false
  }

  const savePermission = async () => {
    const tree = treeRef.value
    if (!tree) return
    if (!props.roleData?.roleId) {
      ElMessage.error('角色数据异常')
      return
    }

    const checkedKeys: string[] = tree.getCheckedKeys()
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

  const toggleExpandAll = () => {
    const tree = treeRef.value
    if (!tree) return
    const nodes = tree.store.nodesMap
    Object.values(nodes).forEach((node: any) => {
      node.expanded = !isExpandAll.value
    })
    isExpandAll.value = !isExpandAll.value
  }

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

  const handleTreeCheck = () => {
    const tree = treeRef.value
    if (!tree) return
    const checkedKeys = tree.getCheckedKeys()
    const allKeys = getAllNodeKeys(processedMenuList.value)
    isSelectAll.value = checkedKeys.length === allKeys.length && allKeys.length > 0
  }
</script>
