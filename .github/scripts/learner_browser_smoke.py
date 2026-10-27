"""Optional browser regression check. Run Vite first; requires Playwright and Chromium.
Uses mocked HTTP responses, so it never sends requests to a real LLM or changes a database.
"""
import copy
import os
from urllib.parse import urlparse
from playwright.sync_api import sync_playwright, expect

BASE_URL = os.environ.get('LEARNER_UI_URL', 'http://127.0.0.1:5173')
VIEWPORT_WIDTH = int(os.environ.get('LEARNER_VIEWPORT_WIDTH', '1280'))
TABLES = '| Угол | sin x |\n| --- | --- |\n| 0 | 0 |\n| π/2 | 1 |\n\nПериоды функций\n\n| Функция | Период |\n| --- | --- |\n| sin x | 2π |\n| tg x | π |'
COURSE = {'id': 42, 'title': 'Математика', 'price': 0, 'status': 'published', 'sections': [{'id': 1, 'title': 'Квадратные уравнения', 'position': 0, 'description': '', 'pages': [
    {'id': 1, 'title': 'Таблицы тригонометрии', 'type': 'theory', 'content': TABLES + ('\n\nОбъяснение темы и пример решения. ' * 70), 'position': 0},
    {'id': 2, 'title': 'Практика: квадратные уравнения', 'type': 'ai_test', 'content': '', 'position': 1},
    {'id': 3, 'title': 'Итоговый тест: весь курс', 'type': 'test', 'content': '', 'position': 2},
]}]}


def make_round(number, review):
    return {'id': 10 + number, 'round_number': number, 'purpose': 'prerequisite_review' if review else 'practice', 'status': 'generated', 'title': 'An English generated title', 'instructions': 'English generated instructions', 'questions': [
        {'id': number * 10 + i, 'position': i, 'question': f'Реши задание {i}', 'topic_title': 'Линейные уравнения' if review else 'Квадратные уравнения', 'difficulty': 2, 'knowledge_basis': 'course', 'options': [{'key': 'A', 'text': 'Первый ответ', 'position': 0}, {'key': 'B', 'text': 'Второй ответ', 'position': 1}]} for i in range(1, 6)
    ]}


