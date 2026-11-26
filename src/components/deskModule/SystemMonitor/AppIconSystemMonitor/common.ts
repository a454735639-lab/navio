import { PanelPanelConfigStyleEnum } from '@/enums'
import { bytesToSize } from '@/utils/cmn'

export function correctionNumberByCardStyle(v: number, cardStyle: PanelPanelConfigStyleEnum): number {
  let keepNum = 0
  if (cardStyle === PanelPanelConfigStyleEnum.small)
    keepNum = 1
  else if (cardStyle === PanelPanelConfigStyleEnum.info)
    keepNum = 2

  return correctionNumber(v, keepNum)
}

export function correctionNumber(v: number, keepNum = 2): number {
  return v === 0 ? 0 : Number(v.toFixed(keepNum))
}

export function allCPUUsages(cpuInfo: SystemMonitor.CPUInfo | null) {
  if (!cpuInfo)
    return 0
  let countPercentage = 0
  for (let i = 0; i < cpuInfo.usages.length; i++) {
    const element = cpuInfo.usages[i]
    countPercentage += element
  }
  return countPercentage / cpuInfo.usages.length
}

/** 内存/磁盘用量文本：单位相同时只显示一个（如 5.3/7.5 GB），窄屏下更省宽度 */

export function sizePairText(used: number, total: number): string {
  const a = bytesToSize(used)
  const b = bytesToSize(total)
  const [na, ua] = a.includes(' ') ? a.split(' ') : [a, '']
  const [nb, ub] = b.includes(' ') ? b.split(' ') : [b, '']
  if (!ua || !ub || ua !== ub)
    return `${a}/${b}`
  return `${na}/${nb} ${ua}`
}
