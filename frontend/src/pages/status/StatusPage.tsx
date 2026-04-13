import { useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { usePageTitle } from '../../lib/usePageTitle'
import { monitorsApi, COMPONENT_STATUS_LABEL, COMPONENT_STATUS_COLOR, COMPONENT_STATUS_DOT, type MonitorDetail, type DailyUptimeStat, type ComponentStatus } from '../../api/monitors'
import type { Incident, IncidentStatus } from '../../api/incidents'
import { format } from '../../lib/format'
import { Footer } from '../../components/ui/Footer'
import { CheckCircle, AlertTriangle, Clock, ChevronDown, ChevronUp } from 'lucide-react'

// ─── Incident status helpers ──────────────────────────────────────────────────

const STATUS_LABEL: Record<IncidentStatus, string> = {
  investigating: 'Investigating',
  identified:    'Identified',
  monitoring:    'Monitoring',
  resolved:      'Resolved',
}

const STATUS_COLOR: Record<IncidentStatus, string> = {
  investigating: 'bg-red-100 text-red-700 border border-red-200',
  identified:    'bg-orange-100 text-orange-700 border border-orange-200',
  monitoring:    'bg-yellow-100 text-yellow-700 border border-yellow-200',
  resolved:      'bg-green-100 text-green-700 border border-green-200',
}

const STATUS_DOT: Record<IncidentStatus, string> = {
  investigating: 'bg-red-500',
  identified:    'bg-orange-500',
  monitoring:    'bg-yellow-500',
  resolved:      'bg-green-500',
}

// ─── Uptime bar ───────────────────────────────────────────────────────────────

function uptimeBarColor(uptime: number): string {
  if (uptime < 0)   return 'bg-gray-200'           // no data
  if (uptime >= 99) return 'bg-green-500'
  if (uptime >= 90) return 'bg-yellow-400'
  return 'bg-red-500'
}

function UptimeBar({ daily }: { daily: DailyUptimeStat[] }) {
  const [hovered, setHovered] = useState<DailyUptimeStat | null>(null)

  return (
    <div className="relative">
      <div className="flex gap-px h-8 items-end">
        {daily.map(d => (
          <div
            key={d.date}
            className="flex-1 relative group"
            onMouseEnter={() => setHovered(d)}
            onMouseLeave={() => setHovered(null)}
          >
            <div className={`w-full h-8 rounded-sm ${uptimeBarColor(d.uptime)} opacity-90 group-hover:opacity-100 transition-opacity`} />
          </div>
        ))}
      </div>

      {/* Tooltip */}
      {hovered && (
        <div className="absolute bottom-10 left-1/2 -translate-x-1/2 z-10 bg-gray-900 text-white text-xs px-2.5 py-1.5 rounded-lg whitespace-nowrap shadow-lg pointer-events-none">
          <p className="font-medium">{hovered.date}</p>
          <p className="text-gray-300">
            {hovered.uptime < 0 ? 'No data' : `${hovered.uptime.toFixed(2)}% uptime`}
          </p>
        </div>
      )}

      {/* Labels */}
      <div className="flex justify-between mt-1">
        <span className="text-xs text-gray-400">90 days ago</span>
        <span className="text-xs text-gray-400">Today</span>
      </div>
    </div>
  )
}

// ─── Monitor row ──────────────────────────────────────────────────────────────

function MonitorRow({ detail }: { detail: MonitorDetail }) {
  const { monitor, uptime, daily_uptime, active_incident } = detail
  const cs = (monitor.component_status ?? 'operational') as ComponentStatus
  const isDown = cs !== 'operational' && cs !== 'under_maintenance'

  return (
    <div className={`bg-white rounded-xl border p-5 ${isDown ? 'border-red-200' : 'border-gray-200'}`}>
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-3">
          <span className={`w-2.5 h-2.5 rounded-full shrink-0 ${COMPONENT_STATUS_DOT[cs]}`} />
          <div>
            <p className="font-semibold text-gray-900">{monitor.name}</p>
            <p className="text-xs text-gray-400 mt-0.5">{monitor.url}</p>
          </div>
        </div>
        <div className="text-right shrink-0">
          <span className={`inline-block text-xs font-medium px-2 py-0.5 rounded-full ${COMPONENT_STATUS_COLOR[cs]}`}>
            {COMPONENT_STATUS_LABEL[cs]}
          </span>
          {monitor.last_checked_at && (
            <p className="text-xs text-gray-400 mt-1">
              Checked {format.timeAgo(monitor.last_checked_at)}
            </p>
          )}
        </div>
      </div>

      {/* 90-day uptime bar */}
      {daily_uptime && daily_uptime.length > 0 && (
        <div className="mb-4">
          <UptimeBar daily={daily_uptime} />
        </div>
      )}

      {/* Uptime stats */}
      <div className="grid grid-cols-4 gap-2 text-center">
        {[
          { label: '24h',  value: uptime.last_24h },
          { label: '7d',   value: uptime.last_7d },
          { label: '30d',  value: uptime.last_30d },
          { label: '90d',  value: uptime.last_90d },
        ].map(({ label, value }) => (
          <div key={label} className="bg-gray-50 rounded-lg py-2">
            <p className={`text-sm font-semibold ${
              value >= 99 ? 'text-green-600' : value >= 95 ? 'text-yellow-600' : 'text-red-600'
            }`}>
              {value.toFixed(2)}%
            </p>
            <p className="text-xs text-gray-400 mt-0.5">{label}</p>
          </div>
        ))}
      </div>

      {/* Active incident banner */}
      {active_incident && (
        <div className="mt-4 flex items-start gap-2.5 px-3.5 py-2.5 bg-red-50 border border-red-100 rounded-lg">
          <AlertTriangle size={14} className="text-red-500 shrink-0 mt-0.5" />
          <div className="min-w-0">
            <p className="text-sm font-medium text-red-800">{active_incident.name}</p>
            <p className="text-xs text-red-600 mt-0.5">
              {STATUS_LABEL[active_incident.status]} · Started {format.timeAgo(active_incident.created_at)}
            </p>
          </div>
        </div>
      )}
    </div>
  )
}

// ─── Incidents section ────────────────────────────────────────────────────────

function IncidentCard({ incident }: { incident: Incident }) {
  const [expanded, setExpanded] = useState(!incident.resolved_at)
  const isResolved = !!incident.resolved_at

  return (
    <div className={`bg-white rounded-xl border ${isResolved ? 'border-gray-200' : 'border-red-200'}`}>
      {/* Incident header */}
      <button
        className="w-full flex items-start justify-between gap-3 p-5 text-left"
        onClick={() => setExpanded(e => !e)}
      >
        <div className="flex items-start gap-3 min-w-0">
          <span className={`mt-0.5 shrink-0 inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_COLOR[incident.status]}`}>
            <span className={`w-1.5 h-1.5 rounded-full ${STATUS_DOT[incident.status]}`} />
            {STATUS_LABEL[incident.status]}
          </span>
          <div className="min-w-0">
            <p className="text-sm font-semibold text-gray-900">{incident.name}</p>
            <p className="text-xs text-gray-500 mt-0.5 flex items-center gap-1">
              <Clock size={11} />
              {isResolved
                ? `Resolved ${format.timeAgo(incident.resolved_at!)} · Duration ${format.duration(incident.created_at, incident.resolved_at!)}`
                : `Started ${format.datetime(incident.created_at)} · ${format.timeAgo(incident.created_at)} ongoing`}
            </p>
          </div>
        </div>
        {expanded ? <ChevronUp size={15} className="text-gray-400 shrink-0 mt-0.5" /> : <ChevronDown size={15} className="text-gray-400 shrink-0 mt-0.5" />}
      </button>

      {/* Timeline */}
      {expanded && incident.updates && incident.updates.length > 0 && (
        <div className="px-5 pb-5 border-t border-gray-100">
          <div className="relative mt-4">
            <div className="absolute left-[7px] top-1 bottom-1 w-px bg-gray-200" />
            <div className="space-y-5">
              {[...incident.updates].reverse().map(u => (
                <div key={u.id} className="flex gap-4 relative">
                  <span className={`w-3.5 h-3.5 rounded-full border-2 border-white shrink-0 mt-0.5 ${STATUS_DOT[u.status]}`} />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap mb-1">
                      <span className={`text-xs font-semibold px-1.5 py-0.5 rounded ${STATUS_COLOR[u.status]}`}>
                        {STATUS_LABEL[u.status]}
                      </span>
                      <span className="text-xs text-gray-400">{format.datetime(u.created_at)}</span>
                    </div>
                    <p className="text-sm text-gray-700 whitespace-pre-line">{u.message}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

// ─── Shell ────────────────────────────────────────────────────────────────────

function StatusShell({ username, children }: { username: string; children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      <div className="flex-1 max-w-2xl w-full mx-auto px-6 py-12">
        <div className="text-center mb-10">
          <h1 className="text-2xl font-bold text-gray-900">{username}</h1>
          <p className="text-sm text-gray-400 mt-1">Service Status</p>
        </div>
        {children}
        <p className="text-center text-xs text-gray-300 mt-12">
          Powered by{' '}
          <Link to="/" className="text-indigo-400 hover:underline">Pingr</Link>
        </p>
      </div>
      <Footer />
    </div>
  )
}

// ─── Page ─────────────────────────────────────────────────────────────────────

export function StatusPage({ slugOverride }: { slugOverride?: string }) {
  const { username } = useParams<{ username: string }>()
  const slug = slugOverride ?? username ?? ''

  usePageTitle(slug ? `${slug} — Status` : 'Status Page')

  const { data: monitors = [], isLoading, isError } = useQuery({
    queryKey: ['status', slug],
    queryFn: () => monitorsApi.statusPage(slug).then(r => r.data ?? []),
    refetchInterval: 60_000,
  })

  if (isLoading) {
    return (
      <StatusShell username={slug}>
        {/* Banner skeleton */}
        <div className="rounded-xl p-5 mb-8 bg-gray-100 dark:bg-gray-800 animate-pulse h-16" />
        {/* Component skeletons */}
        <div className="space-y-3">
          {[1, 2, 3].map(i => (
            <div key={i} className="rounded-xl border border-gray-200 dark:border-gray-700 p-5 animate-pulse">
              <div className="flex items-center justify-between mb-3">
                <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-32" />
                <div className="h-5 bg-gray-200 dark:bg-gray-700 rounded w-20" />
              </div>
              <div className="h-8 bg-gray-100 dark:bg-gray-800 rounded w-full" />
            </div>
          ))}
        </div>
      </StatusShell>
    )
  }

  if (isError) {
    return (
      <StatusShell username={slug}>
        <p className="text-red-400 text-center py-16">Failed to load status page.</p>
      </StatusShell>
    )
  }

  const downCount    = monitors.filter(m => m.monitor.component_status && m.monitor.component_status !== 'operational' && m.monitor.component_status !== 'under_maintenance').length
  const allUp        = downCount === 0
  const allIncidents = monitors.flatMap(m => m.incidents ?? [])
  const activeInc    = allIncidents.filter(i => !i.resolved_at)
  const resolvedInc  = allIncidents
    .filter(i => i.resolved_at)
    .sort((a, b) => new Date(b.resolved_at!).getTime() - new Date(a.resolved_at!).getTime())
    .slice(0, 20)

  // Deduplicate incidents by id (same incident can appear via multiple monitors)
  const deduped = (list: Incident[]) =>
    list.filter((inc, idx, arr) => arr.findIndex(x => x.id === inc.id) === idx)

  const activeDeduped   = deduped(activeInc)
  const resolvedDeduped = deduped(resolvedInc)

  return (
    <StatusShell username={slug}>
      {/* ── Overall banner ── */}
      <div className={`rounded-xl p-5 mb-8 flex items-center gap-4 ${
        allUp
          ? 'bg-green-50 border border-green-200'
          : 'bg-red-50 border border-red-200'
      }`}>
        {allUp
          ? <CheckCircle size={24} className="text-green-500 shrink-0" />
          : <AlertTriangle size={24} className="text-red-500 shrink-0" />
        }
        <div>
          <p className={`font-semibold ${allUp ? 'text-green-800' : 'text-red-800'}`}>
            {allUp
              ? 'All systems operational'
              : `${downCount} system${downCount > 1 ? 's' : ''} experiencing issues`}
          </p>
          <p className="text-xs text-gray-500 mt-0.5">
            Updated {format.time(new Date().toISOString())}
            {monitors.length > 0 && ` · ${monitors.length} component${monitors.length > 1 ? 's' : ''} monitored`}
          </p>
        </div>
      </div>

      {/* ── Monitors ── */}
      {monitors.length === 0 ? (
        <div className="text-center py-16 bg-white rounded-xl border border-gray-200">
          <p className="text-gray-400 text-sm">No monitors configured yet.</p>
        </div>
      ) : (
        <div className="space-y-3 mb-10">
          <h2 className="text-xs font-semibold uppercase tracking-wider text-gray-400 mb-3">
            Components
          </h2>
          {monitors.map(detail => (
            <MonitorRow key={detail.monitor.id} detail={detail} />
          ))}
        </div>
      )}

      {/* ── Active incidents ── */}
      {activeDeduped.length > 0 && (
        <div className="mb-8">
          <h2 className="text-xs font-semibold uppercase tracking-wider text-red-500 mb-3">
            Active Incidents — {activeDeduped.length}
          </h2>
          <div className="space-y-3">
            {activeDeduped.map(inc => (
              <IncidentCard key={inc.id} incident={inc} />
            ))}
          </div>
        </div>
      )}

      {/* ── Incident history ── */}
      {resolvedDeduped.length > 0 && (
        <div>
          <h2 className="text-xs font-semibold uppercase tracking-wider text-gray-400 mb-3">
            Incident History
          </h2>
          <div className="space-y-2">
            {resolvedDeduped.map(inc => (
              <IncidentCard key={inc.id} incident={inc} />
            ))}
          </div>
        </div>
      )}

      {/* No incidents ever */}
      {activeDeduped.length === 0 && resolvedDeduped.length === 0 && monitors.length > 0 && (
        <div className="text-center py-8 bg-white rounded-xl border border-gray-200">
          <CheckCircle size={24} className="text-green-400 mx-auto mb-2" />
          <p className="text-sm text-gray-500">No incidents reported</p>
          <p className="text-xs text-gray-400 mt-1">Everything has been running smoothly.</p>
        </div>
      )}
    </StatusShell>
  )
}
