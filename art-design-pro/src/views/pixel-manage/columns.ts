// 像素管理表格列配置工厂
// 从 index.vue 拆出，保持单文件 < 300 行
import { h } from 'vue'
import { ElTag, ElTooltip, ElIcon } from 'element-plus'
import { Edit } from '@element-plus/icons-vue'
import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
import type { FbPixelItem } from '@/api/facebook'
import type { ColumnOption } from '@/types/component'

interface PixelColumnsOptions {
  t: (key: string) => string
  onEditRemark: (row: FbPixelItem) => void
}

const DASH = () => h('span', { style: { color: '#999' } }, '—')

/** ISO 时间格式化为 YYYY-MM-DD HH:mm:ss（不使用日期库） */
const formatTime = (iso: string | null) => {
  if (!iso) return ''
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ''
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}

export function buildPixelColumns({
  t,
  onEditRemark
}: PixelColumnsOptions): ColumnOption<FbPixelItem>[] {
  const P = 'menus.pixelManage.columns'

  return [
    { type: 'selection', width: 55 },
    {
      prop: 'name',
      label: t(`${P}.pixel`),
      minWidth: 180,
      formatter: (row: FbPixelItem) =>
        h('div', [
          h('span', { style: { fontWeight: 500 } }, row.name || '—'),
          h('div', { style: { color: '#999', fontSize: '12px' } }, row.pixelId)
        ])
    },
    {
      prop: 'ownerBmName',
      label: t(`${P}.ownerBm`),
      minWidth: 120,
      formatter: (row: FbPixelItem) => row.ownerBmName || DASH()
    },
    {
      prop: 'fbOwnerName',
      label: t(`${P}.owner`),
      minWidth: 120,
      formatter: (row: FbPixelItem) => row.fbOwnerName || row.creatorName || DASH()
    },
    {
      prop: 'remark',
      label: t(`${P}.remark`),
      minWidth: 140,
      formatter: (row: FbPixelItem) =>
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
      prop: 'roleNames',
      label: t(`${P}.role`),
      minWidth: 110,
      formatter: (row: FbPixelItem) => {
        const roles = row.roleNames || []
        return roles.length > 0 ? roles.join('、') : DASH()
      }
    },
    {
      prop: 'shared',
      label: t(`${P}.shared`),
      width: 95,
      formatter: (row: FbPixelItem) => {
        const shared = (row.sharedAgencies || []).length > 0
        return h(ElTag, { type: shared ? 'success' : 'info', size: 'small' }, () =>
          shared ? t('menus.pixelManage.sharedYes') : t('menus.pixelManage.sharedNo')
        )
      }
    },
    {
      prop: 'bmShared',
      label: t(`${P}.bmShared`),
      width: 95,
      formatter: (row: FbPixelItem) => {
        // BM 分享与共享合作伙伴同源：分享给其他 BM/代理商即 BM 分享
        const shared = (row.sharedAgencies || []).length > 0
        return h(ElTag, { type: shared ? 'success' : 'info', size: 'small' }, () =>
          shared ? t('menus.pixelManage.sharedYes') : t('menus.pixelManage.sharedNo')
        )
      }
    },
    {
      prop: 'adAccountName',
      label: t(`${P}.adAccount`),
      minWidth: 140,
      formatter: (row: FbPixelItem) =>
        h('div', [
          h('span', {}, row.adAccountName || '—'),
          h('div', { style: { color: '#999', fontSize: '12px' } }, row.adAccountId)
        ])
    },
    {
      prop: 'adminNames',
      label: t(`${P}.admin`),
      minWidth: 110,
      formatter: (row: FbPixelItem) => {
        const names = row.adminNames || []
        if (names.length === 0) return DASH()
        return h(ElTooltip, { content: names.join('、'), placement: 'top' }, () =>
          h(ElTag, { type: 'primary', size: 'small' }, () => `${names[0]} +${names.length - 1}`)
        )
      }
    },
    {
      prop: 'sharedAgencies',
      label: t(`${P}.partners`),
      minWidth: 120,
      formatter: (row: FbPixelItem) => {
        const agencies = row.sharedAgencies || []
        if (agencies.length === 0) return DASH()
        return h(ElTooltip, { content: agencies.join('、'), placement: 'top' }, () =>
          h('span', {}, agencies.join('、'))
        )
      }
    },
    {
      prop: 'active',
      label: t(`${P}.active`),
      width: 95,
      formatter: (row: FbPixelItem) => {
        if (row.isUnavailable === 1)
          return h(ElTag, { type: 'danger', size: 'small' }, () =>
            t('menus.pixelManage.unavailable')
          )
        if (row.lastFiredTime)
          return h(ElTag, { type: 'success', size: 'small' }, () =>
            t('menus.pixelManage.activeYes')
          )
        return h(ElTag, { type: 'warning', size: 'small' }, () => t('menus.pixelManage.activeNo'))
      }
    },
    {
      prop: 'lastFiredTime',
      label: t(`${P}.activeTime`),
      minWidth: 160,
      formatter: (row: FbPixelItem) => formatTime(row.lastFiredTime) || DASH()
    },
    {
      prop: 'operation',
      label: t(`${P}.operation`),
      width: 110,
      fixed: 'right',
      formatter: (row: FbPixelItem) =>
        h('div', { class: 'flex-c' }, [
          h(ArtButtonTable, {
            type: 'edit',
            title: t('menus.pixelManage.editRemark'),
            onClick: () => onEditRemark(row)
          })
        ])
    }
  ]
}
