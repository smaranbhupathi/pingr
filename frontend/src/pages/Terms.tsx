import { Link } from 'react-router-dom'
import { Footer } from '../components/ui/Footer'
import { usePageTitle } from '../lib/usePageTitle'

export function TermsPage() {
  usePageTitle('Terms of Use')

  return (
    <div className="min-h-screen bg-white dark:bg-gray-950 flex flex-col">
      <nav className="border-b border-gray-100 dark:border-gray-800 px-6 py-4 flex items-center justify-between max-w-4xl mx-auto w-full">
        <Link to="/" className="text-xl font-bold text-indigo-600 dark:text-indigo-400">Pingr</Link>
        <Link to="/" className="text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200">← Back</Link>
      </nav>

      <main className="flex-1 max-w-3xl mx-auto w-full px-6 py-12">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">Terms of Use</h1>
        <p className="text-sm text-gray-400 dark:text-gray-500 mb-10">Last updated: April 2026</p>

        <div className="space-y-8 text-sm text-gray-600 dark:text-gray-400 leading-relaxed">

          <section>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-2">1. Acceptance</h2>
            <p>By creating an account or using Pingr, you agree to these Terms of Use. If you do not agree, do not use the service.</p>
          </section>

          <section>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-2">2. The service</h2>
            <p>Pingr provides HTTP uptime monitoring, public status pages, and alert notifications. We offer a free tier with limited monitors and a paid Pro tier (coming soon) with higher limits.</p>
          </section>

          <section>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-2">3. Acceptable use</h2>
            <p>You may only monitor URLs that you own or have explicit permission to monitor. You must not use Pingr to:</p>
            <ul className="list-disc pl-5 space-y-1 mt-2">
              <li>Launch denial-of-service attacks or excessive load against any target.</li>
              <li>Monitor URLs for the purpose of scraping or data harvesting.</li>
              <li>Violate any applicable law or third-party rights.</li>
              <li>Attempt to circumvent plan limits through multiple accounts.</li>
            </ul>
          </section>

          <section>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-2">4. Free tier limits</h2>
            <p>Free accounts are limited to 5 monitors. We reserve the right to adjust these limits with reasonable notice.</p>
          </section>

          <section>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-2">5. Uptime and availability</h2>
            <p>We aim for high availability but do not guarantee uninterrupted service. Pingr is provided "as is" without warranties of any kind. We are not liable for missed alerts or incorrect uptime readings caused by service interruptions.</p>
          </section>

          <section>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-2">6. Account termination</h2>
            <p>You may delete your account at any time. We reserve the right to suspend or terminate accounts that violate these terms.</p>
          </section>

          <section>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-2">7. Limitation of liability</h2>
            <p>To the maximum extent permitted by law, Pingr and its operators shall not be liable for any indirect, incidental, or consequential damages arising from your use of the service.</p>
          </section>

          <section>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-2">8. Changes to terms</h2>
            <p>We may update these terms from time to time. Continued use of the service after changes constitutes acceptance of the new terms.</p>
          </section>

          <section>
            <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-2">9. Contact</h2>
            <p>Questions? Email us at <a href="mailto:legal@getpingr.com" className="text-indigo-500 hover:underline">legal@getpingr.com</a>.</p>
          </section>

        </div>
      </main>

      <Footer />
    </div>
  )
}
