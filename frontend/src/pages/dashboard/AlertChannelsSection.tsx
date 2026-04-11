import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { userApi } from '../../api/user'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Card } from '../../components/ui/Card'
import { Bell, Trash2 } from 'lucide-react'

type ChannelType = 'email' | 'slack' | 'discord'

const CHANNEL_OPTIONS: { type: ChannelType; label: string; icon: string; placeholder: string }[] = [
  { type: 'email',   label: 'Email',   icon: '✉️',  placeholder: 'alerts@example.com' },
  { type: 'slack',   label: 'Slack',   icon: '💬',  placeholder: 'https://hooks.slack.com/services/...' },
  { type: 'discord', label: 'Discord', icon: '🎮',  placeholder: 'https://discord.com/api/webhooks/...' },
]

function channelLabel(ch: { type: string; config: Record<string, string> }) {
  if (ch.type === 'email')   return ch.config.email
  if (ch.type === 'slack')   return 'Slack webhook'
  if (ch.type === 'discord') return 'Discord webhook'
  return ch.type
}

function channelIcon(type: string) {
  return CHANNEL_OPTIONS.find(o => o.type === type)?.icon ?? '🔔'
}

export function AlertChannelsSection() {
  const queryClient = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [type, setType] = useState<ChannelType>('email')
  const [value, setValue] = useState('')
  const [err, setErr] = useState('')

  const { data: channels = [] } = useQuery({
    queryKey: ['alert-channels'],
    queryFn: () => userApi.listAlertChannels().then(r => r.data ?? []),
  })

  const createMutation = useMutation({
    mutationFn: () => {
      const config = type === 'email'
        ? { email: value }
        : { webhook_url: value }
      return userApi.createAlertChannel(type, config, channels.length === 0)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alert-channels'] })
      setValue('')
      setShowForm(false)
      setErr('')
    },
    onError: (e: any) => setErr(e.response?.data?.error ?? 'Failed to create channel'),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => userApi.deleteAlertChannel(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alert-channels'] }),
  })

  const selected = CHANNEL_OPTIONS.find(o => o.type === type)!

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
          {/* Channel type selector */}
          <div className="flex gap-2 mb-4">
            {CHANNEL_OPTIONS.map(opt => (
              <button
                key={opt.type}
                onClick={() => { setType(opt.type); setValue(''); setErr('') }}
                className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium border transition ${
                  type === opt.type
                    ? 'bg-indigo-50 border-indigo-300 text-indigo-700'
                    : 'bg-white border-gray-200 text-gray-600 hover:border-gray-300'
                }`}
              >
                <span>{opt.icon}</span> {opt.label}
              </button>
            ))}
          </div>

          {/* Input */}
          <div className="flex gap-3 items-end">
            <div className="flex-1">
              <Input
                label={selected.label}
                type={type === 'email' ? 'email' : 'url'}
                placeholder={selected.placeholder}
                value={value}
                onChange={e => setValue(e.target.value)}
              />
            </div>
            <div className="flex gap-2 pb-0.5">
              <Button onClick={() => createMutation.mutate()} loading={createMutation.isPending}>
                Save
              </Button>
              <Button variant="secondary" onClick={() => { setShowForm(false); setErr('') }}>
                Cancel
              </Button>
            </div>
          </div>

          {/* Help text */}
          {type === 'slack' && (
            <p className="text-xs text-gray-400 mt-2">
              Slack → Channel settings → Integrations → Add an App → Incoming Webhooks → Add to Slack → copy URL
            </p>
          )}
          {type === 'discord' && (
            <p className="text-xs text-gray-400 mt-2">
              Discord → Channel settings → Integrations → Webhooks → New Webhook → Copy Webhook URL
            </p>
          )}

          {err && <p className="text-sm text-red-500 mt-2">{err}</p>}
        </Card>
      )}

      {channels.length === 0 && !showForm ? (
        <Card className="p-8 text-center">
          <Bell className="mx-auto mb-3 text-gray-300" size={32} />
          <p className="text-gray-500 text-sm font-medium">No alert channels yet</p>
          <p className="text-gray-400 text-xs mt-1">
            Add email, Slack, or Discord to receive alerts when monitors go down or recover
          </p>
        </Card>
      ) : (
        <div className="space-y-2">
          {channels.map(ch => (
            <Card key={ch.id} className="p-4 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <span className="text-lg">{channelIcon(ch.type)}</span>
                <div>
                  <p className="text-sm font-medium text-gray-900">{channelLabel(ch)}</p>
                  <p className="text-xs text-gray-400 capitalize">
                    {ch.type}{ch.is_default ? ' · Default' : ''}
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
