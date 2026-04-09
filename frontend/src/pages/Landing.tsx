import { Link } from 'react-router-dom'
import { Activity, Bell, Globe, BarChart2 } from 'lucide-react'

export function LandingPage() {
  return (
    <div className="min-h-screen bg-white">
      {/* Navbar */}
      <nav className="border-b border-gray-100 px-6 py-4 flex items-center justify-between max-w-6xl mx-auto">
        <span className="text-xl font-bold text-indigo-600">Pingr</span>
        <div className="flex items-center gap-4">
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
          <Link
            to="/status/pingr"
            className="text-gray-500 hover:text-gray-700 text-sm underline"
          >
            See example status page
          </Link>
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
              desc: 'Share pingr.app/status/yourname with your users. Build trust through transparency.',
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
            {/* Free */}
            <div className="bg-white border border-gray-200 rounded-xl p-8">
              <p className="text-sm font-medium text-gray-500 mb-1">Free</p>
              <p className="text-4xl font-bold text-gray-900 mb-1">₹0</p>
              <p className="text-sm text-gray-400 mb-6">Forever</p>
              <ul className="space-y-3 text-sm text-gray-600">
                {['5 monitors', '1-minute checks', 'Email alerts', 'Public status page', '90-day history'].map(f => (
                  <li key={f} className="flex items-center gap-2">
                    <span className="text-green-500">✓</span> {f}
                  </li>
                ))}
              </ul>
              <Link to="/register" className="mt-8 block text-center bg-gray-900 text-white py-2.5 rounded-lg text-sm font-medium hover:bg-gray-700">
                Get started free
              </Link>
            </div>

            {/* Pro — coming soon */}
            <div className="bg-indigo-600 border border-indigo-600 rounded-xl p-8 text-white">
              <p className="text-sm font-medium text-indigo-200 mb-1">Pro</p>
              <p className="text-4xl font-bold mb-1">₹299</p>
              <p className="text-sm text-indigo-200 mb-6">/month · Coming soon</p>
              <ul className="space-y-3 text-sm text-indigo-100">
                {['50 monitors', '30-second checks', 'Slack + Discord alerts', 'Custom domain status page', 'SSL monitoring', 'Multi-region checks'].map(f => (
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

      {/* Footer */}
      <footer className="border-t border-gray-100 px-6 py-6 text-center text-xs text-gray-400">
        © {new Date().getFullYear()} Pingr · Built with ❤️ in Go
      </footer>
    </div>
  )
}
