import { AppRouteRecord } from '@/types/router'

export const adAccountRoutes: AppRouteRecord = {
  path: '/ad-account',
  name: 'AdAccount',
  component: '/index/index',
  meta: {
    title: 'menus.adAccount.title',
    icon: 'ri:advertisement-line',
    roles: ['R_SUPER', 'R_ADMIN']
  },
  children: [
    {
      path: 'list',
      name: 'AdAccountList',
      component: '/ad-account/index',
      meta: {
        title: 'menus.adAccount.list',
        keepAlive: true,
        roles: ['R_SUPER', 'R_ADMIN']
      }
    },
    {
      path: 'manage',
      name: 'AdAccountManage',
      component: '/ad-account/manage/index',
      meta: {
        title: 'menus.adAccount.manage',
        keepAlive: true,
        roles: ['R_SUPER', 'R_ADMIN']
      }
    }
  ]
}
