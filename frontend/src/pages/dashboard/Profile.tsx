import { useRef, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { userApi } from '../../api/user'
import { DashboardLayout } from '../../components/layout/DashboardLayout'
import { User, Mail, Tag, Calendar, ExternalLink, Camera, Loader2 } from 'lucide-react'
import { usePageTitle } from '../../lib/usePageTitle'

const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/webp']
const MAX_SIZE_MB = 2

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })
}

export function ProfilePage() {
  usePageTitle('Profile')
  const queryClient = useQueryClient()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [uploadError, setUploadError] = useState<string | null>(null)

  const { data: profile, isLoading } = useQuery({
    queryKey: ['me'],
    queryFn: () => userApi.me().then(r => r.data),
  })

  // uploadAvatar does the full three-step flow:
  // 1. Ask API for a presigned PUT URL
  // 2. PUT the file directly to R2/S3/MinIO (no API in the path)
  // 3. Tell the API the public URL so it can save it
  const uploadMutation = useMutation({
    mutationFn: async (file: File) => {
      // Step 1 — get presigned URL from our API
      const { data } = await userApi.getAvatarUploadUrl(file.type)

      // Step 2 — PUT file directly to storage (bypass our API)
      // We use fetch here, not axios, because we need to PUT to an external URL
      // without the Authorization header (the presigned URL is the auth).
      const putRes = await fetch(data.upload_url, {
        method: 'PUT',
        body: file,
        headers: { 'Content-Type': file.type },
      })
      if (!putRes.ok) throw new Error('Upload to storage failed')

      // Step 3 — save the public URL in our database
      await userApi.updateAvatar(data.public_url)

      return data.public_url
    },
    onSuccess: () => {
      setUploadError(null)
      // Invalidate the profile query so the new avatar is fetched
      queryClient.invalidateQueries({ queryKey: ['me'] })
    },
    onError: (err: Error) => {
      setUploadError(err.message)
    },
  })

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return

    if (!ALLOWED_TYPES.includes(file.type)) {
      setUploadError('Only JPEG, PNG, and WebP images are supported.')
      return
    }
    if (file.size > MAX_SIZE_MB * 1024 * 1024) {
      setUploadError(`Image must be smaller than ${MAX_SIZE_MB} MB.`)
      return
    }

    setUploadError(null)
    uploadMutation.mutate(file)
    // Reset the input so the same file can be re-selected after an error
    e.target.value = ''
  }

  const isUploading = uploadMutation.isPending

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
              {/* Clickable avatar with camera overlay */}
              <button
                type="button"
                onClick={() => !isUploading && fileInputRef.current?.click()}
                className="relative w-14 h-14 rounded-full shrink-0 group focus:outline-none"
                title="Change profile photo"
                disabled={isUploading}
              >
                {profile.avatar_url ? (
                  <img
                    src={profile.avatar_url}
                    alt={profile.username}
                    className="w-14 h-14 rounded-full object-cover"
                  />
                ) : (
                  <div className="w-14 h-14 rounded-full bg-indigo-100 flex items-center justify-center">
                    <span className="text-xl font-semibold text-indigo-600">
                      {profile.username[0].toUpperCase()}
                    </span>
                  </div>
                )}

                {/* Hover / loading overlay */}
                <div className={`absolute inset-0 rounded-full flex items-center justify-center transition-opacity
                  ${isUploading ? 'bg-black/40 opacity-100' : 'bg-black/40 opacity-0 group-hover:opacity-100'}`}>
                  {isUploading
                    ? <Loader2 size={18} className="text-white animate-spin" />
                    : <Camera size={18} className="text-white" />
                  }
                </div>
              </button>

              <input
                ref={fileInputRef}
                type="file"
                accept={ALLOWED_TYPES.join(',')}
                className="hidden"
                onChange={handleFileChange}
              />

              <div>
                <p className="text-base font-semibold text-gray-900">{profile.username}</p>
                <p className="text-sm text-gray-400">{profile.email}</p>
                {uploadError && (
                  <p className="text-xs text-red-500 mt-1">{uploadError}</p>
                )}
                {isUploading && (
                  <p className="text-xs text-indigo-500 mt-1">Uploading…</p>
                )}
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
