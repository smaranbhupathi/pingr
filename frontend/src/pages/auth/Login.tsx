import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { authApi } from '../../api/auth'
import { saveTokens } from '../../api/client'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Footer } from '../../components/ui/Footer'
import { usePageTitle } from '../../lib/usePageTitle'

export function LoginPage() {
  usePageTitle('Sign in')
  const navigate = useNavigate()
  const [form, setForm] = useState({ email: '', password: '' })
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [apiError, setApiError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleChange = (field: string, value: string) => {
    setForm(f => ({ ...f, [field]: value }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setErrors({})
    setApiError('')
    setLoading(true)
    try {
      const { data } = await authApi.login(form)
      saveTokens(data.access_token, data.refresh_token)
      navigate('/dashboard')
    } catch (err: any) {
      if (err.response?.data?.errors) {
        setErrors(err.response.data.errors)
      } else {
        setApiError(err.response?.data?.error ?? 'Login failed')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      <div className="flex-1 flex items-center justify-center px-4 py-12">
        <div className="w-full max-w-md">
          <div className="text-center mb-8">
            <Link to="/" className="text-2xl font-bold text-indigo-600">Pingr</Link>
            <h1 className="mt-4 text-2xl font-semibold text-gray-900">Welcome back</h1>
            <p className="mt-1 text-sm text-gray-500">Sign in to your account</p>
          </div>

          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-8">
            {apiError && (
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700 font-medium">
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
                onChange={e => handleChange('email', e.target.value)}
              />
              <Input
                label="Password"
                type="password"
                placeholder="••••••••"
                value={form.password}
                error={errors.password}
                onChange={e => handleChange('password', e.target.value)}
              />
              <div className="text-right">
                <Link to="/forgot-password" className="text-sm text-indigo-600 hover:underline">
                  Forgot password?
                </Link>
              </div>
              <Button type="submit" loading={loading} className="w-full">
                Sign in
              </Button>
            </form>
          </div>

          <p className="text-center mt-4 text-sm text-gray-500">
            Don't have an account?{' '}
            <Link to="/register" className="text-indigo-600 font-medium hover:underline">
              Sign up free
            </Link>
          </p>
        </div>
      </div>
      <Footer />
    </div>
  )
}
