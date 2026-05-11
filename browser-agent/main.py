import os
import base64
import json
import asyncio
import httpx
from fastapi import FastAPI, HTTPException, Request
from pydantic import BaseModel
from typing import Dict, Any, Optional
from browser_use import Agent, Browser, BrowserConfig
from langchain_openai import ChatOpenAI
from playwright.async_api import async_playwright

app = FastAPI()

# Configure LLM
api_key = os.getenv("DEEPSEEK_API_KEY")
llm = ChatOpenAI(
    model='deepseek-chat',
    openai_api_key=api_key,
    openai_api_base='https://api.deepseek.com',
)

browser_config = BrowserConfig(
    headless=True,
    disable_security=True,
)

GO_AGENT_URL = os.getenv("GO_AGENT_URL", "https://koola10.fly.dev")

async def emit_event(event_type: str, data: Dict[str, Any]):
    async with httpx.AsyncClient() as client:
        try:
            await client.post(
                f"{GO_AGENT_URL}/events/emit",
                json={
                    "type": event_type,
                    "data": data,
                    "timestamp": "" # Go agent fills this
                }
            )
        except Exception as e:
            print(f"Failed to emit event: {e}")

class NavigateRequest(BaseModel):
    url: str

class FormRequest(BaseModel):
    url: str
    form_data: Dict[str, str]

class ExtractRequest(BaseModel):
    url: str
    instruction: str

@app.get("/health")
async def health():
    return {"status": "ok"}

@app.post("/browser/stripe-live-keys")
async def stripe_live_keys():
    # This is an async background task because it takes time and requires interaction
    async def run_task():
        browser = Browser(config=browser_config)
        agent = Agent(
            task="Go to https://dashboard.stripe.com/login. Login with the credentials provided in the environment (STRIPE_EMAIL, STRIPE_PASSWORD). If you encounter a 2FA screen, STOP and wait for further instructions. Periodically take screenshots so the user can see the 2FA screen and provide the code.",
            llm=llm,
            browser=browser,
        )

        # Override browser-use to emit screenshots
        # This is a bit advanced, for now we will just run a loop and take screenshots manually

        task_future = asyncio.create_task(agent.run())

        while not task_future.done():
            await asyncio.sleep(5)
            try:
                context = await browser.get_context()
                page = await context.get_current_page()
                if page:
                    screenshot = await page.screenshot()
                    screenshot_b64 = base64.b64encode(screenshot).decode('utf-8')
                    await emit_event("browser_screenshot", {
                        "screenshot": screenshot_b64,
                        "url": page.url,
                        "title": await page.title()
                    })
            except Exception as e:
                print(f"Loop screenshot failed: {e}")

        result = await task_future
        await emit_event("browser_task_complete", {"result": str(result)})
        await browser.close()

    asyncio.create_task(run_task())
    return {"status": "started"}

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
    full_instruction = f"Go to {req.url}, " + ", ".join(instructions) + ". After filling everything, stay on the page so I can take a screenshot."

    browser = Browser(config=browser_config)
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
        "agent_result": str(result),
        "screenshot": screenshot_b64
    }

@app.post("/browser/submit-form")
async def submit_form(req: FormRequest):
    instructions = [f"Fill the field '{k}' with value '{v}'" for k, v in req.form_data.items()]
    full_instruction = f"Go to {req.url}, " + ", ".join(instructions) + ". Finally, find and click the submit button. Wait for a confirmation or success message."

    browser = Browser(config=browser_config)
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

    browser = Browser(config=browser_config)
    agent = Agent(
        task=full_instruction,
        llm=llm,
        browser=browser,
    )
    result = await agent.run()
    await browser.close()
    return {"data": str(result)}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
