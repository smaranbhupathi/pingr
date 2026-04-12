import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { userApi } from '../../api/user'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Bell, ChevronRight, Trash2, Upload, Download } from 'lucide-react'
import { ImportExportModal } from './ImportExportModal'

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
  if (ch.type === 'email') return ch.config.email || ''
  const url = ch.config.webhook_url || ''
  try { return new URL(url).hostname } catch { return '' }
}

function channelIcon(type: string) {
  return CHANNEL_OPTIONS.find(o => o.type === type)?.icon ?? '🔔'
}

function channelTypeLabel(type: string) {
  return CHANNEL_OPTIONS.find(o => o.type === type)?.label ?? type
}

function Toggle({ enabled, onChange }: { enabled: boolean; onChange: (v: boolean) => void }) {
  return (
    <button
      type="button"
      onClick={e => { e.preventDefault(); e.stopPropagation(); onChange(!enabled) }}
      className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 focus:outline-none ${
        enabled ? 'bg-indigo-500' : 'bg-gray-200 dark:bg-gray-600'
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
  const [importExportTab, setImportExportTab] = useState<'import' | 'export' | null>(null)

  const { data: channels = [] } = useQuery({
    queryKey: ['alert-channels'],
    queryFn: () => userApi.listAlertChannels().then(r => r.data ?? []),
  })

  const createMutation = useMutation({
    mutationFn: () => {
      const config: Record<string, string> = type === 'email' ? { email: value } : { webhook_url: value }
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
    <div>
      {/* Page header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-semibold text-gray-900 dark:text-white">Alert Channels</h1>
          {channels.length > 0 && (
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
              {channels.filter(c => c.is_enabled).length} active · {channels.length} total
            </p>
          )}
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setImportExportTab('import')}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:border-gray-300 dark:hover:border-gray-600 hover:text-gray-900 dark:hover:text-white transition-colors"
          >
            <Upload size={14} /> Import
          </button>
          <button
            onClick={() => setImportExportTab('export')}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:border-gray-300 dark:hover:border-gray-600 hover:text-gray-900 dark:hover:text-white transition-colors"
          >
            <Download size={14} /> Export
          </button>
          <Button onClick={() => { setShowForm(s => !s); setErr('') }}>
            + Add channel
          </Button>
        </div>
      </div>

      {/* Add channel form */}
      {showForm && (
        <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 mb-4">
          {/* Type selector */}
          <div className="flex gap-2 mb-4">
            {CHANNEL_OPTIONS.map(opt => (
              <button
                key={opt.type}
                onClick={() => { setType(opt.type); setValue(''); setErr('') }}
                className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium border transition-colors ${
                  type === opt.type
                    ? 'bg-indigo-50 dark:bg-indigo-900/30 border-indigo-300 dark:border-indigo-700 text-indigo-700 dark:text-indigo-300'
                    : 'bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:border-gray-300 dark:hover:border-gray-600'
                }`}
              >
                <span>{opt.icon}</span> {opt.label}
              </button>
            ))}
          </div>

          <div className="flex gap-3 items-end">
            <div className="w-40 shrink-0">
              <Input label="Channel name" type="text" placeholder="e.g. Team Slack" value={name} onChange={e => setName(e.target.value)} />
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
              <Button onClick={() => createMutation.mutate()} loading={createMutation.isPending}>Save</Button>
              <Button variant="secondary" onClick={() => { setShowForm(false); setErr('') }}>Cancel</Button>
            </div>
          </div>

          {type === 'slack' && (
            <p className="text-xs text-gray-400 dark:text-gray-500 mt-2">
              Slack → Channel settings → Integrations → Add an App → Incoming Webhooks → Add to Slack → copy URL
            </p>
          )}
          {type === 'discord' && (
            <p className="text-xs text-gray-400 dark:text-gray-500 mt-2">
              Discord → Channel settings → Integrations → Webhooks → New Webhook → Copy Webhook URL
            </p>
          )}
          {err && <p className="text-sm text-red-500 mt-2">{err}</p>}
        </div>
      )}

      {/* Channels table */}
      {channels.length === 0 && !showForm ? (
        <div className="text-center py-24 border border-dashed border-gray-200 dark:border-gray-700 rounded-xl">
          <Bell className="mx-auto mb-3 text-gray-300 dark:text-gray-600" size={36} />
          <p className="text-sm font-medium text-gray-600 dark:text-gray-400">No alert channels yet</p>
          <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">Add email, Slack, or Discord to get notified when monitors go down</p>
        </div>
      ) : channels.length > 0 && (
        <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden">
          {/* Table header */}
          <div className="grid grid-cols-[2fr_1fr_auto_auto] gap-4 px-5 py-3 border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
            <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Channel</span>
            <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Type</span>
            <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Enabled</span>
            <span />
          </div>

          <div className="divide-y divide-gray-100 dark:divide-gray-800">
            {channels.map(ch => (
              <div key={ch.id} className={`grid grid-cols-[2fr_1fr_auto_auto] gap-4 items-center px-5 py-3.5 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors group ${!ch.is_enabled ? 'opacity-60' : ''}`}>
                {/* Name */}
                <Link to={`/dashboard/alert-channels/${ch.id}`} className="flex items-center gap-3 min-w-0">
                  <span className="text-lg shrink-0">{channelIcon(ch.type)}</span>
                  <div className="min-w-0">
                    <div className="flex items-center gap-2">
                      <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                        {channelDisplayName(ch)}
                      </p>
                      {ch.is_default && (
                        <span className="shrink-0 text-xs font-medium text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-900/30 px-1.5 py-0.5 rounded">
                          Default
                        </span>
                      )}
                    </div>
                    {channelSubtitle(ch) && (
                      <p className="text-xs text-gray-400 dark:text-gray-500 truncate">{channelSubtitle(ch)}</p>
                    )}
                  </div>
                </Link>

                {/* Type */}
                <span className="text-sm text-gray-500 dark:text-gray-400">{channelTypeLabel(ch.type)}</span>

                {/* Toggle */}
                <Toggle
                  enabled={ch.is_enabled}
                  onChange={enabled => toggleMutation.mutate({ id: ch.id, enabled })}
                />

                {/* Actions */}
                <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <Link to={`/dashboard/alert-channels/${ch.id}`}>
                    <span className="p-1.5 rounded text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 inline-block">
                      <ChevronRight size={15} />
                    </span>
                  </Link>
                  <button
                    onClick={() => deleteMutation.mutate(ch.id)}
                    className="p-1.5 rounded text-gray-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20"
                    title="Delete channel"
                  >
                    <Trash2 size={15} />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {importExportTab && (
        <ImportExportModal
          channels={channels}
          defaultTab={importExportTab}
          onClose={() => setImportExportTab(null)}
        />
      )}
    </div>
  )
}
