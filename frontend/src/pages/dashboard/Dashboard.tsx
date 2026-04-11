import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { monitorsApi, type Monitor } from '../../api/monitors'
import { DashboardLayout } from '../../components/layout/DashboardLayout'
import { StatusBadge } from '../../components/ui/StatusBadge'
import { Button } from '../../components/ui/Button'
import { AddMonitorModal } from './AddMonitorModal'
import { Globe, Trash2, PauseCircle, PlayCircle } from 'lucide-react'
import { usePageTitle } from '../../lib/usePageTitle'

function timeAgo(iso: string | null | undefined): string {
  if (!iso) return '—'
  const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000)
  if (diff < 60) return `${diff}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return `${Math.floor(diff / 86400)}d ago`
}

export function DashboardPage() {
  usePageTitle('Monitors')
  const [showAdd, setShowAdd] = useState(false)
  const queryClient = useQueryClient()

  const { data: monitors = [], isLoading } = useQuery({
    queryKey: ['monitors'],
    queryFn: () => monitorsApi.list().then(r => r.data ?? []),
    refetchInterval: 30_000,
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => monitorsApi.delete(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['monitors'] }),
  })

  const toggleMutation = useMutation({
    mutationFn: ({ id, is_active }: { id: string; is_active: boolean }) =>
      monitorsApi.update(id, { is_active }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['monitors'] }),
  })

  const up = monitors.filter(m => m.status === 'up').length
  const down = monitors.filter(m => m.status === 'down').length

  return (
    <DashboardLayout>
      {/* Page header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-semibold text-gray-900 dark:text-white">Monitors</h1>
          {monitors.length > 0 && (
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
              {up} up · {down} down · {monitors.length} total
            </p>
          )}
        </div>
        <Button onClick={() => setShowAdd(true)}>+ Add monitor</Button>
      </div>

      {/* Table */}
      {isLoading ? (
        <div className="text-center py-20 text-gray-400 text-sm">Loading…</div>
      ) : monitors.length === 0 ? (
        <div className="text-center py-24 border border-dashed border-gray-200 dark:border-gray-700 rounded-xl">
          <Globe className="mx-auto mb-3 text-gray-300 dark:text-gray-600" size={36} />
          <p className="text-sm font-medium text-gray-600 dark:text-gray-400">No monitors yet</p>
          <p className="text-xs text-gray-400 dark:text-gray-500 mt-1 mb-4">Add a URL to start tracking uptime</p>
          <Button onClick={() => setShowAdd(true)}>Add monitor</Button>
        </div>
      ) : (
        <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden">
          {/* Table header */}
          <div className="grid grid-cols-[2fr_1fr_1fr_1fr_auto] gap-4 px-5 py-3 border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
            <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Monitor</span>
            <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Status</span>
            <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Interval</span>
            <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Last checked</span>
            <span />
          </div>

          {/* Rows */}
          <div className="divide-y divide-gray-100 dark:divide-gray-800">
            {monitors.map(m => (
              <MonitorRow
                key={m.id}
                monitor={m}
                onDelete={() => deleteMutation.mutate(m.id)}
                onToggle={() => toggleMutation.mutate({ id: m.id, is_active: !m.is_active })}
              />
            ))}
          </div>
        </div>
      )}

      {showAdd && <AddMonitorModal onClose={() => setShowAdd(false)} />}
    </DashboardLayout>
  )
}

function MonitorRow({
  monitor: m,
  onDelete,
  onToggle,
}: {
  monitor: Monitor
  onDelete: () => void
  onToggle: () => void
}) {
  return (
    <div className="grid grid-cols-[2fr_1fr_1fr_1fr_auto] gap-4 items-center px-5 py-3.5 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors group">
      {/* Name + URL */}
      <Link to={`/dashboard/monitors/${m.id}`} className="min-w-0">
        <p className="text-sm font-medium text-gray-900 dark:text-white truncate">{m.name}</p>
        <p className="text-xs text-gray-400 dark:text-gray-500 truncate mt-0.5">{m.url}</p>
      </Link>

      {/* Status */}
      <div>
        <StatusBadge status={m.is_active ? m.status : 'paused'} />
      </div>

      {/* Interval */}
      <span className="text-sm text-gray-500 dark:text-gray-400">
        {m.interval_seconds < 60 ? `${m.interval_seconds}s` : `${m.interval_seconds / 60}m`}
      </span>

      {/* Last checked */}
      <span className="text-sm text-gray-400 dark:text-gray-500">{timeAgo(m.last_checked_at)}</span>

      {/* Actions — visible on hover */}
      <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
        <button
          onClick={onToggle}
          className="p-1.5 rounded text-gray-400 hover:text-indigo-600 hover:bg-indigo-50"
          title={m.is_active ? 'Pause' : 'Resume'}
        >
          {m.is_active ? <PauseCircle size={15} /> : <PlayCircle size={15} />}
        </button>
        <button
          onClick={onDelete}
          className="p-1.5 rounded text-gray-400 hover:text-red-500 hover:bg-red-50"
          title="Delete"
        >
          <Trash2 size={15} />
        </button>
      </div>
    </div>
  )
}
