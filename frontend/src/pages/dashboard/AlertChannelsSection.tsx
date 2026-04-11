import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { userApi, AlertChannel } from '../../api/user'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Card } from '../../components/ui/Card'
import { Bell, ChevronRight, Trash2 } from 'lucide-react'

type ChannelType = 'email' | 'slack' | 'discord'

const CHANNEL_OPTIONS: { type: ChannelType; label: string; icon: string; placeholder: string }[] = [
  { type: 'email',   label: 'Email',   icon: '✉️',  placeholder: 'alerts@example.com' },
  { type: 'slack',   label: 'Slack',   icon: '💬',  placeholder: 'https://hooks.slack.com/services/...' },
  { type: 'discord', label: 'Discord', icon: '🎮',  placeholder: 'https://discord.com/api/webhooks/...' },
]

function channelDisplayName(ch: { name: string; type: string; config: Record<string, string> }) {
  if (ch.name) return ch.name
  if (ch.type === 'email') return ch.config.email || 'Email channel'
  const label = CHANNEL_OPTIONS.find(o => o.type === ch.type)?.label ?? ch.type
  return `${label} webhook`
}

function channelSubtitle(ch: { name: string; type: string; config: Record<string, string> }) {
  const typeLabel = CHANNEL_OPTIONS.find(o => o.type === ch.type)?.label ?? ch.type
  if (ch.type === 'email') return `${typeLabel} · ${ch.config.email || ''}`
  const url = ch.config.webhook_url || ''
  // Show just the domain part of the webhook URL to keep it short
  const short = url ? new URL(url).hostname : ''
  return `${typeLabel}${short ? ` · ${short}` : ''}`
}

function channelIcon(type: string) {
  return CHANNEL_OPTIONS.find(o => o.type === type)?.icon ?? '🔔'
}

function Toggle({ enabled, onChange }: { enabled: boolean; onChange: (v: boolean) => void }) {
  return (
    <button
      type="button"
      onClick={e => { e.preventDefault(); e.stopPropagation(); onChange(!enabled) }}
      className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 focus:outline-none ${
        enabled ? 'bg-indigo-500' : 'bg-gray-200'
      }`}
      title={enabled ? 'Disable channel' : 'Enable channel'}
    >
      <span className={`pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow transform transition-transform duration-200 ${
        enabled ? 'translate-x-4' : 'translate-x-0'
      }`} />
    </button>
  )
}

export function AlertChannelsSection() {
  const queryClient = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [type, setType] = useState<ChannelType>('email')
  const [name, setName] = useState('')
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
      return userApi.createAlertChannel(name, type, config, channels.length === 0)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alert-channels'] })
      setName('')
      setValue('')
      setShowForm(false)
      setErr('')
    },
    onError: (e: any) => setErr(e.response?.data?.error ?? 'Failed to create channel'),
  })

  const toggleMutation = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      userApi.toggleAlertChannel(id, enabled),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alert-channels'] }),
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
                onClick={() => { setType(opt.type); setValue(''); setErr(''); }}
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

          {/* Name + value inputs */}
          <div className="flex gap-3 items-end">
            <div className="w-40 shrink-0">
              <Input
                label="Channel name"
                type="text"
                placeholder="e.g. Team Slack"
                value={name}
                onChange={e => setName(e.target.value)}
              />
            </div>
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
            <Card key={ch.id} className={`p-4 flex items-center justify-between group hover:border-indigo-200 transition-colors ${!ch.is_enabled ? 'opacity-60' : ''}`}>
              <Link
                to={`/dashboard/alert-channels/${ch.id}`}
                className="flex items-center gap-3 flex-1 min-w-0"
              >
                <span className="text-lg shrink-0">{channelIcon(ch.type)}</span>
                <div className="min-w-0">
                  <p className="text-sm font-medium text-gray-900 truncate">
                    {channelDisplayName(ch)}
                    {ch.is_default && (
                      <span className="ml-2 text-xs font-normal text-indigo-500 bg-indigo-50 px-1.5 py-0.5 rounded">Default</span>
                    )}
                    {!ch.is_enabled && (
                      <span className="ml-2 text-xs font-normal text-gray-400 bg-gray-100 px-1.5 py-0.5 rounded">Paused</span>
                    )}
                  </p>
                  <p className="text-xs text-gray-400 truncate">{channelSubtitle(ch)}</p>
                </div>
                <ChevronRight size={14} className="text-gray-300 group-hover:text-gray-400 shrink-0 ml-auto mr-2" />
              </Link>
              <div className="flex items-center gap-2 shrink-0">
                <Toggle
                  enabled={ch.is_enabled}
                  onChange={enabled => toggleMutation.mutate({ id: ch.id, enabled })}
                />
                <button
                  onClick={() => deleteMutation.mutate(ch.id)}
                  className="p-1.5 rounded text-gray-400 hover:text-red-500 hover:bg-red-50"
                  title="Delete channel"
                >
                  <Trash2 size={16} />
                </button>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
