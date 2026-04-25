import os
import base64
from fastapi import FastAPI, HTTPException
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
    browser = Browser(config=browser_config)
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
