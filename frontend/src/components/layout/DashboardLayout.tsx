import { NavLink, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { userApi } from '../../api/user'
import { Activity, AlertTriangle, Bell, BookOpen, ExternalLink, Layers, Settings } from 'lucide-react'
import { Footer } from '../ui/Footer'
import { UserMenu } from '../ui/UserMenu'

const navItems = [
  { to: '/dashboard',                label: 'Monitors',       icon: Activity,       exact: true },
  { to: '/dashboard/components',     label: 'Components',     icon: Layers,         exact: false },
  { to: '/dashboard/incidents',      label: 'Incidents',      icon: AlertTriangle,  exact: false },
  { to: '/dashboard/alert-channels', label: 'Alert Channels', icon: Bell,           exact: false },
  { to: '/docs',                     label: 'Docs',           icon: BookOpen,       exact: false },
  { to: '/dashboard/settings',       label: 'Settings',       icon: Settings,       exact: false },
]

export function DashboardLayout({ children }: { children: React.ReactNode }) {
  const { data: profile } = useQuery({
    queryKey: ['me'],
    queryFn: () => userApi.me().then(r => r.data),
  })

  return (
    <div className="flex h-screen overflow-hidden bg-gray-50 dark:bg-gray-950">
      {/* Sidebar — fixed height, never scrolls with content */}
      <aside className="w-56 h-full bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-700/50 flex flex-col shrink-0 overflow-y-auto">
        {/* Brand */}
        <div className="px-5 py-4 border-b border-gray-100 dark:border-gray-700/50">
          <Link to="/" className="text-xl font-bold text-indigo-600 dark:text-indigo-400">Pingr</Link>
        </div>

        {/* Nav links */}
        <nav className="flex-1 px-3 py-4 flex flex-col gap-0.5">
          {navItems.map(({ to, label, icon: Icon, exact }) => (
            <NavLink
              key={to}
              to={to}
              end={exact}
              className={({ isActive }) =>
                `flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-300'
                    : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-white'
                }`
              }
            >
              <Icon size={15} />
              {label}
            </NavLink>
          ))}

          {profile && (
            <>
              <div className="my-2 border-t border-gray-100 dark:border-gray-800" />
              <a
                href={`/status/${profile.username}`}
                target="_blank"
                rel="noreferrer"
                className="flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm font-medium text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-white transition-colors"
              >
                <ExternalLink size={15} />
                Status Page
              </a>
            </>
          )}
        </nav>
      </aside>

      {/* Right side — header fixed, only main scrolls */}
      <div className="flex-1 min-w-0 flex flex-col h-full overflow-hidden">
        {/* Top header */}
        <header className="h-14 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700/50 px-6 flex items-center justify-end shrink-0 z-10">
          <UserMenu />
        </header>

        {/* Scrollable content area only */}
        <div className="flex-1 overflow-y-auto">
          <main className="px-8 py-8">
            {children}
          </main>
          <Footer />
        </div>
      </div>
    </div>
  )
}
