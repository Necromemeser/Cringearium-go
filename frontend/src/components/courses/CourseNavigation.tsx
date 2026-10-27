import { useEffect, useRef, useState, type ReactNode } from 'react'

export default function CourseNavigation({ children, progress }: { children: ReactNode; progress: number }) {
  const [mobile, setMobile] = useState(() => window.matchMedia('(max-width: 700px)').matches)
  const [open, setOpen] = useState(false)
  const panel = useRef<HTMLElement>(null)
  const toggle = useRef<HTMLButtonElement>(null)
  const restoreFocus = useRef(true)
  useEffect(() => {
    const query = window.matchMedia('(max-width: 700px)')
    const resize = () => { setMobile(query.matches); setOpen(false) }
    query.addEventListener('change', resize)
    return () => query.removeEventListener('change', resize)
  }, [])
  useEffect(() => {
    if (!mobile || !open) return
    const previous = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    panel.current?.querySelector<HTMLButtonElement>('button')?.focus()
    const keydown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') { setOpen(false); return }
      if (event.key !== 'Tab') return
      const items = Array.from(panel.current?.querySelectorAll<HTMLElement>('button:not(:disabled), a[href]') ?? [])
      const first = items[0]; const last = items[items.length - 1]
      if (event.shiftKey && (document.activeElement === first || !panel.current?.contains(document.activeElement))) { event.preventDefault(); last?.focus() }
      else if (!event.shiftKey && (document.activeElement === last || !panel.current?.contains(document.activeElement))) { event.preventDefault(); first?.focus() }
    }
    document.addEventListener('keydown', keydown)
    return () => {
      document.body.style.overflow = previous
      document.removeEventListener('keydown', keydown)
      if (restoreFocus.current && toggle.current?.isConnected && toggle.current.getClientRects().length) toggle.current.focus({ preventScroll: true })
    }
  }, [mobile, open])
  return <>
    <div className="course-mobile-toolbar"><button type="button" ref={toggle} aria-expanded={open} aria-controls="course-navigation" onClick={() => { restoreFocus.current = true; setOpen(true) }}>Содержание курса <span aria-hidden="true">☰</span></button><span>{progress}% пройдено</span></div>
    {mobile && open && <button type="button" className="course-menu-backdrop" tabIndex={-1} aria-label="Закрыть оглавление" onClick={() => setOpen(false)} />}
    <aside ref={panel} id="course-navigation" className={`course-sidebar${mobile && open ? ' is-open' : ''}`} role={mobile ? 'dialog' : undefined} aria-modal={mobile && open ? true : undefined} aria-label="Содержание курса" onClick={event => { if (mobile && (event.target as Element).closest('a, .course-page-button')) { restoreFocus.current = false; setOpen(false) } }}>
      <div className="course-drawer-heading"><strong>Содержание курса</strong><button type="button" aria-label="Закрыть меню курса" onClick={() => setOpen(false)}>×</button></div>
      {children}
    </aside>
  </>
}
