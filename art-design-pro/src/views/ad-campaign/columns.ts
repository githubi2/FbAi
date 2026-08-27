// 广告投放表格列配置工厂（从 index.vue 拆出，保持单文件 < 300 行）
import { h } from 'vue'
import { ElTag, ElTooltip, ElButton } from 'element-plus'
import type { FbCampaign, FbAdSet, FbAd, FbInsight } from '@/api/facebook'
import type { ColumnOption } from '@/types/component'

const DASH = () => h('span', { style: { color: '#999' } }, '—')

interface StatusConfig {
  type: 'success' | 'warning' | 'info' | 'danger' | 'primary'
  label: string
}

// 状态映射：ACTIVE=投放中 PAUSED=已暂停 ARCHIVED=已归档 DELETED=已删除
export function getStatusConfig(status: string, t: (key: string) => string): StatusConfig {
  const P = 'menus.adCampaign.status'
  switch (status) {
    case 'ACTIVE':
      return { type: 'success', label: t(`${P}.ACTIVE`) }
    case 'PAUSED':
      return { type: 'warning', label: t(`${P}.PAUSED`) }
    case 'ARCHIVED':
      return { type: 'info', label: t(`${P}.ARCHIVED`) }
    case 'DELETED':
      return { type: 'danger', label: t(`${P}.DELETED`) }
    default:
      return { type: 'primary', label: status || '—' }
  }
}

const statusTag = (config: StatusConfig) =>
  h(ElTag, { type: config.type, size: 'small' }, () => config.label)

// 账户被封禁时（FB 后台口径"账户已停用"），campaign/adset/ad 状态统一显示停用
const accountDisabledTag = (t: (key: string) => string) =>
  statusTag({ type: 'danger', label: t('menus.adCampaign.status.ACCOUNT_DISABLED') })

// 金额：FB 返回字符串，"0"/空 = 未设置
const budgetText = (v: string | undefined) => {
  if (!v || v === '0') return DASH()
  return h('span', `$${v}`)
}

const insightCell = (ins: FbInsight | undefined, key: string) => {
  if (!ins) return DASH()
  return h('span', ins[key as keyof FbInsight] || '—')
}

interface CampaignColsOptions {
  t: (key: string) => string
  isAccountDisabled: (row: any) => boolean
  onViewAdSets: (row: FbCampaign) => void
}
export function buildCampaignColumns({
  t,
  isAccountDisabled,
  onViewAdSets
}: CampaignColsOptions): ColumnOption<FbCampaign>[] {
  const P = 'menus.adCampaign.columns'
  return [
    {
      prop: 'accountName',
      label: t(`${P}.account`),
      minWidth: 140,
      formatter: (row: FbCampaign) =>
        h(
          'span',
          row.accountName
            ? `${row.accountName}${row.accountBm ? '（' + row.accountBm + '）' : ''}`
            : '—'
        )
    },
    {
      prop: 'name',
      label: t(`${P}.name`),
      minWidth: 180,
      formatter: (row: FbCampaign) =>
        h(ElTooltip, { content: row.id, placement: 'top' }, () =>
          h('span', { style: { fontWeight: 500 } }, row.name || '—')
        )
    },
    {
      prop: 'status',
      label: t(`${P}.status`),
      width: 90,
      formatter: (row: FbCampaign) =>
        isAccountDisabled(row) ? accountDisabledTag(t) : statusTag(getStatusConfig(row.status, t))
    },
    {
      prop: 'objective',
      label: t(`${P}.objective`),
      minWidth: 130,
      formatter: (row: FbCampaign) => row.objective || '—'
    },
    {
      prop: 'dailyBudget',
      label: t(`${P}.budget`),
      width: 100,
      formatter: (row: FbCampaign) => budgetText(row.dailyBudget || row.lifetimeBudget)
    },
    {
      prop: 'bidStrategy',
      label: t(`${P}.bidStrategy`),
      minWidth: 140,
      formatter: (row: FbCampaign) => row.bidStrategy || '—'
    },
    {
      prop: 'spend',
      label: t(`${P}.spend`),
      width: 100,
      formatter: (row: FbCampaign) => insightCell(row.insight, 'spend')
    },
    {
      prop: 'impressions',
      label: t(`${P}.impressions`),
      width: 95,
      formatter: (row: FbCampaign) => insightCell(row.insight, 'impressions')
    },
    {
      prop: 'clicks',
      label: t(`${P}.clicks`),
      width: 80,
      formatter: (row: FbCampaign) => insightCell(row.insight, 'clicks')
    },
    {
      prop: 'ctr',
      label: t(`${P}.ctr`),
      width: 80,
      formatter: (row: FbCampaign) => insightCell(row.insight, 'ctr')
    },
    {
      prop: 'cpc',
      label: t(`${P}.cpc`),
      width: 80,
      formatter: (row: FbCampaign) => insightCell(row.insight, 'cpc')
    },
    {
      prop: 'startTime',
      label: t(`${P}.startTime`),
      minWidth: 160,
      formatter: (row: FbCampaign) => row.startTime || '—'
    },
    {
      prop: 'createdTime',
      label: t(`${P}.createdTime`),
      minWidth: 160,
      formatter: (row: FbCampaign) => row.createdTime || '—'
    },
    {
      prop: 'operation',
      label: t('menus.adCampaign.adSetTab'),
      width: 90,
      fixed: 'right',
      formatter: (row: FbCampaign) =>
        h(
          ElButton,
          { size: 'small', link: true, type: 'primary', onClick: () => onViewAdSets(row) },
          () => t('menus.adCampaign.adSetTab')
        )
    }
  ]
}

