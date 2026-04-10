import { useState, useRef, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { UserCircle, LogOut } from 'lucide-react'
import { clearTokens } from '../../api/client'
import { useQuery } from '@tanstack/react-query'
import { userApi } from '../../api/user'

export function UserMenu() {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  const navigate = useNavigate()

  const { data: profile } = useQuery({
    queryKey: ['me'],
    queryFn: () => userApi.me().then(r => r.data),
  })

  useEffect(() => {
    function onClickOutside(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', onClickOutside)
    return () => document.removeEventListener('mousedown', onClickOutside)
  }, [])

  const logout = () => {
    clearTokens()
    navigate('/login')
  }

  const initial = profile?.username?.[0]?.toUpperCase() ?? '?'

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen(o => !o)}
        className="w-8 h-8 rounded-full bg-indigo-100 flex items-center justify-center text-sm font-semibold text-indigo-600 hover:bg-indigo-200 transition"
        title={profile?.username}
      >
        {initial}
      </button>

      {open && (
        <div className="absolute right-0 top-10 w-56 bg-white border border-gray-200 rounded-xl shadow-lg z-50 overflow-hidden">
          {/* User info */}
          <div className="px-4 py-3 border-b border-gray-100">
            <p className="text-sm font-semibold text-gray-900">{profile?.username}</p>
            <p className="text-xs text-gray-400 truncate">{profile?.email}</p>
          </div>

          {/* Profile */}
          <Link
            to="/dashboard/profile"
            onClick={() => setOpen(false)}
            className="flex items-center gap-2.5 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition"
          >
            <UserCircle size={15} className="text-gray-400" />
            Profile
          </Link>

          {/* Sign out */}
          <div className="border-t border-gray-100">
            <button
              onClick={logout}
              className="flex items-center gap-2.5 w-full px-4 py-2.5 text-sm text-red-600 hover:bg-red-50 transition"
            >
              <LogOut size={15} />
              Sign out
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
