import { Link } from 'react-router-dom'
import { isLoggedIn } from '../api/client'
import { usePageTitle } from '../lib/usePageTitle'
import { DashboardLayout } from '../components/layout/DashboardLayout'
import {
  Activity, Bell, CheckCircle, ChevronRight, Clock, ExternalLink,
  Globe, Mail, PauseCircle, PlayCircle, Plus, ShieldCheck, Zap,
} from 'lucide-react'

const sections = [
  { id: 'what-is-pingr',  label: 'What is Pingr?' },
  { id: 'quick-start',    label: 'Quick start' },
  { id: 'monitors',       label: 'Monitors' },
  { id: 'alert-channels', label: 'Alert channels' },
  { id: 'subscriptions',  label: 'Subscriptions' },
  { id: 'status-page',    label: 'Public status page' },
  { id: 'incidents',      label: 'Incidents & history' },
  { id: 'tips',           label: 'Tips & best practices' },
]

function DocsTOC({ loggedIn }: { loggedIn: boolean }) {
  return (
    <aside className="w-52 shrink-0 hidden lg:block">
      <div className="sticky top-6 space-y-1">
        <p className="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">On this page</p>
        {sections.map(s => (
          <a
            key={s.id}
            href={`#${s.id}`}
            className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 py-1 transition-colors"
          >
            <ChevronRight size={12} />
            {s.label}
          </a>
        ))}

        <div className="pt-6 border-t border-gray-100 dark:border-gray-700 mt-6">
          {loggedIn ? (
            <Link to="/dashboard" className="flex items-center gap-2 text-sm font-medium text-indigo-600 dark:text-indigo-400 hover:text-indigo-700">
              <Activity size={14} /> Go to dashboard
            </Link>
          ) : (
            <Link to="/register" className="flex items-center gap-2 text-sm font-medium text-indigo-600 dark:text-indigo-400 hover:text-indigo-700">
              <Zap size={14} /> Start for free
            </Link>
          )}
        </div>
      </div>
    </aside>
  )
}

