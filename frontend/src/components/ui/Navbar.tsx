import { Link, useNavigate } from 'react-router-dom'
import { clearTokens } from '../../api/client'
import { useQuery } from '@tanstack/react-query'
import { userApi } from '../../api/user'

export function Navbar() {
  const navigate = useNavigate()
  const { data: profile } = useQuery({
    queryKey: ['me'],
    queryFn: () => userApi.me().then(r => r.data),
  })

  const logout = () => {
    clearTokens()
    navigate('/login')
  }

  return (
    <nav className="bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between">
      <Link to="/dashboard" className="text-lg font-bold text-indigo-600">Pingr</Link>
      <div className="flex items-center gap-4">
        {profile && (
          <Link
            to={`/status/${profile.username}`}
            target="_blank"
            className="text-sm text-gray-500 hover:text-indigo-600"
          >
            Status page ↗
          </Link>
        )}
        <span className="text-sm text-gray-500">{profile?.email}</span>
        <button onClick={logout} className="text-sm text-gray-500 hover:text-red-500">
          Sign out
        </button>
      </div>
    </nav>
  )
}
