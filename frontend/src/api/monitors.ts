import { client } from './client'
import type { Incident } from './incidents'

export type ComponentStatus =
  | 'operational'
  | 'degraded_performance'
  | 'partial_outage'
  | 'major_outage'
  | 'under_maintenance'

export interface Monitor {
  id: string
  user_id: string
  name: string
  description: string
  url: string
  type: string
  interval_seconds: number
  is_active: boolean
  status: 'up' | 'down' | 'paused' | 'pending'
  component_status: ComponentStatus
  component_id: string | null
  last_checked_at: string | null
  created_at: string
}

export interface MonitorCheck {
  id: string
  monitor_id: string
  checked_at: string
  is_up: boolean
  status_code: number | null
  response_time_ms: number
  error_message: string
  region: string
}

export type { Incident }

export interface UptimeStats {
  last_24h: number
  last_7d: number
  last_30d: number
  last_90d: number
}

export interface DailyUptimeStat {
  date: string    // "2026-04-11"
  uptime: number  // 0–100, or -1 = no data
}

export interface MonitorDetail {
  monitor: Monitor
  uptime: UptimeStats
  daily_uptime: DailyUptimeStat[]
  recent_check: MonitorCheck | null
  incidents: Incident[]
  active_incident: Incident | null
}

export interface CheckDataPoint {
  timestamp: string
  response_time_ms: number
  is_up: boolean
}

export interface CreateMonitorInput {
  name: string
  url: string
  interval_seconds: number
}

export const COMPONENT_STATUS_LABEL: Record<ComponentStatus, string> = {
  operational:          'Operational',
  degraded_performance: 'Degraded Performance',
  partial_outage:       'Partial Outage',
  major_outage:         'Major Outage',
  under_maintenance:    'Under Maintenance',
}

export const COMPONENT_STATUS_COLOR: Record<ComponentStatus, string> = {
  operational:          'bg-green-100 text-green-700 border border-green-200',
  degraded_performance: 'bg-yellow-100 text-yellow-700 border border-yellow-200',
  partial_outage:       'bg-orange-100 text-orange-700 border border-orange-200',
  major_outage:         'bg-red-100 text-red-700 border border-red-200',
  under_maintenance:    'bg-blue-100 text-blue-700 border border-blue-200',
}

export const COMPONENT_STATUS_DOT: Record<ComponentStatus, string> = {
  operational:          'bg-green-500',
  degraded_performance: 'bg-yellow-400',
  partial_outage:       'bg-orange-500',
  major_outage:         'bg-red-500',
  under_maintenance:    'bg-blue-500',
}

export const monitorsApi = {
  list: () =>
    client.get<Monitor[]>('/monitors'),

  get: (id: string) =>
    client.get<MonitorDetail>(`/monitors/${id}`),

  create: (data: CreateMonitorInput) =>
    client.post<Monitor>('/monitors', data),

  update: (id: string, data: Partial<{ name: string; interval_seconds: number; is_active: boolean; component_id: string | null }>) =>
    client.patch<Monitor>(`/monitors/${id}`, data),

  updateMeta: (id: string, data: { name: string; description: string; component_id: string | null }) =>
    client.patch<Monitor>(`/monitors/${id}/meta`, data),

  delete: (id: string) =>
    client.delete(`/monitors/${id}`),

  graph: (id: string) =>
    client.get<CheckDataPoint[]>(`/monitors/${id}/graph`),

  statusPage: (username: string) =>
    client.get<MonitorDetail[]>(`/status/${username}`),
}
