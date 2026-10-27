import { useLayoutEffect } from 'react'

// Scroll after the new lesson/round/result has committed, not before its height changes.
// On mobile the course menu precedes the lesson, so the document's top isn't the content top.
export function useScrollToLesson(transition: string | number | null | undefined, enabled: boolean) {
  useLayoutEffect(() => {
    if (!enabled) return
    const target = document.querySelector<HTMLElement>('.lesson-header')
    if (!target) return
    const headerHeight = document.querySelector('.app-shell > .header')?.getBoundingClientRect().height ?? 0
    window.scrollTo({ top: Math.max(0, window.scrollY + target.getBoundingClientRect().top - headerHeight - 16), behavior: 'instant' })
    target.querySelector<HTMLElement>('h2')?.focus({ preventScroll: true })
  }, [transition, enabled])
}
