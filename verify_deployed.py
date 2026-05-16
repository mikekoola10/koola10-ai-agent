import asyncio
from playwright.async_api import async_playwright
import time

async def verify():
    async with async_playwright() as p:
        browser = await p.chromium.launch()
        page = await browser.new_page()

        print("Navigating to dashboard...")
        await page.goto("https://ceo-dashboard-three-nu.vercel.app")

        # Wait for potential initial load
        await page.wait_for_timeout(2000)

        print("Triggering test notification via API...")
        # Use run_in_bash to trigger the notification since we want to see it on this page
        # But wait, we can't easily trigger it from within the same browser session's background
        # unless we use another request.
        import requests
        requests.post("https://koola10.fly.dev/events/emit", json={
            "type": "jarvis_notification",
            "data": {
                "alert_id": "test_alert_verify",
                "title": "Jarvis Verification",
                "message": "Deployment verification in progress."
            }
        })

        print("Waiting for notification panel...")
        try:
            # The NotificationPanel should appear
            await page.wait_for_selector("text=Jarvis Verification", timeout=10000)
            print("NotificationPanel found!")
        except:
            print("NotificationPanel not found, checking for buttons anyway...")

        # Check for Proceed button
        proceed_btn = page.locator("button:has-text('Proceed')")
        if await proceed_btn.is_visible():
            print("Proceed button found, clicking...")
            await proceed_btn.click()

            # After clicking proceed, it should show "Analyzing..."
            print("Waiting for ConversationPanel (Analyzing state)...")
            await page.wait_for_selector("text=Analyzing situation", timeout=5000)
            print("ConversationPanel is visible (Analyzing state).")
        else:
            print("Proceed button NOT found.")

        # Final screenshot
        await page.screenshot(path="deployed_verification.png")
        print("Screenshot saved to deployed_verification.png")

        await browser.close()

if __name__ == "__main__":
    asyncio.run(verify())
