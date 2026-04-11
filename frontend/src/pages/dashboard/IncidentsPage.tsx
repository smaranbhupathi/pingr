import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Bot, CheckCircle, ChevronRight, AlertTriangle } from 'lucide-react'
import { DashboardLayout } from '../../components/layout/DashboardLayout'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { incidentsApi, type Incident, type IncidentStatus } from '../../api/incidents'
import { monitorsApi } from '../../api/monitors'
import { STATUS_LABEL, STATUS_COLOR, STATUS_DOT } from '../../lib/incidents'
import { format } from '../../lib/format'
import { usePageTitle } from '../../lib/usePageTitle'

export function IncidentsPage() {
  usePageTitle('Incidents')
  const [showCreate, setShowCreate] = useState(false)

  const { data: incidents = [], isLoading } = useQuery({
    queryKey: ['incidents'],
    queryFn: () => incidentsApi.list().then(r => r.data ?? []),
  })

  const active   = incidents.filter(i => !i.resolved_at)
  const resolved = incidents.filter(i =>  i.resolved_at)

  return (
    <DashboardLayout>
      {/* Page header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-semibold text-gray-900 dark:text-white">Incidents</h1>
          {incidents.length > 0 && (
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
              {active.length} active · {resolved.length} resolved
            </p>
          )}
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <Plus size={15} /> New incident
        </Button>
      </div>

      {isLoading ? (
        <div className="text-center py-20 text-gray-400 text-sm">Loading…</div>
      ) : incidents.length === 0 ? (
        <div className="text-center py-24 border border-dashed border-gray-200 dark:border-gray-700 rounded-xl">
          <CheckCircle className="mx-auto mb-3 text-gray-300 dark:text-gray-600" size={36} />
          <p className="text-sm font-medium text-gray-600 dark:text-gray-400">All systems operational</p>
          <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">No incidents yet. Create one to communicate issues to your users.</p>
        </div>
      ) : (
        <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden">
          {/* Table header */}
          <div className="grid grid-cols-[2fr_1fr_1fr_1fr_auto] gap-4 px-5 py-3 border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
            <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Incident</span>
            <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Status</span>
            <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Affected</span>
            <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Time</span>
            <span />
          </div>

          {active.length > 0 && (
            <>
              <div className="px-5 py-2 bg-red-50/50 dark:bg-red-900/10 border-b border-gray-100 dark:border-gray-800">
                <span className="text-xs font-semibold text-red-600 dark:text-red-400 uppercase tracking-wider">
                  Active — {active.length}
                </span>
              </div>
              <div className="divide-y divide-gray-100 dark:divide-gray-800">
                {active.map(inc => <IncidentRow key={inc.id} incident={inc} />)}
              </div>
            </>
          )}

          {resolved.length > 0 && (
            <>
              <div className="px-5 py-2 bg-gray-50 dark:bg-gray-800/30 border-y border-gray-100 dark:border-gray-800">
                <span className="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">
                  Resolved — {resolved.length}
                </span>
              </div>
              <div className="divide-y divide-gray-100 dark:divide-gray-800">
                {resolved.map(inc => <IncidentRow key={inc.id} incident={inc} />)}
              </div>
            </>
          )}
        </div>
      )}

      {showCreate && <CreateIncidentModal onClose={() => setShowCreate(false)} />}
    </DashboardLayout>
  )
}

