import { useEffect, useMemo, useState } from 'react'
import { createAdaptiveSession, getAdaptiveSession, submitAdaptiveAnswers, type AdaptiveSession } from '../../api/adaptiveTests'
import './AdaptiveTestPage.css'

type Props = { courseId: number; topicPageId: number; token: string; onCompleted: () => void }

const storageKey = (courseId: number, topicPageId: number) => `adaptive-test:${courseId}:${topicPageId}`

export default function AdaptiveTestPage({ courseId, topicPageId, token, onCompleted }: Props) {
  const [session, setSession] = useState<AdaptiveSession | null>(null)
  const [answers, setAnswers] = useState<Record<number, string>>({})
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  const round = useMemo(() => session?.rounds.find(r => r.round_number === session.current_round && r.status === 'generated') ?? null, [session])

  useEffect(() => {
    let cancelled = false

    const restore = async () => {
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

      if (data.status === 'completed') {
        localStorage.removeItem(storageKey(courseId, topicPageId))
        onCompleted()
      }
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
        <p>Пять вопросов по этой теме. После первого раунда следующий набор заданий изменится в зависимости от твоих ответов.</p>
        {error && <div className="form-error">{error}</div>}
        <button type="button" className="button button-primary" onClick={start} disabled={loading}>{loading ? 'Готовим тест...' : 'Начать AI-тест'}</button>
      </div>
    </div>
  )

  if (session.status === 'completed') return (
    <div className="adaptive-test-container">
      <div className="adaptive-feedback">
        <span className="adaptive-icon">✓</span>
        <h3>AI-тест завершён</h3>
        {session.feedback?.summary && <p>{session.feedback.summary}</p>}
        {session.feedback?.mastered_topics.length ? <section><h4>Что получилось</h4><ul>{session.feedback.mastered_topics.map((x, i) => <li key={i}><strong>{x.title}</strong><span>{x.reason}</span></li>)}</ul></section> : null}
        {session.feedback?.topics_to_review.length ? <section><h4>Что стоит повторить</h4><ul>{session.feedback.topics_to_review.map((x, i) => <li key={i}><strong>{x.title}</strong><span>{x.reason}</span></li>)}</ul></section> : null}
        {session.feedback?.next_steps.length ? <section><h4>Следующие шаги</h4><ul>{session.feedback.next_steps.map((x, i) => <li key={i}><strong>{x.title}</strong><span>{x.description}</span></li>)}</ul></section> : null}
      </div>
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

  return (
    <div className="adaptive-test-container">
      <div className="adaptive-meta">
        <div>
          <span>Раунд {session.current_round} из 2</span>
          <strong>{round.title || 'Адаптивный тест'}</strong>
        </div>
        <span>{answered} / {round.questions.length}</span>
      </div>

      {round.instructions && <p className="adaptive-instructions">{round.instructions}</p>}
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
          {submitting ? 'Проверяем...' : session.current_round === 1 ? 'Завершить раунд' : 'Завершить тест'}
        </button>
      </div>
    </div>
  )
}
