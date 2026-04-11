import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { userApi } from '../../api/user'
import { DashboardLayout } from '../../components/layout/DashboardLayout'
import { Card } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { ArrowLeft, Trash2, Check, Pencil } from 'lucide-react'
import { usePageTitle } from '../../lib/usePageTitle'
import { format } from '../../lib/format'

const TYPE_META: Record<string, { label: string; icon: string; configKey: string; configLabel: string }> = {
  email:   { label: 'Email',   icon: '✉️',  configKey: 'email',       configLabel: 'Email address' },
  slack:   { label: 'Slack',   icon: '💬',  configKey: 'webhook_url', configLabel: 'Webhook URL' },
  discord: { label: 'Discord', icon: '🎮',  configKey: 'webhook_url', configLabel: 'Webhook URL' },
}

export function AlertChannelDetailPage() {
  usePageTitle('Alert channel')
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const { data: ch, isLoading } = useQuery({
    queryKey: ['alert-channel', id],
    queryFn: () => userApi.getAlertChannel(id!).then(r => r.data),
  })

  const [editingName, setEditingName] = useState(false)
  const [draftName, setDraftName] = useState('')

  useEffect(() => {
    if (ch) setDraftName(ch.name)
  }, [ch])

  const renameMutation = useMutation({
    mutationFn: () => userApi.updateAlertChannel(id!, draftName),
    onSuccess: () => {
      setEditingName(false)
      queryClient.invalidateQueries({ queryKey: ['alert-channel', id] })
      queryClient.invalidateQueries({ queryKey: ['alert-channels'] })
    },
  })

  const toggleMutation = useMutation({
    mutationFn: (enabled: boolean) => userApi.toggleAlertChannel(id!, enabled),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alert-channel', id] })
      queryClient.invalidateQueries({ queryKey: ['alert-channels'] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: () => userApi.deleteAlertChannel(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alert-channels'] })
      navigate('/dashboard/alert-channels')
    },
  })

  if (isLoading || !ch) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-64 text-gray-400 text-sm">Loading…</div>
      </DashboardLayout>
    )
  }

  const meta = TYPE_META[ch.type] ?? { label: ch.type, icon: '🔔', configKey: 'webhook_url', configLabel: 'Config' }
  const configValue = ch.config[meta.configKey] ?? ''

  // Mask the middle of webhook URLs — keep scheme+host visible
  function maskWebhook(url: string) {
    try {
      const u = new URL(url)
      return `${u.protocol}//${u.hostname}/••••••`
    } catch {
      return url
    }
  }

  const displayConfig = ch.type === 'email' ? configValue : maskWebhook(configValue)

  return (
    <DashboardLayout>
      <div className="max-w-xl mx-auto">
        <Link
          to="/dashboard/alert-channels"
          className="inline-flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700 mb-6"
        >
          <ArrowLeft size={14} /> Back to channels
        </Link>

        {/* Name + type header */}
        <div className="flex items-start justify-between mb-6">
          <div className="flex items-center gap-3">
            <span className="text-3xl">{meta.icon}</span>
            <div>
              {editingName ? (
                <div className="flex items-center gap-2">
                  <input
                    autoFocus
                    value={draftName}
                    onChange={e => setDraftName(e.target.value)}
                    onKeyDown={e => {
                      if (e.key === 'Enter') renameMutation.mutate()
                      if (e.key === 'Escape') { setEditingName(false); setDraftName(ch.name) }
                    }}
                    className="text-xl font-semibold text-gray-900 border-b border-indigo-400 focus:outline-none bg-transparent w-56"
                  />
                  <button
                    onClick={() => renameMutation.mutate()}
                    disabled={renameMutation.isPending || !draftName}
                    className="p-1 rounded text-indigo-500 hover:bg-indigo-50 disabled:opacity-50"
                    title="Save"
                  >
                    <Check size={16} />
                  </button>
                </div>
              ) : (
                <div className="flex items-center gap-2">
                  <h1 className="text-xl font-semibold text-gray-900">{ch.name || `${meta.label} channel`}</h1>
                  <button
                    onClick={() => setEditingName(true)}
                    className="p-1 rounded text-gray-400 hover:text-gray-600 hover:bg-gray-100"
                    title="Rename"
                  >
                    <Pencil size={14} />
                  </button>
                </div>
              )}
              <p className="text-sm text-gray-400 mt-0.5">
                {meta.label}{ch.is_default ? ' · Default channel' : ''}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2">
              <span className="text-xs text-gray-500">{ch.is_enabled ? 'Enabled' : 'Paused'}</span>
              <button
                type="button"
                onClick={() => toggleMutation.mutate(!ch.is_enabled)}
                disabled={toggleMutation.isPending}
                className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 disabled:opacity-50 ${
                  ch.is_enabled ? 'bg-indigo-500' : 'bg-gray-200'
                }`}
              >
                <span className={`pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow transform transition-transform duration-200 ${
                  ch.is_enabled ? 'translate-x-4' : 'translate-x-0'
                }`} />
              </button>
            </div>
            <button
              onClick={() => {
                if (confirm('Delete this alert channel? Existing monitor subscriptions will also be removed.')) {
                  deleteMutation.mutate()
                }
              }}
              disabled={deleteMutation.isPending}
              className="flex items-center gap-1.5 px-3 py-2 rounded-lg text-sm border border-red-200 text-red-500 hover:bg-red-50 disabled:opacity-50 transition-colors"
            >
              <Trash2 size={14} /> Delete
            </button>
          </div>
        </div>

        {/* Config detail */}
        <Card className="p-6 mb-4 space-y-4">
          <div>
            <p className="text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">{meta.configLabel}</p>
            <p className="text-sm text-gray-800 font-mono bg-gray-50 rounded px-3 py-2 break-all">{displayConfig}</p>
            {ch.type !== 'email' && (
              <p className="text-xs text-gray-400 mt-1">URL is partially masked for security</p>
            )}
          </div>

          <div>
            <p className="text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">Created</p>
            <p className="text-sm text-gray-700">{format.datetime(ch.created_at)}</p>
          </div>
        </Card>

        {/* Instructions */}
        {ch.type === 'slack' && (
          <Card className="p-4">
            <p className="text-xs font-medium text-gray-500 mb-1">How to update the Slack webhook</p>
            <p className="text-xs text-gray-400 leading-relaxed">
              Slack → Channel settings → Integrations → Incoming Webhooks → copy a new URL,
              then delete this channel and create a new one.
            </p>
          </Card>
        )}
        {ch.type === 'discord' && (
          <Card className="p-4">
            <p className="text-xs font-medium text-gray-500 mb-1">How to update the Discord webhook</p>
            <p className="text-xs text-gray-400 leading-relaxed">
              Discord → Channel settings → Integrations → Webhooks → edit or copy the URL,
              then delete this channel and create a new one.
            </p>
          </Card>
        )}
      </div>
    </DashboardLayout>
  )
}
