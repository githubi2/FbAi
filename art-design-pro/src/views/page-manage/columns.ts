// 主页管理表格列配置工厂
// 从 index.vue 拆出，保持单文件 < 300 行
import { h } from 'vue'
import { ElTag, ElTooltip, ElIcon } from 'element-plus'
import { Edit } from '@element-plus/icons-vue'
import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
import type { FbPageItem } from '@/api/facebook'
import type { ColumnOption } from '@/types/component'

interface PageColumnsOptions {
  t: (key: string) => string
  onEditRemark: (row: FbPageItem) => void
}

// FB 官方 API 未提供的字段统一显示 —
const DASH = () => h('span', { style: { color: '#999' } }, '—')

export function buildPageColumns({
  t,
  onEditRemark
}: PageColumnsOptions): ColumnOption<FbPageItem>[] {
  const P = 'menus.pageManage.columns'

  return [
    { type: 'selection', width: 55 },
    {
      prop: 'name',
      label: t(`${P}.name`),
      minWidth: 180,
      formatter: (row: FbPageItem) =>
        h(ElTooltip, { content: row.link || row.name, placement: 'top' }, () =>
          h('span', { style: { fontWeight: 500 } }, row.name || '—')
        )
    },
    {
      prop: 'pushStatus',
      label: t(`${P}.pushStatus`),
      width: 95,
      formatter: (row: FbPageItem) => row.pushStatus || '—'
    },
    {
      prop: 'remark',
      label: t(`${P}.remark`),
      minWidth: 140,
      formatter: (row: FbPageItem) =>
        h('div', { style: 'display:flex;align-items:center;gap:6px' }, [
          h('span', { style: row.remark ? '' : 'color:#999' }, row.remark || '—'),
          h(
            ElIcon,
            {
              style: 'cursor:pointer;color:#409eff;font-size:14px',
              onClick: () => onEditRemark(row)
            },
            () => h(Edit)
          )
        ])
    },
    {
      prop: 'pageId',
      label: t(`${P}.pageId`),
      minWidth: 150,
      formatter: (row: FbPageItem) => row.pageId || '—'
    },
    {
      prop: 'isPublished',
      label: t(`${P}.publishStatus`),
      width: 100,
      formatter: (row: FbPageItem) =>
        h(ElTag, { type: row.isPublished ? 'success' : 'info', size: 'small' }, () =>
          row.isPublished ? t('menus.pageManage.published') : t('menus.pageManage.unpublished')
        )
    },
    {
      prop: 'hideProfanity',
      label: t(`${P}.hideProfanity`),
      minWidth: 130,
      formatter: (row: FbPageItem) => {
        const map: Record<string, string> = {
          none: t('menus.pageManage.profanityNone'),
          medium: t('menus.pageManage.profanityMedium'),
          strong: t('menus.pageManage.profanityStrong')
        }
        const label = map[row.profanityFilter]
        if (!label) return DASH()
        return h(
          ElTag,
          { type: row.profanityFilter === 'none' ? 'info' : 'success', size: 'small' },
          () => label
        )
      }
    },
    {
      prop: 'verificationStatus',
      label: t(`${P}.verification`),
      width: 105,
      formatter: (row: FbPageItem) => {
        const verified = row.verificationStatus === 'verified'
        return h(ElTag, { type: verified ? 'success' : 'info', size: 'small' }, () =>
          verified ? t('menus.pageManage.verified') : t('menus.pageManage.notVerified')
        )
      }
    },
    {
      prop: 'adPerm',
      label: t(`${P}.adPerm`),
      minWidth: 170,
      formatter: (row: FbPageItem) => {
        if (row.adPerm === 1)
          return h(ElTag, { type: 'success', size: 'small' }, () => t('menus.pageManage.adPermOk'))
        if (row.adPerm === 0)
          return h(ElTag, { type: 'danger', size: 'small' }, () => t('menus.pageManage.adPermNone'))
        return DASH()
      }
    },
    {
      prop: 'bmName',
      label: t(`${P}.bm`),
      minWidth: 120,
      formatter: (row: FbPageItem) => row.bmName || DASH()
    },
    {
      prop: 'adminNames',
      label: t(`${P}.admin`),
      minWidth: 110,
      formatter: (row: FbPageItem) => {
        const names = row.adminNames || []
        if (names.length === 0) return DASH()
        return h(ElTooltip, { content: names.join('、'), placement: 'top' }, () =>
          h(ElTag, { type: 'primary', size: 'small' }, () => `${names[0]} +${names.length - 1}`)
        )
      }
    },
    {
      prop: 'blacklist',
      label: t(`${P}.blacklist`),
      minWidth: 110,
      formatter: (row: FbPageItem) => String(row.blockedCount ?? 0)
    },
    {
      prop: 'address',
      label: t(`${P}.address`),
      minWidth: 160,
      formatter: (row: FbPageItem) => row.address || '—'
    },
    {
      prop: 'phone',
      label: t(`${P}.phone`),
      minWidth: 120,
      formatter: (row: FbPageItem) => row.phone || '—'
    },
    {
      prop: 'email',
      label: t(`${P}.email`),
      minWidth: 160,
      formatter: (row: FbPageItem) => row.email || '—'
    },
    {
      prop: 'website',
      label: t(`${P}.website`),
      minWidth: 160,
      formatter: (row: FbPageItem) =>
        row.website ? h('span', { style: { color: '#409eff' } }, row.website) : '—'
    },
    {
      prop: 'category',
      label: t(`${P}.category`),
      minWidth: 130,
      formatter: (row: FbPageItem) => row.category || '—'
    },
    {
      prop: 'fanCount',
      label: t(`${P}.fanCount`),
      width: 90,
      formatter: (row: FbPageItem) => String(row.fanCount ?? 0)
    },
    {
      prop: 'followersCount',
      label: t(`${P}.followersCount`),
      width: 90,
      formatter: (row: FbPageItem) => String(row.followersCount ?? 0)
    },
    {
      prop: 'operation',
      label: t(`${P}.operation`),
      width: 110,
      fixed: 'right',
      formatter: (row: FbPageItem) =>
        h('div', { class: 'flex-c' }, [
          h(ArtButtonTable, {
            type: 'edit',
            title: t('menus.pageManage.editRemark'),
            onClick: () => onEditRemark(row)
          })
        ])
    }
  ]
}
