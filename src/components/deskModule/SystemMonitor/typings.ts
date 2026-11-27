export enum MonitorType {
  'cpu' = 'cpu', // 图标风格
  'memory' = 'memory', // 详情风格
  'disk' = 'disk',
  'network' = 'network', // 网络流量
}

export interface CardStyle {
  background: string
}

export interface MonitorData {
  monitorType: MonitorType
  extendParam?: { [key: string]: [value:any] } | any
  description?: string
  // cardStyle: CardStyle
}

export interface ProgressStyle {
  color: string
  railColor: string
  height: number
}

export interface GenericProgressStyleExtendParam {
  progressColor: string
  progressRailColor: string
  color: string
  backgroundColor: string
  /** 警告阈值：CPU/内存/磁盘为百分比(0-100)，网络为 KB/s。0 或不填表示不启用 */
  warningThreshold?: number
  /** 达到/超过阈值时使用的警告色 */
  warningColor?: string
}

export interface DiskExtendParam extends GenericProgressStyleExtendParam {
  path: string
  title: string
}
