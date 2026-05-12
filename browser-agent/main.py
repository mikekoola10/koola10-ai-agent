import os
import base64
import asyncio
import logging
import re
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Dict, Any, Optional
from browser_use import Agent, Browser, BrowserProfile
from langchain_openai import ChatOpenAI
from playwright.async_api import async_playwright, TimeoutError as PlaywrightTimeoutError

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

PROFILE_DIR = "/data/browser-profile"
os.makedirs(PROFILE_DIR, exist_ok=True)

app = FastAPI()

# Configure LLM
api_key = os.getenv("DEEPSEEK_API_KEY")
llm = ChatOpenAI(
    model='deepseek-chat',
    openai_api_key=api_key,
    openai_api_base='https://api.deepseek.com',
)

browser_profile = BrowserProfile(
    headless=os.getenv("BROWSER_HEADLESS", "true").lower() == "true",
    disable_security=True,
)
# We'll create a new browser instance for each task to ensure clean state and session persistence within task
# browser = Browser(config=browser_config)

class NavigateRequest(BaseModel):
    url: str

class FormRequest(BaseModel):
    url: str
    form_data: Dict[str, str]

class ExtractRequest(BaseModel):
    url: str
    instruction: str

class StripeKeysRequest(BaseModel):
    otp: Optional[str] = None

async def get_screenshot_base64(page):
    try:
        screenshot = await page.screenshot()
        return base64.b64encode(screenshot).decode('utf-8')
    except Exception as e:
        logger.error(f"Failed to take screenshot: {e}")
        return None

@app.get("/health")
async def health():
    return {"status": "ok"}

@app.post("/browser/navigate")
async def navigate(req: NavigateRequest):
    async with async_playwright() as p:
        browser_pw = await p.chromium.launch()
        page = await browser_pw.new_page()
        await page.goto(req.url)
        title = await page.title()
        screenshot = await page.screenshot()
        await browser_pw.close()
        return {
            "title": title,
            "screenshot": base64.b64encode(screenshot).decode('utf-8')
        }

@app.post("/browser/fill-form")
async def fill_form(req: FormRequest):
    instructions = [f"Fill the field '{k}' with value '{v}'" for k, v in req.form_data.items()]
    # Task includes returning a screenshot of the filled form.
    # browser-use Agent can take screenshots, but to return it via API we need to capture it after the task.
    full_instruction = f"Go to {req.url}, " + ", ".join(instructions) + ". After filling everything, stay on the page so I can take a screenshot."

    # Using a local browser instance to ensure we can capture the final state
    browser = Browser(profile=browser_profile)
    agent = Agent(
        task=full_instruction,
        llm=llm,
        browser=browser,
    )
    result = await agent.run()

    # Capture screenshot from the same session
    # browser-use manages sessions internally. We can access the playwright page via the browser instance.
    playwright_browser = await browser.get_playwright_browser()
    # This is slightly tricky as browser-use abstracts the page.
    # Actually, browser-use Agent has a .run() that returns a result.
    # Let's try to get the active page.

    screenshot_b64 = ""
    try:
        # Get the underlying playwright page from the browser manager
        context = (await browser.get_context())
        page = await context.get_current_page()
        if page:
            screenshot = await page.screenshot()
            screenshot_b64 = base64.b64encode(screenshot).decode('utf-8')
    except Exception as e:
        print(f"Screenshot failed: {e}")

    await browser.close()
    return {
        "status": "success",
        "agent_result": str(result),
        "screenshot": screenshot_b64
    }

@app.post("/browser/submit-form")
async def submit_form(req: FormRequest):
    instructions = [f"Fill the field '{k}' with value '{v}'" for k, v in req.form_data.items()]
    full_instruction = f"Go to {req.url}, " + ", ".join(instructions) + ". Finally, find and click the submit button. Wait for a confirmation or success message."

    browser = Browser(profile=browser_profile)
    agent = Agent(
        task=full_instruction,
        llm=llm,
        browser=browser,
    )
    result = await agent.run()

    screenshot_b64 = ""
    try:
        context = (await browser.get_context())
        page = await context.get_current_page()
        if page:
            screenshot = await page.screenshot()
            screenshot_b64 = base64.b64encode(screenshot).decode('utf-8')
    except Exception as e:
        print(f"Screenshot failed: {e}")

    await browser.close()
    return {
        "status": "success",
        "confirmation": str(result),
        "screenshot": screenshot_b64
    }

