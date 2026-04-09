import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { monitorsApi, type Monitor } from '../../api/monitors'
import { Navbar } from '../../components/ui/Navbar'
import { StatusBadge } from '../../components/ui/StatusBadge'
import { Button } from '../../components/ui/Button'
import { Card } from '../../components/ui/Card'
import { AddMonitorModal } from './AddMonitorModal'
import { Globe, Trash2, PauseCircle, PlayCircle } from 'lucide-react'

export function DashboardPage() {
  const [showAdd, setShowAdd] = useState(false)
  const queryClient = useQueryClient()

  const { data: monitors = [], isLoading } = useQuery({
    queryKey: ['monitors'],
    queryFn: () => monitorsApi.list().then(r => r.data),
    refetchInterval: 30_000, // refresh every 30s
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
    <div className="min-h-screen bg-gray-50">
      <Navbar />

      <main className="max-w-5xl mx-auto px-6 py-8">
        {/* Stats */}
        <div className="grid grid-cols-3 gap-4 mb-8">
          <Card className="p-5">
            <p className="text-sm text-gray-500">Total monitors</p>
            <p className="text-3xl font-semibold text-gray-900 mt-1">{monitors.length}</p>
          </Card>
          <Card className="p-5">
            <p className="text-sm text-gray-500">Up</p>
            <p className="text-3xl font-semibold text-green-600 mt-1">{up}</p>
          </Card>
          <Card className="p-5">
            <p className="text-sm text-gray-500">Down</p>
            <p className="text-3xl font-semibold text-red-600 mt-1">{down}</p>
          </Card>
        </div>

        {/* Header */}
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-gray-900">Monitors</h2>
          <Button onClick={() => setShowAdd(true)}>+ Add monitor</Button>
        </div>

        {/* Monitor list */}
        {isLoading ? (
          <Card className="p-12 text-center text-gray-400">Loading...</Card>
        ) : monitors.length === 0 ? (
          <Card className="p-12 text-center">
            <Globe className="mx-auto mb-3 text-gray-300" size={40} />
            <p className="text-gray-500 font-medium">No monitors yet</p>
            <p className="text-gray-400 text-sm mt-1">Add your first URL to start monitoring</p>
            <Button className="mt-4" onClick={() => setShowAdd(true)}>Add monitor</Button>
          </Card>
        ) : (
          <div className="space-y-3">
            {monitors.map(m => (
              <MonitorRow
                key={m.id}
                monitor={m}
                onDelete={() => deleteMutation.mutate(m.id)}
                onToggle={() => toggleMutation.mutate({ id: m.id, is_active: !m.is_active })}
              />
            ))}
          </div>
        )}
      </main>

      {showAdd && <AddMonitorModal onClose={() => setShowAdd(false)} />}
    </div>
  )
}

function MonitorRow({
  monitor,
  onDelete,
  onToggle,
}: {
  monitor: Monitor
  onDelete: () => void
  onToggle: () => void
}) {
  return (
    <Card className="p-4 flex items-center justify-between hover:shadow-md transition-shadow">
      <Link to={`/dashboard/monitors/${monitor.id}`} className="flex items-center gap-3 flex-1 min-w-0">
        <StatusBadge status={monitor.status} />
        <div className="min-w-0">
          <p className="font-medium text-gray-900 truncate">{monitor.name}</p>
          <p className="text-sm text-gray-400 truncate">{monitor.url}</p>
        </div>
      </Link>
      <div className="flex items-center gap-2 ml-4 shrink-0">
        <span className="text-xs text-gray-400">
          Every {monitor.interval_seconds / 60}m
        </span>
        <button
          onClick={onToggle}
          className="p-1.5 rounded text-gray-400 hover:text-indigo-600 hover:bg-indigo-50"
          title={monitor.is_active ? 'Pause' : 'Resume'}
        >
          {monitor.is_active ? <PauseCircle size={16} /> : <PlayCircle size={16} />}
        </button>
        <button
          onClick={onDelete}
          className="p-1.5 rounded text-gray-400 hover:text-red-500 hover:bg-red-50"
          title="Delete"
        >
          <Trash2 size={16} />
        </button>
      </div>
    </Card>
  )
}
