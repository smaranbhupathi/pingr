import { client } from './client'

export interface UserProfile {
  id: string
  email: string
  username: string
  plan: string
  created_at: string
}

export interface AlertChannel {
  id: string
  user_id: string
  type: string
  config: Record<string, string>
  is_default: boolean
  created_at: string
}

export const userApi = {
  me: () =>
    client.get<UserProfile>('/me'),

  listAlertChannels: () =>
    client.get<AlertChannel[]>('/alert-channels'),

  createAlertChannel: (type: string, config: Record<string, string>, is_default = false) =>
    client.post<AlertChannel>('/alert-channels', { type, config, is_default }),

  deleteAlertChannel: (id: string) =>
    client.delete(`/alert-channels/${id}`),

  subscribeMonitor: (monitorId: string, alertChannelId: string) =>
    client.post(`/monitors/${monitorId}/subscribe`, { alert_channel_id: alertChannelId }),

  unsubscribeMonitor: (monitorId: string, channelId: string) =>
    client.delete(`/monitors/${monitorId}/subscriptions/${channelId}`),

  listMonitorSubscriptions: (monitorId: string) =>
    client.get<AlertChannel[]>(`/monitors/${monitorId}/subscriptions`),
}
