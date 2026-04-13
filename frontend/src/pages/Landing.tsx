import { Link } from 'react-router-dom'
import { Activity, Bell, Globe, BarChart2, ArrowRight, Plus, ExternalLink } from 'lucide-react'
import { isLoggedIn } from '../api/client'
import { Footer } from '../components/ui/Footer'
import { UserMenu } from '../components/ui/UserMenu'
import { useQuery } from '@tanstack/react-query'
import { userApi } from '../api/user'
import { monitorsApi } from '../api/monitors'

// ─── Logged-in home ──────────────────────────────────────────────────────────

function LoggedInHome() {
  const { data: profile } = useQuery({
    queryKey: ['me'],
    queryFn: () => userApi.me().then(r => r.data),
  })

  const { data: monitors = [] } = useQuery({
    queryKey: ['monitors'],
    queryFn: () => monitorsApi.list().then(r => r.data),
  })

  const up = monitors.filter(m => m.status === 'up').length
  const down = monitors.filter(m => m.status === 'down').length
  const total = monitors.length

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      {/* Top bar */}
      <nav className="bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between">
        <span className="text-xl font-bold text-indigo-600">Pingr</span>
        <UserMenu />
      </nav>

      <main className="flex-1 max-w-4xl mx-auto w-full px-6 py-12">
        {/* Greeting */}
        <div className="mb-10">
          <h1 className="text-3xl font-bold text-gray-900">
            Welcome back{profile?.username ? `, ${profile.username}` : ''}
          </h1>
          <p className="text-gray-500 mt-1">Here's a quick overview of your monitoring.</p>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-3 gap-4 mb-10">
          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <p className="text-sm text-gray-500 mb-1">Total monitors</p>
            <p className="text-4xl font-bold text-gray-900">{total}</p>
          </div>
          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <p className="text-sm text-gray-500 mb-1">Up</p>
            <p className="text-4xl font-bold text-green-600">{up}</p>
          </div>
          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <p className="text-sm text-gray-500 mb-1">Down</p>
            <p className="text-4xl font-bold text-red-500">{down}</p>
          </div>
        </div>

        {/* Quick actions */}
        <h2 className="text-sm font-semibold text-gray-500 uppercase tracking-wider mb-4">Quick actions</h2>
        <div className="grid sm:grid-cols-2 gap-3">
          <Link
            to="/dashboard"
            className="bg-white border border-gray-200 rounded-xl p-5 flex items-center justify-between hover:border-indigo-300 hover:shadow-sm transition group"
          >
            <div className="flex items-center gap-3">
              <div className="w-9 h-9 bg-indigo-50 rounded-lg flex items-center justify-center">
                <Activity size={18} className="text-indigo-600" />
              </div>
              <div>
                <p className="text-sm font-medium text-gray-900">View monitors</p>
                <p className="text-xs text-gray-400">See all your monitors</p>
              </div>
            </div>
            <ArrowRight size={16} className="text-gray-400 group-hover:text-indigo-500 transition" />
          </Link>

          <Link
            to="/dashboard"
            state={{ openAdd: true }}
            className="bg-white border border-gray-200 rounded-xl p-5 flex items-center justify-between hover:border-indigo-300 hover:shadow-sm transition group"
          >
            <div className="flex items-center gap-3">
              <div className="w-9 h-9 bg-green-50 rounded-lg flex items-center justify-center">
                <Plus size={18} className="text-green-600" />
              </div>
              <div>
                <p className="text-sm font-medium text-gray-900">Add a monitor</p>
                <p className="text-xs text-gray-400">Start monitoring a new URL</p>
              </div>
            </div>
            <ArrowRight size={16} className="text-gray-400 group-hover:text-indigo-500 transition" />
          </Link>

          <Link
            to="/dashboard/alert-channels"
            className="bg-white border border-gray-200 rounded-xl p-5 flex items-center justify-between hover:border-indigo-300 hover:shadow-sm transition group"
          >
            <div className="flex items-center gap-3">
              <div className="w-9 h-9 bg-amber-50 rounded-lg flex items-center justify-center">
                <Bell size={18} className="text-amber-500" />
              </div>
              <div>
                <p className="text-sm font-medium text-gray-900">Alert channels</p>
                <p className="text-xs text-gray-400">Manage where you get notified</p>
              </div>
            </div>
            <ArrowRight size={16} className="text-gray-400 group-hover:text-indigo-500 transition" />
          </Link>

          {profile && (
            <a
              href={
                profile.status_page_slug
                  ? `https://${profile.status_page_slug}.getpingr.com`
                  : `/status/${profile.username}`
              }
              target="_blank"
              rel="noreferrer"
              className="bg-white border border-gray-200 rounded-xl p-5 flex items-center justify-between hover:border-indigo-300 hover:shadow-sm transition group"
            >
              <div className="flex items-center gap-3">
                <div className="w-9 h-9 bg-sky-50 rounded-lg flex items-center justify-center">
                  <Globe size={18} className="text-sky-500" />
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-900">Your status page</p>
                  <p className="text-xs text-gray-400">
                    {profile.status_page_slug
                      ? `${profile.status_page_slug}.getpingr.com`
                      : `getpingr.com/status/${profile.username}`}
                  </p>
                </div>
              </div>
              <ExternalLink size={16} className="text-gray-400 group-hover:text-indigo-500 transition" />
            </a>
          )}
        </div>
      </main>

      <Footer />
    </div>
  )
}

