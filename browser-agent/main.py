import os
# Force Playwright to use the browsers installed in the image
os.environ["PLAYWRIGHT_BROWSERS_PATH"] = "/ms-playwright"
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

class PSPlusPurchaseRequest(BaseModel):
    email: str
    password: str
    card_details: Dict[str, str]
    otp: Optional[str] = None

class PrivacyCreateCardRequest(BaseModel):
    email: str
    password: str
    amount_cents: int
    merchant: str
    memo: str
    otp: Optional[str] = None

async def get_screenshot_base64(page):
    try:
        # Use animations disabled to speed up and avoid font timeouts
        screenshot = await page.screenshot(animations="disabled", timeout=10000)
        return base64.b64encode(screenshot).decode('utf-8')
    except Exception as e:
        logger.error(f"Failed to take screenshot: {e}")
        try:
            # Fallback to very basic screenshot
            screenshot = await page.screenshot(timeout=5000)
            return base64.b64encode(screenshot).decode('utf-8')
        except:
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

@app.post("/browser/privacy/create-card")
async def privacy_create_card(req: PrivacyCreateCardRequest):
    logger.info(f"Starting Privacy card creation for {req.merchant}")

    PROFILE_DIR_PRIVACY = "/data/privacy-browser-profile"
    os.makedirs(PROFILE_DIR_PRIVACY, exist_ok=True)

    async with async_playwright() as p:
        context = await p.chromium.launch_persistent_context(
            user_data_dir=PROFILE_DIR_PRIVACY,
            headless=os.getenv("BROWSER_HEADLESS", "false").lower() == "true",
            args=[
                "--disable-blink-features=AutomationControlled",
                "--no-sandbox",
                "--disable-setuid-sandbox"
            ],
            viewport={"width": 1280, "height": 800}
        )
        page = context.pages[0] if context.pages else await context.new_page()
        page.set_default_timeout(60000)

        try:
            # Step 1: Login
            logger.info("Logging into Privacy.com...")
            await page.goto("https://app.privacy.com/login", wait_until="networkidle")

            if await page.locator('input[name="email"]').is_visible(timeout=5000):
                await page.fill('input[name="email"]', req.email)
                await page.fill('input[name="password"]', req.password)
                await page.click('button[type="submit"]')
                await page.wait_for_load_state("networkidle")

            # Handle 2FA if required
            if "2fa" in page.url or await page.locator('input[name="otp"]').is_visible(timeout=5000):
                if req.otp:
                    logger.info("Filling Privacy 2FA...")
                    await page.fill('input[name="otp"]', req.otp)
                    await page.click('button:has-text("Verify"), button:has-text("Submit")')
                    await page.wait_for_load_state("networkidle")
                else:
                    logger.info("Privacy 2FA Required")
                    screenshot = await get_screenshot_base64(page)
                    await context.close()
                    return {"status": "2FA_REQUIRED", "screenshot": screenshot}

            # Step 2: Handle paused account
            logger.info("Checking account status...")
            is_paused = await page.is_visible("text=account is paused") or await page.is_visible("text=temporarily frozen")
            if is_paused:
                logger.info("Account paused – attempting auto-reactivation...")
                reactivate_btn = page.locator("button:has-text('Reactivate'), a:has-text('Reactivate')").first
                if await reactivate_btn.is_visible(timeout=5000):
                    await reactivate_btn.click()
                    await page.wait_for_selector("text=Account active", timeout=30000)
                    logger.info("Account reactivated successfully.")
                else:
                    logger.info("Contacting support for reactivation...")
                    await page.click("button:has-text('Contact Support')")
                    await page.fill("textarea", "Please reactivate my account. It was paused due to inactivity.")
                    await page.click("button:has-text('Send')")
                    logger.info("Support request sent.")
                    # We can't wait forever, so we'll just try to proceed or fail gracefully
                    await asyncio.sleep(5)

            # Step 3: Create Card
            logger.info("Navigating to card creation...")
            create_btn = page.locator("a:has-text('Create Card'), button:has-text('New Card'), button:has-text('Create Card')").first
            await create_btn.click()
            await page.wait_for_load_state("networkidle")

            # Fill card details
            logger.info(f"Filling card details: {req.merchant}, ${req.amount_cents/100:.2f}")
            await page.fill("input[placeholder*='limit']", f"{req.amount_cents/100:.2f}")

            # Merchant lock
            merchant_lock = page.locator("text=Merchant lock, label:has-text('Merchant lock')").first
            if await merchant_lock.is_visible(timeout=3000):
                await merchant_lock.click()
                await page.fill("input[placeholder='Merchant name']", req.merchant)

            await page.fill("input[placeholder*='memo']", req.memo)

            # Submit
            await page.click("button:has-text('Create Card')")

            # Step 4: Extract details
            logger.info("Waiting for card details...")
            # These selectors might need adjustment based on real UI
            await page.wait_for_selector("[data-testid='card-number'], .card-number, text=4111", timeout=30000)

            async def get_text(sel):
                try: return await page.locator(sel).first.inner_text()
                except: return ""

            card_number = await get_text("[data-testid='card-number']")
            expiry = await get_text("[data-testid='expiry']")
            cvv = await get_text("[data-testid='cvv']")

            if not card_number:
                # Fallback extraction from page content
                content = await page.content()
                # Simple regex for PAN-like structures if needed

            # Split expiry MM/YYYY or MM/YY
            exp_month, exp_year = "", ""
            if expiry and "/" in expiry:
                parts = expiry.split("/")
                exp_month = parts[0].strip()
                exp_year = parts[1].strip()
                if len(exp_year) == 2:
                    exp_year = "20" + exp_year

            await context.close()
            return {
                "status": "success",
                "card": {
                    "pan": card_number,
                    "exp_month": exp_month,
                    "exp_year": exp_year,
                    "cvv": cvv
                }
            }

        except Exception as e:
            logger.error(f"Privacy automation failed: {e}")
            screenshot = await get_screenshot_base64(page)
            await context.close()
            return {
                "status": "error",
                "error": str(e),
                "screenshot": screenshot
            }

