import { client } from './client'

export interface Component {
  id: string
  user_id: string
  name: string
  description: string
  sort_order: number
  created_at: string
  updated_at: string
}

export const componentsApi = {
  list: () =>
    client.get<Component[]>('/components'),

  create: (name: string, description: string) =>
    client.post<Component>('/components', { name, description }),

  update: (id: string, data: { name?: string; description?: string }) =>
    client.patch<Component>(`/components/${id}`, data),

  delete: (id: string) =>
    client.delete(`/components/${id}`),
}
