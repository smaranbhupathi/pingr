import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { monitorsApi } from '../../api/monitors'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { X } from 'lucide-react'

const INTERVALS = [
  { label: 'Every 1 minute', value: 60 },
  { label: 'Every 3 minutes', value: 180 },
  { label: 'Every 5 minutes', value: 300 },
]

export function AddMonitorModal({ onClose }: { onClose: () => void }) {
  const queryClient = useQueryClient()
  const [form, setForm] = useState({ name: '', url: 'https://', interval_seconds: 60 })
  const [errors, setErrors] = useState<Record<string, string>>({})

  const mutation = useMutation({
    mutationFn: () => monitorsApi.create(form),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitors'] })
      onClose()
    },
    onError: (err: any) => {
      if (err.response?.data?.errors) setErrors(err.response.data.errors)
    },
  })

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50 px-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-md p-6">
        <div className="flex items-center justify-between mb-5">
          <h3 className="font-semibold text-gray-900 text-lg">Add monitor</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600"><X size={20} /></button>
        </div>

        <div className="space-y-4">
          <Input
            label="Name"
            placeholder="My Website"
            value={form.name}
            error={errors.name}
            onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
          />
          <Input
            label="URL"
            placeholder="https://example.com"
            value={form.url}
            error={errors.url}
            onChange={e => setForm(f => ({ ...f, url: e.target.value }))}
          />
          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium text-gray-700">Check interval</label>
            <select
              value={form.interval_seconds}
              onChange={e => setForm(f => ({ ...f, interval_seconds: Number(e.target.value) }))}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              {INTERVALS.map(i => (
                <option key={i.value} value={i.value}>{i.label}</option>
              ))}
            </select>
          </div>
        </div>

        <div className="flex gap-3 mt-6">
          <Button variant="secondary" onClick={onClose} className="flex-1">Cancel</Button>
          <Button
            onClick={() => mutation.mutate()}
            loading={mutation.isPending}
            className="flex-1"
          >
            Add monitor
          </Button>
        </div>
      </div>
    </div>
  )
}
