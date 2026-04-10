import { useEffect } from 'react'

export function usePageTitle(title: string) {
  useEffect(() => {
    document.title = `${title} · Pingr`
    return () => { document.title = 'Pingr' }
  }, [title])
}
