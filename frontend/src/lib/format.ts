export const format = {
  time: (iso: string) =>
    new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),

  datetime: (iso: string) =>
    new Date(iso).toLocaleString([], {
      month: 'short', day: 'numeric',
      hour: '2-digit', minute: '2-digit',
    }),

  duration: (start: string, end: string) => {
    const ms = new Date(end).getTime() - new Date(start).getTime()
    const mins = Math.floor(ms / 60000)
    if (mins < 60) return `${mins}m`
    const hrs = Math.floor(mins / 60)
    return `${hrs}h ${mins % 60}m`
  },

  timeAgo: (iso: string) => {
    const secs = Math.floor((Date.now() - new Date(iso).getTime()) / 1000)
    if (secs < 60) return `${secs}s ago`
    const mins = Math.floor(secs / 60)
    if (mins < 60) return `${mins}m ago`
    const hrs = Math.floor(mins / 60)
    return `${hrs}h ${mins % 60}m ago`
  },

  ongoingDuration: (start: string) => {
    const ms = Date.now() - new Date(start).getTime()
    const mins = Math.floor(ms / 60000)
    if (mins < 1) return 'just started'
    if (mins < 60) return `${mins}m`
    const hrs = Math.floor(mins / 60)
    return `${hrs}h ${mins % 60}m`
  },
}
