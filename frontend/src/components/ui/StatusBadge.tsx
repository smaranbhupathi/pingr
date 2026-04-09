type Status = 'up' | 'down' | 'paused' | 'pending'

const config: Record<Status, { label: string; className: string; dot: string }> = {
  up:      { label: 'Up',      className: 'bg-green-100 text-green-700',  dot: 'bg-green-500' },
  down:    { label: 'Down',    className: 'bg-red-100 text-red-700',      dot: 'bg-red-500' },
  paused:  { label: 'Paused', className: 'bg-gray-100 text-gray-600',    dot: 'bg-gray-400' },
  pending: { label: 'Pending', className: 'bg-yellow-100 text-yellow-700', dot: 'bg-yellow-400' },
}

export function StatusBadge({ status }: { status: Status }) {
  const { label, className, dot } = config[status] ?? config.pending
  return (
    <span className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium ${className}`}>
      <span className={`w-1.5 h-1.5 rounded-full ${dot}`} />
      {label}
    </span>
  )
}
