import { BrowserRouter, MemoryRouter, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { isLoggedIn } from './api/client'
import { ThemeProvider } from './lib/theme'

import { LandingPage } from './pages/Landing'
import { LoginPage } from './pages/auth/Login'
import { RegisterPage } from './pages/auth/Register'
import { ForgotPasswordPage } from './pages/auth/ForgotPassword'
import { ResetPasswordPage } from './pages/auth/ResetPassword'
import { VerifyEmailPage } from './pages/auth/VerifyEmail'
import { DashboardPage } from './pages/dashboard/Dashboard'
import { MonitorDetailPage } from './pages/dashboard/MonitorDetail'
import { AlertChannelsPage } from './pages/dashboard/AlertChannels'
import { AlertChannelDetailPage } from './pages/dashboard/AlertChannelDetail'
import { ProfilePage } from './pages/dashboard/Profile'
import { SettingsPage } from './pages/dashboard/Settings'
import { IncidentsPage } from './pages/dashboard/IncidentsPage'
import { IncidentDetailPage } from './pages/dashboard/IncidentDetail'
import { ComponentsPage } from './pages/dashboard/ComponentsPage'
import { StatusPage } from './pages/status/StatusPage'
import { DocsPage } from './pages/Docs'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30_000 },
  },
})

function RequireAuth({ children }: { children: React.ReactNode }) {
  return isLoggedIn() ? <>{children}</> : <Navigate to="/login" replace />
}

function RequireGuest({ children }: { children: React.ReactNode }) {
  return !isLoggedIn() ? <>{children}</> : <Navigate to="/dashboard" replace />
}

// If the app is loaded on a *.getpingr.com subdomain (e.g. acme-corp.getpingr.com),
// render the status page for that slug directly — React Router sees "/" so we
// intercept before routing kicks in.
function SubdomainStatusPage() {
  const host = window.location.hostname // e.g. "acme-corp.getpingr.com"
  const slug = host.replace('.getpingr.com', '')
  return <StatusPage slugOverride={slug} />
}

const ROOT_DOMAIN = 'getpingr.com'
const isSubdomain = (
  window.location.hostname.endsWith(`.${ROOT_DOMAIN}`) &&
  window.location.hostname !== `www.${ROOT_DOMAIN}`
)

export default function App() {
  if (isSubdomain) {
    return (
      <ThemeProvider>
        <QueryClientProvider client={queryClient}>
          <MemoryRouter>
            <SubdomainStatusPage />
          </MemoryRouter>
        </QueryClientProvider>
      </ThemeProvider>
    )
  }

  return (
    <ThemeProvider>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <Routes>
            {/* Public */}
            <Route path="/" element={<LandingPage />} />
            <Route path="/docs" element={<DocsPage />} />
            <Route path="/status/:username" element={<StatusPage />} />
            <Route path="/verify-email" element={<VerifyEmailPage />} />

            {/* Guest only */}
            <Route path="/login" element={<RequireGuest><LoginPage /></RequireGuest>} />
            <Route path="/register" element={<RequireGuest><RegisterPage /></RequireGuest>} />
            <Route path="/forgot-password" element={<RequireGuest><ForgotPasswordPage /></RequireGuest>} />
            <Route path="/reset-password" element={<RequireGuest><ResetPasswordPage /></RequireGuest>} />

            {/* Protected */}
            <Route path="/dashboard" element={<RequireAuth><DashboardPage /></RequireAuth>} />
            <Route path="/dashboard/monitors/:id" element={<RequireAuth><MonitorDetailPage /></RequireAuth>} />
            <Route path="/dashboard/alert-channels" element={<RequireAuth><AlertChannelsPage /></RequireAuth>} />
            <Route path="/dashboard/alert-channels/:id" element={<RequireAuth><AlertChannelDetailPage /></RequireAuth>} />
            <Route path="/dashboard/profile" element={<RequireAuth><ProfilePage /></RequireAuth>} />
            <Route path="/dashboard/settings" element={<RequireAuth><SettingsPage /></RequireAuth>} />
            <Route path="/dashboard/incidents" element={<RequireAuth><IncidentsPage /></RequireAuth>} />
            <Route path="/dashboard/incidents/:id" element={<RequireAuth><IncidentDetailPage /></RequireAuth>} />
            <Route path="/dashboard/components" element={<RequireAuth><ComponentsPage /></RequireAuth>} />

            {/* Fallback */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </QueryClientProvider>
    </ThemeProvider>
  )
}
