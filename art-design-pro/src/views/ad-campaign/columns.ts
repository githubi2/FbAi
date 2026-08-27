// 广告投放表格列配置工厂（从 index.vue 拆出，保持单文件 < 300 行）
import { h } from 'vue'
import { ElTag, ElTooltip, ElButton } from 'element-plus'
import type { FbCampaign, FbAdSet, FbAd, FbInsight, FbAction } from '@/api/facebook'
import type { ColumnOption } from '@/types/component'

const DASH = () => h('span', { style: { color: '#999' } }, '—')

/** 列分组（FB 广告管理工具"编辑列"分组口径） */
export const COLUMN_GROUPS = [
  { key: 'basic', labelKey: 'menus.adCampaign.groups.basic', icon: 'ri:list-unordered' },
  {
    key: 'results',
    labelKey: 'menus.adCampaign.groups.results',
    icon: 'ri:money-dollar-circle-line'
  },
  { key: 'reach', labelKey: 'menus.adCampaign.groups.reach', icon: 'ri:eye-line' }
] as const

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

// 金额：FB 预算单位为账户货币最小单位（分/cent，官方示例 daily_budget=1000 即 $10.00）
const budgetText = (v: string | undefined) => {
  if (!v || v === '0') return DASH()
  const n = parseFloat(v)
  if (Number.isNaN(n)) return h('span', `$${v}`)
  return h('span', `$${(n / 100).toFixed(2)}`)
}

// 指标单元格：空/无数据显示 —
const insightCell = (ins: FbInsight | undefined, key: string) => {
  if (!ins) return DASH()
  const v = ins[key as keyof FbInsight]
  if (typeof v !== 'string' || !v) return DASH()
  return h('span', v)
}

// 动作细分数量（购买数/消息数/线索数…）：action_type 可能带前缀，模糊匹配
const actionOf = (actions: FbAction[] | undefined, type: string) => {
  if (!actions) return ''
  const hit = actions.find(
    (a) =>
      a.type === type ||
      a.type.endsWith('.' + type) ||
      a.type.endsWith('_' + type) ||
      a.type.includes(type)
  )
  return hit?.value || ''
}

// 动作金额（购买金额等，来自 action_values）
const actionValueOf = (values: FbAction[] | undefined, type: string) => {
  if (!values) return ''
  const hit = values.find(
    (a) =>
      a.type === type ||
      a.type.endsWith('.' + type) ||
      a.type.endsWith('_' + type) ||
      a.type.includes(type)
  )
  if (!hit) return ''
  return hit.value2 || hit.value
}

const actionCell = (actions: FbAction[] | undefined, type: string) => {
  const v = actionOf(actions, type)
  return v ? h('span', v) : DASH()
}

const moneyCell = (values: FbAction[] | undefined, type: string) => {
  const v = actionValueOf(values, type)
  return v ? h('span', `$${parseFloat(v).toFixed(2)}`) : DASH()
}

// ===== 排序（客户端）：虚拟 prop 列（insight/actions 提取）需自定义 sortMethod =====
const insightSort = (key: string) => (a: any, b: any) =>
  (parseFloat(a.insight?.[key]) || 0) - (parseFloat(b.insight?.[key]) || 0)
const actionSort = (type: string) => (a: any, b: any) =>
  (parseFloat(actionOf(a.insight?.actions, type)) || 0) -
  (parseFloat(actionOf(b.insight?.actions, type)) || 0)

const SORT_MAP: Record<string, (a: any, b: any) => number> = {
  dailyBudget: (a, b) =>
    (parseFloat(a.dailyBudget || a.lifetimeBudget) || 0) -
    (parseFloat(b.dailyBudget || b.lifetimeBudget) || 0),
  insightSpend: insightSort('spend'),
  insightResults: insightSort('results'),
  insightCostPerResult: insightSort('costPerResult'),
  insightResultRate: insightSort('resultRate'),
  insightImpressions: insightSort('impressions'),
  insightReach: insightSort('reach'),
  insightFrequency: insightSort('frequency'),
  insightClicks: insightSort('clicks'),
  insightCtr: insightSort('ctr'),
  insightCpc: insightSort('cpc'),
  insightCpm: insightSort('cpm'),
  insightCpp: insightSort('cpp'),
  purchaseCount: actionSort('purchase'),
  purchaseValue: actionSort('purchase'),
  messagingCount: actionSort('messaging'),
  leadCount: actionSort('lead'),
  linkClickCount: actionSort('link_click')
}

const withSorting = <T>(cols: ColumnOption<T>[]): ColumnOption<T>[] =>
  cols.map((c) => {
    if (c.type === 'selection' || c.type === 'expand' || c.prop === 'operation') return c
    const sm = SORT_MAP[c.prop as string]
    return sm ? { ...c, sortable: true, sortMethod: sm } : { ...c, sortable: true }
  })

