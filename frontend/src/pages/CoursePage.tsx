import { useEffect, useState } from 'react'
import { completePage, enrollCourse, formatCoursePrice, getCourse, getCourseProgress, type CourseDetails } from '../api/courses'
import type { User } from '../api/auth'
import { TOKEN_KEY } from '../constants/auth'
import Link from '../components/common/Link'
import Markdown from '../components/markdown/Markdown'
import TestPage from '../components/tests/TestPage'
import { navigate } from '../utils/navigation'
export default function CoursePage({ courseId, user }: { courseId: number; user: User | null }) {
  const [course, setCourse] = useState<CourseDetails | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [enrolling, setEnrolling] = useState(false)
  const [enrolled, setEnrolled] = useState(false)
  const [accessLoading, setAccessLoading] = useState(true)
  const [progress, setProgress] = useState<number[]>([])
  const [selectedPage, setSelectedPage] = useState<number | null>(null)
  const [progressLoading, setProgressLoading] = useState(false)
  const [actionError, setActionError] = useState('')

  const token = localStorage.getItem(TOKEN_KEY)

  useEffect(() => {
    getCourse(courseId)
      .then(setCourse)
      .catch((e) =>
        setError(
          e instanceof Error
            ? e.message
            : 'Не удалось загрузить курс',
        ),
      )
      .finally(() => setLoading(false))
  }, [courseId])

  useEffect(() => {
    if (!user || !token) {
      setEnrolled(false)
      setProgress([])
      setAccessLoading(false)
      return
    }

    setAccessLoading(true)

    fetch(`/api/courses/${courseId}/access`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
      .then(async (response) => {
        if (!response.ok) {
          throw new Error('Не удалось проверить доступ к курсу')
        }

        return response.json() as Promise<{
          has_access: boolean
        }>
      })
      .then((data) => setEnrolled(data.has_access))
      .catch(() => setEnrolled(false))
      .finally(() => setAccessLoading(false))
  }, [courseId, user, token])

  useEffect(() => {
    if (!user || !token || !enrolled) {
      return
    }

    getCourseProgress(courseId, token)
      .then(setProgress)
      .catch(() => setProgress([]))
  }, [courseId, user, token, enrolled])

  const handleEnroll = async () => {
    if (!user || !token) {
      navigate('/login')
      return
    }

    setEnrolling(true)
    setActionError('')

    try {
      await enrollCourse(courseId, token)
      setEnrolled(true)
    } catch (e) {
      setActionError(
        e instanceof Error
          ? e.message
          : 'Не удалось записаться на курс',
      )
    } finally {
      setEnrolling(false)
    }
  }

  const markComplete = async (pageId: number) => {
    if (!user || !token) {
      navigate('/login')
      return
    }

    if (progress.includes(pageId)) {
      return
    }

    setProgressLoading(true)
    setActionError('')

    try {
      await completePage(pageId, token)

      setProgress((old) =>
        old.includes(pageId)
          ? old
          : [...old, pageId],
      )
    } catch (e) {
      setActionError(
        e instanceof Error
          ? e.message
          : 'Не удалось сохранить прогресс',
      )
    } finally {
      setProgressLoading(false)
    }
  }

  const handlePageSelect = (pageId: number) => {
    setSelectedPage(pageId)
  }

  const handleTestPassed = (pageId: number) => {
    setProgress((old) =>
      old.includes(pageId)
        ? old
        : [...old, pageId],
    )
  }

  useEffect(() => {
    if (!course || !enrolled || selectedPage !== null) {
      return
    }

    const firstPage = course.sections
      .flatMap((section) => section.pages)[0]

    if (firstPage) {
      setSelectedPage(firstPage.id)
    }
  }, [course, enrolled, selectedPage])

  useEffect(() => {
    if (!course || !enrolled || selectedPage === null) {
      return
    }

    const page = course.sections
      .flatMap((section) => section.pages)
      .find((page) => page.id === selectedPage)

    if (!page || page.type !== 'theory') {
      return
    }

    if (progress.includes(page.id)) {
      return
    }

    void markComplete(page.id)
  }, [course, enrolled, selectedPage, progress])

  if (loading) {
    return (
      <main className="page">
        <div className="container">
          <div className="loading-screen">
            Загружаем курс...
          </div>
        </div>
      </main>
    )
  }

  if (error || !course) {
    return (
      <main className="page">
        <div className="container empty-page">
          <span className="eyebrow">ОШИБКА</span>

          <h1>Не удалось открыть курс</h1>

          <p>{error || 'Курс не найден.'}</p>

          <Link
            href="/courses"
            className="button button-primary"
          >
            Вернуться в каталог
          </Link>
        </div>
      </main>
    )
  }

  const pages = course.sections.flatMap(
    (section) => section.pages,
  )

  const currentPage =
    pages.find((page) => page.id === selectedPage) ||
    pages[0]

  const progressPercent = pages.length
    ? Math.round(
        (progress.filter((id) =>
          pages.some((page) => page.id === id),
        ).length /
          pages.length) *
          100,
      )
    : 0

  const isFree = course.price === 0

  if (accessLoading) {
    return (
      <main className="page">
        <div className="container">
          <div className="loading-screen">
            Проверяем доступ к курсу...
          </div>
        </div>
      </main>
    )
  }

  if (!enrolled) {
    return (
      <main className="page">
        <div className="container">
          <div className="course-detail-hero">
            <div>
              <span className="eyebrow">
                {course.theme || 'ОБУЧЕНИЕ'}
              </span>

              <h1>{course.title}</h1>

              <p>
                {course.description ||
                  'Описание курса пока не добавлено.'}
              </p>
            </div>

            <div className="course-detail-action">
              <strong>
                {formatCoursePrice(course.price)}
              </strong>

              {isFree ? (
                user ? (
                  <button
                    type="button"
                    className="button button-primary"
                    onClick={handleEnroll}
                    disabled={enrolling}
                  >
                    {enrolling
                      ? 'Записываем...'
                      : 'Записаться на курс'}
                  </button>
                ) : (
                  <Link
                    href="/login"
                    className="button button-primary"
                  >
                    Войти и записаться
                  </Link>
                )
              ) : (
                <span className="course-coming-soon">
                  Оплата появится позже
                </span>
              )}

              {actionError && (
                <div className="form-error">
                  {actionError}
                </div>
              )}
            </div>
          </div>
        </div>
      </main>
    )
  }

  return (
    <main className="page">
      <div className="container">
        <section className="course-learning">
          <aside className="course-sidebar">
            <div className="course-progress">
              <span>Прогресс</span>

              <strong>{progressPercent}%</strong>

              <div className="progress-bar">
                <span
                  style={{
                    width: `${progressPercent}%`,
                  }}
                />
              </div>
            </div>

            {course.sections.map((section) => (
              <div
                key={section.id}
                className="course-section"
              >
                <div className="course-section-title">
                  <span>{section.position + 1}</span>

                  <div>
                    <strong>{section.title}</strong>

                    {section.description && (
                      <small>
                        {section.description}
                      </small>
                    )}
                  </div>
                </div>

                <div className="course-page-list">
                  {section.pages.map((page) => (
                    <button
                      key={page.id}
                      type="button"
                      className={`course-page-button ${
                        selectedPage === page.id
                          ? 'active'
                          : ''
                      } ${
                        progress.includes(page.id)
                          ? 'completed'
                          : ''
                      }`}
                      onClick={() =>
                        handlePageSelect(page.id)
                      }
                    >
                      <span className="page-status">
                        {progress.includes(page.id)
                          ? '✓'
                          : page.type === 'test'
                            ? '◉'
                            : page.type === 'ai_test'
                              ? '✦'
                              : '○'}
                      </span>

                      <span>{page.title}</span>
                    </button>
                  ))}
                </div>
              </div>
            ))}
          </aside>

          <article className="lesson-panel">
            {currentPage ? (
              <>
                <div className="lesson-header">
                  <span className="lesson-type">
                    {currentPage.type === 'theory'
                      ? 'ТЕОРИЯ'
                      : currentPage.type === 'test'
                        ? 'ТЕСТ'
                        : 'AI-ТЕСТ'}
                  </span>

                  <h2>{currentPage.title}</h2>
                </div>

                {currentPage.type === 'theory' ? (
                  <div className="lesson-content">
                    <Markdown
                      content={
                        currentPage.content || ''
                      }
                    />
                  </div>
                ) : currentPage.type === 'test' ? (
                  <TestPage
                    pageId={currentPage.id}
                    token={token || ''}
                    onPassed={() =>
                      handleTestPassed(currentPage.id)
                    }
                  />
                ) : (
                  <div className="lesson-placeholder">
                    <div className="placeholder-icon">
                      ✦
                    </div>

                    <h3>Персональный AI-тест</h3>

                    <p>
                      AI-тест будет формироваться на
                      основе твоего прогресса и
                      результатов обучения.
                    </p>
                  </div>
                )}

                {currentPage.type === 'test' ? null : currentPage.type === 'theory' ? (
                  progress.includes(currentPage.id) ? (
                    <div className="lesson-completed">
                      ✓ Страница завершена
                    </div>
                  ) : (
                    <div className="lesson-completed">
                      Сохраняем прогресс...
                    </div>
                  )
                ) : progress.includes(currentPage.id) ? (
                  <div className="lesson-completed">
                    ✓ Страница завершена
                  </div>
                ) : (
                  <button
                    type="button"
                    className="button button-primary lesson-complete"
                    onClick={() =>
                      markComplete(currentPage.id)
                    }
                    disabled={progressLoading}
                  >
                    {progressLoading
                      ? 'Сохраняем...'
                      : 'Отметить как пройденное'}
                  </button>
                )}

                {currentPage.type !== 'test' && (
                  <div className="lesson-navigation">
                    <button
                      type="button"
                      className="button button-secondary"
                      onClick={() => {
                        const index = pages.findIndex(
                          (page) =>
                            page.id ===
                            currentPage.id,
                        )

                        if (index > 0) {
                          handlePageSelect(
                            pages[index - 1].id,
                          )
                        }
                      }}
                      disabled={
                        pages.findIndex(
                          (page) =>
                            page.id ===
                            currentPage.id,
                        ) <= 0
                      }
                    >
                      ← Назад
                    </button>

                    <button
                      type="button"
                      className="button button-primary"
                      onClick={() => {
                        const index = pages.findIndex(
                          (page) =>
                            page.id ===
                            currentPage.id,
                        )

                        if (
                          index <
                          pages.length - 1
                        ) {
                          handlePageSelect(
                            pages[index + 1].id,
                          )
                        }
                      }}
                      disabled={
                        pages.findIndex(
                          (page) =>
                            page.id ===
                            currentPage.id,
                        ) >=
                        pages.length - 1
                      }
                    >
                      Далее →
                    </button>
                  </div>
                )}

                {currentPage.type === 'test' && (
                  <div className="lesson-navigation">
                    <button
                      type="button"
                      className="button button-secondary"
                      onClick={() => {
                        const index = pages.findIndex(
                          (page) =>
                            page.id ===
                            currentPage.id,
                        )

                        if (index > 0) {
                          handlePageSelect(
                            pages[index - 1].id,
                          )
                        }
                      }}
                      disabled={
                        pages.findIndex(
                          (page) =>
                            page.id ===
                            currentPage.id,
                        ) <= 0
                      }
                    >
                      ← Назад
                    </button>

                    <button
                      type="button"
                      className="button button-primary"
                      onClick={() => {
                        const index = pages.findIndex(
                          (page) =>
                            page.id ===
                            currentPage.id,
                        )

                        if (
                          index <
                            pages.length - 1 &&
                          progress.includes(
                            currentPage.id,
                          )
                        ) {
                          handlePageSelect(
                            pages[index + 1].id,
                          )
                        }
                      }}
                      disabled={
                        !progress.includes(
                          currentPage.id,
                        ) ||
                        pages.findIndex(
                          (page) =>
                            page.id ===
                            currentPage.id,
                        ) >=
                          pages.length - 1
                      }
                    >
                      Далее →
                    </button>
                  </div>
                )}
              </>
            ) : (
              <div className="empty-state">
                <h3>
                  В этом курсе пока нет страниц
                </h3>
              </div>
            )}
          </article>
        </section>
      </div>
    </main>
  )
}