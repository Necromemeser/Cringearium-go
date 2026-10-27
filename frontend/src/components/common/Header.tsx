import { useEffect, useRef, useState } from 'react'
import type { User } from '../../api/auth'
import Link from './Link'
import { navigate, usePath } from '../../utils/navigation'

export default function Header({ user, onLogout }: { user: User | null; onLogout: () => void }) {
  const [open, setOpen] = useState(false)
  const header = useRef<HTMLElement>(null)
  const toggle = useRef<HTMLButtonElement>(null)
  const path = usePath()
  useEffect(() => { setOpen(false) }, [path])
  useEffect(() => {
    if (!open) return
    const outside = (event: PointerEvent) => { if (!header.current?.contains(event.target as Node)) setOpen(false) }
    const escape = (event: KeyboardEvent) => { if (event.key === 'Escape') { setOpen(false); toggle.current?.focus() } }
    const desktop = window.matchMedia('(min-width: 981px)')
    const resize = () => { if (desktop.matches) setOpen(false) }
    document.addEventListener('pointerdown', outside)
    document.addEventListener('keydown', escape)
    desktop.addEventListener('change', resize)
    return () => { document.removeEventListener('pointerdown', outside); document.removeEventListener('keydown', escape); desktop.removeEventListener('change', resize) }
  }, [open])
  return <header className="header" ref={header}>
    <div className="container nav-container">
      <a href="/" className="brand" aria-label="Cringearium" onClick={event => { event.preventDefault(); setOpen(false); navigate('/') }}><span className="brand-full">Cringearium</span><span className="brand-short" aria-hidden="true">C</span></a>
      <button ref={toggle} type="button" className="mobile-nav-toggle" aria-expanded={open} aria-controls="site-navigation" onClick={() => setOpen(value => !value)}>{open ? 'Закрыть меню' : 'Меню'}<span aria-hidden="true">{open ? '×' : '☰'}</span></button>
      <div id="site-navigation" className={`header-menu${open ? ' is-open' : ''}`} onClick={event => { if ((event.target as Element).closest('a')) setOpen(false) }}>
        <nav className="nav-links" aria-label="Основная навигация"><Link href="/courses">Каталог курсов</Link><Link href="/chat">Кринжик</Link><Link href="/about">О нас</Link></nav>
        <nav className="nav-actions" aria-label="Аккаунт">{user ? <>{user.role === 'admin' && <Link href="/admin">Админ</Link>}<Link href="/profile" className="profile-link">{user.username}</Link><button type="button" className="nav-button" onClick={() => { setOpen(false); onLogout() }}>Выйти</button></> : <Link href="/login" className="login-link">Войти</Link>}</nav>
      </div>
    </div>
  </header>
}
