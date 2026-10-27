import { useEffect, useMemo, useState } from 'react'
import { createAdaptiveSession, getAdaptiveSession, submitAdaptiveAnswers, type AdaptiveSession } from '../../api/adaptiveTests'
import './AdaptiveTestPage.css'
import AdaptiveResults from './AdaptiveResults'
import { useScrollToLesson } from '../../utils/learningScroll'

type Props = { courseId: number; topicPageId: number; token: string; onCompleted: () => void }

const storageKey = (courseId: number, topicPageId: number) => `adaptive-test:${courseId}:${topicPageId}`

export default function AdaptiveTestPage({ courseId, topicPageId, token, onCompleted }: Props) {
  const [session, setSession] = useState<AdaptiveSession | null>(null)
  const [answers, setAnswers] = useState<Record<number, string>>({})
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  const round = useMemo(() => session?.rounds.find(r => r.round_number === session.current_round && r.status === 'generated') ?? null, [session])

  useScrollToLesson(session ? `${session.id}:${session.current_round}:${session.status}` : null, session !== null && !loading)

  useEffect(() => {
    let cancelled = false

    const restore = async () => {
      setSession(null)
      setAnswers({})
      setLoading(true)
      setError('')

      const savedSessionID = localStorage.getItem(storageKey(courseId, topicPageId))
      if (!savedSessionID || !token) {
        if (!cancelled) setLoading(false)
        return
      }

      try {
        const restored = await getAdaptiveSession(savedSessionID, token)
        if (cancelled) return

        if (restored.status === 'failed') {
          localStorage.removeItem(storageKey(courseId, topicPageId))
          setError('Предыдущий AI-тест завершился с ошибкой. Можно начать новый.')
        } else {
          setSession(restored)
          if (restored.status === 'completed') onCompleted()
        }
      } catch {
        localStorage.removeItem(storageKey(courseId, topicPageId))
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    void restore()

    return () => {
      cancelled = true
    }
  }, [courseId, topicPageId, token])

  const start = async () => {
    setLoading(true)
    setError('')
    setAnswers({})

    try {
      const created = await createAdaptiveSession(courseId, topicPageId, 5, token)
      localStorage.setItem(storageKey(courseId, topicPageId), created.id)
      setSession(created)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось создать AI-тест')
    } finally {
      setLoading(false)
    }
  }

  const submit = async () => {
    if (!session || !round || submitting) return

    if (Object.keys(answers).length !== round.questions.length) {
      setError('Ответь на все вопросы перед отправкой.')
      return
    }

    setSubmitting(true)
    setError('')

    try {
      const data = await submitAdaptiveAnswers(session.id, round.id, answers, token)
      setSession(data)
      setAnswers({})

      if (data.status === 'completed') onCompleted()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось отправить ответы')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return <div className="adaptive-loading">Загружаем AI-тест...</div>
  }

  if (!session) return (
    <div className="adaptive-test-container">
      <div className="adaptive-intro">
        <span className="adaptive-icon">✦</span>
        <h3>Персональный AI-тест</h3>
        <p>Два коротких раунда по пять вопросов. Если перед этой темой нужно освежить основы, сначала будет отдельное повторение, затем — основная практика.</p>
        {error && <div className="form-error">{error}</div>}
        <button type="button" className="button button-primary" onClick={start} disabled={loading}>{loading ? 'Готовим тест...' : 'Начать AI-тест'}</button>
      </div>
    </div>
  )

  if (session.status === 'completed') return (
    <div className="adaptive-test-container">
      <AdaptiveResults session={session} />
    </div>
  )

  if (!round) return (
    <div className="adaptive-test-container">
      <div className="adaptive-feedback">
        <h3>Не удалось найти активный раунд</h3>
        <p>Сессия больше не содержит доступного раунда.</p>
        <button type="button" className="button button-primary" onClick={() => {
          localStorage.removeItem(storageKey(courseId, topicPageId))
          setSession(null)
          setAnswers({})
          setError('')
        }}>
          Начать новый тест
        </button>
      </div>
    </div>
  )

  const answered = Object.keys(answers).length
  const reviewRound = round.purpose === 'prerequisite_review'

  return (
    <div className="adaptive-test-container">
      <div className="adaptive-meta">
        <div>
          <span>Раунд {session.current_round} из 2</span>
          <strong>{reviewRound ? 'Повторение основ' : round.purpose === 'mixed' ? 'Практика с повторением' : 'Основная практика'}</strong>
        </div>
        <span>{answered} / {round.questions.length}</span>
      </div>

      {reviewRound ? (
        <div className="adaptive-review-notice" role="note">
          <strong>Сначала освежим основы</strong>
          <p>В предыдущих заданиях эти темы вызвали трудности. Это отдельный раунд повторения: после него вернёмся к теме текущего раздела.</p>
        </div>
      ) : <p className="adaptive-instructions">Решаем задания по теме текущего раздела. Выбери один ответ в каждом вопросе.</p>}
      {error && <div className="form-error">{error}</div>}

      <div className="adaptive-questions">
        {round.questions.slice().sort((a, b) => a.position - b.position).map((q, i) => (
          <section key={q.id} className="adaptive-question">
            <div className="adaptive-question-number">Вопрос {i + 1}{q.topic_title ? ` · ${q.topic_title}` : ''}</div>
            <h3>{q.question}</h3>

            <div className="adaptive-options">
              {q.options.slice().sort((a, b) => a.position - b.position).map(option => {
                const selected = answers[q.id] === option.key

                return (
                  <label key={option.key} className={`adaptive-option ${selected ? 'selected' : ''}`}>
                    <input
                      type="radio"
                      name={`adaptive-question-${q.id}`}
                      checked={selected}
                      disabled={submitting}
                      onChange={() => setAnswers(current => ({ ...current, [q.id]: option.key }))}
                    />
                    <span className="adaptive-option-marker" />
                    <span>{option.text}</span>
                  </label>
                )
              })}
            </div>
          </section>
        ))}
      </div>

      <div className="adaptive-submit">
        <span>{answered === round.questions.length ? 'Все вопросы отвечены' : 'Ответь на все вопросы'}</span>
        <button
          type="button"
          className="button button-primary"
          onClick={submit}
          disabled={submitting || answered !== round.questions.length}
        >
          {submitting ? 'Проверяем...' : session.current_round === 1 ? reviewRound ? 'Перейти к основной практике' : 'Завершить раунд' : 'Завершить тест'}
        </button>
      </div>
    </div>
  )
}
