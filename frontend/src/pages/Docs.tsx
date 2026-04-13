import { Link } from 'react-router-dom'
import { isLoggedIn } from '../api/client'
import { usePageTitle } from '../lib/usePageTitle'
import { DashboardLayout } from '../components/layout/DashboardLayout'
import {
  Activity, AlertTriangle, Bell, CheckCircle, ChevronRight, Clock,
  Download, ExternalLink, Globe, Layers, Mail, PauseCircle, PlayCircle,
  Plus, Upload, Zap,
} from 'lucide-react'

const sections = [
  { id: 'what-is-pingr',   label: 'What is Pingr?' },
  { id: 'quick-start',     label: 'Quick start' },
  { id: 'monitors',        label: 'Monitors' },
  { id: 'components',      label: 'Components' },
  { id: 'alert-channels',  label: 'Alert channels' },
  { id: 'import-export',   label: 'Import & export' },
  { id: 'subscriptions',   label: 'Subscriptions' },
  { id: 'incidents',       label: 'Incidents' },
  { id: 'status-page',     label: 'Public status page' },
  { id: 'tips',            label: 'Tips & best practices' },
]

// ─── Shared primitives ────────────────────────────────────────────────────────

function Field({ name, desc }: { name: string; desc: string }) {
  return (
    <li className="flex gap-3 text-sm">
      <span className="font-mono font-medium text-indigo-700 dark:text-indigo-300 bg-indigo-50 dark:bg-indigo-900/30 px-2 py-0.5 rounded text-xs h-fit mt-0.5 shrink-0">{name}</span>
      <span className="text-gray-600 dark:text-gray-400">{desc}</span>
    </li>
  )
}

function Note({ children }: { children: React.ReactNode }) {
  return (
    <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-700 rounded-lg px-4 py-3 text-sm text-blue-800 dark:text-blue-300">
      {children}
    </div>
  )
}

function Tip({ children }: { children: React.ReactNode }) {
  return (
    <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-700 rounded-lg px-4 py-3 text-sm text-amber-800 dark:text-amber-300">
      {children}
    </div>
  )
}

function SectionHeading({ icon: Icon, title }: { icon: React.ElementType; title: string }) {
  return (
    <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
      <Icon size={20} className="text-indigo-500" /> {title}
    </h2>
  )
}

function SubHeading({ title }: { title: string }) {
  return <h3 className="text-base font-semibold text-gray-800 dark:text-gray-200 mb-2">{title}</h3>
}

// ─── TOC ─────────────────────────────────────────────────────────────────────

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

// ─── Article ──────────────────────────────────────────────────────────────────

