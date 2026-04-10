import { useQuery } from '@tanstack/react-query'
import { userApi } from '../../api/user'
import { DashboardLayout } from '../../components/layout/DashboardLayout'
import { User, Mail, Tag, Calendar, ExternalLink } from 'lucide-react'
import { usePageTitle } from '../../lib/usePageTitle'

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })
}

export function ProfilePage() {
  usePageTitle('Profile')
  const { data: profile, isLoading } = useQuery({
    queryKey: ['me'],
    queryFn: () => userApi.me().then(r => r.data),
  })

  return (
    <DashboardLayout>
      <div className="max-w-2xl mx-auto">
        <h1 className="text-xl font-semibold text-gray-900 mb-6">Profile</h1>

        {isLoading ? (
          <div className="text-sm text-gray-400 py-10 text-center">Loading…</div>
        ) : profile ? (
          <div className="bg-white border border-gray-200 rounded-xl overflow-hidden">
            {/* Avatar block */}
            <div className="px-6 py-6 border-b border-gray-100 flex items-center gap-4">
              <div className="w-14 h-14 rounded-full bg-indigo-100 flex items-center justify-center shrink-0">
                <span className="text-xl font-semibold text-indigo-600">
                  {profile.username[0].toUpperCase()}
                </span>
              </div>
              <div>
                <p className="text-base font-semibold text-gray-900">{profile.username}</p>
                <p className="text-sm text-gray-400">{profile.email}</p>
              </div>
            </div>

            {/* Info rows */}
            <div className="divide-y divide-gray-100">
              <ProfileRow icon={<User size={15} />} label="Username" value={profile.username} />
              <ProfileRow icon={<Mail size={15} />} label="Email" value={profile.email} />
              <ProfileRow
                icon={<Tag size={15} />}
                label="Plan"
                value={
                  <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-indigo-50 text-indigo-700 capitalize">
                    {profile.plan}
                  </span>
                }
              />
              <ProfileRow
                icon={<Calendar size={15} />}
                label="Member since"
                value={formatDate(profile.created_at)}
              />
              <ProfileRow
                icon={<ExternalLink size={15} />}
                label="Status page"
                value={
                  <a
                    href={`/status/${profile.username}`}
                    target="_blank"
                    rel="noreferrer"
                    className="text-indigo-600 hover:underline text-sm"
                  >
                    /status/{profile.username} ↗
                  </a>
                }
              />
            </div>
          </div>
        ) : null}
      </div>
    </DashboardLayout>
  )
}

function ProfileRow({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode
  label: string
  value: React.ReactNode
}) {
  return (
    <div className="flex items-center px-6 py-4 gap-4">
      <div className="flex items-center gap-2 w-36 shrink-0 text-gray-400 text-sm">
        {icon}
        <span>{label}</span>
      </div>
      <div className="text-sm text-gray-700">{value}</div>
    </div>
  )
}
