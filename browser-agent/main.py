import os
import base64
import asyncio
import logging
import re
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import Dict, Any, Optional
from playwright.async_api import async_playwright, TimeoutError as PlaywrightTimeoutError

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

browser_pw = None
context = None

@app.on_event("startup")
async def startup():
    global browser_pw, context
    p = await async_playwright().start()
    browser_pw = await p.chromium.launch(
        headless=True,
        args=["--no-sandbox", "--disable-setuid-sandbox", "--disable-dev-shm-usage"]
    )
    context = await browser_pw.new_context(viewport={'width': 1280, 'height': 800})

@app.on_event("shutdown")
async def shutdown():
    global browser_pw
    if browser_pw:
        await browser_pw.close()

class NavigateRequest(BaseModel):
    url: str

class ExtractRequest(BaseModel):
    url: str
    instruction: str

@app.get("/health")
async def health():
    return {"status": "ok"}

async def get_screenshot_base64(page):
    try:
        screenshot = await page.screenshot()
        return base64.b64encode(screenshot).decode('utf-8')
    except Exception as e:
        logger.error(f"Failed to take screenshot: {e}")
        return None

@app.get("/browser/live-screenshot")
async def live_screenshot():
    if not context:
        raise HTTPException(status_code=503, detail="Browser not initialized")
    try:
        page = context.pages[0] if context.pages else await context.new_page()
        screenshot = await page.screenshot()
        return {"screenshot": base64.b64encode(screenshot).decode('utf-8')}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Screenshot failed: {str(e)}")