interface CampaignColsOptions {
  t: (key: string) => string
  isAccountDisabled: (row: any) => boolean
  onViewAdSets: () => void
}
export function buildCampaignColumns({
  t,
  isAccountDisabled,
  onViewAdSets
}: CampaignColsOptions): ColumnOption<FbCampaign>[] {
  const P = 'menus.adCampaign.columns'
  return withSorting([
    // ===== 基础 =====
    {
      prop: 'accountName',
      label: t(`${P}.account`),
      group: 'basic',
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
      group: 'basic',
      minWidth: 180,
      formatter: (row: FbCampaign) =>
        h(ElTooltip, { content: row.id, placement: 'top' }, () =>
          h('span', { style: { fontWeight: 500 } }, row.name || '—')
        )
    },
    {
      prop: 'status',
      label: t(`${P}.status`),
      group: 'basic',
      width: 90,
      formatter: (row: FbCampaign) =>
        isAccountDisabled(row) ? accountDisabledTag(t) : statusTag(getStatusConfig(row.status, t))
    },
    {
      prop: 'objective',
      label: t(`${P}.objective`),
      group: 'basic',
      minWidth: 130,
      formatter: (row: FbCampaign) => row.objective || '—'
    },
    {
      prop: 'dailyBudget',
      label: t(`${P}.budget`),
      group: 'basic',
      width: 100,
      formatter: (row: FbCampaign) => budgetText(row.dailyBudget || row.lifetimeBudget)
    },
    {
      prop: 'bidStrategy',
      label: t(`${P}.bidStrategy`),
      group: 'basic',
      width: 130,
      formatter: (row: FbCampaign) => row.bidStrategy || '—'
    },
    {
      prop: 'startTime',
      label: t(`${P}.startTime`),
      group: 'basic',
      minWidth: 160,
      formatter: (row: FbCampaign) => row.startTime || '—'
    },
    {
      prop: 'createdTime',
      label: t(`${P}.createdTime`),
      group: 'basic',
      minWidth: 160,
      formatter: (row: FbCampaign) => row.createdTime || '—'
    },
    // ===== 成效与花费 =====
    {
      prop: 'insightSpend',
      label: t(`${P}.spend`),
      group: 'results',
      width: 100,
      formatter: (row: FbCampaign) => insightCell(row.insight, 'spend')
    },
    {
      prop: 'insightResults',
      label: t(`${P}.results`),
      group: 'results',
      width: 90,
      formatter: (row: FbCampaign) => insightCell(row.insight, 'results')
    },
    {
      prop: 'insightCostPerResult',
      label: t(`${P}.costPerResult`),
      group: 'results',
      width: 110,
      formatter: (row: FbCampaign) => insightCell(row.insight, 'costPerResult')
    },
    {
      prop: 'insightResultRate',
      label: t(`${P}.resultRate`),
      group: 'results',
      width: 100,
      formatter: (row: FbCampaign) => insightCell(row.insight, 'resultRate')
    },
    {
      prop: 'purchaseCount',
      label: t(`${P}.purchases`),
      group: 'results',
      width: 90,
      formatter: (row: FbCampaign) => actionCell(row.insight?.actions, 'purchase')
    },
    {
      prop: 'purchaseValue',
      label: t(`${P}.purchaseValue`),
      group: 'results',
      width: 110,
      formatter: (row: FbCampaign) => moneyCell(row.insight?.actionValues, 'purchase')
    },
    {
      prop: 'messagingCount',
      label: t(`${P}.messages`),
      group: 'results',
      width: 90,
      formatter: (row: FbCampaign) => actionCell(row.insight?.actions, 'messaging')
    },
    {
      prop: 'leadCount',
      label: t(`${P}.leads`),
      group: 'results',
      width: 90,
      formatter: (row: FbCampaign) => actionCell(row.insight?.actions, 'lead')
    },
    {
      prop: 'linkClickCount',
      label: t(`${P}.linkClicks`),
      group: 'results',
      width: 100,
      formatter: (row: FbCampaign) => actionCell(row.insight?.actions, 'link_click')
    },
    // ===== 传播 =====
    {
      prop: 'insightImpressions',
      label: t(`${P}.impressions`),
      group: 'reach',
      width: 100,
      formatter: (row: FbCampaign) => insightCell(row.insight, 'impressions')
    },
    {
      prop: 'insightReach',
      label: t(`${P}.reach`),
      group: 'reach',
      width: 100,
      formatter: (row: FbCampaign) => insightCell(row.insight, 'reach')
    },
    {
      prop: 'insightFrequency',
      label: t(`${P}.frequency`),
      group: 'reach',
      width: 90,
      formatter: (row: FbCampaign) => insightCell(row.insight, 'frequency')
    },
    {
      prop: 'insightClicks',
      label: t(`${P}.clicks`),
      group: 'reach',
      width: 90,
      formatter: (row: FbCampaign) => insightCell(row.insight, 'clicks')
    },
    {
      prop: 'insightCtr',
      label: t(`${P}.ctr`),
      group: 'reach',
      width: 80,
      formatter: (row: FbCampaign) => insightCell(row.insight, 'ctr')
    },
    {
      prop: 'insightCpc',
      label: t(`${P}.cpc`),
      group: 'reach',
      width: 90,
      formatter: (row: FbCampaign) => insightCell(row.insight, 'cpc')
    },
    {
      prop: 'insightCpm',
      label: t(`${P}.cpm`),
      group: 'reach',
      width: 90,
      formatter: (row: FbCampaign) => insightCell(row.insight, 'cpm')
    },
    {
      prop: 'insightCpp',
      label: t(`${P}.cpp`),
      group: 'reach',
      width: 90,
      formatter: (row: FbCampaign) => insightCell(row.insight, 'cpp')
    },
    {
      prop: 'operation',
      label: t('menus.adCampaign.adSetTab'),
      group: 'basic',
      width: 80,
      fixed: 'right',
      formatter: () =>
        h(
          ElButton,
          { size: 'small', link: true, type: 'primary', onClick: () => onViewAdSets() },
          () => t('menus.adCampaign.adSetTab')
        )
    }
  ])
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
  return withSorting([
    {
      prop: 'accountName',
      label: t(`${P}.account`),
      group: 'basic',
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
      group: 'basic',
      minWidth: 150,
      formatter: (row: FbAdSet) => row.campaignName || '—'
    },
    {
      prop: 'name',
      label: t(`${P}.name`),
      group: 'basic',
      minWidth: 180,
      formatter: (row: FbAdSet) =>
        h(ElTooltip, { content: row.id, placement: 'top' }, () =>
          h('span', { style: { fontWeight: 500 } }, row.name || '—')
        )
    },
    {
      prop: 'status',
      label: t(`${P}.status`),
      group: 'basic',
      width: 90,
      formatter: (row: FbAdSet) =>
        isAccountDisabled(row) ? accountDisabledTag(t) : statusTag(getStatusConfig(row.status, t))
    },
    {
      prop: 'optimizationGoal',
      label: t(`${P}.optimizationGoal`),
      group: 'basic',
      minWidth: 140,
      formatter: (row: FbAdSet) => row.optimizationGoal || '—'
    },
    {
      prop: 'billingEvent',
      label: t(`${P}.billingEvent`),
      group: 'basic',
      width: 110,
      formatter: (row: FbAdSet) => row.billingEvent || '—'
    },
    {
      prop: 'dailyBudget',
      label: t(`${P}.budget`),
      group: 'basic',
      width: 100,
      formatter: (row: FbAdSet) => budgetText(row.dailyBudget || row.lifetimeBudget)
    },
    {
      prop: 'startTime',
      label: t(`${P}.startTime`),
      group: 'basic',
      minWidth: 160,
      formatter: (row: FbAdSet) => row.startTime || '—'
    },
    {
      prop: 'createdTime',
      label: t(`${P}.createdTime`),
      group: 'basic',
      minWidth: 160,
      formatter: (row: FbAdSet) => row.createdTime || '—'
    }
  ])
}

