import { useLayoutEffect } from 'react'

// Includes both sticky bars; selecting a lesson must reveal its heading on small screens.
export function scrollToLesson() {
  const target = document.querySelector<HTMLElement>('.lesson-header')
  if (!target) return
  const headerHeight = document.querySelector('.app-shell > .header')?.getBoundingClientRect().height ?? 0
  const toolbarHeight = document.querySelector('.course-mobile-toolbar')?.getBoundingClientRect().height ?? 0
  window.scrollTo({ top: Math.max(0, window.scrollY + target.getBoundingClientRect().top - headerHeight - toolbarHeight - 16), behavior: 'instant' })
  target.querySelector<HTMLElement>('h2')?.focus({ preventScroll: true })
}

export function useScrollToLesson(transition: string | number | null | undefined, enabled: boolean) {
  useLayoutEffect(() => { if (enabled) scrollToLesson() }, [transition, enabled])
}
