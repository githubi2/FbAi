import { AppRouteRecord } from '@/types/router'

export const bmManageRoutes: AppRouteRecord = {
  path: '/bm-manage',
  name: 'AdAccountBm',
  component: '/index/index',
  meta: {
    title: 'menus.adAccount.bm',
    icon: 'ri:building-2-line',
    roles: ['R_SUPER', 'R_ADMIN']
  },
  children: [
    {
      path: 'list',
      name: 'AdAccountBmList',
      component: '/ad-account/bm/index',
      meta: {
        title: 'menus.adAccount.bmList',
        keepAlive: true,
        roles: ['R_SUPER', 'R_ADMIN']
      }
    }
  ]
}
