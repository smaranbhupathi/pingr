import { Link } from 'react-router-dom'
import { Footer } from '../components/ui/Footer'
import { usePageTitle } from '../lib/usePageTitle'

export function PrivacyPage() {
  usePageTitle('Privacy Policy')

  return (
    <div className="min-h-screen bg-white dark:bg-gray-950 flex flex-col">
      <nav className="border-b border-gray-100 dark:border-gray-800 px-6 py-4 flex items-center justify-between max-w-4xl mx-auto w-full">
        <Link to="/" className="text-xl font-bold text-indigo-600 dark:text-indigo-400">Pingr</Link>
        <Link to="/" className="text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200">← Back</Link>
      </nav>

      <main className="flex-1 max-w-3xl mx-auto w-full px-6 py-12">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">Privacy Policy</h1>
        <p className="text-sm text-gray-400 dark:text-gray-500 mb-10">Last updated: April 2026</p>

        <div className="prose prose-gray dark:prose-invert max-w-none space-y-8 text-sm text-gray-600 dark:text-gray-400 leading-relaxed">

          <section>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-2">1. What we collect</h2>
            <p>When you create an account, we collect your email address, a username, and a hashed password. We never store your plain-text password.</p>
            <p className="mt-2">When you add a monitor, we store the URL, check interval, and the results of each uptime check (response time, status code, timestamp).</p>
            <p className="mt-2">When you configure alert channels, we store the delivery addresses or webhook URLs you provide (email addresses, Slack/Discord webhook URLs).</p>
          </section>

          <section>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-2">2. How we use it</h2>
            <ul className="list-disc pl-5 space-y-1">
              <li>To run uptime checks on the URLs you register.</li>
              <li>To send you downtime and recovery alerts via the channels you configure.</li>
              <li>To display your public status page to anyone who visits it.</li>
              <li>To send transactional emails (email verification, password reset).</li>
            </ul>
            <p className="mt-2">We do not sell your data, run advertising, or share your information with third parties except the service providers listed below.</p>
          </section>

          <section>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-2">3. Third-party services</h2>
            <ul className="list-disc pl-5 space-y-1">
              <li><strong>Resend</strong> — used to deliver transactional emails.</li>
              <li><strong>Neon</strong> — hosted PostgreSQL database where your data is stored.</li>
              <li><strong>Cloudflare R2</strong> — used to store profile avatars.</li>
              <li><strong>Railway</strong> — hosts the API and worker processes.</li>
              <li><strong>Cloudflare Pages</strong> — hosts the frontend.</li>
            </ul>
            <p className="mt-2">Each provider processes only the minimum data needed to deliver the service.</p>
          </section>

          <section>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-2">4. Data retention</h2>
            <p>Uptime check results are retained for 90 days. Account data is retained for as long as your account is active. You can request deletion at any time by emailing us.</p>
          </section>

          <section>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-2">5. Cookies</h2>
            <p>Pingr does not use tracking or advertising cookies. Authentication tokens are stored in your browser's localStorage, not cookies.</p>
          </section>

          <section>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-2">6. Your rights</h2>
            <p>You can request access to, correction of, or deletion of your personal data at any time. To do so, email us at <a href="mailto:privacy@getpingr.com" className="text-indigo-500 hover:underline">privacy@getpingr.com</a>.</p>
          </section>

          <section>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-2">7. Changes</h2>
            <p>We may update this policy from time to time. We will notify registered users of material changes by email.</p>
          </section>

          <section>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-2">8. Contact</h2>
            <p>Questions? Email us at <a href="mailto:privacy@getpingr.com" className="text-indigo-500 hover:underline">privacy@getpingr.com</a>.</p>
          </section>

        </div>
      </main>

      <Footer />
    </div>
  )
}
