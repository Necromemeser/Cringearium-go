"""Optional public frontend check. Requires Vite, Playwright and Chromium; no writes or LLM calls."""
import os
from playwright.sync_api import sync_playwright, expect

BASE = os.environ.get('LEARNER_UI_URL', 'http://127.0.0.1:5173')


def main():
    errors = []
    with sync_playwright() as p:
        browser = p.chromium.launch(executable_path=os.environ.get('CHROMIUM_PATH', '/usr/bin/chromium'), headless=True, args=['--no-sandbox'])
        page = browser.new_page(viewport={'width': 1440, 'height': 1000})
        page.on('pageerror', lambda error: errors.append(str(error)))
        page.on('response', lambda response: errors.append(response.url) if '/images/about/' in response.url and response.status >= 400 else None)
        for width in [1440, 768, 390, 320]:
            page.set_viewport_size({'width': width, 'height': 1000})
            page.goto(BASE)
            expect(page.get_by_role('heading', name='Сложные темы. Понятные шаги.')).to_be_visible()
            expect(page.get_by_role('heading', name='Популярные направления')).to_have_count(0)
            expect(page.get_by_role('heading', name='Курс дня')).to_have_count(0)
            assert page.evaluate('document.documentElement.scrollWidth <= innerWidth + 1'), ('home overflow', width)
            if width == 1440: page.screenshot(path='/tmp/home-redesign-desktop.png', full_page=True)
            if width == 390: page.screenshot(path='/tmp/home-redesign-mobile.png', full_page=True)
            page.evaluate('window.scrollTo(0, document.body.scrollHeight)')
            page.locator('.footer').get_by_role('link', name='О нас', exact=True).click()
            expect(page.get_by_role('heading', name='О нашей образовательной платформе')).to_be_visible()
            page.wait_for_function('window.scrollY === 0')
            expect(page.get_by_role('heading', name='Наша миссия')).to_be_visible()
            expect(page.locator('.about-member')).to_have_count(3)
            expect(page.get_by_text('146%', exact=True)).to_be_attached()
            for img in page.locator('main img').all():
                img.scroll_into_view_if_needed()
                page.wait_for_function('(src) => Array.from(document.images).some(img => img.src === src && img.complete && img.naturalWidth > 0)', arg=img.evaluate('img => img.src'))
            assert page.evaluate('document.documentElement.scrollWidth <= innerWidth + 1'), ('about overflow', width)
            page.evaluate('window.scrollTo({top: 0, behavior: "instant"})')
            if width == 1440: page.screenshot(path='/tmp/about-migrated-desktop.png', full_page=True)
            if width == 390: page.screenshot(path='/tmp/about-migrated-mobile.png', full_page=True)
            header = page.locator('.app-shell > .header')
            assert not any(symbol in header.inner_text() for symbol in ['⌂', '📚', '🤖'])
            header.get_by_role('link', name='Cringearium', exact=True).click()
            page.wait_for_function('window.scrollY === 0')
            expect(page.get_by_role('link', name='Выбрать курс')).to_be_visible()
        assert not errors, errors
        browser.close()
    print('PASS public pages: home layout, migrated about text and four images, route scroll reset, plain header labels, no overflow at 320/390/768/1440px')


if __name__ == '__main__':
    main()
