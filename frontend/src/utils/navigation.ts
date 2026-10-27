import { useEffect, useState } from 'react'

export function navigate(path: string, options: { replace?: boolean } = {}) {
  if (options.replace) window.history.replaceState({}, '', path)
  else window.history.pushState({}, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

export function usePath() {
  const [path, setPath] = useState(window.location.pathname)

  useEffect(() => {
    const handle = () => setPath(window.location.pathname)
    window.addEventListener('popstate', handle)
    return () => window.removeEventListener('popstate', handle)
  }, [])

  return path
}
