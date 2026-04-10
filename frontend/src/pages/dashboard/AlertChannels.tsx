import { DashboardLayout } from '../../components/layout/DashboardLayout'
import { AlertChannelsSection } from './AlertChannelsSection'
import { usePageTitle } from '../../lib/usePageTitle'

export function AlertChannelsPage() {
  usePageTitle('Alert Channels')
  return (
    <DashboardLayout>
      <h1 className="text-xl font-semibold text-gray-900 mb-6">Alert Channels</h1>
      <AlertChannelsSection />
    </DashboardLayout>
  )
}
