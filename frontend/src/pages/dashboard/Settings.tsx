import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { DashboardLayout } from '../../components/layout/DashboardLayout'
import { Card } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Sun, Moon, Globe } from 'lucide-react'
import { useTheme } from '../../lib/theme'
import { usePageTitle } from '../../lib/usePageTitle'
import { userApi } from '../../api/user'

export function SettingsPage() {
  usePageTitle('Settings')
  const { theme, setTheme } = useTheme()
  const queryClient = useQueryClient()

  const { data: profile } = useQuery({
    queryKey: ['me'],
    queryFn: () => userApi.me().then(r => r.data),
  })

  const [slug, setSlug] = useState('')
  const [slugError, setSlugError] = useState('')
  const [slugSuccess, setSlugSuccess] = useState(false)

  const slugMutation = useMutation({
    mutationFn: () => userApi.setSlug(slug),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['me'] })
      setSlugSuccess(true)
      setSlugError('')
      setTimeout(() => setSlugSuccess(false), 3000)
    },
    onError: (err: any) => {
      const msg = err?.response?.data?.error
      setSlugError(msg === 'this URL is already taken'
        ? 'That URL is already taken — try another.'
        : msg || 'Something went wrong.')
    },
  })

  function handleSlugChange(val: string) {
    setSlug(val.toLowerCase().replace(/[^a-z0-9-]/g, ''))
    setSlugError('')
    setSlugSuccess(false)
  }

  function handleSlugSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (slug.length < 3) { setSlugError('Must be at least 3 characters.'); return }
    slugMutation.mutate()
  }

  return (
    <DashboardLayout>
      <div className="max-w-2xl mx-auto">
        <div className="mb-6">
          <h1 className="text-xl font-semibold text-gray-900 dark:text-white">Settings</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Manage your account preferences</p>
        </div>

        {/* Status Page URL */}
        <Card className="p-5 mb-4">
          <div className="flex items-center gap-2 mb-1">
            <Globe size={15} className="text-indigo-500" />
            <h2 className="text-sm font-semibold text-gray-900 dark:text-white">Status Page URL</h2>
          </div>
          {profile?.status_page_slug && (
            <p className="text-xs text-gray-500 dark:text-gray-400 mb-3">
              Current:{' '}
              <a
                href={`https://${profile.status_page_slug}.getpingr.com`}
                target="_blank"
                rel="noreferrer"
                className="text-indigo-500 hover:underline"
              >
                {profile.status_page_slug}.getpingr.com
              </a>
            </p>
          )}
          <form onSubmit={handleSlugSubmit} className="flex gap-2">
            <div className="flex items-center flex-1 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden text-sm">
              <span className="px-3 py-2 bg-gray-50 dark:bg-gray-800 text-gray-500 dark:text-gray-400 border-r border-gray-200 dark:border-gray-700 whitespace-nowrap select-none text-xs">
                getpingr.com/status/
              </span>
              <Input
                value={slug}
                onChange={e => handleSlugChange(e.target.value)}
                placeholder={profile?.status_page_slug ?? 'acme-corp'}
                className="border-0 rounded-none focus:ring-0 bg-transparent flex-1"
                maxLength={50}
              />
            </div>
            <Button type="submit" disabled={slug.length < 3 || slugMutation.isPending}>
              {slugMutation.isPending ? 'Saving…' : 'Save'}
            </Button>
          </form>
          {slugError && <p className="text-xs text-red-500 mt-1.5">{slugError}</p>}
          {slugSuccess && <p className="text-xs text-green-500 mt-1.5">Saved!</p>}
          <p className="text-xs text-gray-400 dark:text-gray-500 mt-2">
            Lowercase letters, numbers, and hyphens only (3–50 characters).
          </p>
        </Card>

        {/* Appearance */}
        <Card className="p-5 mb-4">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-white mb-4">Appearance</h2>

          <div className="flex gap-3">
            <ThemeOption
              label="Light"
              icon={<Sun size={18} />}
              active={theme === 'light'}
              onClick={() => setTheme('light')}
              preview="bg-white border-gray-200"
            />
            <ThemeOption
              label="Dark"
              icon={<Moon size={18} />}
              active={theme === 'dark'}
              onClick={() => setTheme('dark')}
              preview="bg-gray-900 border-gray-700"
            />
          </div>
        </Card>

        <Card className="p-5">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-white mb-1">Notifications</h2>
          <p className="text-sm text-gray-400 dark:text-gray-500">
            Manage alert channels and subscriptions in{' '}
            <a href="/dashboard/alert-channels" className="text-indigo-500 hover:underline">Alert Channels</a>.
          </p>
        </Card>
      </div>
    </DashboardLayout>
  )
}

function ThemeOption({
  label,
  icon,
  active,
  onClick,
  preview,
}: {
  label: string
  icon: React.ReactNode
  active: boolean
  onClick: () => void
  preview: string
}) {
  return (
    <button
      onClick={onClick}
      className={`flex-1 flex flex-col items-center gap-2 p-4 rounded-xl border-2 transition-all ${
        active
          ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/20'
          : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
      }`}
    >
      {/* Mini preview */}
      <div className={`w-full h-16 rounded-lg border ${preview} flex flex-col p-2 gap-1`}>
        <div className={`h-2 w-12 rounded ${preview.includes('gray-900') ? 'bg-gray-700' : 'bg-gray-200'}`} />
        <div className={`h-2 w-8 rounded ${preview.includes('gray-900') ? 'bg-gray-600' : 'bg-gray-100'}`} />
      </div>
      <div className={`flex items-center gap-1.5 text-sm font-medium ${
        active ? 'text-indigo-600 dark:text-indigo-400' : 'text-gray-600 dark:text-gray-400'
      }`}>
        {icon}
        {label}
      </div>
    </button>
  )
}
