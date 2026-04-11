import { DashboardLayout } from '../../components/layout/DashboardLayout'
import { AlertChannelsSection } from './AlertChannelsSection'
import { usePageTitle } from '../../lib/usePageTitle'

export function AlertChannelsPage() {
  usePageTitle('Alert Channels')
  return (
    <DashboardLayout>
      <AlertChannelsSection />
    </DashboardLayout>
  )
}
