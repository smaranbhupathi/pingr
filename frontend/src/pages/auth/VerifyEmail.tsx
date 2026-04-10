import { useEffect, useRef, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { authApi } from '../../api/auth'
import { Footer } from '../../components/ui/Footer'

export function VerifyEmailPage() {
  const [params] = useSearchParams()
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading')
  const called = useRef(false)

  useEffect(() => {
    if (called.current) return
    called.current = true
    const token = params.get('token')
    if (!token) { setStatus('error'); return }
    authApi.verifyEmail(token)
      .then(() => setStatus('success'))
      .catch(() => setStatus('error'))
  }, [])

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      <div className="flex-1 flex items-center justify-center px-4">
      <div className="text-center max-w-md">
        {status === 'loading' && (
          <>
            <div className="text-4xl mb-3 animate-spin">⏳</div>
            <p className="text-gray-500">Verifying your email...</p>
          </>
        )}
        {status === 'success' && (
          <>
            <div className="text-5xl mb-4">✅</div>
            <h2 className="text-2xl font-semibold text-gray-900 mb-2">Email verified!</h2>
            <p className="text-gray-500 text-sm mb-6">Your account is ready. Start monitoring your services.</p>
            <Link to="/login" className="bg-indigo-600 text-white px-6 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700">
              Go to login
            </Link>
          </>
        )}
        {status === 'error' && (
          <>
            <div className="text-5xl mb-4">❌</div>
            <h2 className="text-2xl font-semibold text-gray-900 mb-2">Verification failed</h2>
            <p className="text-gray-500 text-sm mb-4">
              This link may have already been used or something went wrong. Try registering again.
            </p>
            <Link to="/register" className="text-indigo-600 text-sm font-medium hover:underline">
              Back to register
            </Link>
          </>
        )}
      </div>
      </div>
      <Footer />
    </div>
  )
}
