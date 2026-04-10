import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { userApi } from '../../api/user'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Card } from '../../components/ui/Card'
import { Bell, Trash2 } from 'lucide-react'

export function AlertChannelsSection() {
  const queryClient = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [email, setEmail] = useState('')
  const [err, setErr] = useState('')

  const { data: channels = [] } = useQuery({
    queryKey: ['alert-channels'],
    queryFn: () => userApi.listAlertChannels().then(r => r.data ?? []),
  })

  const createMutation = useMutation({
    mutationFn: () => userApi.createAlertChannel('email', { email }, channels.length === 0),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alert-channels'] })
      setEmail('')
      setShowForm(false)
      setErr('')
    },
    onError: () => setErr('Failed to create channel'),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => userApi.deleteAlertChannel(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alert-channels'] }),
  })

  return (
    <div className="mt-10">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-gray-900">Alert Channels</h2>
        <Button variant="secondary" onClick={() => { setShowForm(s => !s); setErr('') }}>
          + Add channel
        </Button>
      </div>

      {showForm && (
        <Card className="p-4 mb-3">
          <p className="text-sm font-medium text-gray-700 mb-3">New email alert channel</p>
          <div className="flex gap-3 items-end">
            <div className="flex-1">
              <Input
                label="Email address"
                type="email"
                placeholder="alerts@example.com"
                value={email}
                onChange={e => setEmail(e.target.value)}
              />
            </div>
            <div className="flex gap-2 pb-0.5">
              <Button
                onClick={() => createMutation.mutate()}
                loading={createMutation.isPending}
              >
                Save
              </Button>
              <Button variant="secondary" onClick={() => { setShowForm(false); setErr('') }}>
                Cancel
              </Button>
            </div>
          </div>
          {err && <p className="text-sm text-red-500 mt-2">{err}</p>}
        </Card>
      )}

      {channels.length === 0 && !showForm ? (
        <Card className="p-8 text-center">
          <Bell className="mx-auto mb-3 text-gray-300" size={32} />
          <p className="text-gray-500 text-sm font-medium">No alert channels yet</p>
          <p className="text-gray-400 text-xs mt-1">
            Add an email to receive alerts when monitors go down or recover
          </p>
        </Card>
      ) : (
        <div className="space-y-2">
          {channels.map(ch => (
            <Card key={ch.id} className="p-4 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <Bell size={16} className="text-indigo-500 shrink-0" />
                <div>
                  <p className="text-sm font-medium text-gray-900">{ch.config.email}</p>
                  <p className="text-xs text-gray-400">
                    Email{ch.is_default ? ' · Default' : ''}
                  </p>
                </div>
              </div>
              <button
                onClick={() => deleteMutation.mutate(ch.id)}
                className="p-1.5 rounded text-gray-400 hover:text-red-500 hover:bg-red-50"
                title="Delete channel"
              >
                <Trash2 size={16} />
              </button>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