@app.post("/browser/extract")
async def extract(req: ExtractRequest):
    full_instruction = f"Go to {req.url} and {req.instruction}"

    browser = Browser(profile=browser_profile)
    agent = Agent(
        task=full_instruction,
        llm=llm,
        browser=browser,
    )
    result = await agent.run()
    await browser.close()
    return {"data": str(result)}

@app.post("/browser/stripe-live-keys")
async def get_stripe_keys(req: StripeKeysRequest):
    email = os.getenv("STRIPE_LOGIN_EMAIL")
    password = os.getenv("STRIPE_LOGIN_PASSWORD")

    if not email or not password:
        logger.error("Credentials not set")
        raise HTTPException(status_code=500, detail="STRIPE_LOGIN_EMAIL or STRIPE_LOGIN_PASSWORD not set")

    logger.info(f"Starting Stripe key extraction for {email}")

    async with async_playwright() as p:
        # Launch persistent context — saves cookies, localStorage, session tokens
        context = await p.chromium.launch_persistent_context(
            user_data_dir=PROFILE_DIR,
            headless=os.getenv("BROWSER_HEADLESS", "false").lower() == "true",  # Default to False for Stripe
            args=["--disable-blink-features=AutomationControlled", "--no-sandbox", "--disable-setuid-sandbox"],
            viewport={"width": 1280, "height": 800}
        )
        page = context.pages[0] if context.pages else await context.new_page()
        page.set_default_timeout(60000)

        try:
            # Step 1: Check if already logged in or need login
            logger.info("Checking Stripe session...")
            await page.goto("https://dashboard.stripe.com/dashboard", wait_until="load")

            # If redirected to login, we need to authenticate
            if "login" in page.url:
                logger.info("Session expired or not found. Logging in...")
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
            else:
                logger.info("Already logged in.")

            # Step 2: Handle 2FA or dashboard redirect
            logger.info("Waiting for dashboard or 2FA...")
            success = False
            for _ in range(30):
                if "dashboard" in page.url and "login" not in page.url:
                    logger.info("Successfully reached dashboard")
                    success = True
                    break

                # Check for 2FA markers
                is_2fa = await page.get_by_text("Verification code").is_visible() or \
                         await page.locator('input[name="otp"]').is_visible() or \
                         await page.locator('input[id="otp"]').is_visible() or \
                         await page.get_by_text("Enter the code").is_visible()

                if is_2fa:
                    if req.otp:
                        logger.info("OTP provided, filling...")
                        otp_field = page.locator('input[name="otp"], input[id="otp"]').first
                        await otp_field.fill(req.otp)
                        # Usually it auto-submits, but let's check for a button just in case
                        submit_otp = page.locator('button:has-text("Continue"), button:has-text("Submit")').first
                        if await submit_otp.is_visible(timeout=2000):
                            await submit_otp.click()
                        req.otp = None # Clear it so we don't try again if it fails
                        await asyncio.sleep(5)
                        continue
                    else:
                        logger.info("2FA detected, but no OTP provided")
                        screenshot = await get_screenshot_base64(page)
                        await context.close()
                        return {
                            "message": "2FA_REQUIRED",
                            "screenshot": screenshot
                        }

                # Handle anti-bot challenge
                if await page.get_by_text("Please drag the element").is_visible() or \
                   await page.get_by_text("Verify you are human").is_visible():
                    logger.info("Bot challenge detected")
                    screenshot = await get_screenshot_base64(page)
                    await context.close()
                    return {
                        "message": "2FA_REQUIRED",
                        "screenshot": screenshot,
                        "info": "Anti-bot challenge detected. Manual intervention required."
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
                screenshot = await get_screenshot_base64(page)
                await context.close()
                return {
                    "message": "2FA_REQUIRED",
                    "screenshot": screenshot,
                    "url": page.url,
                    "info": "Timed out waiting for dashboard. Screenshot attached."
                }

            # Step 3: Extract Secret Key
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

            # Step 4: Extract Webhook Secret
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

            await context.close()
            return {
                "secret_key": secret_key,
                "webhook_secret": webhook_secret
            }

        except Exception as e:
            logger.error(f"Error during extraction: {e}")
            screenshot = await get_screenshot_base64(page)
            current_url = page.url
            await context.close()
            return {
                "message": "ERROR",
                "error": str(e),
                "screenshot": screenshot,
                "url": current_url
            }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
