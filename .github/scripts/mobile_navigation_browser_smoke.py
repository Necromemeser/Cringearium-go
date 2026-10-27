"""Optional navigation regression check with mocked API; no database writes or LLM calls."""
import os
import sys
sys.dont_write_bytecode = True
from urllib.parse import urlparse
from playwright.sync_api import sync_playwright, expect
from learner_browser_smoke import COURSE

BASE = os.environ.get('LEARNER_UI_URL', 'http://127.0.0.1:5173')


def main():
    calls = []
    errors = []
    def api(route):
        req = route.request
        path = urlparse(req.url).path
        if not path.startswith('/api/'):
            route.continue_()
            return
        calls.append(path)
        if path == '/api/auth/me':
            if req.headers.get('authorization') == 'Bearer expired':
                route.fulfill(status=401, body='Expired')
                return
            body = {'id': 7, 'username': 'ОченьДлинноеИмяПользователяДляПроверкиШапки', 'email': 'test@example.test', 'role': 'student'}
        elif path == '/api/auth/login': body = {'token': 'valid'}
        elif path == '/api/chats': body = []
        elif path == '/api/courses': body = []
        elif path == '/api/courses/42': body = COURSE
        elif path == '/api/courses/42/access': body = {'has_access': True}
        elif path == '/api/courses/42/progress': body = {'completed_page_ids': [1]}
        elif path == '/api/pages/1/complete':
            route.fulfill(status=204, body='')
            return
        else:
            errors.append(path)
            route.fulfill(status=500, body='Unexpected API')
            return
        route.fulfill(status=200, json=body)

    with sync_playwright() as p:
        browser = p.chromium.launch(executable_path=os.environ.get('CHROMIUM_PATH', '/usr/bin/chromium'), headless=True, args=['--no-sandbox'])
        guest = browser.new_context(viewport={'width': 390, 'height': 844})
        guest.route('**/api/**', api)
        page = guest.new_page()
        page.on('pageerror', lambda error: errors.append(str(error)))
        page.goto(BASE + '/chat')
        expect(page.get_by_role('heading', name='Вход в Cringearium')).to_be_visible()
        assert urlparse(page.url).path == '/login'
        assert '/api/chats' not in calls
        page.get_by_label('Email', exact=True).fill('test@example.test')
        page.get_by_label('Пароль', exact=True).fill('password')
        page.locator('.auth-form').get_by_role('button', name='Войти', exact=True).click()
        page.wait_for_url('**/chat')
        page.wait_for_function("document.querySelector('.chat-layout') !== null")
        assert '/api/chats' in calls
        for width in [320, 390, 768, 980]:
            page.set_viewport_size({'width': width, 'height': 844})
            page.goto(BASE)
            header = page.locator('.app-shell > .header')
            toggle = header.get_by_role('button', name='Меню', exact=True)
            expect(toggle).to_be_visible()
            assert header.bounding_box()['height'] <= 62
            if width <= 420:
                expect(header.locator('.brand-short')).to_be_visible()
                expect(header.locator('.brand-full')).not_to_be_visible()
            toggle.click()
            expect(header.get_by_role('link', name='Каталог курсов', exact=True)).to_be_visible()
            expect(header.get_by_role('link', name='Кринжик', exact=True)).to_be_visible()
            assert page.evaluate('document.documentElement.scrollWidth <= innerWidth + 1')
            page.keyboard.press('Escape')
            expect(toggle).to_be_focused()
            toggle.click()
            header.get_by_role('link', name='Каталог курсов', exact=True).click()
            expect(page.get_by_role('heading', name='Каталог курсов')).to_be_visible()
            expect(toggle).to_have_attribute('aria-expanded', 'false')
        page.set_viewport_size({'width': 390, 'height': 844})
        page.goto(BASE + '/courses/42')
        expect(page.locator('.lesson-header')).to_be_visible()
        drawer = page.get_by_role('dialog', name='Содержание курса')
        expect(drawer).not_to_be_visible()
        trigger = page.locator('.course-mobile-toolbar button')
        trigger.click()
        expect(drawer).to_be_visible()
        assert page.evaluate("document.body.style.overflow === 'hidden'")
        expect(drawer.get_by_role('button', name='Закрыть меню курса')).to_be_focused()
        page.keyboard.press('Shift+Tab')
        expect(drawer.get_by_role('button', name='Итоговый тест: весь курс')).to_be_focused()
        page.keyboard.press('Tab')
        expect(drawer.get_by_role('button', name='Закрыть меню курса')).to_be_focused()
        page.wait_for_function("document.querySelector('.course-sidebar').getBoundingClientRect().left >= -1")
        page.screenshot(path='/tmp/mobile-course-drawer.png')
        page.keyboard.press('Escape')
        expect(drawer).not_to_be_visible()
        expect(trigger).to_be_focused()
        page.wait_for_function("document.body.style.overflow !== 'hidden'")
        trigger.click()
        page.locator('.course-menu-backdrop').click(position={'x': 380, 'y': 500}, force=True)
        expect(drawer).not_to_be_visible()
        trigger.click()
        drawer.get_by_role('button', name='Практика: квадратные уравнения').click()
        expect(drawer).not_to_be_visible()
        expect(page.get_by_role('button', name='Начать AI-тест')).to_be_visible()
        page.wait_for_function("document.body.style.overflow !== 'hidden'")
        trigger.click()
        page.set_viewport_size({'width': 1280, 'height': 900})
        expect(page.locator('.course-sidebar')).to_be_visible()
        expect(trigger).not_to_be_visible()
        page.wait_for_function("document.body.style.overflow !== 'hidden'")
        expect(page.locator('.footer').get_by_role('link', name='Кринжик', exact=True)).to_be_visible()
        expired = browser.new_context()
        expired.add_init_script("localStorage.setItem('cringearium_token','expired')")
        expired.route('**/api/**', api)
        expired_page = expired.new_page()
        before = calls.count('/api/chats')
        expired_page.goto(BASE + '/chat')
        expect(expired_page.get_by_role('heading', name='Вход в Cringearium')).to_be_visible()
        assert calls.count('/api/chats') == before
        assert not errors, errors
        browser.close()
    print('PASS mobile navigation: guest/expired chat redirects, return after login, responsive dropdown, compact brand, course drawer/focus/Escape/backdrop/selection/resize, footer link')


if __name__ == '__main__':
    main()
