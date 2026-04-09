import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { isLoggedIn } from './api/client'

import { LandingPage } from './pages/Landing'
import { LoginPage } from './pages/auth/Login'
import { RegisterPage } from './pages/auth/Register'
import { ForgotPasswordPage } from './pages/auth/ForgotPassword'
import { ResetPasswordPage } from './pages/auth/ResetPassword'
import { VerifyEmailPage } from './pages/auth/VerifyEmail'
import { DashboardPage } from './pages/dashboard/Dashboard'
import { MonitorDetailPage } from './pages/dashboard/MonitorDetail'
import { StatusPage } from './pages/status/StatusPage'

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

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          {/* Public */}
          <Route path="/" element={<LandingPage />} />
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

          {/* Fallback */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
