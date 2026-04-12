import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Pencil, Trash2, Layers, ChevronRight } from 'lucide-react'
import { DashboardLayout } from '../../components/layout/DashboardLayout'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { componentsApi, type Component } from '../../api/components'
import { monitorsApi } from '../../api/monitors'
import { usePageTitle } from '../../lib/usePageTitle'

export function ComponentsPage() {
  usePageTitle('Components')
  const [showCreate, setShowCreate] = useState(false)
  const [editing, setEditing] = useState<Component | null>(null)

  const qc = useQueryClient()

  const { data: components = [], isLoading } = useQuery({
    queryKey: ['components'],
    queryFn: () => componentsApi.list().then(r => r.data ?? []),
  })

  const { data: monitors = [] } = useQuery({
    queryKey: ['monitors'],
    queryFn: () => monitorsApi.list().then(r => r.data ?? []),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => componentsApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['components'] }),
  })

  const monitorsInComponent = (id: string) => monitors.filter(m => m.component_id === id)

  return (
    <DashboardLayout>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-semibold text-gray-900 dark:text-white">Components</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
            Group monitors for your public status page
          </p>
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <Plus size={15} /> New component
        </Button>
      </div>

      {isLoading ? (
        <div className="text-center py-20 text-gray-400 text-sm">Loading…</div>
      ) : components.length === 0 ? (
        <div className="text-center py-24 border border-dashed border-gray-200 dark:border-gray-700 rounded-xl">
          <Layers className="mx-auto mb-3 text-gray-300 dark:text-gray-600" size={36} />
          <p className="text-sm font-medium text-gray-600 dark:text-gray-400">No components yet</p>
          <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">Create a component to group monitors on your status page.</p>
        </div>
      ) : (
        <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden">
          <div className="grid grid-cols-[2fr_1fr_auto] gap-4 px-5 py-3 border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
            <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Component</span>
            <span className="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wide">Monitors</span>
            <span />
          </div>
          <div className="divide-y divide-gray-100 dark:divide-gray-800">
            {components.map(c => {
              const count = monitorsInComponent(c.id).length
              return (
                <div key={c.id} className="grid grid-cols-[2fr_1fr_auto] gap-4 items-center px-5 py-3.5">
                  <div className="min-w-0">
                    <p className="text-sm font-medium text-gray-900 dark:text-white">{c.name}</p>
                    {c.description && (
                      <p className="text-xs text-gray-400 dark:text-gray-500 mt-0.5 truncate">{c.description}</p>
                    )}
                  </div>
                  <span className="text-sm text-gray-500 dark:text-gray-400">
                    {count} monitor{count !== 1 ? 's' : ''}
                  </span>
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => setEditing(c)}
                      className="p-1.5 text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                    >
                      <Pencil size={13} />
                    </button>
                    <button
                      onClick={() => deleteMutation.mutate(c.id)}
                      className="p-1.5 text-gray-400 hover:text-red-600 rounded-md hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
                    >
                      <Trash2 size={13} />
                    </button>
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {showCreate && <ComponentModal onClose={() => setShowCreate(false)} />}
      {editing  && <ComponentModal component={editing} onClose={() => setEditing(null)} />}
    </DashboardLayout>
  )
}

function ComponentModal({ component, onClose }: { component?: Component; onClose: () => void }) {
  const qc = useQueryClient()
  const [name, setName] = useState(component?.name ?? '')
  const [description, setDescription] = useState(component?.description ?? '')

  const mutation = useMutation({
    mutationFn: () => component
      ? componentsApi.update(component.id, { name, description })
      : componentsApi.create(name, description),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['components'] })
      onClose()
    },
  })

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="w-full max-w-md bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xl">
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100 dark:border-gray-800">
          <h2 className="text-base font-semibold text-gray-900 dark:text-white flex items-center gap-2">
            <Layers size={15} className="text-indigo-500" />
            {component ? 'Edit Component' : 'New Component'}
          </h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-xl leading-none">×</button>
        </div>
        <div className="p-6 space-y-4">
          <Input
            label="Name"
            value={name}
            onChange={e => setName(e.target.value)}
            placeholder="e.g. API, Database, CDN"
          />
          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Description <span className="text-gray-400">(optional)</span></label>
            <textarea
              value={description}
              onChange={e => setDescription(e.target.value)}
              rows={2}
              placeholder="Short description shown on status page"
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
            />
          </div>
        </div>
        <div className="flex justify-end gap-3 px-6 py-4 border-t border-gray-100 dark:border-gray-800">
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button onClick={() => mutation.mutate()} loading={mutation.isPending} disabled={!name}>
            {component ? 'Save' : 'Create'}
          </Button>
        </div>
      </div>
    </div>
  )
}

// Re-export for use in monitors page
export { ChevronRight }
