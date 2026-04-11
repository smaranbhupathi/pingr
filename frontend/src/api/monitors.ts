import { client } from './client'
import type { Incident } from './incidents'

export interface Monitor {
  id: string
  user_id: string
  name: string
  url: string
  type: string
  interval_seconds: number
  is_active: boolean
  status: 'up' | 'down' | 'paused' | 'pending'
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

export interface MonitorDetail {
  monitor: Monitor
  uptime: UptimeStats
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

export const monitorsApi = {
  list: () =>
    client.get<Monitor[]>('/monitors'),

  get: (id: string) =>
    client.get<MonitorDetail>(`/monitors/${id}`),

  create: (data: CreateMonitorInput) =>
    client.post<Monitor>('/monitors', data),

  update: (id: string, data: Partial<{ name: string; interval_seconds: number; is_active: boolean }>) =>
    client.patch<Monitor>(`/monitors/${id}`, data),

  delete: (id: string) =>
    client.delete(`/monitors/${id}`),

  graph: (id: string) =>
    client.get<CheckDataPoint[]>(`/monitors/${id}/graph`),

  statusPage: (username: string) =>
    client.get<MonitorDetail[]>(`/status/${username}`),
}