// ─── Guest / marketing page ───────────────────────────────────────────────────

function GuestHome() {
  return (
    <div className="min-h-screen bg-white">
      {/* Navbar */}
      <nav className="border-b border-gray-100 px-6 py-4 flex items-center justify-between max-w-6xl mx-auto">
        <span className="text-xl font-bold text-indigo-600">Pingr</span>
        <div className="flex items-center gap-4">
          <Link to="/docs" className="text-sm text-gray-600 hover:text-gray-900">Docs</Link>
          <Link to="/login" className="text-sm text-gray-600 hover:text-gray-900">Sign in</Link>
          <Link to="/register" className="bg-indigo-600 text-white text-sm px-4 py-2 rounded-lg hover:bg-indigo-700">
            Get started free
          </Link>
        </div>
      </nav>

      {/* Hero */}
      <section className="max-w-4xl mx-auto px-6 pt-20 pb-16 text-center">
        <div className="inline-flex items-center gap-2 bg-indigo-50 text-indigo-600 text-xs font-medium px-3 py-1 rounded-full mb-6">
          <span className="w-1.5 h-1.5 bg-indigo-500 rounded-full" />
          Free forever · No credit card required
        </div>
        <h1 className="text-5xl font-bold text-gray-900 leading-tight tracking-tight mb-6">
          Know when your website
          <br />
          goes down <span className="text-indigo-600">before your users do</span>
        </h1>
        <p className="text-xl text-gray-500 mb-8 max-w-2xl mx-auto">
          Pingr monitors your websites every minute and alerts you instantly when something breaks.
          Share a beautiful status page with your users — free.
        </p>
        <div className="flex items-center justify-center gap-4">
          <Link
            to="/register"
            className="bg-indigo-600 text-white px-8 py-3 rounded-xl font-medium hover:bg-indigo-700 text-lg"
          >
            Start monitoring free →
          </Link>
          <a
            href="https://smaran-pingr.getpingr.com"
            target="_blank"
            rel="noreferrer"
            className="text-gray-500 hover:text-gray-700 text-sm underline"
          >
            See example status page
          </a>
        </div>
        <p className="text-xs text-gray-400 mt-4">5 monitors free · No setup required</p>
      </section>

      {/* Features */}
      <section className="max-w-5xl mx-auto px-6 py-16">
        <h2 className="text-3xl font-bold text-gray-900 text-center mb-12">
          Everything you need, nothing you don't
        </h2>
        <div className="grid md:grid-cols-2 gap-6">
          {[
            {
              icon: <Activity className="text-indigo-600" size={24} />,
              title: 'Uptime monitoring',
              desc: 'We check your website every minute from multiple regions. Never miss downtime again.',
            },
            {
              icon: <Bell className="text-indigo-600" size={24} />,
              title: 'Instant alerts',
              desc: 'Get notified by email the moment your site goes down — and again when it recovers.',
            },
            {
              icon: <Globe className="text-indigo-600" size={24} />,
              title: 'Public status page',
              desc: 'Share yourname.getpingr.com with your users. Build trust through transparency.',
            },
            {
              icon: <BarChart2 className="text-indigo-600" size={24} />,
              title: 'Response time analytics',
              desc: 'Track response time trends and uptime history. Know your SLA before your clients do.',
            },
          ].map(f => (
            <div key={f.title} className="border border-gray-200 rounded-xl p-6 hover:shadow-md transition-shadow">
              <div className="w-10 h-10 bg-indigo-50 rounded-lg flex items-center justify-center mb-4">
                {f.icon}
              </div>
              <h3 className="font-semibold text-gray-900 mb-2">{f.title}</h3>
              <p className="text-gray-500 text-sm leading-relaxed">{f.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Pricing */}
      <section className="bg-gray-50 py-16">
        <div className="max-w-3xl mx-auto px-6 text-center">
          <h2 className="text-3xl font-bold text-gray-900 mb-3">Simple, honest pricing</h2>
          <p className="text-gray-500 mb-10">Start free. Upgrade only when you need more.</p>
          <div className="grid md:grid-cols-2 gap-6 text-left">
            <div className="bg-white border border-gray-200 rounded-xl p-8">
              <p className="text-sm font-medium text-gray-500 mb-1">Free</p>
              <p className="text-4xl font-bold text-gray-900 mb-1">₹0</p>
              <p className="text-sm text-gray-400 mb-6">Forever</p>
              <ul className="space-y-3 text-sm text-gray-600">
                {[
                  '5 monitors',
                  '30s – 24h check intervals',
                  'Email, Slack & Discord alerts',
                  'Public status page (yourname.getpingr.com)',
                  'Incident management',
                  'Component groups',
                  '90-day uptime history',
                  'Import / export alert channels',
                ].map(f => (
                  <li key={f} className="flex items-center gap-2">
                    <span className="text-green-500">✓</span> {f}
                  </li>
                ))}
              </ul>
              <Link to="/register" className="mt-8 block text-center bg-gray-900 text-white py-2.5 rounded-lg text-sm font-medium hover:bg-gray-700">
                Get started free
              </Link>
            </div>

            <div className="bg-indigo-600 border border-indigo-600 rounded-xl p-8 text-white">
              <p className="text-sm font-medium text-indigo-200 mb-1">Pro</p>
              <p className="text-4xl font-bold mb-1">₹299</p>
              <p className="text-sm text-indigo-200 mb-6">/month · Coming soon</p>
              <ul className="space-y-3 text-sm text-indigo-100">
                {[
                  '50 monitors',
                  '30-second checks',
                  'Microsoft Teams alerts',
                  'PagerDuty & webhook integrations',
                  'Custom domain status page',
                  'SSL certificate monitoring',
                  'Multi-region checks',
                  'API access',
                  'Priority support',
                ].map(f => (
                  <li key={f} className="flex items-center gap-2">
                    <span className="text-indigo-300">✓</span> {f}
                  </li>
                ))}
              </ul>
              <button disabled className="mt-8 w-full bg-white/20 text-white py-2.5 rounded-lg text-sm font-medium cursor-not-allowed">
                Coming soon
              </button>
            </div>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="max-w-2xl mx-auto px-6 py-20 text-center">
        <h2 className="text-3xl font-bold text-gray-900 mb-4">
          Start monitoring in 2 minutes
        </h2>
        <p className="text-gray-500 mb-8">No credit card. No setup. Just paste your URL and go.</p>
        <Link
          to="/register"
          className="bg-indigo-600 text-white px-10 py-3 rounded-xl font-medium hover:bg-indigo-700 text-lg inline-block"
        >
          Create free account →
        </Link>
      </section>

      <Footer />
    </div>
  )
}

// ─── Entry point ──────────────────────────────────────────────────────────────

export function LandingPage() {
  return isLoggedIn() ? <LoggedInHome /> : <GuestHome />
}