interface AdSetColsOptions {
  t: (key: string) => string
  isAccountDisabled: (row: any) => boolean
}
export function buildAdSetColumns({
  t,
  isAccountDisabled
}: AdSetColsOptions): ColumnOption<FbAdSet>[] {
  const P = 'menus.adCampaign.columns'
  return [
    {
      prop: 'accountName',
      label: t(`${P}.account`),
      minWidth: 140,
      formatter: (row: FbAdSet) =>
        h(
          'span',
          row.accountName
            ? `${row.accountName}${row.accountBm ? '（' + row.accountBm + '）' : ''}`
            : '—'
        )
    },
    {
      prop: 'campaignName',
      label: t(`${P}.campaignName`),
      minWidth: 150,
      formatter: (row: FbAdSet) => row.campaignName || '—'
    },
    {
      prop: 'name',
      label: t(`${P}.name`),
      minWidth: 180,
      formatter: (row: FbAdSet) =>
        h(ElTooltip, { content: row.id, placement: 'top' }, () =>
          h('span', { style: { fontWeight: 500 } }, row.name || '—')
        )
    },
    {
      prop: 'status',
      label: t(`${P}.status`),
      width: 90,
      formatter: (row: FbAdSet) =>
        isAccountDisabled(row) ? accountDisabledTag(t) : statusTag(getStatusConfig(row.status, t))
    },
    {
      prop: 'optimizationGoal',
      label: t(`${P}.optimizationGoal`),
      minWidth: 140,
      formatter: (row: FbAdSet) => row.optimizationGoal || '—'
    },
    {
      prop: 'billingEvent',
      label: t(`${P}.billingEvent`),
      width: 110,
      formatter: (row: FbAdSet) => row.billingEvent || '—'
    },
    {
      prop: 'dailyBudget',
      label: t(`${P}.budget`),
      width: 100,
      formatter: (row: FbAdSet) => budgetText(row.dailyBudget || row.lifetimeBudget)
    },
    {
      prop: 'startTime',
      label: t(`${P}.startTime`),
      minWidth: 160,
      formatter: (row: FbAdSet) => row.startTime || '—'
    },
    {
      prop: 'createdTime',
      label: t(`${P}.createdTime`),
      minWidth: 160,
      formatter: (row: FbAdSet) => row.createdTime || '—'
    }
  ]
}

export function buildAdColumns(
  t: (key: string) => string,
  isAccountDisabled: (row: any) => boolean
): ColumnOption<FbAd>[] {
  const P = 'menus.adCampaign.columns'
  return [
    {
      prop: 'accountName',
      label: t(`${P}.account`),
      minWidth: 130,
      formatter: (row: FbAd) =>
        h(
          'span',
          row.accountName
            ? `${row.accountName}${row.accountBm ? '（' + row.accountBm + '）' : ''}`
            : '—'
        )
    },
    {
      prop: 'campaignName',
      label: t(`${P}.campaignName`),
      minWidth: 140,
      formatter: (row: FbAd) => row.campaignName || '—'
    },
    {
      prop: 'adsetName',
      label: t(`${P}.adsetName`),
      minWidth: 140,
      formatter: (row: FbAd) => row.adsetName || '—'
    },
    {
      prop: 'name',
      label: t(`${P}.name`),
      minWidth: 180,
      formatter: (row: FbAd) =>
        h(ElTooltip, { content: row.id, placement: 'top' }, () =>
          h('span', { style: { fontWeight: 500 } }, row.name || '—')
        )
    },
    {
      prop: 'status',
      label: t(`${P}.status`),
      width: 90,
      formatter: (row: FbAd) =>
        isAccountDisabled(row) ? accountDisabledTag(t) : statusTag(getStatusConfig(row.status, t))
    },
    {
      prop: 'creativeName',
      label: t(`${P}.creative`),
      minWidth: 220,
      formatter: (row: FbAd) =>
        h(ElTooltip, { content: row.creativeId || '', placement: 'top' }, () =>
          h('span', row.creativeName || '—')
        )
    },
    {
      prop: 'createdTime',
      label: t(`${P}.createdTime`),
      minWidth: 160,
      formatter: (row: FbAd) => row.createdTime || '—'
    },
    {
      prop: 'updatedTime',
      label: t(`${P}.updatedTime`),
      minWidth: 160,
      formatter: (row: FbAd) => row.updatedTime || '—'
    }
  ]
}
