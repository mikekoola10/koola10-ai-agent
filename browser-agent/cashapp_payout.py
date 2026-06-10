import asyncio
from playwright.async_api import async_playwright
import os
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

async def cashapp_payout(amount: float, target_tag: str, card_number: str, card_expiry: str, card_cvv: str) -> dict:
    """
    Automates Cash App payout:
    - Logs into Cash App (email/password from env)
    - Adds cash using the provided virtual card
    - Sends the specified amount to the target $tag
    """
    email = os.getenv("CASHAPP_EMAIL")
    password = os.getenv("CASHAPP_PASSWORD")
    if not email or not password:
        return {"success": False, "message": "Missing Cash App credentials in environment"}

    async with async_playwright() as p:
        # Launch browser (headless=False for debugging, set True in production)
        browser = await p.chromium.launch(headless=True)
        context = await browser.new_context()
        page = await context.new_page()

        try:
            # 1. Go to Cash App login page
            await page.goto("https://cash.app/login")
            await page.wait_for_selector("input[type='email']", timeout=10000)
            await page.fill("input[type='email']", email)
            await page.click("button:has-text('Continue')")
            await page.wait_for_selector("input[type='password']", timeout=5000)
            await page.fill("input[type='password']", password)
            await page.click("button:has-text('Sign In')")

            # Handle any 2FA if needed (e.g., SMS code) – you may need to extend
            # For now, assume no 2FA or use a saved session.

            # 2. Navigate to "Add Cash" (usually from the main balance screen)
            # Wait for main dashboard to load
            await page.wait_for_selector("text=Add Cash", timeout=15000)
            await page.click("text=Add Cash")

            # 3. Enter amount
            await page.fill("input[placeholder='Amount']", str(amount))
            await page.click("button:has-text('Add')")

            # 4. Select "Debit Card" as funding source (if prompted)
            await page.wait_for_selector("text=Debit Card", timeout=5000)
            await page.click("text=Debit Card")

            # 5. Fill card details
            await page.fill("input[name='cardnumber']", card_number)
            await page.fill("input[name='expiry']", card_expiry)  # MM/YY
            await page.fill("input[name='cvv']", card_cvv)
            await page.click("button:has-text('Add Card')")

            # 6. Confirm add cash
            await page.click("button:has-text('Confirm')")
            # Wait for success message
            await page.wait_for_selector("text=Added", timeout=10000)

            # 7. Now send payout to target tag
            # Navigate to "Pay" or "Send"
            await page.goto("https://cash.app/pay")
            await page.fill("input[placeholder='$Cashtag, phone, or email']", target_tag)
            await page.click("button:has-text('Pay')")

            # 8. Enter amount
            await page.fill("input[placeholder='Amount']", str(amount))
            await page.click("button:has-text('Pay')")

            # 9. Confirm payment (assuming no additional prompts)
            await page.wait_for_selector("button:has-text('Confirm')", timeout=5000)
            await page.click("button:has-text('Confirm')")

            # Wait for success
            await page.wait_for_selector("text=Sent", timeout=15000)

            # Extract transaction ID (if possible) from the receipt
            tx_id = None
            # You may need to parse the page; here we return a dummy ID
            tx_id = f"cashapp_{int(asyncio.get_event_loop().time())}"

            return {"success": True, "message": "Payout completed", "tx_id": tx_id}

        except Exception as e:
            logger.exception("Cash App payout failed")
            return {"success": False, "message": str(e)}
        finally:
            await browser.close()
