import { useEffect, useMemo, useState } from 'react'
import { getTest, submitTest, type CourseTest, type TestResult } from '../api/courses'

type TestPageProps = {
  pageId: number
  token: string
  onPassed: () => void
}

export default function TestPage({ pageId, token, onPassed }: TestPageProps) {
  const [test, setTest] = useState<CourseTest | null>(null)
  const [answers, setAnswers] = useState<Record<number, number>>({})
  const [result, setResult] = useState<TestResult | null>(null)
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError('')
    setResult(null)
    setAnswers({})

    getTest(pageId, token)
      .then((data) => {
        if (!cancelled) setTest(data)
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Не удалось загрузить тест')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => { cancelled = true }
  }, [pageId, token])

  const answeredCount = Object.keys(answers).length
  const questionCount = test?.questions.length ?? 0
  const allAnswered = questionCount > 0 && answeredCount === questionCount

  const orderedQuestions = useMemo(
    () => test ? [...test.questions].sort((a, b) => a.position - b.position) : [],
    [test],
  )

  const handleSubmit = async () => {
    if (!test || !allAnswered || submitting) return

    setSubmitting(true)
    setError('')

    try {
      const submission = orderedQuestions.map((question) => ({
        question_id: question.id,
        answer_id: answers[question.id],
      }))
      const data = await submitTest(test.id, submission, token)
      setResult(data)
      if (data.passed) onPassed()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось отправить тест')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) return <div className="test-loading">Загружаем тест...</div>

  if (error && !test) {
    return (
      <div className="test-error">
        <div className="test-result-icon">!</div>
        <h3>Не удалось загрузить тест</h3>
        <p>{error}</p>
      </div>
    )
  }

  if (!test || questionCount === 0) {
    return (
      <div className="test-error">
        <div className="test-result-icon">!</div>
        <h3>В тесте пока нет вопросов</h3>
        <p>Попробуй открыть эту страницу позже.</p>
      </div>
    )
  }

  if (result) {
    return (
      <div className={`test-result ${result.passed ? 'passed' : 'failed'}`}>
        <div className="test-result-icon">{result.passed ? '✓' : '×'}</div>
        <span className="test-result-label">РЕЗУЛЬТАТ</span>
        <div className="test-score">{result.score}%</div>
        <h3>{result.passed ? 'Тест пройден' : 'Тест не пройден'}</h3>
        <p>Для прохождения нужно набрать не менее {result.passing_score}%.</p>
        {!result.passed && (
          <button type="button" className="button button-primary" onClick={() => { setResult(null); setAnswers({}); setError('') }}>
            Попробовать снова
          </button>
        )}
      </div>
    )
  }

  return (
    <div className="test-container">
      <div className="test-meta">
        <span>Вопросов: {questionCount}</span>
        <strong>{answeredCount} / {questionCount} отвечено</strong>
      </div>

      {error && <div className="form-error test-form-error">{error}</div>}

      <div className="test-questions">
        {orderedQuestions.map((question, index) => {
          const selectedAnswer = answers[question.id]
          const orderedAnswers = [...question.answers].sort((a, b) => a.position - b.position)

          return (
            <section key={question.id} className="test-question">
              <div className="test-question-number">Вопрос {index + 1}</div>
              <h3>{question.question}</h3>
              <div className="test-answers">
                {orderedAnswers.map((answer) => (
                  <label key={answer.id} className={`test-answer ${selectedAnswer === answer.id ? 'selected' : ''}`}>
                    <input
                      type="radio"
                      name={`question-${question.id}`}
                      value={answer.id}
                      checked={selectedAnswer === answer.id}
                      onChange={() => setAnswers((current) => ({ ...current, [question.id]: answer.id }))}
                    />
                    <span className="test-answer-marker" />
                    <span>{answer.text}</span>
                  </label>
                ))}
              </div>
            </section>
          )
        })}
      </div>

      <div className="test-submit">
        <div>
          <strong>{answeredCount} из {questionCount}</strong>
          <span>{allAnswered ? 'Все вопросы отвечены' : 'Ответь на все вопросы'}</span>
        </div>
        <button type="button" className="button button-primary" onClick={handleSubmit} disabled={!allAnswered || submitting}>
          {submitting ? 'Проверяем...' : 'Завершить тест'}
        </button>
      </div>
    </div>
  )
}