function IncidentRow({ incident }: { incident: Incident }) {
  const latestUpdate = incident.updates?.[incident.updates.length - 1]
  const monitorNames = incident.monitors?.map(m => m.name).join(', ') || '—'

  return (
    <Link
      to={`/dashboard/incidents/${incident.id}`}
      className="grid grid-cols-[2fr_1fr_1fr_1fr_auto] gap-4 items-center px-5 py-3.5 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors group"
    >
      {/* Name */}
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <p className="text-sm font-medium text-gray-900 dark:text-white truncate">{incident.name}</p>
          {incident.source === 'auto' && (
            <span className="shrink-0 inline-flex items-center gap-1 text-xs px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400">
              <Bot size={9} /> Auto
            </span>
          )}
        </div>
        {latestUpdate && (
          <p className="text-xs text-gray-400 dark:text-gray-500 truncate mt-0.5">{latestUpdate.message}</p>
        )}
      </div>

      {/* Status badge */}
      <div>
        <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_COLOR[incident.status]}`}>
          <span className={`w-1.5 h-1.5 rounded-full ${STATUS_DOT[incident.status]}`} />
          {STATUS_LABEL[incident.status]}
        </span>
      </div>

      {/* Affected monitors */}
      <span className="text-sm text-gray-500 dark:text-gray-400 truncate">{monitorNames}</span>

      {/* Time */}
      <span className="text-sm text-gray-400 dark:text-gray-500">
        {incident.resolved_at
          ? `Resolved ${format.timeAgo(incident.resolved_at)}`
          : `Started ${format.timeAgo(incident.created_at)}`}
      </span>

      {/* Chevron */}
      <ChevronRight size={15} className="text-gray-300 dark:text-gray-600 group-hover:text-gray-400 dark:group-hover:text-gray-500" />
    </Link>
  )
}

function CreateIncidentModal({ onClose }: { onClose: () => void }) {
  const queryClient = useQueryClient()
  const [name, setName] = useState('')
  const [status, setStatus] = useState<IncidentStatus>('investigating')
  const [message, setMessage] = useState('')
  const [notify, setNotify] = useState(false)
  const [selectedMonitors, setSelectedMonitors] = useState<string[]>([])

  const { data: monitors = [] } = useQuery({
    queryKey: ['monitors'],
    queryFn: () => monitorsApi.list().then(r => r.data ?? []),
  })

  const mutation = useMutation({
    mutationFn: () => incidentsApi.create({ name, status, message, monitor_ids: selectedMonitors, notify }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['incidents'] })
      onClose()
    },
  })

  const toggleMonitor = (id: string) =>
    setSelectedMonitors(prev => prev.includes(id) ? prev.filter(m => m !== id) : [...prev, id])

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="w-full max-w-lg bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xl">
        {/* Modal header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100 dark:border-gray-800">
          <h2 className="text-base font-semibold text-gray-900 dark:text-white flex items-center gap-2">
            <AlertTriangle size={15} className="text-orange-500" />
            New Incident
          </h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-xl leading-none">×</button>
        </div>

        <div className="p-6 space-y-4">
          <Input
            label="Incident name"
            value={name}
            onChange={e => setName(e.target.value)}
            placeholder="e.g. API latency degradation"
          />

          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Status</label>
            <select
              value={status}
              onChange={e => setStatus(e.target.value as IncidentStatus)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              {(Object.keys(STATUS_LABEL) as IncidentStatus[]).map(s => (
                <option key={s} value={s}>{STATUS_LABEL[s]}</option>
              ))}
            </select>
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Message</label>
            <textarea
              value={message}
              onChange={e => setMessage(e.target.value)}
              rows={3}
              placeholder="Describe what's happening and what you're doing about it…"
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder:text-gray-400 dark:placeholder:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
            />
          </div>

          {monitors.length > 0 && (
            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Affected components</label>
              <div className="space-y-1 max-h-36 overflow-y-auto border border-gray-200 dark:border-gray-700 rounded-lg p-1">
                {monitors.map(m => (
                  <label key={m.id} className="flex items-center gap-2 px-3 py-2 rounded-md hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={selectedMonitors.includes(m.id)}
                      onChange={() => toggleMonitor(m.id)}
                      className="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                    />
                    <span className="text-sm text-gray-700 dark:text-gray-300">{m.name}</span>
                    <span className="text-xs text-gray-400 dark:text-gray-500 truncate ml-auto">{m.url}</span>
                  </label>
                ))}
              </div>
            </div>
          )}

          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={notify}
              onChange={e => setNotify(e.target.checked)}
              className="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
            />
            <span className="text-sm text-gray-700 dark:text-gray-300">Send notification to alert channels</span>
          </label>
        </div>

        <div className="flex justify-end gap-3 px-6 py-4 border-t border-gray-100 dark:border-gray-800">
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button onClick={() => mutation.mutate()} loading={mutation.isPending} disabled={!name || !message}>
            Create incident
          </Button>
        </div>
      </div>
    </div>
  )
}
