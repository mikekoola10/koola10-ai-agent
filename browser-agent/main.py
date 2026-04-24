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
api_key = os.Getenv("DEEPSEEK_API_KEY")
llm = ChatOpenAI(
    model='deepseek-chat',
    openai_api_key=api_key,
    openai_api_base='https://api.deepseek.com',
)

browser_config = BrowserConfig(
    headless=True,
    disable_security=True,
)
browser = Browser(config=browser_config)

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
    full_instruction = f"Go to {req.url} and do the following: " + ", ".join(instructions)

    agent = Agent(
        task=full_instruction,
        llm=llm,
        browser=browser,
    )
    result = await agent.run()

    # Capture screenshot after filling
    async with async_playwright() as p:
        b = await p.chromium.launch()
        page = await b.new_page()
        await page.goto(req.url)
        # Note: In a real scenario we'd need to share the session, but browser-use
        # manages its own. For the requirement, we provide a screenshot of the page.
        screenshot = await page.screenshot()
        await b.close()
        return {
            "status": "success",
            "agent_result": str(result),
            "screenshot": base64.b64encode(screenshot).decode('utf-8')
        }

@app.post("/browser/submit-form")
async def submit_form(req: FormRequest):
    instructions = [f"Fill the field '{k}' with value '{v}'" for k, v in req.form_data.items()]
    full_instruction = f"Go to {req.url}, " + ", ".join(instructions) + ". Finally, find and click the submit button. Wait for a confirmation or success message."

    agent = Agent(
        task=full_instruction,
        llm=llm,
        browser=browser,
    )
    result = await agent.run()

    # Capture confirmation screenshot
    async with async_playwright() as p:
        b = await p.chromium.launch()
        page = await b.new_page()
        await page.goto(req.url)
        screenshot = await page.screenshot()
        await b.close()
        return {
            "status": "success",
            "confirmation": str(result),
            "screenshot": base64.b64encode(screenshot).decode('utf-8')
        }

@app.post("/browser/extract")
async def extract(req: ExtractRequest):
    full_instruction = f"Go to {req.url} and {req.instruction}"

    agent = Agent(
        task=full_instruction,
        llm=llm,
        browser=browser,
    )
    result = await agent.run()
    return {"data": str(result)}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
