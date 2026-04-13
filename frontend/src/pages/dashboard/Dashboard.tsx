import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { monitorsApi, COMPONENT_STATUS_LABEL, COMPONENT_STATUS_DOT, type Monitor } from '../../api/monitors'
import { componentsApi, type Component } from '../../api/components'
import { userApi } from '../../api/user'
import { DashboardLayout } from '../../components/layout/DashboardLayout'
import { StatusBadge } from '../../components/ui/StatusBadge'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { AddMonitorModal } from './AddMonitorModal'
import { SlugSetupModal } from './SlugSetupModal'
import { Globe, Trash2, PauseCircle, PlayCircle, ChevronDown, ChevronRight, Pencil, Layers } from 'lucide-react'
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
  const [editingMonitor, setEditingMonitor] = useState<Monitor | null>(null)
  const [slugDismissed, setSlugDismissed] = useState(false)
  const queryClient = useQueryClient()

  const { data: profile } = useQuery({
    queryKey: ['me'],
    queryFn: () => userApi.me().then(r => r.data),
  })

  const showSlugModal = !slugDismissed && profile !== undefined && profile?.status_page_slug == null

  const { data: monitors = [], isLoading } = useQuery({
    queryKey: ['monitors'],
    queryFn: () => monitorsApi.list().then(r => r.data ?? []),
    refetchInterval: 30_000,
  })

  const { data: components = [] } = useQuery({
    queryKey: ['components'],
    queryFn: () => componentsApi.list().then(r => r.data ?? []),
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

  const up   = monitors.filter(m => m.status === 'up').length
  const down = monitors.filter(m => m.status === 'down').length

  // Separate monitors into grouped and ungrouped
  const grouped   = components.map(c => ({ component: c, monitors: monitors.filter(m => m.component_id === c.id) }))
  const ungrouped = monitors.filter(m => !m.component_id)

  const rowProps = (m: Monitor) => ({
    monitor: m,
    onDelete:  () => deleteMutation.mutate(m.id),
    onToggle:  () => toggleMutation.mutate({ id: m.id, is_active: !m.is_active }),
    onEdit:    () => setEditingMonitor(m),
  })

  return (
    <DashboardLayout>
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
        <div className="space-y-3">
          {/* Grouped by component */}
          {grouped.filter(g => g.monitors.length > 0).map(({ component, monitors: cms }) => (
            <ComponentGroup
              key={component.id}
              component={component}
              monitors={cms}
              rowProps={rowProps}
            />
          ))}

          {/* Ungrouped */}
          {ungrouped.length > 0 && (
            <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden">
              {(grouped.filter(g => g.monitors.length > 0).length > 0) && (
                <div className="px-5 py-2.5 border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/30">
                  <span className="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">Ungrouped</span>
                </div>
              )}
              <TableHeader />
              <div className="divide-y divide-gray-100 dark:divide-gray-800">
                {ungrouped.map(m => <MonitorRow key={m.id} {...rowProps(m)} />)}
              </div>
            </div>
          )}
        </div>
      )}

      {showSlugModal && <SlugSetupModal onDone={() => setSlugDismissed(true)} />}
      {showAdd && <AddMonitorModal onClose={() => setShowAdd(false)} />}
      {editingMonitor && (
        <EditMonitorModal
          monitor={editingMonitor}
          components={components}
          onClose={() => setEditingMonitor(null)}
        />
      )}
    </DashboardLayout>
  )
}

// ─── Component group ──────────────────────────────────────────────────────────

function ComponentGroup({ component, monitors, rowProps }: {
  component: Component
  monitors: Monitor[]
  rowProps: (m: Monitor) => React.ComponentProps<typeof MonitorRow>
}) {
  const [open, setOpen] = useState(true)
  const downCount = monitors.filter(m => m.status === 'down').length

  return (
    <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden">
      {/* Group header */}
      <button
        onClick={() => setOpen(o => !o)}
        className="w-full flex items-center gap-3 px-5 py-3 border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50 hover:bg-gray-100 dark:hover:bg-gray-800/70 transition-colors text-left"
      >
        {open ? <ChevronDown size={14} className="text-gray-400 shrink-0" /> : <ChevronRight size={14} className="text-gray-400 shrink-0" />}
        <Layers size={13} className="text-indigo-400 shrink-0" />
        <span className="text-sm font-semibold text-gray-700 dark:text-gray-300">{component.name}</span>
        {component.description && (
          <span className="text-xs text-gray-400 dark:text-gray-500 hidden sm:inline truncate">{component.description}</span>
        )}
        <div className="ml-auto flex items-center gap-2 shrink-0">
          {downCount > 0 && (
            <span className="text-xs px-2 py-0.5 rounded-full bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400 font-medium">
              {downCount} down
            </span>
          )}
          <span className="text-xs text-gray-400">{monitors.length} monitor{monitors.length !== 1 ? 's' : ''}</span>
        </div>
      </button>

      {open && (
        <>
          <TableHeader />
          <div className="divide-y divide-gray-100 dark:divide-gray-800">
            {monitors.map(m => <MonitorRow key={m.id} {...rowProps(m)} />)}
          </div>
        </>
      )}
    </div>
  )
}

