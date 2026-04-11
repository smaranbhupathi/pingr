import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, ChevronRight, AlertTriangle, CheckCircle, Bot } from 'lucide-react'
import { DashboardLayout } from '../../components/layout/DashboardLayout'
import { Card } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { incidentsApi, type Incident, type IncidentStatus } from '../../api/incidents'
import { monitorsApi } from '../../api/monitors'
import { format } from '../../lib/format'
import { usePageTitle } from '../../lib/usePageTitle'

const STATUS_LABEL: Record<IncidentStatus, string> = {
  investigating: 'Investigating',
  identified: 'Identified',
  monitoring: 'Monitoring',
  resolved: 'Resolved',
}

const STATUS_COLOR: Record<IncidentStatus, string> = {
  investigating: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
  identified: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
  monitoring: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
  resolved: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
}

export function IncidentsPage() {
  usePageTitle('Incidents')
  const [showCreate, setShowCreate] = useState(false)

  const { data: incidents = [], isLoading } = useQuery({
    queryKey: ['incidents'],
    queryFn: () => incidentsApi.list().then(r => r.data ?? []),
  })

  const active = incidents.filter(i => !i.resolved_at)
  const resolved = incidents.filter(i => i.resolved_at)

  return (
    <DashboardLayout>
      <div className="max-w-3xl mx-auto">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold text-gray-900 dark:text-white">Incidents</h1>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
              Communicate outages and maintenance to your users
            </p>
          </div>
          <Button onClick={() => setShowCreate(true)}>
            <Plus size={15} /> New incident
          </Button>
        </div>

        {isLoading ? (
          <div className="text-center py-16 text-gray-400 text-sm">Loading…</div>
        ) : incidents.length === 0 ? (
          <Card className="p-12 text-center">
            <CheckCircle size={32} className="text-green-400 mx-auto mb-3" />
            <p className="text-sm font-medium text-gray-700 dark:text-gray-300">All systems operational</p>
            <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">No incidents yet. Create one to communicate issues to your users.</p>
          </Card>
        ) : (
          <div className="space-y-6">
            {active.length > 0 && (
              <div>
                <h2 className="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">
                  Active — {active.length}
                </h2>
                <div className="space-y-2">
                  {active.map(inc => <IncidentRow key={inc.id} incident={inc} />)}
                </div>
              </div>
            )}
            {resolved.length > 0 && (
              <div>
                <h2 className="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">
                  Resolved — {resolved.length}
                </h2>
                <div className="space-y-2">
                  {resolved.map(inc => <IncidentRow key={inc.id} incident={inc} />)}
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {showCreate && <CreateIncidentModal onClose={() => setShowCreate(false)} />}
    </DashboardLayout>
  )
}

function IncidentRow({ incident }: { incident: Incident }) {
  const latestUpdate = incident.updates?.[incident.updates.length - 1]

  return (
    <Link to={`/dashboard/incidents/${incident.id}`}>
      <Card className="p-4 hover:border-indigo-300 dark:hover:border-indigo-700 transition-colors cursor-pointer">
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-start gap-3 min-w-0">
            <span className={`mt-0.5 shrink-0 inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_COLOR[incident.status]}`}>
              {STATUS_LABEL[incident.status]}
            </span>
            <div className="min-w-0">
              <p className="text-sm font-medium text-gray-900 dark:text-white truncate">{incident.name}</p>
              {latestUpdate && (
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5 truncate">
                  {latestUpdate.message}
                </p>
              )}
              <div className="flex items-center gap-2 mt-1 flex-wrap">
                {incident.source === 'auto' && (
                  <span className="inline-flex items-center gap-1 text-xs px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400">
                    <Bot size={9} /> Auto-detected
                  </span>
                )}
                {incident.monitors && incident.monitors.length > 0 && (
                  <span className="text-xs text-gray-400 dark:text-gray-500">
                    {incident.monitors.map(m => m.name).join(', ')}
                  </span>
                )}
                <span className="text-xs text-gray-400 dark:text-gray-500">
                  {incident.resolved_at
                    ? `Resolved ${format.timeAgo(incident.resolved_at)}`
                    : `Started ${format.timeAgo(incident.created_at)}`}
                </span>
              </div>
            </div>
          </div>
          <ChevronRight size={16} className="text-gray-400 shrink-0 mt-0.5" />
        </div>
      </Card>
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
    mutationFn: () => incidentsApi.create({
      name,
      status,
      message,
      monitor_ids: selectedMonitors,
      notify,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['incidents'] })
      onClose()
    },
  })

  const toggleMonitor = (id: string) => {
    setSelectedMonitors(prev =>
      prev.includes(id) ? prev.filter(m => m !== id) : [...prev, id]
    )
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <Card className="w-full max-w-lg p-6 dark:bg-gray-800">
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-base font-semibold text-gray-900 dark:text-white flex items-center gap-2">
            <AlertTriangle size={16} className="text-orange-500" />
            New Incident
          </h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-xl leading-none">×</button>
        </div>

        <div className="space-y-4">
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
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="investigating">Investigating</option>
              <option value="identified">Identified</option>
              <option value="monitoring">Monitoring</option>
              <option value="resolved">Resolved</option>
            </select>
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Message</label>
            <textarea
              value={message}
              onChange={e => setMessage(e.target.value)}
              rows={3}
              placeholder="Describe what's happening and what you're doing about it…"
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder:text-gray-400 dark:placeholder:text-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
            />
          </div>

          {monitors.length > 0 && (
            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Affected components</label>
              <div className="space-y-1 max-h-36 overflow-y-auto">
                {monitors.map(m => (
                  <label key={m.id} className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={selectedMonitors.includes(m.id)}
                      onChange={() => toggleMonitor(m.id)}
                      className="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                    />
                    <span className="text-sm text-gray-700 dark:text-gray-300">{m.name}</span>
                    <span className="text-xs text-gray-400 dark:text-gray-500 truncate">{m.url}</span>
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
            <span className="text-sm text-gray-700 dark:text-gray-300">
              Send notification to alert channels
            </span>
          </label>
        </div>

        <div className="flex justify-end gap-3 mt-6">
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button
            onClick={() => mutation.mutate()}
            loading={mutation.isPending}
            disabled={!name || !message}
          >
            Create incident
          </Button>
        </div>
      </Card>
    </div>
  )
}
