import { DashboardLayout } from '../../components/layout/DashboardLayout'
import { Card } from '../../components/ui/Card'
import { Sun, Moon } from 'lucide-react'
import { useTheme } from '../../lib/theme'
import { usePageTitle } from '../../lib/usePageTitle'

export function SettingsPage() {
  usePageTitle('Settings')
  const { theme, setTheme } = useTheme()

  return (
    <DashboardLayout>
      <div className="max-w-xl mx-auto">
        <h1 className="text-xl font-semibold text-gray-900 dark:text-white mb-6">Settings</h1>

        {/* Appearance */}
        <Card className="p-6 mb-4">
          <h2 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-4">Appearance</h2>

          <div className="flex gap-3">
            <ThemeOption
              label="Light"
              icon={<Sun size={18} />}
              active={theme === 'light'}
              onClick={() => setTheme('light')}
              preview="bg-white border-gray-200"
              textPreview="text-gray-800"
            />
            <ThemeOption
              label="Dark"
              icon={<Moon size={18} />}
              active={theme === 'dark'}
              onClick={() => setTheme('dark')}
              preview="bg-gray-900 border-gray-700"
              textPreview="text-gray-100"
            />
          </div>
        </Card>

        {/* More settings sections can go here */}
        <Card className="p-6">
          <h2 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-1">Notifications</h2>
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
  textPreview,
}: {
  label: string
  icon: React.ReactNode
  active: boolean
  onClick: () => void
  preview: string
  textPreview: string
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
