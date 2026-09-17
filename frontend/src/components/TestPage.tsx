import { useEffect, useMemo, useState } from 'react'
import { getTest, submitTest, type CourseTest, type TestResult } from '../api/courses'
import './TestPage.css'

type TestPageProps = {
  pageId: number
  token: string
  onPassed: () => void
}

export default function TestPage({ pageId, token, onPassed }: TestPageProps) {
  const [test, setTest] = useState<CourseTest | null>(null)
  const [answers, setAnswers] = useState<Record<number, number>>({})
  const [result, setResult] = useState<TestResult | null>(null)
  const [locked, setLocked] = useState(false)
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError('')
    setResult(null)
    setAnswers({})
    setLocked(false)

    getTest(pageId, token)
      .then((data) => {
        if (cancelled) return
        setTest(data)
        if (data.completed) {
          setLocked(true)
          setAnswers(Object.fromEntries((data.answers ?? []).map((answer) => [answer.question_id, answer.answer_id])))
          setResult({
            attempt_id: data.attempt_id ?? 0,
            score: data.score ?? 0,
            passed: true,
            passing_score: data.passing_score,
            answers: data.answers ?? [],
          })
          onPassed()
        }
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Не удалось загрузить тест')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => { cancelled = true }
  }, [pageId, token])

  const questionCount = test?.questions.length ?? 0
  const answeredCount = Object.keys(answers).length
  const allAnswered = questionCount > 0 && answeredCount === questionCount
  const orderedQuestions = useMemo(() => test ? [...test.questions].sort((a, b) => a.position - b.position) : [], [test])
  const submittedByQuestion = useMemo(() => new Map((result?.answers ?? []).map((answer) => [answer.question_id, answer])), [result])
  const hasSubmittedAnswers = submittedByQuestion.size > 0

  const handleAnswerChange = (questionId: number, answerId: number) => {
    if (locked) return
    setResult(null)
    setError('')
    setAnswers((current) => ({ ...current, [questionId]: answerId }))
  }

  const handleSubmit = async () => {
    if (!test || !allAnswered || submitting || locked) return
    setSubmitting(true)
    setError('')
    try {
      const submission = orderedQuestions.map((question) => ({ question_id: question.id, answer_id: answers[question.id] }))
      const data = await submitTest(test.id, submission, token)
      setResult(data)
      if (data.passed) {
        setLocked(true)
        onPassed()
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось отправить тест')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) return <div className="test-loading">Загружаем тест...</div>
  if (error && !test) return <div className="test-error"><div className="test-result-icon">!</div><h3>Не удалось загрузить тест</h3><p>{error}</p></div>
  if (!test || questionCount === 0) return <div className="test-error"><div className="test-result-icon">!</div><h3>В тесте пока нет вопросов</h3><p>Попробуй открыть эту страницу позже.</p></div>

  return <div className="test-container">
    <div className="test-meta">
      <span>Вопросов: {questionCount}</span>
      <strong>{locked ? `Результат: ${result?.score ?? 0}%` : `${answeredCount} / ${questionCount} отвечено`}</strong>
    </div>

    {locked && <div className="test-completed-note">Тест уже пройден. Ответы доступны только для просмотра.</div>}
    {result && !locked && <div className={`test-result-summary ${result.passed ? 'passed' : 'failed'}`}>
      <strong>{result.passed ? 'Тест пройден' : 'Тест не пройден'}</strong>
      <span>{result.score}% · для прохождения нужно {result.passing_score}%</span>
    </div>}
    {error && <div className="form-error test-form-error">{error}</div>}

    <div className="test-questions">
      {orderedQuestions.map((question, index) => {
        const selectedAnswer = answers[question.id]
        const submitted = submittedByQuestion.get(question.id)
        const orderedAnswers = [...question.answers].sort((a, b) => a.position - b.position)

        return <section key={question.id} className="test-question">
          <div className="test-question-number">Вопрос {index + 1}</div>
          <h3>{question.question}</h3>
          <div className="test-answers">
            {orderedAnswers.map((answer) => {
              const isSelected = selectedAnswer === answer.id
              const isCorrect = hasSubmittedAnswers && submitted?.correct_answer_id === answer.id
              const isWrongSelected = hasSubmittedAnswers && isSelected && submitted?.is_correct === false
              const stateClass = isCorrect ? 'correct' : isWrongSelected ? 'wrong' : ''

              return <label key={answer.id} className={`test-answer ${isSelected ? 'selected' : ''} ${stateClass} ${locked ? 'locked' : ''}`}>
                <input type="radio" name={`question-${question.id}`} value={answer.id} checked={isSelected} disabled={locked || submitting} onChange={() => handleAnswerChange(question.id, answer.id)} />
                <span className="test-answer-marker" />
                <span>{answer.text}</span>
              </label>
            })}
          </div>
        </section>
      })}
    </div>

    {!locked && <div className="test-submit">
      <div><strong>{answeredCount} из {questionCount}</strong><span>{allAnswered ? 'Все вопросы отвечены' : 'Ответь на все вопросы'}</span></div>
      <button type="button" className="button button-primary" onClick={handleSubmit} disabled={!allAnswered || submitting}>{submitting ? 'Проверяем...' : result ? 'Отправить снова' : 'Завершить тест'}</button>
    </div>}
  </div>
}
