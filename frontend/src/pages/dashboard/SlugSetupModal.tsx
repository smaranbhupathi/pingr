import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { userApi } from '../../api/user'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Globe } from 'lucide-react'

interface Props {
  onDone: () => void
}

export function SlugSetupModal({ onDone }: Props) {
  const [slug, setSlug] = useState('')
  const [error, setError] = useState('')
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: () => userApi.setSlug(slug),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['me'] })
      onDone()
    },
    onError: (err: any) => {
      const msg = err?.response?.data?.error
      setError(msg === 'this URL is already taken'
        ? 'That URL is already taken — try another.'
        : msg || 'Something went wrong.')
    },
  })

  function handleChange(val: string) {
    // Auto-lowercase and strip invalid chars as they type
    setSlug(val.toLowerCase().replace(/[^a-z0-9-]/g, ''))
    setError('')
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (slug.length < 3) {
      setError('Must be at least 3 characters.')
      return
    }
    mutation.mutate()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className="w-full max-w-md bg-white dark:bg-gray-900 rounded-2xl shadow-2xl p-8">
        {/* Icon */}
        <div className="flex items-center justify-center w-12 h-12 rounded-xl bg-indigo-100 dark:bg-indigo-900/40 mb-5">
          <Globe size={22} className="text-indigo-600 dark:text-indigo-400" />
        </div>

        <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-1">
          Set your status page URL
        </h2>
        <p className="text-sm text-gray-500 dark:text-gray-400 mb-6">
          This is the URL you share with customers. Choose something that represents your project or company.
        </p>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* URL preview */}
          <div className="flex items-center rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden text-sm">
            <span className="px-3 py-2.5 bg-gray-50 dark:bg-gray-800 text-gray-500 dark:text-gray-400 border-r border-gray-200 dark:border-gray-700 whitespace-nowrap select-none">
              getpingr.com/status/
            </span>
            <Input
              value={slug}
              onChange={e => handleChange(e.target.value)}
              placeholder="acme-corp"
              className="border-0 rounded-none focus:ring-0 bg-transparent flex-1"
              maxLength={50}
              autoFocus
            />
          </div>

          {error && (
            <p className="text-xs text-red-500">{error}</p>
          )}

          <p className="text-xs text-gray-400 dark:text-gray-500">
            Lowercase letters, numbers, and hyphens only. You can change this later in Settings.
          </p>

          <div className="flex gap-3 pt-1">
            <Button
              type="submit"
              className="flex-1"
              disabled={slug.length < 3 || mutation.isPending}
            >
              {mutation.isPending ? 'Saving…' : 'Set URL'}
            </Button>
            <Button
              type="button"
              variant="secondary"
              onClick={onDone}
            >
              Skip for now
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
