import { client as api } from './client'

export type IncidentStatus = 'investigating' | 'identified' | 'monitoring' | 'resolved'

export interface IncidentUpdate {
  id: string
  incident_id: string
  status: IncidentStatus
  message: string
  notify: boolean
  created_at: string
}

export interface Incident {
  id: string
  user_id: string
  name: string
  status: IncidentStatus
  source: 'manual' | 'auto'
  resolved_at: string | null
  created_at: string
  updated_at: string
  updates: IncidentUpdate[]
  monitor_ids: string[]
}

export const incidentsApi = {
  list: () => api.get<Incident[]>('/incidents'),
  get: (id: string) => api.get<Incident>(`/incidents/${id}`),
  create: (body: {
    name: string
    status: IncidentStatus
    message: string
    monitor_ids: string[]
    notify: boolean
  }) => api.post<Incident>('/incidents', body),
  postUpdate: (id: string, body: {
    status: IncidentStatus
    message: string
    notify: boolean
  }) => api.post<Incident>(`/incidents/${id}/updates`, body),
}
