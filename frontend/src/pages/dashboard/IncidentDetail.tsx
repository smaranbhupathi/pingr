import { useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, Clock, Bot, Activity } from 'lucide-react'
import { DashboardLayout } from '../../components/layout/DashboardLayout'
import { Card } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { incidentsApi, type IncidentStatus } from '../../api/incidents'
import { format } from '../../lib/format'
import { usePageTitle } from '../../lib/usePageTitle'

const STATUS_LABEL: Record<IncidentStatus, string> = {
  investigating: 'Investigating',
  identified: 'Identified',
  monitoring: 'Monitoring',
  resolved: 'Resolved',
}

const STATUS_COLOR: Record<IncidentStatus, string> = {
  investigating: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400 border-red-200 dark:border-red-800',
  identified: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400 border-orange-200 dark:border-orange-800',
  monitoring: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400 border-yellow-200 dark:border-yellow-800',
  resolved: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 border-green-200 dark:border-green-800',
}

const STATUS_DOT: Record<IncidentStatus, string> = {
  investigating: 'bg-red-500',
  identified: 'bg-orange-500',
  monitoring: 'bg-yellow-500',
  resolved: 'bg-green-500',
}

export function IncidentDetailPage() {
  usePageTitle('Incident')
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()

  const { data: incident, isLoading } = useQuery({
    queryKey: ['incidents', id],
    queryFn: () => incidentsApi.get(id!).then(r => r.data),
  })

  const [updateStatus, setUpdateStatus] = useState<IncidentStatus>('investigating')
  const [updateMessage, setUpdateMessage] = useState('')
  const [notify, setNotify] = useState(false)

  // Sync status selector with current incident status
  const currentStatus = incident?.status ?? 'investigating'

  const postUpdate = useMutation({
    mutationFn: () => incidentsApi.postUpdate(id!, {
      status: updateStatus,
      message: updateMessage,
      notify,
    }),
    onSuccess: () => {
      setUpdateMessage('')
      setNotify(false)
      queryClient.invalidateQueries({ queryKey: ['incidents', id] })
      queryClient.invalidateQueries({ queryKey: ['incidents'] })
    },
  })

  if (isLoading || !incident) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-64 text-gray-400 text-sm">Loading…</div>
      </DashboardLayout>
    )
  }

  const isResolved = !!incident.resolved_at

  return (
    <DashboardLayout>
      <div className="max-w-2xl mx-auto">
        <Link to="/dashboard/incidents" className="inline-flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 mb-6">
          <ArrowLeft size={14} /> Back to incidents
        </Link>

        {/* Header */}
        <div className="mb-6">
          <div className="flex items-start gap-3 mb-2">
            <span className={`mt-0.5 shrink-0 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${STATUS_COLOR[incident.status]}`}>
              <span className={`w-1.5 h-1.5 rounded-full mr-1.5 ${STATUS_DOT[incident.status]}`} />
              {STATUS_LABEL[incident.status]}
            </span>
            {incident.source === 'auto' && (
              <span className="mt-0.5 inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 border border-gray-200 dark:border-gray-700">
                <Bot size={10} /> Auto-detected
              </span>
            )}
          </div>
          <h1 className="text-2xl font-semibold text-gray-900 dark:text-white">{incident.name}</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1 flex items-center gap-1">
            <Clock size={13} />
            {isResolved
              ? `Resolved ${format.datetime(incident.resolved_at!)} · Duration: ${format.duration(incident.created_at, incident.resolved_at!)}`
              : `Started ${format.datetime(incident.created_at)} · ${format.timeAgo(incident.created_at)} ongoing`}
          </p>
        </div>

        {/* Affected monitors */}
        {incident.monitors && incident.monitors.length > 0 && (
          <Card className="p-5 mb-6">
            <h2 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
              <Activity size={14} className="text-indigo-500" />
              Affected components
            </h2>
            <div className="space-y-2">
              {incident.monitors.map(m => (
                <Link
                  key={m.id}
                  to={`/dashboard/monitors/${m.id}`}
                  className="flex items-center justify-between px-3 py-2 rounded-lg bg-gray-50 dark:bg-gray-700/50 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                >
                  <div className="flex items-center gap-2">
                    <span className="w-2 h-2 rounded-full bg-indigo-400 shrink-0" />
                    <span className="text-sm font-medium text-gray-800 dark:text-gray-200">{m.name}</span>
                  </div>
                  <span className="text-xs text-gray-400 dark:text-gray-500 truncate ml-3">{m.url}</span>
                </Link>
              ))}
            </div>
          </Card>
        )}

        {/* Post update form — only for active incidents */}
        {!isResolved && (
          <Card className="p-5 mb-6">
            <h2 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-4">Post an update</h2>
            <div className="space-y-3">
              <div className="flex flex-col gap-1">
                <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Status</label>
                <select
                  value={updateStatus}
                  onChange={e => setUpdateStatus(e.target.value as IncidentStatus)}
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
                  value={updateMessage}
                  onChange={e => setUpdateMessage(e.target.value)}
                  rows={3}
                  placeholder="What's the latest update?"
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder:text-gray-400 dark:placeholder:text-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
                />
              </div>
              <div className="flex items-center justify-between">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={notify}
                    onChange={e => setNotify(e.target.checked)}
                    className="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                  />
                  <span className="text-sm text-gray-600 dark:text-gray-400">Send notification to alert channels</span>
                </label>
                <Button
                  onClick={() => postUpdate.mutate()}
                  loading={postUpdate.isPending}
                  disabled={!updateMessage}
                >
                  Post update
                </Button>
              </div>
            </div>
          </Card>
        )}

        {/* Timeline */}
        <Card className="p-5">
          <h2 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-5">Timeline</h2>
          {!incident.updates || incident.updates.length === 0 ? (
            <p className="text-sm text-gray-400 dark:text-gray-500 text-center py-4">No updates yet.</p>
          ) : (
            <div className="relative">
              {/* Vertical line */}
              <div className="absolute left-[7px] top-2 bottom-2 w-px bg-gray-200 dark:bg-gray-700" />

              <div className="space-y-6">
                {[...incident.updates].reverse().map(update => (
                  <div key={update.id} className="flex gap-4 relative">
                    <span className={`w-3.5 h-3.5 rounded-full border-2 border-white dark:border-gray-900 shrink-0 mt-0.5 ${STATUS_DOT[update.status]}`} />
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1 flex-wrap">
                        <span className={`text-xs font-semibold px-1.5 py-0.5 rounded ${STATUS_COLOR[update.status]}`}>
                          {STATUS_LABEL[update.status]}
                        </span>
                        {update.source === 'auto' ? (
                          <span className="inline-flex items-center gap-1 text-xs px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400">
                            <Bot size={10} /> Auto
                          </span>
                        ) : (
                          <span className="text-xs text-gray-400 dark:text-gray-500">Manual</span>
                        )}
                        <span className="text-xs text-gray-400 dark:text-gray-500">
                          {format.datetime(update.created_at)}
                        </span>
                      </div>
                      <p className="text-sm text-gray-700 dark:text-gray-300 whitespace-pre-line">{update.message}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </Card>
      </div>
    </DashboardLayout>
  )
}
