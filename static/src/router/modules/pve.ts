import { AppRouteRecord } from '@/types/router'

export const pveRoutes: AppRouteRecord = {
  path: '/pve',
  name: 'PVE',
  component: '/index/index',
  meta: {
    title: 'menus.pve.title',
    icon: 'ri:leaf-line',
    roles: ['super_admin', 'admin', 'fc', 'srp', 'user']
  },
  children: [
    {
      path: 'aoe-notification',
      name: 'AOENotification',
      component: '/pve/aoe-notification',
      meta: {
        title: 'menus.pve.aoeNotification',
        keepAlive: true,
        roles: ['super_admin', 'admin', 'fc', 'srp', 'user']
      }
    }
  ]
}