export function buildAdColumns(
  t: (key: string) => string,
  isAccountDisabled: (row: any) => boolean
): ColumnOption<FbAd>[] {
  const P = 'menus.adCampaign.columns'
  return withSorting([
    {
      prop: 'accountName',
      label: t(`${P}.account`),
      group: 'basic',
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
      group: 'basic',
      minWidth: 140,
      formatter: (row: FbAd) => row.campaignName || '—'
    },
    {
      prop: 'adsetName',
      label: t(`${P}.adsetName`),
      group: 'basic',
      minWidth: 140,
      formatter: (row: FbAd) => row.adsetName || '—'
    },
    {
      prop: 'name',
      label: t(`${P}.name`),
      group: 'basic',
      minWidth: 180,
      formatter: (row: FbAd) =>
        h(ElTooltip, { content: row.id, placement: 'top' }, () =>
          h('span', { style: { fontWeight: 500 } }, row.name || '—')
        )
    },
    {
      prop: 'status',
      label: t(`${P}.status`),
      group: 'basic',
      width: 90,
      formatter: (row: FbAd) =>
        isAccountDisabled(row) ? accountDisabledTag(t) : statusTag(getStatusConfig(row.status, t))
    },
    {
      prop: 'creativeName',
      label: t(`${P}.creative`),
      group: 'basic',
      minWidth: 220,
      formatter: (row: FbAd) =>
        h(ElTooltip, { content: row.creativeId || '', placement: 'top' }, () =>
          h('span', row.creativeName || '—')
        )
    },
    {
      prop: 'createdTime',
      label: t(`${P}.createdTime`),
      group: 'basic',
      minWidth: 160,
      formatter: (row: FbAd) => row.createdTime || '—'
    },
    {
      prop: 'updatedTime',
      label: t(`${P}.updatedTime`),
      group: 'basic',
      minWidth: 160,
      formatter: (row: FbAd) => row.updatedTime || '—'
    }
  ])
}