function DocsArticle({ loggedIn }: { loggedIn: boolean }) {
  return (
    <article className="flex-1 min-w-0 space-y-14">

      {/* ── What is Pingr ── */}
      <section id="what-is-pingr" className="scroll-mt-20">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-3">Pingr Documentation</h1>
        <p className="text-lg text-gray-500 dark:text-gray-400 mb-8">
          Everything you need to monitor your services and communicate with your users when things go wrong.
        </p>
        <SectionHeading icon={Globe} title="What is Pingr?" />
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Pingr is an open-source uptime monitoring tool. It pings your URLs on a regular schedule,
          tracks response times, detects outages, and sends alerts the moment something goes down —
          then again when it recovers. A public status page lets your users know what's happening in real time.
        </p>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          {[
            { icon: Activity,       title: 'Monitor URLs',        desc: 'HTTP checks every 30s – 24h' },
            { icon: Bell,           title: 'Get alerted',         desc: 'Email, Slack, or Discord' },
            { icon: ExternalLink,   title: 'Public status page',  desc: 'Share uptime with your users' },
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

      {/* ── Quick start ── */}
      <section id="quick-start" className="scroll-mt-20">
        <SectionHeading icon={Zap} title="Quick start" />
        <p className="text-gray-600 dark:text-gray-400 mb-6">Get up and running in under 2 minutes:</p>
        <ol className="space-y-5">
          {[
            { step: '1', title: 'Create an account',       desc: 'Sign up with your email and verify the link to activate your account.',                                                            action: loggedIn ? null : { label: 'Create account', to: '/register' } },
            { step: '2', title: 'Add an alert channel',    desc: 'Go to Alert Channels in the sidebar and add your email, Slack, or Discord webhook so you can receive notifications.',              action: loggedIn ? { label: 'Alert Channels', to: '/dashboard/alert-channels' } : null },
            { step: '3', title: 'Add your first monitor',  desc: 'Click "+ Add monitor" on the dashboard. Enter the URL you want to watch and pick a check interval.',                              action: loggedIn ? { label: 'Dashboard', to: '/dashboard' } : null },
            { step: '4', title: 'Subscribe to alerts',     desc: 'Open the monitor detail page and connect your alert channel. You can subscribe multiple channels to one monitor.',                 action: null },
            { step: '5', title: 'Share your status page',  desc: 'Click "Status Page ↗" in the sidebar and send the link to your team or customers.',                                               action: null },
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

      {/* ── Monitors ── */}
      <section id="monitors" className="scroll-mt-20">
        <SectionHeading icon={Activity} title="Monitors" />
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          A monitor is a scheduled HTTP check. Pingr sends a GET request to your URL and records whether
          the server responded with a 2xx status code and how long it took.
        </p>

        <SubHeading title="Creating a monitor" />
        <ul className="space-y-2 mb-6">
          {[
            { field: 'Name',              desc: 'A friendly label shown on the dashboard and status page (e.g. "Production API")' },
            { field: 'URL',               desc: 'Full URL including https:// — must be publicly reachable from the internet' },
            { field: 'Check interval',    desc: 'How often to ping: 30s, 1m, 5m, 10m, 15m, 30m, or 60m' },
            { field: 'Failure threshold', desc: 'Consecutive failures before an incident is opened (default 1)' },
            { field: 'Description',       desc: 'Optional text shown below the monitor name on your public status page' },
          ].map(f => <Field key={f.field} name={f.field} desc={f.desc} />)}
        </ul>

        <SubHeading title="Editing a monitor" />
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Hover over any monitor row on the dashboard and click the <strong>pencil icon</strong> to edit
          its name, description, or component group. URL and check interval are set at creation time.
        </p>

        <SubHeading title="Monitor statuses" />
        <div className="space-y-2 mb-6">
          {[
            { badge: 'up',      color: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',    desc: 'Last check returned a 2xx response.' },
            { badge: 'down',    color: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',            desc: 'Check failed. An incident has been opened.' },
            { badge: 'paused',  color: 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400',           desc: 'Monitoring is paused — no checks running.' },
            { badge: 'pending', color: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',desc: 'Newly created — no checks run yet.' },
          ].map(({ badge, color, desc }) => (
            <div key={badge} className="flex items-center gap-3 text-sm">
              <span className={`${color} px-2 py-0.5 rounded-full text-xs font-medium w-16 text-center`}>{badge}</span>
              <span className="text-gray-600 dark:text-gray-400">{desc}</span>
            </div>
          ))}
        </div>

        <SubHeading title="Pausing & resuming" />
        <p className="text-gray-600 dark:text-gray-400 mb-3">
          On the dashboard, hover a monitor row and click <span className="inline-flex items-center gap-1 font-medium"><PauseCircle size={13} className="text-amber-500" /> Pause</span> to
          stop checks temporarily, then <span className="inline-flex items-center gap-1 font-medium"><PlayCircle size={13} className="text-green-500" /> Resume</span> to restart.
        </p>
        <Tip><strong>Tip:</strong> Pause monitors before planned maintenance so you don't get false DOWN alerts and your uptime stats stay clean.</Tip>
      </section>

      {/* ── Components ── */}
      <section id="components" className="scroll-mt-20">
        <SectionHeading icon={Layers} title="Components" />
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Components let you group related monitors together — for example, grouping your API monitor,
          database monitor, and CDN monitor under a "Backend" component. Groups appear as collapsible
          sections on the dashboard and as named service blocks on your public status page.
        </p>

        <SubHeading title="Creating a component" />
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Go to <strong>Components</strong> in the sidebar → click <strong>"+ New component"</strong> → give it a name and optional description.
          Then assign monitors to it by editing each monitor (pencil icon on the dashboard) and selecting the component.
        </p>

        <SubHeading title="Component status" />
        <p className="text-gray-600 dark:text-gray-400 mb-3">
          Each monitor has a <strong>component status</strong> that controls what is shown on the public status page.
          This is separate from the internal up/down check result — you can override it manually via incidents.
        </p>
        <div className="space-y-2 mb-4">
          {[
            { badge: 'Operational',          color: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' },
            { badge: 'Degraded Performance', color: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400' },
            { badge: 'Partial Outage',       color: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400' },
            { badge: 'Major Outage',         color: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' },
            { badge: 'Under Maintenance',    color: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' },
          ].map(({ badge, color }) => (
            <span key={badge} className={`inline-block mr-2 mb-2 ${color} px-2.5 py-0.5 rounded-full text-xs font-medium`}>{badge}</span>
          ))}
        </div>
        <Note>
          <strong>How it works:</strong> When the worker detects a monitor going <strong>down</strong>, it automatically sets the component status
          to <em>Major Outage</em>. When the monitor <strong>recovers</strong>, it resets to <em>Operational</em>.
          You can override this at any time by creating or updating an incident.
        </Note>
      </section>

      {/* ── Alert channels ── */}
      <section id="alert-channels" className="scroll-mt-20">
        <SectionHeading icon={Bell} title="Alert channels" />
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          An alert channel is a destination where Pingr sends notifications when a monitor goes down or recovers.
          Supported types: <strong>Email</strong>, <strong>Slack</strong>, and <strong>Discord</strong>.
        </p>

        <SubHeading title="Adding a channel" />
        <ul className="space-y-2 mb-6">
          {[
            { field: 'Email',    desc: 'Enter any email address. Pingr sends a formatted HTML alert.' },
            { field: 'Slack',    desc: 'Go to your Slack channel → Integrations → Incoming Webhooks → copy the webhook URL.' },
            { field: 'Discord',  desc: 'Go to your Discord channel → Settings → Integrations → Webhooks → New Webhook → copy the URL.' },
          ].map(f => <Field key={f.field} name={f.field} desc={f.desc} />)}
        </ul>

        <SubHeading title="Default channel" />
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Marking a channel as <em>default</em> means new monitors you create are automatically subscribed to it.
          Useful if you always want all monitors to notify the same Slack channel.
        </p>

        <SubHeading title="Enabling / disabling" />
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Toggle the switch on any channel row to temporarily pause notifications from that channel without
          deleting it or removing subscriptions.
        </p>

        <Note><strong>Note:</strong> You can subscribe any monitor to any combination of channels.</Note>
      </section>

      {/* ── Import & Export ── */}
      <section id="import-export" className="scroll-mt-20">
        <SectionHeading icon={Download} title="Import & export alert channels" />
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          You can export all your alert channels to a file and import them back — useful for backups,
          migrating between accounts, or bulk-adding many channels at once.
          Go to <strong>Alert Channels</strong> in the sidebar to find the Import and Export buttons.
        </p>

        <SubHeading title="Exporting" />
        <p className="text-gray-600 dark:text-gray-400 mb-2">
          Click <span className="inline-flex items-center gap-1 font-medium"><Download size={13} className="text-gray-500" /> Export</span> and
          choose <strong>CSV</strong> (opens in Excel / Google Sheets) or <strong>JSON</strong> (round-trip import ready).
          All channels are exported including disabled ones. No API call is made — the download happens instantly in your browser.
        </p>

        <SubHeading title="Importing" />
        <p className="text-gray-600 dark:text-gray-400 mb-3">
          Click <span className="inline-flex items-center gap-1 font-medium"><Upload size={13} className="text-gray-500" /> Import</span> and
          drop a <code className="font-mono text-xs">.csv</code> or <code className="font-mono text-xs">.json</code> file.
          Pingr parses it in your browser immediately and shows a preview of each row before anything is saved.
        </p>

        <SubHeading title="File format" />
        <div className="mb-4">
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-2">CSV:</p>
          <pre className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-3 text-xs overflow-x-auto leading-relaxed">{`name,type,value,enabled
My Email Alert,email,alerts@example.com,true
Team Slack,slack,https://hooks.slack.com/services/...,true
Discord Alerts,discord,https://discord.com/api/webhooks/...,false`}</pre>
        </div>
        <div className="mb-4">
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-2">JSON:</p>
          <pre className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-3 text-xs overflow-x-auto leading-relaxed">{`[
  { "name": "My Email", "type": "email",   "value": "alerts@example.com",                  "enabled": true },
  { "name": "Slack",    "type": "slack",   "value": "https://hooks.slack.com/services/...", "enabled": true },
  { "name": "Discord",  "type": "discord", "value": "https://discord.com/api/webhooks/...", "enabled": false }
]`}</pre>
        </div>
        <ul className="space-y-1 text-sm text-gray-600 dark:text-gray-400 mb-4 list-disc list-inside">
          <li><code className="font-mono text-xs">type</code> must be <code className="font-mono text-xs">email</code>, <code className="font-mono text-xs">slack</code>, or <code className="font-mono text-xs">discord</code></li>
          <li><code className="font-mono text-xs">value</code> is the email address for email channels, webhook URL for Slack/Discord</li>
          <li><code className="font-mono text-xs">enabled</code> is optional — defaults to <code className="font-mono text-xs">true</code></li>
        </ul>

        <SubHeading title="Conflict handling" />
        <p className="text-gray-600 dark:text-gray-400 mb-3">
          A conflict occurs when an imported channel has the same type and value as an existing one.
          Pingr detects this in the preview and gives you two options:
        </p>
        <div className="space-y-2 mb-4">
          {[
            { label: 'Skip conflicts',      desc: 'Leave existing channels untouched. Only new channels are imported.' },
            { label: 'Overwrite conflicts', desc: 'Update the name and enabled state of the existing channel. The destination (email/webhook) stays the same.' },
          ].map(({ label, desc }) => (
            <div key={label} className="flex gap-3 text-sm">
              <CheckCircle size={15} className="text-green-500 shrink-0 mt-0.5" />
              <div>
                <span className="font-medium text-gray-900 dark:text-white">{label}</span>
                <span className="text-gray-500 dark:text-gray-400"> — {desc}</span>
              </div>
            </div>
          ))}
        </div>
        <Tip>
          <strong>Security:</strong> Your export file contains webhook URLs. Treat it like a password —
          don't share it publicly or commit it to version control.
        </Tip>
      </section>

      {/* ── Subscriptions ── */}
      <section id="subscriptions" className="scroll-mt-20">
        <SectionHeading icon={Mail} title="Subscriptions" />
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          A subscription links a monitor to an alert channel. When the monitor goes down or recovers,
          a notification is sent to every subscribed channel.
        </p>
        <SubHeading title="Managing subscriptions" />
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Open any monitor from the dashboard → scroll to the <strong>Alert Channels</strong> section →
          subscribe or unsubscribe channels individually. You can subscribe multiple channels to one monitor
          and one channel to multiple monitors.
        </p>
        <SubHeading title="What notifications you receive" />
        <div className="space-y-3">
          {[
            { icon: Bell,        label: 'DOWN alert',     desc: 'Sent as soon as the monitor crosses the failure threshold.' },
            { icon: CheckCircle, label: 'RECOVERY alert', desc: 'Sent when the monitor comes back up after being down.' },
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

      {/* ── Incidents ── */}
      <section id="incidents" className="scroll-mt-20">
        <SectionHeading icon={AlertTriangle} title="Incidents" />
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Incidents are the public-facing communication layer. They appear on your status page and let
          you communicate what's happening, what you know, and when things are resolved.
        </p>

        <SubHeading title="Auto vs manual incidents" />
        <div className="space-y-3 mb-6">
          <div className="border border-gray-200 dark:border-gray-700 rounded-xl p-4 bg-white dark:bg-gray-800">
            <p className="text-sm font-semibold text-gray-900 dark:text-white mb-1 flex items-center gap-2">
              <span className="text-xs bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400 px-1.5 py-0.5 rounded">Auto</span>
              Created by the worker
            </p>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              When a monitor crosses its failure threshold, the worker automatically opens an incident
              and posts an initial "Investigating" update. When the monitor recovers, the incident is
              automatically resolved.
            </p>
          </div>
          <div className="border border-gray-200 dark:border-gray-700 rounded-xl p-4 bg-white dark:bg-gray-800">
            <p className="text-sm font-semibold text-gray-900 dark:text-white mb-1 flex items-center gap-2">
              <span className="text-xs bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 px-1.5 py-0.5 rounded">Manual</span>
              Created by you
            </p>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              Go to <strong>Incidents</strong> in the sidebar → <strong>"+ New incident"</strong>.
              Use these for planned maintenance, partial degradation that doesn't trip the check threshold,
              or any situation you want to communicate proactively.
            </p>
          </div>
        </div>

        <SubHeading title="Incident statuses" />
        <div className="space-y-2 mb-6">
          {[
            { badge: 'Investigating', color: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',         desc: 'You are aware of the issue and investigating the cause.' },
            { badge: 'Identified',    color: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400', desc: 'Root cause has been found. Fix is in progress.' },
            { badge: 'Monitoring',    color: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400', desc: 'Fix deployed. Watching to confirm recovery.' },
            { badge: 'Resolved',      color: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',  desc: 'Incident is over. Everything is back to normal.' },
          ].map(({ badge, color, desc }) => (
            <div key={badge} className="flex items-start gap-3 text-sm">
              <span className={`${color} px-2 py-0.5 rounded-full text-xs font-medium shrink-0 mt-0.5`}>{badge}</span>
              <span className="text-gray-600 dark:text-gray-400">{desc}</span>
            </div>
          ))}
        </div>

        <SubHeading title="Posting updates" />
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Open an incident → scroll to <strong>Post an update</strong>. Each update has a status and a message.
          Updates are shown as a timeline on your public status page, newest first.
          You can also change the component status of affected monitors with each update.
        </p>

        <SubHeading title="Affecting components" />
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          When creating or updating an incident, you can select which monitors are affected and set
          a <strong>component status</strong> for each one (e.g. set one monitor to <em>Partial Outage</em> and
          another to <em>Degraded Performance</em>). This is what visitors see on your public status page.
        </p>

        <SubHeading title="Notifications" />
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          When creating or posting an update, check <strong>"Send notification to alert channels"</strong> to
          push the update to all subscribed channels for the affected monitors.
        </p>

        <Note>
          <strong>Note:</strong> Auto-resolved incidents are not overwritten if you've already manually
          resolved them. Manual resolution always takes priority.
        </Note>
      </section>

      {/* ── Status page ── */}
      <section id="status-page" className="scroll-mt-20">
        <SectionHeading icon={ExternalLink} title="Public status page" />
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Every account gets a public status page. No configuration needed — it's live the moment
          you add your first monitor.
        </p>
        <div className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg px-4 py-3 font-mono text-sm text-indigo-700 dark:text-indigo-300 mb-4">
          https://getpingr.com/status/<span className="font-bold">your-username</span>
        </div>

        <SubHeading title="What it shows" />
        <div className="space-y-2 mb-4">
          {[
            'Overall system status banner (all operational vs issues detected)',
            'Each monitor with its component status and 90-day uptime bar',
            '24h / 7d / 30d / 90d uptime percentages per monitor',
            'Active incidents with full update timeline',
            'Incident history (last 20 resolved incidents)',
          ].map(item => (
            <div key={item} className="flex gap-2 text-sm">
              <CheckCircle size={15} className="text-green-500 shrink-0 mt-0.5" />
              <span className="text-gray-600 dark:text-gray-400">{item}</span>
            </div>
          ))}
        </div>

        <SubHeading title="Accessing it" />
        <p className="text-gray-600 dark:text-gray-400 mb-3">
          Click <strong>Status Page ↗</strong> in the bottom of the left sidebar. The page is fully public —
          no login required for your users.
        </p>
        <Tip><strong>Tip:</strong> Add your status page URL to your app's footer or README so users check it before submitting support tickets.</Tip>
      </section>

      {/* ── Tips ── */}
      <section id="tips" className="scroll-mt-20">
        <SectionHeading icon={Clock} title="Tips & best practices" />
        <div className="space-y-4">
          {[
            { title: 'Monitor a health endpoint, not the homepage',   desc: 'Create a /health route that checks your DB and dependencies. A static homepage can return 200 even when your API is broken.' },
            { title: 'Use components to mirror your architecture',     desc: 'Group monitors the same way your system is divided — Frontend, API, Database, CDN. Your status page becomes much easier to read.' },
            { title: 'Lower failure threshold for critical services',  desc: 'Threshold of 1 = alert on first failure. Raise it to 2-3 for services prone to transient blips to reduce noise.' },
            { title: 'Pause before maintenance windows',              desc: 'Prevents false-positive DOWN alerts and keeps your uptime stats accurate.' },
            { title: 'Create a manual incident for planned downtime',  desc: 'Set component status to "Under Maintenance" so users know it\'s intentional, not an outage.' },
            { title: 'Subscribe team-shared channels',                desc: "Add a team Slack channel or group email so alerts aren't siloed in one person's inbox." },
            { title: 'Export channels before migrating accounts',     desc: 'Use the Export → CSV feature to back up all your alert channels before making big changes.' },
            { title: 'Watch the response time chart',                 desc: 'Gradual latency increases often precede outages. Investigate a climb before it becomes a full incident.' },
            { title: 'Share your status page publicly',              desc: 'Link it from your app footer, docs, and README. Transparent communication builds user trust.' },
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

      {/* CTA — guests only */}
      {!loggedIn && (
        <div className="border border-indigo-100 dark:border-indigo-900/50 bg-indigo-50 dark:bg-indigo-900/20 rounded-2xl p-8 text-center">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">Ready to start monitoring?</h3>
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">Set up your first monitor in under 2 minutes. Free, no credit card required.</p>
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

// ─── Page shell ───────────────────────────────────────────────────────────────

export function DocsPage() {
  usePageTitle('Documentation')
  const loggedIn = isLoggedIn()

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

  return (
    <div className="min-h-screen bg-white dark:bg-gray-950">
      <header className="sticky top-0 z-10 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700">
        <div className="max-w-6xl mx-auto px-6 h-14 flex items-center justify-between">
          <Link to="/" className="text-xl font-bold text-indigo-600">Pingr</Link>
          <div className="flex items-center gap-4">
            <Link to="/login" className="text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white">Sign in</Link>
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