// ─── Table header ─────────────────────────────────────────────────────────────

function TableHeader() {
  return (
    <div className="grid grid-cols-[2fr_1fr_1fr_1fr_1fr_auto] gap-4 px-5 py-2.5 bg-white dark:bg-gray-900 border-b border-gray-100 dark:border-gray-800">
      <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Monitor</span>
      <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Status</span>
      <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Component</span>
      <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Interval</span>
      <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Last checked</span>
      <span />
    </div>
  )
}

// ─── Monitor row ──────────────────────────────────────────────────────────────

function MonitorRow({ monitor: m, onDelete, onToggle, onEdit }: {
  monitor: Monitor
  onDelete: () => void
  onToggle: () => void
  onEdit: () => void
}) {
  const dot = COMPONENT_STATUS_DOT[m.component_status]
  const label = COMPONENT_STATUS_LABEL[m.component_status]

  return (
    <div className="grid grid-cols-[2fr_1fr_1fr_1fr_1fr_auto] gap-4 items-center px-5 py-3.5 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors group">
      <Link to={`/dashboard/monitors/${m.id}`} className="min-w-0">
        <p className="text-sm font-medium text-gray-900 dark:text-white truncate">{m.name}</p>
        <p className="text-xs text-gray-400 dark:text-gray-500 truncate mt-0.5">{m.url}</p>
      </Link>
      <div><StatusBadge status={m.is_active ? m.status : 'paused'} /></div>
      <div className="flex items-center gap-1.5 min-w-0">
        <span className={`w-1.5 h-1.5 rounded-full shrink-0 ${dot}`} />
        <span className="text-xs text-gray-500 dark:text-gray-400 truncate">{label}</span>
      </div>
      <span className="text-sm text-gray-500 dark:text-gray-400">
        {m.interval_seconds < 60 ? `${m.interval_seconds}s` : `${m.interval_seconds / 60}m`}
      </span>
      <span className="text-sm text-gray-400 dark:text-gray-500">{timeAgo(m.last_checked_at)}</span>
      <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
        <button onClick={onEdit} className="p-1.5 rounded text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-900/20" title="Edit">
          <Pencil size={13} />
        </button>
        <button onClick={onToggle} className="p-1.5 rounded text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-900/20" title={m.is_active ? 'Pause' : 'Resume'}>
          {m.is_active ? <PauseCircle size={14} /> : <PlayCircle size={14} />}
        </button>
        <button onClick={onDelete} className="p-1.5 rounded text-gray-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20" title="Delete">
          <Trash2 size={14} />
        </button>
      </div>
    </div>
  )
}

// ─── Edit monitor modal ───────────────────────────────────────────────────────

function EditMonitorModal({ monitor, components, onClose }: {
  monitor: Monitor
  components: Component[]
  onClose: () => void
}) {
  const qc = useQueryClient()
  const [name, setName] = useState(monitor.name)
  const [description, setDescription] = useState(monitor.description ?? '')
  const [componentId, setComponentId] = useState<string>(monitor.component_id ?? '')

  const mutation = useMutation({
    mutationFn: () => monitorsApi.updateMeta(monitor.id, {
      name,
      description,
      component_id: componentId || null,
    }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['monitors'] })
      onClose()
    },
  })

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="w-full max-w-md bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xl">
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100 dark:border-gray-800">
          <h2 className="text-base font-semibold text-gray-900 dark:text-white">Edit Monitor</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-xl leading-none">×</button>
        </div>
        <div className="p-6 space-y-4">
          <Input label="Name" value={name} onChange={e => setName(e.target.value)} />
          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Description <span className="text-gray-400">(optional)</span></label>
            <textarea
              value={description}
              onChange={e => setDescription(e.target.value)}
              rows={2}
              placeholder="Shown on status page below monitor name"
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
            />
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Component</label>
            <select
              value={componentId}
              onChange={e => setComponentId(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="">— No component —</option>
              {components.map(c => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </select>
          </div>
        </div>
        <div className="flex justify-end gap-3 px-6 py-4 border-t border-gray-100 dark:border-gray-800">
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button onClick={() => mutation.mutate()} loading={mutation.isPending} disabled={!name}>Save</Button>
        </div>
      </div>
    </div>
  )
}
