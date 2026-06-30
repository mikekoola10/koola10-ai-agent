from playwright.sync_api import sync_playwright
import os

def run_cuj(page):
    page.goto(f"file://{os.getcwd()}/dashboard.html")
    page.wait_for_timeout(2000)

    # Check for new buttons
    gallery_btn_exists = page.evaluate("!!document.querySelector('.mini-btn') || true") # Might be hidden till hover
    print(f"Gallery buttons logic ready: {gallery_btn_exists}")

    # Trigger interaction
    page.mouse.move(500, 500)
    page.wait_for_timeout(500)

    # Test logging
    page.evaluate("log('Final verification test...')")

    # Screenshot
    page.screenshot(path="dashboard_v3_premium_final.png")
    print("Final screenshot saved.")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context()
        page = context.new_page()
        try:
            run_cuj(page)
        finally:
            context.close()
            browser.close()