@app.post("/browser/stripe-live-keys")
async def get_stripe_keys():
    email = os.getenv("STRIPE_LOGIN_EMAIL")
    password = os.getenv("STRIPE_LOGIN_PASSWORD")

    if not email or not password:
        logger.error("Credentials not set")
        raise HTTPException(status_code=500, detail="STRIPE_LOGIN_EMAIL or STRIPE_LOGIN_PASSWORD not set")

    if not context:
        raise HTTPException(status_code=503, detail="Browser not initialized")

    logger.info(f"Starting Stripe key extraction for {email}")

    try:
        page = context.pages[0] if context.pages else await context.new_page()
        page.set_default_timeout(60000)

        # Step 1: Login
        logger.info("Navigating to login page...")
        await page.goto("https://dashboard.stripe.com/login", wait_until="load")

        # Dismiss cookies
        try:
            cookie_button = page.locator('button:has-text("Accept all"), button:has-text("Accept"), button:has-text("Reject non-essential")').first
            if await cookie_button.is_visible(timeout=3000):
                await cookie_button.click()
                logger.info("Dismissed cookies banner")
        except:
            pass

        logger.info("Filling email...")
        email_field = page.locator('input[name="email"], input[type="email"]').first
        await email_field.wait_for(state="visible", timeout=30000)
        await email_field.fill(email)

        password_field = page.locator('input[name="password"], input[type="password"]').first
        if not await password_field.is_visible(timeout=2000):
            logger.info("Password field not visible, clicking Continue...")
            await page.get_by_role("button", name="Continue").click()
            await password_field.wait_for(state="visible", timeout=10000)

        logger.info("Filling password...")
        await password_field.fill(password)

        # Click Sign in
        submit_button = page.locator('button[type="submit"], button:has-text("Sign in")').first
        await submit_button.click()
        logger.info("Login submitted")

        # Wait for either the dashboard to load OR a 2FA prompt
        logger.info("Waiting for dashboard or 2FA...")
        success = False
        for _ in range(30):
            if "dashboard" in page.url and "login" not in page.url:
                logger.info("Successfully reached dashboard")
                success = True
                break

            # Check for 2FA markers or if we are stuck on login
            if await page.get_by_text("Verification code").is_visible() or \
               await page.locator('input[name="otp"]').is_visible() or \
               await page.locator('input[id="otp"]').is_visible() or \
               await page.get_by_text("Enter the code").is_visible() or \
               await page.get_by_text("Please drag the element").is_visible() or \
               await page.get_by_text("Verify you are human").is_visible():
                logger.info("2FA or Challenge detected")
                screenshot = await get_screenshot_base64(page)
                return {
                    "message": "2FA_REQUIRED",
                    "screenshot": screenshot
                }

            # Handle intermediate pages
            if "select-account" in page.url:
                logger.info("Account selection detected, picking first account...")
                await page.locator('button').first.click()
            elif await page.get_by_text("Stay signed in").is_visible():
                await page.get_by_text("Yes").click()

            await asyncio.sleep(2)

        if not success:
            logger.warning(f"Failed to reach dashboard. Current URL: {page.url}")
            # Return screenshot as 2FA_REQUIRED so the user can see what's wrong
            screenshot = await get_screenshot_base64(page)
            return {
                "message": "2FA_REQUIRED",
                "screenshot": screenshot,
                "url": page.url,
                "info": "Timed out waiting for dashboard. Screenshot attached."
            }

        # Step 2: Extract Secret Key
        logger.info("Navigating to API keys page...")
        await page.goto("https://dashboard.stripe.com/apikeys", wait_until="load")

        # Look for the Reveal button in the Secret key row
        logger.info("Looking for secret key...")
        secret_key = None
        try:
            await page.wait_for_selector('text="Secret key"', timeout=20000)

            reveal_button = page.locator('button:has-text("Reveal live key")').first
            if await reveal_button.is_visible(timeout=10000):
                await reveal_button.click()
                logger.info("Clicked Reveal live key")

            await page.wait_for_selector('text=/sk_live_[a-zA-Z0-9]+/', timeout=15000)
            content = await page.content()
            match = re.search(r"(sk_live_[a-zA-Z0-9]+)", content)
            if match:
                secret_key = match.group(1)
            logger.info(f"Extracted secret key: {secret_key[:10]}...")
        except Exception as e:
            logger.error(f"Failed to find secret key: {e}")

        # Step 3: Extract Webhook Secret
        logger.info("Navigating to Webhooks page...")
        await page.goto("https://dashboard.stripe.com/webhooks", wait_until="load")

        webhook_url = "https://koola10.fly.dev/stripe/webhook"
        webhook_secret = None

        try:
            await page.wait_for_selector('text="Endpoints"', timeout=20000)

            webhook_row = page.get_by_role("link", name=webhook_url).first
            if await webhook_row.is_visible(timeout=10000):
                logger.info("Found existing webhook row, clicking...")
                await webhook_row.click()
            else:
                logger.info("Webhook not found in list, creating new one...")
                add_button = page.locator('button:has-text("Add endpoint"), a:has-text("Add endpoint")').first
                await add_button.click()

                await page.wait_for_selector('input[name="url"]', state="visible")
                await page.fill('input[name="url"]', webhook_url)

                await page.click('button:has-text("Select events")')
                await page.wait_for_selector('text="Select events to listen to"', timeout=10000)

                events = ["checkout.session.completed", "invoice.payment_succeeded"]
                for event in events:
                    search_box = page.get_by_label("Search events")
                    await search_box.fill(event)
                    await page.get_by_role("checkbox", name=event).check()
                    logger.info(f"Selected event: {event}")

                await page.click('button:has-text("Add events")')
                await page.click('button:has-text("Add endpoint")')
                logger.info("New webhook endpoint added")
                await page.wait_for_url("**/webhooks/**", timeout=20000)
        except Exception as e:
            logger.error(f"Error handling webhook list/creation: {e}")

        logger.info("Extracting signing secret...")
        try:
            reveal_wh_button = page.locator('button:has-text("Reveal")').first
            if await reveal_wh_button.is_visible(timeout=10000):
                await reveal_wh_button.click()
                logger.info("Clicked Reveal signing secret")

            await page.wait_for_selector('text=/whsec_[a-zA-Z0-9]+/', timeout=15000)
            content = await page.content()
            match = re.search(r"(whsec_[a-zA-Z0-9]+)", content)
            if match:
                webhook_secret = match.group(1)
            logger.info(f"Extracted webhook secret: {webhook_secret[:10]}...")
        except Exception as e:
            logger.error(f"Failed to find webhook secret: {e}")

        return {
            "secret_key": secret_key,
            "webhook_secret": webhook_secret
        }

    except Exception as e:
        logger.error(f"Error during extraction: {e}")
        screenshot = await get_screenshot_base64(page)
        current_url = page.url
        return {
            "message": "ERROR",
            "error": str(e),
            "screenshot": screenshot,
            "url": current_url
        }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
