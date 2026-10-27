"""Chat draft regression with synthetic API; no LLM calls or database changes."""
import os
from urllib.parse import urlparse
from playwright.sync_api import sync_playwright, expect

BASE = os.environ.get('LEARNER_UI_URL', 'http://127.0.0.1:5173')


def main():
    calls = []
    errors = []
    state = {'fail': True, 'chats': [{'id': 1, 'chatName': 'История'}]}
    def api(route):
        req = route.request
        path = urlparse(req.url).path
        if not path.startswith('/api/'):
            route.continue_()
            return
        if path == '/api/auth/me': body = {'id': 7, 'username': 'Анна', 'email': 'test@example.test', 'role': 'student'}
        elif path == '/api/chats' and req.method == 'GET': body = state['chats']
        elif path == '/api/chats' and req.method == 'POST':
            calls.append(path)
            if state['fail']:
                state['fail'] = False
                route.fulfill(status=503, body='Попробуйте ещё раз')
                return
            draft = {'id': 2, 'chatName': 'Новый чат'}
            state['chats'] = [draft, {'id': 1, 'chatName': 'История'}]
            body = draft
        elif path == '/api/chats/1/messages': body = [{'id': 10, 'content': 'Начатая переписка', 'isAiResponse': False, 'timestamp': '2026-10-27T10:00:00Z'}]
        elif path == '/api/chats/2/messages': body = []
        else:
            errors.append(path)
            route.fulfill(status=500, body='Unexpected API')
            return
        route.fulfill(status=200, json=body)
    with sync_playwright() as p:
        browser = p.chromium.launch(executable_path=os.environ.get('CHROMIUM_PATH', '/usr/bin/chromium'), headless=True, args=['--no-sandbox'])
        context = browser.new_context()
        context.add_init_script("localStorage.setItem('cringearium_token','test')")
        context.route('**/api/**', api)
        page = context.new_page()
        page.on('pageerror', lambda error: errors.append(str(error)))
        page.goto(BASE + '/chat')
        expect(page.get_by_text('Начатая переписка', exact=True)).to_be_visible()
        new = page.locator('.chat-new')
        expect(new).to_be_enabled()
        new.click()
        expect(page.get_by_role('alert')).to_contain_text('Попробуйте ещё раз')
        expect(new).to_be_enabled()
        before = len(calls)
        page.evaluate("for(let i=0;i<30;i++) document.querySelector('.chat-new').click()")
        expect(page.locator('.chat-item')).to_have_count(2)
        expect(page.locator('.chat-item-wrapper.active')).to_contain_text('Новый чат')
        expect(new).to_be_enabled()
        assert len(calls) == before + 1, calls
        page.get_by_role('textbox', name='Сообщение', exact=True).fill('Не потерять черновик')
        for _ in range(8): new.click()
        assert len(calls) == before + 1
        expect(page.get_by_role('textbox', name='Сообщение', exact=True)).to_have_value('Не потерять черновик')
        page.locator('.chat-item').filter(has_text='История').click()
        expect(page.get_by_text('Начатая переписка', exact=True)).to_be_visible()
        expect(new).to_be_enabled()
        new.click()
        expect(page.locator('.chat-item-wrapper.active')).to_contain_text('Новый чат')
        expect(page.locator('.chat-item')).to_have_count(2)
        page.reload()
        expect(page.locator('.chat-item-wrapper.active')).to_contain_text('Новый чат')
        expect(new).to_be_enabled()
        before = len(calls)
        new.click()
        assert len(calls) == before
        assert not errors, errors
        browser.close()
    print('PASS chat drafts: failed creation retry, rapid-click guard, empty draft reuse, input preservation, server-reused ID deduplication, reload')


if __name__ == '__main__':
    main()