def main():
    state = {'progress': [1], 'session': None, 'test_done': False, 'saved_ai': False}
    errors = []

    def api(route):
        request = route.request
        path = urlparse(request.url).path
        if not path.startswith('/api/'):
            route.continue_()
            return
        body = None
        if path == '/api/auth/me':
            body = {'id': 7, 'username': 'Анна', 'email': 'browser@example.test', 'role': 'student'}
        elif path == '/api/courses/42':
            body = COURSE
        elif path == '/api/courses/42/access':
            body = {'has_access': True}
        elif path == '/api/courses/42/progress':
            body = {'completed_page_ids': state['progress']}
        elif path == '/api/pages/2/complete':
            state['progress'] = sorted(set(state['progress'] + [2]))
            state['saved_ai'] = True
            route.fulfill(status=200, body='')
            return
        elif path == '/api/pages/1/complete':
            route.fulfill(status=200, body='')
            return
        elif path == '/api/adaptive-tests' and request.method == 'POST':
            state['session'] = {'id': 'browser-session', 'course_id': 42, 'topic_page_id': 2, 'status': 'in_progress', 'current_round': 1, 'question_count': 5, 'rounds': [make_round(1, True)]}
            body = state['session']
        elif path == '/api/adaptive-tests/browser-session/answers':
            session = state['session']
            current = session['rounds'][-1]
            current.update(status='completed', answered_count=5, correct_count=2 if session['current_round'] == 1 else 4)
            if session['current_round'] == 1:
                session['current_round'] = 2
                session['rounds'].append(make_round(2, False))
            else:
                session['status'] = 'completed'
                session['feedback'] = {'summary': 'Практика завершена. Основы повторены, можно двигаться дальше.', 'mastered_topics': [{'title': 'Дискриминант', 'reason': 'Верных ответов: 4 из 5.'}], 'topics_to_review': [], 'next_steps': []}
            body = session
        elif path == '/api/adaptive-tests/browser-session':
            body = state['session']
        elif path == '/api/pages/3/test':
            body = {'id': 900, 'page_id': 3, 'passing_score': 50, 'completed': state['test_done'], 'score': 100 if state['test_done'] else None, 'questions': [{'id': 901, 'question': 'Реши уравнение 2x = 4', 'position': 0, 'answers': [{'id': 902, 'text': 'x = 2', 'position': 0}, {'id': 903, 'text': 'x = 3', 'position': 1}]}]}
        elif path == '/api/tests/900/attempts':
            correct = request.post_data_json['answers'][0]['answer_id'] == 902
            state['test_done'] = correct
            if correct: state['progress'] = [1, 2, 3]
            body = {'attempt_id': 2 if correct else 1, 'score': 100 if correct else 0, 'passed': correct, 'passing_score': 50, 'answers': [{'question_id': 901, 'answer_id': 902 if correct else 903, 'correct_answer_id': 902, 'is_correct': correct}]}
        else:
            errors.append(f'Unexpected API request: {request.method} {path}')
            route.fulfill(status=500, body='Unexpected test request')
            return
        route.fulfill(status=200, json=copy.deepcopy(body))

    with sync_playwright() as p:
        browser = p.chromium.launch(executable_path=os.environ.get('CHROMIUM_PATH', '/usr/bin/chromium'), headless=True, args=['--no-sandbox'])
        context = browser.new_context(viewport={'width': VIEWPORT_WIDTH, 'height': 900})
        context.add_init_script("localStorage.setItem('cringearium_token', 'browser-test-token')")
        context.route('**/api/**', api)
        page = context.new_page()
        page.on('pageerror', lambda error: errors.append(str(error)))
        def at_lesson_top():
            page.wait_for_function("""() => {
                const heading = document.querySelector('.lesson-header').getBoundingClientRect();
                const header = document.querySelector('.header').getBoundingClientRect();
                return heading.top >= header.bottom - 1 && heading.top <= header.bottom + 24;
            }""")
        def scroll_down():
            page.evaluate('window.scrollTo({top: document.body.scrollHeight, behavior: "instant"})')
            assert page.evaluate('window.scrollY > 100')
        page.goto(BASE_URL + '/courses/42')
        try:
            expect(page.locator('.markdown-content table')).to_have_count(2)
        except Exception:
            print('Browser diagnostics:', errors, page.locator('body').inner_text()[:1800])
            raise
        page.set_viewport_size({'width': 390, 'height': 844})
        expect(page.locator('.markdown-content table')).to_have_count(2)
        assert page.evaluate('document.documentElement.scrollWidth <= innerWidth + 1'), 'Table overflowed the mobile viewport'
        page.evaluate('window.scrollTo({top: 0, behavior: "instant"})')
        assert page.evaluate("Math.abs(document.querySelector('.course-learning').getBoundingClientRect().top - document.querySelector('.header').getBoundingClientRect().bottom) <= 1"), 'Gap above course on mobile'
        # On mobile the lesson must appear below the sticky header, past the course menu.
        scroll_down()
        page.get_by_role('button', name='Практика: квадратные уравнения').click()
        at_lesson_top()
        page.get_by_role('button', name='Таблицы тригонометрии').click()
        at_lesson_top()
        page.set_viewport_size({'width': VIEWPORT_WIDTH, 'height': 900})
        scroll_down()
        page.get_by_role('button', name='Далее →', exact=True).click()
        at_lesson_top()
        expect(page.get_by_text('Твои навыки', exact=True)).to_have_count(0)
        page.get_by_role('button', name='Начать AI-тест').click()
        expect(page.get_by_text('Сначала освежим основы')).to_be_visible()
        expect(page.locator('.adaptive-meta')).to_contain_text('Повторение основ')
        at_lesson_top()
        for question in page.locator('.adaptive-question').all():
            question.locator('label').first.click()
        page.get_by_role('button', name='Перейти к основной практике').click()
        expect(page.locator('.adaptive-meta')).to_contain_text('Основная практика')
        at_lesson_top()
        expect(page.locator('.adaptive-review-notice')).to_have_count(0)
        assert 'Линейные' not in page.locator('.adaptive-questions').inner_text()
        for question in page.locator('.adaptive-question').all():
            question.locator('label').first.click()
        page.get_by_role('button', name='Завершить тест', exact=True).click()
        expect(page.get_by_role('heading', name='Практика завершена')).to_be_visible()
        at_lesson_top()
        expect(page.get_by_text('40%', exact=True)).to_be_visible()
        expect(page.get_by_text('80%', exact=True)).to_be_visible()
        expect(page.get_by_text('Верных ответов: 4 из 5.', exact=True)).to_be_visible()
        expect(page.get_by_text('English generated instructions', exact=True)).to_have_count(0)
        page.reload()
        expect(page.get_by_role('button', name='Практика: квадратные уравнения')).to_be_visible()
        page.get_by_role('button', name='Практика: квадратные уравнения').click()
        expect(page.get_by_role('heading', name='Практика завершена')).to_be_visible()
        at_lesson_top()
        assert state['saved_ai'], 'AI completion was not persisted'
        page.get_by_role('button', name='Далее →', exact=True).click()
        at_lesson_top()
        page.locator('.test-answers label').last.click()
        scroll_down()
        page.get_by_role('button', name='Завершить тест', exact=True).click()
        expect(page.locator('.test-result-summary')).to_contain_text('Тест не пройден')
        at_lesson_top()
        page.locator('.test-answers label').first.click()
        scroll_down()
        page.get_by_role('button', name='Завершить тест', exact=True).click()
        expect(page.locator('.test-meta')).to_contain_text('Результат: 100%')
        at_lesson_top()
        assert page.evaluate("Math.abs(document.querySelector('.course-learning').getBoundingClientRect().bottom - document.querySelector('.footer').getBoundingClientRect().top) <= 1"), 'Gap before footer'
        finish = page.get_by_role('button', name='Завершить курс →')
        expect(finish).to_be_enabled()
        finish.click()
        expect(page.get_by_role('heading', name='Спасибо, что прошёл наш курс, Анна!')).to_be_visible()
        page.reload()
        expect(page.get_by_role('heading', name='Спасибо, что прошёл наш курс, Анна!')).to_be_visible()
        page.screenshot(path='/tmp/learner-course-completion.png', full_page=True)
        state['progress'] = [1]
        page.reload()
        expect(page.get_by_role('heading', name='До встречи на финише!')).to_be_visible()
        assert not errors, errors
        browser.close()
    print('PASS learner browser flow: scroll on lesson/round/result transitions, no footer gap, tables/mobile, separate review, Russian results, saved AI progress, final-test navigation, named completion, reload, incomplete-course guard')


if __name__ == '__main__':
    main()
