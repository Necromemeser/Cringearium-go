import type { ReactNode } from 'react'
import { navigate } from '../../utils/navigation'

export default function Link({ href, children, className = '' }: {
  href: string
  children: ReactNode
  className?: string
}) {
  return (
    <a href={href} className={className} onClick={(event) => {
      if (href.startsWith('/')) {
        event.preventDefault()
        navigate(href)
      }
    }}>
      {children}
    </a>
  )
}
