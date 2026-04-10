import { useState } from 'react'
import { Link } from 'react-router-dom'
import { authApi } from '../../api/auth'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Footer } from '../../components/ui/Footer'
import { usePageTitle } from '../../lib/usePageTitle'

export function ForgotPasswordPage() {
  usePageTitle('Forgot password')
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [done, setDone] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    await authApi.forgotPassword(email).catch(() => {})
    setDone(true)
    setLoading(false)
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      <div className="flex-1 flex items-center justify-center px-4 py-12">
        <div className="w-full max-w-md">
          <div className="text-center mb-8">
            <Link to="/" className="text-2xl font-bold text-indigo-600">Pingr</Link>
            <h1 className="mt-4 text-2xl font-semibold text-gray-900">Reset your password</h1>
          </div>

          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-8">
            {done ? (
              <div className="text-center">
                <div className="text-4xl mb-3">📬</div>
                <p className="text-gray-900 font-medium">Check your inbox</p>
                <p className="text-gray-500 text-sm mt-1">
                  We sent a password reset link to <span className="font-medium text-gray-700">{email}</span>.
                  If that address is registered, it'll arrive shortly.
                </p>
              </div>
            ) : (
              <form onSubmit={handleSubmit} className="space-y-4">
                <Input
                  label="Email"
                  type="email"
                  placeholder="you@example.com"
                  value={email}
                  onChange={e => setEmail(e.target.value)}
                />
                <Button type="submit" loading={loading} className="w-full">
                  Send reset link
                </Button>
              </form>
            )}
          </div>

          <p className="text-center mt-4 text-sm text-gray-500">
            <Link to="/login" className="text-indigo-600 hover:underline">Back to login</Link>
          </p>
        </div>
      </div>
      <Footer />
    </div>
  )
}
