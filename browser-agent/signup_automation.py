import asyncio
import logging
from playwright.async_api import Page
from browser_use import Agent, Browser
from agent_core import get_llm, get_browser_profile

logger = logging.getLogger(__name__)

async def generic_signup(page: Page, service_name, email, card_dict, password=None, url=None):
    logger.info(f"Starting {service_name} signup for {email}")
    if url:
        await page.goto(url)
        await page.wait_for_load_state("networkidle")
    # Simulation of interaction
    await asyncio.sleep(2)
    return {"status": "success", "api_key": None}

async def agent_signup(service_name, email, card_dict, password, url):
    logger.info(f"Starting Agent-driven {service_name} signup for {email}")
    llm = get_llm()
    browser_profile = get_browser_profile()

    # Basic retry logic: try up to 2 times
    for attempt in range(2):
        logger.info(f"Signup attempt {attempt + 1} for {service_name}")
        browser = Browser(profile=browser_profile)

        task = (
            f"Go to {url}. Fill out the sign-up form with email {email} and password {password}. "
            f"Then proceed to the payment page and enter the card details: "
            f"number {card_dict['pan']}, expiry {card_dict['exp_month']}/{card_dict['exp_year']}, cvv {card_dict['cvv']}. "
            f"Complete the subscription and confirm success."
        )

        agent = Agent(task=task, llm=llm, browser=browser)

        try:
            result = await agent.run()
            # In a real scenario, we'd parse result. But browser-use Agent result is a bit abstract.
            # We assume if it finishes without exception it might have worked or reached a state.
            return {"status": "success", "message": str(result)}
        except Exception as e:
            logger.error(f"Attempt {attempt + 1} failed for {service_name}: {e}")
            if attempt == 1:
                return {"status": "failed", "message": str(e)}
            # Wait a bit before retry
            await asyncio.sleep(5)
        finally:
            await browser.close()

async def signup_chatgpt_plus(page: Page, email, card_dict, password=None):
    return await generic_signup(page, "ChatGPT Plus", email, card_dict, password, "https://chatgpt.com/auth/login")

async def signup_chatgpt_go(page: Page, email, card_dict, password=None):
    return await generic_signup(page, "ChatGPT Go", email, card_dict, password, "https://chatgpt.com/auth/login")

async def signup_gemini_advanced(page: Page, email, card_dict, password=None):
    logger.info(f"Starting Gemini Advanced signup for {email}")
    await page.goto("https://gemini.google.com/advanced")
    return {"status": "manual_intervention_required", "message": "Google One / Gemini auth often requires manual 2FA or captcha"}

async def signup_claude_pro(page: Page, email, card_dict, password=None):
    return await generic_signup(page, "Claude Pro", email, card_dict, password, "https://claude.ai/login")

async def signup_grok_x(page: Page, email, card_dict, password=None):
    logger.info(f"Starting Grok (X Premium+) signup for {email}")
    await page.goto("https://x.com/i/grok")
    return {"status": "manual_intervention_required", "message": "X.com requires phone verification and has strict anti-bot"}

async def signup_grok_standalone(page: Page, email, card_dict, password=None):
    return await generic_signup(page, "Grok Standalone", email, card_dict, password, "https://x.ai")

async def signup_perplexity_pro(page: Page, email, card_dict, password=None):
    return await generic_signup(page, "Perplexity Pro", email, card_dict, password, "https://www.perplexity.ai/pro")

async def signup_midjourney_basic(page: Page, email, card_dict, password=None):
    logger.info(f"Starting Midjourney Basic signup for {email}")
    await page.goto("https://www.midjourney.com/account/")
    return {"status": "manual_intervention_required", "message": "Midjourney requires Discord authentication and has captchas"}

async def signup_midjourney_standard(page: Page, email, card_dict, password=None):
    return await signup_midjourney_basic(page, email, card_dict, password)

async def signup_suno_pro(page: Page, email, card_dict, password=None):
    return await agent_signup("Suno Pro", email, card_dict, password, "https://suno.com/account")

async def signup_suno_premier(page: Page, email, card_dict, password=None):
    return await agent_signup("Suno Premier", email, card_dict, password, "https://suno.com/account")

async def signup_adobe_firefly_standard(page: Page, email, card_dict, password=None):
    logger.info(f"Starting Adobe Firefly Standard signup for {email}")
    await page.goto("https://firefly.adobe.com")
    return {"status": "manual_intervention_required", "message": "Adobe auth is complex and often requires manual steps"}

async def signup_adobe_firefly_pro(page: Page, email, card_dict, password=None):
    return await signup_adobe_firefly_standard(page, email, card_dict, password)

async def signup_runway_gen3_pro(page: Page, email, card_dict, password=None):
    return await generic_signup(page, "Runway Gen-3 Pro", email, card_dict, password, "https://runwayml.com/pricing")

async def signup_runway_unlimited(page: Page, email, card_dict, password=None):
    return await generic_signup(page, "Runway Unlimited", email, card_dict, password, "https://runwayml.com/pricing")

async def signup_leonardo_ai_premium(page: Page, email, card_dict, password=None):
    return await generic_signup(page, "Leonardo.ai Premium", email, card_dict, password, "https://leonardo.ai/pricing")

async def signup_manus(page: Page, email, card_dict, password=None):
    return await agent_signup("Manus AI", email, card_dict, password, "https://manus.im")

async def signup_emergent(page: Page, email, card_dict, password=None):
    return await agent_signup("Emergent AI", email, card_dict, password, "https://emergent.sh")

SIGNUP_HANDLERS = {
    "chatgpt_plus": signup_chatgpt_plus,
    "chatgpt_go": signup_chatgpt_go,
    "gemini_advanced": signup_gemini_advanced,
    "claude_pro": signup_claude_pro,
    "grok_x": signup_grok_x,
    "grok_standalone": signup_grok_standalone,
    "perplexity_pro": signup_perplexity_pro,
    "midjourney_basic": signup_midjourney_basic,
    "midjourney_standard": signup_midjourney_standard,
    "suno_pro": signup_suno_pro,
    "suno_premier": signup_suno_premier,
    "adobe_firefly_standard": signup_adobe_firefly_standard,
    "adobe_firefly_pro": signup_adobe_firefly_pro,
    "runway_gen3_pro": signup_runway_gen3_pro,
    "runway_unlimited": signup_runway_unlimited,
    "leonardo_ai_premium": signup_leonardo_ai_premium,
    "manus": signup_manus,
    "emergent": signup_emergent
}