@app.post("/browser/psplus-purchase")
async def psplus_purchase(req: PSPlusPurchaseRequest):
    logger.info(f"Starting PS Plus purchase for {req.email}")

    # Use a specific profile for PS Plus to avoid permission issues with existing root-owned folders
    PROFILE_DIR_PS = "/data/psplus-browser-profile"
    os.makedirs(PROFILE_DIR_PS, exist_ok=True)

    async with async_playwright() as p:
        context = await p.chromium.launch_persistent_context(
            user_data_dir=PROFILE_DIR_PS,
            headless=os.getenv("BROWSER_HEADLESS", "false").lower() == "true",
            args=[
                "--disable-blink-features=AutomationControlled",
                "--no-sandbox",
                "--disable-setuid-sandbox",
                "--disable-dev-shm-usage",
                "--disable-web-security",
                "--disable-features=IsolateOrigins,site-per-process"
            ],
            user_agent="Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
            viewport={"width": 1280, "height": 800}
        )
        page = context.pages[0] if context.pages else await context.new_page()
        page.set_default_timeout(60000)

        try:
            # 1. Login Flow
            logger.info("Navigating to PlayStation Store...")
            await page.goto("https://store.playstation.com/en-us/pages/latest", wait_until="networkidle", timeout=60000)

            # Dismiss cookie banner if present
            logger.info("Handling potential cookie banner...")
            await asyncio.sleep(2)
            try:
                await page.evaluate("""
                    () => {
                        const clickBtn = (id) => document.getElementById(id)?.click();
                        clickBtn('onetrust-accept-btn-handler');

                        const removeEls = (sel) => document.querySelectorAll(sel).forEach(el => el.remove());
                        ['#onetrust-banner-sdk', '.ot-sdk-container', '.ot-sdk-overlay'].forEach(removeEls);

                        document.body.style.overflow = 'auto';
                        document.body.classList.remove('ot-sdk-no-scroll');
                        document.documentElement.classList.remove('ot-sdk-no-scroll');
                    }
                """)
                logger.info("Cleanup cookie banner elements")
            except:
                pass

            # Check if signed in
            signin_button = page.locator('button:has-text("Sign In"), [data-qa="web-server-nav-primary-button"]').first
            if await signin_button.is_visible(timeout=10000):
                logger.info("Clicking Sign In...")

                # Handle popup or redirect
                try:
                    async with context.expect_page(timeout=10000) as new_page_info:
                        await signin_button.click(force=True)
                    page = await new_page_info.value
                    logger.info("Detected new page/popup for sign-in")
                except:
                    logger.info("No popup detected, continuing on current page")
                    # If no popup, it might be a redirect. Let's just wait for the fields.

                # Wait for any of the common Sony login markers
                logger.info(f"Current URL: {page.url}")
                logger.info("Waiting for sign-in page content...")

                email_selectors = [
                    'input[name="loginId"]',
                    'input[type="email"]',
                    '#signin-id',
                    '[data-qa="loginId-input-field"]',
                    'input[name="username"]'
                ]

                # Wait for at least one to be visible
                email_field = None
                for _ in range(12): # 60 seconds total
                    for sel in email_selectors:
                        try:
                            el = page.locator(sel).first
                            if await el.is_visible(timeout=1000):
                                email_field = el
                                break
                        except: continue
                    if email_field: break
                    await asyncio.sleep(5)
                    logger.info("Still waiting for email field...")

                if not email_field:
                    raise Exception("Timeout waiting for Sony login email field")

                logger.info("Filling email...")
                await email_field.fill(req.email)

                next_btn = page.locator('button:has-text("Next"), button:has-text("Continue"), [data-qa="loginId-next-button"]').first
                await next_btn.click()

                logger.info("Filling password...")
                password_field = page.locator('input[type="password"], input[name="password"]').first
                await password_field.wait_for(state="visible", timeout=20000)
                await password_field.fill(req.password)

                signin_btn = page.locator('button:has-text("Sign In"), button:has-text("Log In"), [data-qa="password-signin-button"]').first
                await signin_btn.click()

                # Check for 2FA
                await asyncio.sleep(5)
                is_2fa = "2step" in page.url or await page.locator('input[name="verificationCode"]').is_visible(timeout=5000)

                if is_2fa:
                    if req.otp:
                        logger.info("Filling 2FA code...")
                        await page.fill('input[name="verificationCode"]', req.otp)
                        await page.click('button:has-text("Verify")')
                        await asyncio.sleep(5)
                    else:
                        logger.info("2FA Required")
                        screenshot = await get_screenshot_base64(page)
                        await context.close()
                        return {"status": "2FA_REQUIRED", "screenshot": screenshot}

            # 2. Navigate to Premium Subscription
            logger.info("Navigating to PS Plus Premium page...")
            await page.goto("https://store.playstation.com/en-us/product/IP9102-NPIA90006_01-PSPLUSPREMIUM12M", wait_until="load")

            # 3. Add to Cart / Subscribe
            subscribe_btn = page.locator('button:has-text("Add to Cart"), button:has-text("Subscribe"), button:has-text("Select")').first
            await subscribe_btn.click()
            logger.info("Clicked Subscribe/Add to Cart")

            # 4. Checkout Flow
            # Navigate to cart/checkout if not automatic
            if "cart" not in page.url and "checkout" not in page.url:
                await page.goto("https://store.playstation.com/en-us/cart", wait_until="load")

            checkout_btn = page.locator('button:has-text("Proceed to Checkout")').first
            await checkout_btn.click()

            # 5. Payment Method
            # Find "Add a credit/debit card"
            add_card_btn = page.locator('button:has-text("Add a credit/debit card"), button:has-text("Add a New Card")').first
            if await add_card_btn.is_visible(timeout=10000):
                await add_card_btn.click()

                # Enter card details
                # Note: PlayStation often uses iframes for card details
                logger.info("Filling card details...")

                async def fill_field(selector, value):
                    # Try main page
                    try:
                        field = page.locator(selector)
                        if await field.is_visible(timeout=2000):
                            await field.fill(value)
                            return True
                    except:
                        pass

                    # Try iframes
                    for frame in page.frames:
                        try:
                            field = frame.locator(selector)
                            if await field.is_visible(timeout=1000):
                                await field.fill(value)
                                return True
                        except:
                            continue
                    return False

                await fill_field('input[name="cardNumber"], #cardNumber', req.card_details["number"])
                await fill_field('input[name="expiryDate"], #expiryDate', f"{req.card_details['expiry_month']}/{req.card_details['expiry_year'][-2:]}")
                await fill_field('input[name="cvv"], #cvv', req.card_details["cvc"])
                await fill_field('input[name="cardholderName"], #cardholderName', "Mike Koola")

                save_btn = page.locator('button:has-text("Save")').first
                await save_btn.click()
                await asyncio.sleep(3)

            # 6. Confirm Purchase
            confirm_btn = page.locator('button:has-text("Confirm Purchase"), button:has-text("Order & Pay")').first
            await confirm_btn.click()
            logger.info("Purchase confirmed")

            await page.wait_for_selector('text="Thank you"', timeout=30000)
            screenshot = await get_screenshot_base64(page)

            await context.close()
            return {
                "status": "success",
                "message": "PlayStation Plus Premium purchased successfully",
                "screenshot": screenshot
            }

        except Exception as e:
            logger.error(f"Purchase failed: {e}")
            screenshot = await get_screenshot_base64(page)
            await context.close()
            return {
                "status": "error",
                "error": str(e),
                "screenshot": screenshot
            }

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
    uvicorn.run(app, host="0.0.0.0", port=8081)
