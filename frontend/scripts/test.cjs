const assert = require('node:assert/strict')
const { execFileSync } = require('node:child_process')
const fs = require('node:fs')
const path = require('node:path')
const React = require('react')
const { renderToStaticMarkup } = require('react-dom/server')

const root = path.resolve(__dirname, '..')
const output = path.join(root, 'node_modules/.cache/learner-ui-tests')
fs.mkdirSync(output, { recursive: true })
fs.writeFileSync(path.join(output, 'package.json'), '{"type":"commonjs"}')
execFileSync(process.execPath, [require.resolve('typescript/bin/tsc'),
  '--ignoreConfig', '--ignoreDeprecations', '6.0', '--module', 'commonjs', '--moduleResolution', 'node',
  '--jsx', 'react-jsx', '--target', 'es2022', '--lib', 'es2022,dom', '--esModuleInterop', '--skipLibCheck',
  '--outDir', output, '--rootDir', 'src',
  'src/components/markdown/Markdown.tsx', 'src/components/tests/AdaptiveResults.tsx', 'src/pages/CourseCompletionPage.tsx', 'src/components/admin/ResearchCharts.tsx',
], { cwd: root, stdio: 'inherit' })
const Markdown = require(path.join(output, 'components/markdown/Markdown.js')).default
const Results = require(path.join(output, 'components/tests/AdaptiveResults.js')).default
const { CourseCompletionContent } = require(path.join(output, 'pages/CourseCompletionPage.js'))
const { GainChart, ScoreScatter, ParticipationChart } = require(path.join(output, 'components/admin/ResearchCharts.js'))
const render = (component, props) => renderToStaticMarkup(React.createElement(component, props))
let count = 0
function test(name, body) { body(); count++; console.log(`PASS ${name}`) }

test('GFM tables support header, body, alignment and inline emphasis', () => {
  const html = render(Markdown, { content: '| Угол | sin x |\n| :--- | ---: |\n| 0 | **0** |\n| π/2 | 1 |' })
  assert.match(html, /<table>/)
  assert.equal((html.match(/<th /g) ?? []).length, 2)
  assert.match(html, /text-align:right/)
  assert.match(html, /<strong>0<\/strong>/)
  assert.equal((html.match(/<tr>/g) ?? []).length, 3)
})
test('multiple tables and optional outer pipes are rendered separately', () => {
  const html = render(Markdown, { content: 'Угол | cos x\n--- | ---\n0 | 1\n\nТекст между таблицами\n\n| Угол | sin x |\n| --- | --- |\n| 0 | 0 |' })
  assert.equal((html.match(/<table>/g) ?? []).length, 2)
  assert.match(html, /<p>Текст между таблицами<\/p>/)
})
test('escaped pipes and inline code stay in the same cell', () => {
  const html = render(Markdown, { content: '| `a|b` | a\\|b |\n| --- | --- |\n| 1 | 2 |' })
  assert.equal((html.match(/<th /g) ?? []).length, 2)
  assert.match(html, /<code>a\|b<\/code>/)
})
test('code fences containing table syntax stay code, including an unclosed fence', () => {
  for (const ending of ['', '\n```']) {
    const html = render(Markdown, { content: '```text\n| x | y |\n| --- | --- |\n| 1 | 2 |' + ending })
    assert.doesNotMatch(html, /<table>/)
    assert.match(html, /<pre>/)
    assert.match(html, /\| x \| y \|/)
  }
})
test('feedback is readable, separate review scores are accurate, empty cards are hidden', () => {
  const html = render(Results, { session: { rounds: [
    { id: 1, status: 'completed', purpose: 'prerequisite_review', answered_count: 5, correct_count: 2 },
    { id: 2, status: 'completed', purpose: 'practice', answered_count: 5, correct_count: 4 },
  ], feedback: { summary: 'Практика завершена.', mastered_topics: [{ title: 'Дискриминант', reason: 'Все ответы верны.' }], topics_to_review: [{ title: '', reason: '' }], next_steps: [] } } })
  for (const text of ['Повторение основ', 'Основная практика', '40%', '80%', 'Дискриминант', 'Все ответы верны.']) assert.ok(html.includes(text), text)
  assert.doesNotMatch(html, /Что стоит повторить|Следующие шаги|Твои навыки/)
})
test('legacy English feedback does not appear as a Russian result', () => {
  const html = render(Results, { session: { rounds: [], feedback: { summary: 'Round 1 was strong', mastered_topics: [], topics_to_review: [], next_steps: [] } } })
  assert.doesNotMatch(html, /Round 1/)
  assert.match(html, /Ты завершил практику/)
})
test('legacy null feedback groups do not crash the result screen', () => {
  const html = render(Results, { session: { rounds: [], feedback: { summary: 'Практика завершена.', mastered_topics: null, topics_to_review: null, next_steps: null } } })
  assert.match(html, /Практика завершена/)
  assert.doesNotMatch(html, /Что получилось|Что стоит повторить|Следующие шаги/)
})
test('course completion thanks the account user and safely escapes their name', () => {
  const html = render(CourseCompletionContent, { username: 'Аня <script>', courseTitle: 'Математика', courseId: 42, pageCount: 12 })
  assert.match(html, /Спасибо, что прошёл наш курс, Аня &lt;script&gt;!/)
  assert.match(html, /Пройдено страниц: 12/)
  assert.match(html, /href="\/courses\/42"/)
  assert.doesNotMatch(html, /<script>/)
})
test('research charts do not invent zeros for unavailable comparisons', () => {
  assert.match(render(GainChart, { groups: [{ arm: 'control', gain: { mean: null } }] }), /нужны полные сопоставимые пары/)
  assert.match(render(ScoreScatter, { participants: [{ eligible: false, pre_score: 0, post_score: null }] }), /Нет полных пар/)
})
test('research plots retain legitimate zero scores, negative change and full denominators', () => {
  const gain = render(GainChart, { groups: [{ arm: 'control', paired: 3, gain: { mean: -10, ci_low: -30, ci_high: 10 } }] })
  assert.match(gain, /-10/)
  assert.match(gain, /Полных пар: 3/)
  const scatter = render(ScoreScatter, { participants: [{ participant_id: 'p', arm: 'adaptive', eligible: true, pre_score: 0, post_score: 50, gain: 50 }] })
  assert.match(scatter, /<circle/)
  assert.match(scatter, /входной 0%/)
  const participation = render(ParticipationChart, { groups: [{ arm: 'control', assigned: 10, with_pre: 8, with_post: 5, paired: 4 }] })
  assert.match(participation, /Назначено/)
  assert.match(participation, />10<\/strong>/)
  assert.match(participation, />4<\/strong>/)
})
console.log(`${count} learner UI checks passed`)
