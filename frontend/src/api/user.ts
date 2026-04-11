import { client } from './client'

export interface UserProfile {
  id: string
  email: string
  username: string
  plan: string
  avatar_url: string | null
  created_at: string
}

export interface AvatarUploadResult {
  upload_url: string
  public_url: string
}

export interface AlertChannel {
  id: string
  user_id: string
  name: string
  type: string
  config: Record<string, string>
  is_default: boolean
  created_at: string
}

export const userApi = {
  me: () =>
    client.get<UserProfile>('/me'),

  getAvatarUploadUrl: (contentType: string) =>
    client.post<AvatarUploadResult>('/me/avatar-upload-url', { content_type: contentType }),

  updateAvatar: (avatarUrl: string) =>
    client.patch('/me/avatar', { avatar_url: avatarUrl }),

  listAlertChannels: () =>
    client.get<AlertChannel[]>('/alert-channels'),

  createAlertChannel: (name: string, type: string, config: Record<string, string>, is_default = false) =>
    client.post<AlertChannel>('/alert-channels', { name, type, config, is_default }),

  getAlertChannel: (id: string) =>
    client.get<AlertChannel>(`/alert-channels/${id}`),

  updateAlertChannel: (id: string, name: string) =>
    client.patch(`/alert-channels/${id}`, { name }),

  deleteAlertChannel: (id: string) =>
    client.delete(`/alert-channels/${id}`),

  subscribeMonitor: (monitorId: string, alertChannelId: string) =>
    client.post(`/monitors/${monitorId}/subscribe`, { alert_channel_id: alertChannelId }),

  unsubscribeMonitor: (monitorId: string, channelId: string) =>
    client.delete(`/monitors/${monitorId}/subscriptions/${channelId}`),

  listMonitorSubscriptions: (monitorId: string) =>
    client.get<AlertChannel[]>(`/monitors/${monitorId}/subscriptions`),
}
