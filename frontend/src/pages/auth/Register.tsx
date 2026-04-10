import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { authApi } from '../../api/auth'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Footer } from '../../components/ui/Footer'
import { usePageTitle } from '../../lib/usePageTitle'

export function RegisterPage() {
  usePageTitle('Create account')
  const navigate = useNavigate()
  const [form, setForm] = useState({ email: '', username: '', password: '' })
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [apiError, setApiError] = useState('')
  const [loading, setLoading] = useState(false)
  const [done, setDone] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setApiError('')
    setErrors({})
    setLoading(true)
    try {
      await authApi.register(form)
      setDone(true)
    } catch (err: any) {
      if (err.response?.data?.errors) {
        setErrors(err.response.data.errors)
      } else {
        setApiError(err.response?.data?.error ?? 'Registration failed')
      }
    } finally {
      setLoading(false)
    }
  }

  if (done) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
        <div className="text-center max-w-md">
          <div className="text-5xl mb-4">📬</div>
          <h2 className="text-2xl font-semibold text-gray-900 mb-2">Check your email</h2>
          <p className="text-gray-500 text-sm">
            We sent a verification link to <strong>{form.email}</strong>. Click it to activate your account.
          </p>
          <Button variant="ghost" className="mt-6" onClick={() => navigate('/login')}>
            Back to login
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      <div className="flex-1 flex items-center justify-center px-4 py-12">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <Link to="/" className="text-2xl font-bold text-indigo-600">Pingr</Link>
          <h1 className="mt-4 text-2xl font-semibold text-gray-900">Create your account</h1>
          <p className="mt-1 text-sm text-gray-500">Free forever. No credit card required.</p>
        </div>

        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-8">
          {apiError && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-600">
              {apiError}
            </div>
          )}
          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              label="Email"
              type="email"
              placeholder="you@example.com"
              value={form.email}
              error={errors.email}
              onChange={e => setForm(f => ({ ...f, email: e.target.value }))}
            />
            <Input
              label="Username"
              placeholder="yourname"
              value={form.username}
              error={errors.username}
              onChange={e => setForm(f => ({ ...f, username: e.target.value }))}
            />
            <p className="text-xs text-gray-400 -mt-2">
              Your status page will be at pingr.app/status/{form.username || 'yourname'}
            </p>
            <Input
              label="Password"
              type="password"
              placeholder="Min. 8 characters"
              value={form.password}
              error={errors.password}
              onChange={e => setForm(f => ({ ...f, password: e.target.value }))}
            />
            <Button type="submit" loading={loading} className="w-full">
              Create account
            </Button>
          </form>
        </div>

        <p className="text-center mt-4 text-sm text-gray-500">
          Already have an account?{' '}
          <Link to="/login" className="text-indigo-600 font-medium hover:underline">
            Sign in
          </Link>
        </p>
      </div>
      </div>
      <Footer />
    </div>
  )
}
