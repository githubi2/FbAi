import { AppRouteRecord } from '@/types/router'

export const tenantSystemRoutes: AppRouteRecord = {
  path: '/tenant-system',
  name: 'TenantSystem',
  component: '/index/index',
  meta: {
    title: 'menus.tenantSystem.title',
    icon: 'ri:building-2-line',
    roles: []
  },
  children: [
    {
      path: 'user',
      name: 'TenantUser',
      component: '/tenant-system/user',
      meta: {
        title: 'menus.tenantSystem.user',
        keepAlive: true,
        roles: []
      }
    },
    {
      path: 'role',
      name: 'TenantRole',
      component: '/tenant-system/role',
      meta: {
        title: 'menus.tenantSystem.role',
        keepAlive: true,
        roles: []
      }
    },
    {
      path: 'menu',
      name: 'TenantMenu',
      component: '/tenant-system/menu',
      meta: {
        title: 'menus.tenantSystem.menu',
        keepAlive: true,
        roles: []
      }
    }
  ]
}