function DocsArticle({ loggedIn }: { loggedIn: boolean }) {
  return (
    <article className="flex-1 min-w-0 space-y-14">

      {/* What is Pingr */}
      <section id="what-is-pingr" className="scroll-mt-20">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-3">Pingr Documentation</h1>
        <p className="text-lg text-gray-500 dark:text-gray-400 mb-8">
          Everything you need to monitor your services and get alerted when things go wrong.
        </p>
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
          <Globe size={20} className="text-indigo-500" /> What is Pingr?
        </h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Pingr is an uptime monitoring tool. It pings your URLs on a regular schedule, tracks
          response times, detects outages, and sends you an alert the moment something goes down —
          then again when it recovers.
        </p>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-4">
          {[
            { icon: Activity,     title: 'Monitor URLs',       desc: 'HTTP checks every 1–60 minutes' },
            { icon: Bell,         title: 'Get alerted',        desc: 'Email, Slack, or Discord' },
            { icon: ExternalLink, title: 'Public status page', desc: 'Share uptime with your users' },
          ].map(({ icon: Icon, title, desc }) => (
            <div key={title} className="border border-gray-200 dark:border-gray-700 rounded-xl p-4 flex gap-3 items-start bg-white dark:bg-gray-800">
              <Icon size={18} className="text-indigo-500 mt-0.5 shrink-0" />
              <div>
                <p className="text-sm font-semibold text-gray-900 dark:text-white">{title}</p>
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{desc}</p>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* Quick start */}
      <section id="quick-start" className="scroll-mt-20">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
          <Zap size={20} className="text-indigo-500" /> Quick start
        </h2>
        <p className="text-gray-600 dark:text-gray-400 mb-6">Get up and running in under 2 minutes:</p>
        <ol className="space-y-4">
          {[
            { step: '1', title: 'Create an account',      desc: 'Sign up with your email. Verify the link to activate your account.',                                                  action: loggedIn ? null : { label: 'Create account', to: '/register' } },
            { step: '2', title: 'Add an alert channel',   desc: 'Go to Alert Channels in the sidebar and add your email, Slack, or Discord webhook.',                                  action: loggedIn ? { label: 'Alert Channels', to: '/dashboard/alert-channels' } : null },
            { step: '3', title: 'Add your first monitor', desc: 'Click "+ Add Monitor" on the dashboard. Enter the URL you want to watch and set a check interval.',                  action: loggedIn ? { label: 'Dashboard', to: '/dashboard' } : null },
            { step: '4', title: 'Subscribe to alerts',    desc: 'Open the monitor detail page and connect your alert channel. You can subscribe multiple channels to one monitor.',    action: null },
          ].map(({ step, title, desc, action }) => (
            <li key={step} className="flex gap-4">
              <div className="w-8 h-8 rounded-full bg-indigo-600 text-white text-sm font-bold flex items-center justify-center shrink-0 mt-0.5">{step}</div>
              <div>
                <p className="text-sm font-semibold text-gray-900 dark:text-white">{title}</p>
                <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">{desc}</p>
                {action && (
                  <Link to={action.to} className="inline-flex items-center gap-1 text-xs text-indigo-600 dark:text-indigo-400 hover:underline mt-1">
                    {action.label} <ChevronRight size={11} />
                  </Link>
                )}
              </div>
            </li>
          ))}
        </ol>
      </section>

      {/* Monitors */}
      <section id="monitors" className="scroll-mt-20">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
          <Activity size={20} className="text-indigo-500" /> Monitors
        </h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          A monitor is a scheduled HTTP check. Pingr sends a GET request and records whether the server
          responded with a 2xx status code and how long it took.
        </p>
        <h3 className="text-base font-semibold text-gray-800 dark:text-gray-200 mb-2">Creating a monitor</h3>
        <ul className="space-y-2 mb-6">
          {[
            { field: 'Name',             desc: 'A friendly label (e.g. "Production API")' },
            { field: 'URL',              desc: 'Full URL including https:// — must be publicly reachable' },
            { field: 'Check interval',   desc: 'How often to ping: 1, 5, 10, 15, 30, or 60 minutes' },
            { field: 'Failure threshold',desc: 'Consecutive failures before an incident is opened (default 1)' },
          ].map(({ field, desc }) => (
            <li key={field} className="flex gap-3 text-sm">
              <span className="font-mono font-medium text-indigo-700 dark:text-indigo-300 bg-indigo-50 dark:bg-indigo-900/30 px-2 py-0.5 rounded text-xs h-fit mt-0.5 shrink-0">{field}</span>
              <span className="text-gray-600 dark:text-gray-400">{desc}</span>
            </li>
          ))}
        </ul>
        <h3 className="text-base font-semibold text-gray-800 dark:text-gray-200 mb-2">Monitor statuses</h3>
        <div className="space-y-2 mb-4">
          {[
            { badge: 'up',      color: 'bg-green-100 text-green-700',  desc: 'Last check returned a 2xx response.' },
            { badge: 'down',    color: 'bg-red-100 text-red-700',      desc: 'Check failed. An incident is open.' },
            { badge: 'paused',  color: 'bg-gray-100 text-gray-600',    desc: 'Monitoring is paused.' },
            { badge: 'pending', color: 'bg-yellow-100 text-yellow-700',desc: 'Newly created — no checks run yet.' },
          ].map(({ badge, color, desc }) => (
            <div key={badge} className="flex items-center gap-3 text-sm">
              <span className={`${color} px-2 py-0.5 rounded-full text-xs font-medium w-16 text-center`}>{badge}</span>
              <span className="text-gray-600 dark:text-gray-400">{desc}</span>
            </div>
          ))}
        </div>
        <h3 className="text-base font-semibold text-gray-800 dark:text-gray-200 mb-2">Pausing & resuming</h3>
        <p className="text-gray-600 dark:text-gray-400 mb-2">
          Open a monitor's detail page and click <span className="inline-flex items-center gap-1 font-medium"><PauseCircle size={13} className="text-amber-500" /> Pause</span> to stop checks
          temporarily, then <span className="inline-flex items-center gap-1 font-medium"><PlayCircle size={13} className="text-green-500" /> Resume</span> to restart.
        </p>
        <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-700 rounded-lg px-4 py-3 text-sm text-amber-800 dark:text-amber-300">
          <strong>Tip:</strong> Pausing a monitor does not delete its history.
        </div>
      </section>

      {/* Alert channels */}
      <section id="alert-channels" className="scroll-mt-20">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
          <Bell size={20} className="text-indigo-500" /> Alert channels
        </h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          An alert channel is a destination where Pingr sends notifications — email, Slack webhook, or Discord webhook.
        </p>
        <h3 className="text-base font-semibold text-gray-800 dark:text-gray-200 mb-2">Default channel</h3>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Marking a channel as <em>default</em> means new monitors you create will automatically be subscribed to it.
        </p>
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-700 rounded-lg px-4 py-3 text-sm text-blue-800 dark:text-blue-300">
          <strong>Note:</strong> You can subscribe any monitor to any combination of channels.
        </div>
      </section>

      {/* Subscriptions */}
      <section id="subscriptions" className="scroll-mt-20">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
          <Mail size={20} className="text-indigo-500" /> Subscriptions
        </h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          A subscription links a monitor to an alert channel. When the monitor goes down or recovers, a notification
          is sent to every subscribed channel.
        </p>
        <h3 className="text-base font-semibold text-gray-800 dark:text-gray-200 mb-2">What notifications you receive</h3>
        <div className="space-y-3">
          {[
            { icon: Bell,        label: 'DOWN alert',                desc: 'Sent as soon as a monitor crosses the failure threshold.' },
            { icon: CheckCircle, label: 'RECOVERY alert',            desc: 'Sent when the monitor comes back up.' },
            { icon: Mail,        label: 'Subscription confirmation', desc: 'Sent when you subscribe a channel to a monitor.' },
          ].map(({ icon: Icon, label, desc }) => (
            <div key={label} className="flex gap-3 text-sm">
              <Icon size={16} className="text-indigo-500 mt-0.5 shrink-0" />
              <div>
                <span className="font-medium text-gray-900 dark:text-white">{label}</span>
                <span className="text-gray-500 dark:text-gray-400"> — {desc}</span>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* Status page */}
      <section id="status-page" className="scroll-mt-20">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
          <ExternalLink size={20} className="text-indigo-500" /> Public status page
        </h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">Every Pingr account gets a public status page at:</p>
        <div className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg px-4 py-3 font-mono text-sm text-indigo-700 dark:text-indigo-300 mb-4">
          https://your-pingr-app.com/status/<span className="font-bold">your-username</span>
        </div>
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-700 rounded-lg px-4 py-3 text-sm text-blue-800 dark:text-blue-300">
          <strong>Access:</strong> Click <strong>Status Page ↗</strong> in the left sidebar.
        </div>
      </section>

      {/* Incidents */}
      <section id="incidents" className="scroll-mt-20">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
          <ShieldCheck size={20} className="text-indigo-500" /> Incidents & history
        </h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          An incident is opened when a monitor fails for <em>N</em> consecutive checks. It is automatically
          resolved the next time the monitor returns a successful response.
        </p>
        <h3 className="text-base font-semibold text-gray-800 dark:text-gray-200 mb-2">Uptime percentages</h3>
        <p className="text-gray-600 dark:text-gray-400">
          The 24h / 7d / 30d / 90d tiles are calculated from all checks in that window.
          100.00% means every single check passed.
        </p>
      </section>

      {/* Tips */}
      <section id="tips" className="scroll-mt-20">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
          <Clock size={20} className="text-indigo-500" /> Tips & best practices
        </h2>
        <div className="space-y-4">
          {[
            { title: 'Monitor a health endpoint, not the homepage',    desc: 'Create a /health route that checks your DB and dependencies — more signal than a static page.' },
            { title: 'Lower failure threshold for critical services',  desc: 'Threshold of 1 = alert on first failure. Higher values reduce noise from transient blips.' },
            { title: 'Pause before maintenance windows',               desc: "Prevents false-positive DOWN alerts and keeps uptime metrics clean." },
            { title: 'Subscribe team-shared channels',                 desc: "Add a team email or Slack channel so outage notifications aren't siloed in one person's inbox." },
            { title: 'Share your status page',                         desc: 'Add the link to your app footer or README. Customers will check it before submitting a support ticket.' },
            { title: 'Watch the response time chart',                  desc: 'Gradual increases often precede outages. Investigate a climb before it becomes an incident.' },
          ].map(({ title, desc }) => (
            <div key={title} className="flex gap-3">
              <CheckCircle size={16} className="text-green-500 mt-0.5 shrink-0" />
              <div>
                <p className="text-sm font-semibold text-gray-900 dark:text-white">{title}</p>
                <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">{desc}</p>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* CTA — only for guests */}
      {!loggedIn && (
        <div className="border border-indigo-100 bg-indigo-50 rounded-2xl p-8 text-center">
          <h3 className="text-lg font-semibold text-gray-900 mb-2">Ready to start monitoring?</h3>
          <p className="text-sm text-gray-500 mb-4">Set up your first monitor in under 2 minutes.</p>
          <Link
            to="/register"
            className="inline-flex items-center gap-2 bg-indigo-600 text-white px-5 py-2.5 rounded-lg text-sm font-medium hover:bg-indigo-700 transition-colors"
          >
            <Plus size={15} /> Create free account
          </Link>
        </div>
      )}
    </article>
  )
}

export function DocsPage() {
  usePageTitle('Documentation')
  const loggedIn = isLoggedIn()

  // When logged in — use the dashboard layout so the sidebar stays consistent
  if (loggedIn) {
    return (
      <DashboardLayout>
        <div className="flex gap-10 max-w-4xl">
          <DocsTOC loggedIn={loggedIn} />
          <DocsArticle loggedIn={loggedIn} />
        </div>
      </DashboardLayout>
    )
  }

  // Public / guest layout
  return (
    <div className="min-h-screen bg-white dark:bg-gray-950">
      <header className="sticky top-0 z-10 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700">
        <div className="max-w-6xl mx-auto px-6 h-14 flex items-center justify-between">
          <Link to="/" className="text-xl font-bold text-indigo-600">Pingr</Link>
          <div className="flex items-center gap-4">
            <Link to="/login" className="text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900">Sign in</Link>
            <Link to="/register" className="text-sm bg-indigo-600 text-white px-4 py-1.5 rounded-lg hover:bg-indigo-700 transition-colors">
              Get started free
            </Link>
          </div>
        </div>
      </header>
      <div className="max-w-6xl mx-auto px-6 py-10 flex gap-10">
        <DocsTOC loggedIn={loggedIn} />
        <DocsArticle loggedIn={loggedIn} />
      </div>
    </div>
  )
}
