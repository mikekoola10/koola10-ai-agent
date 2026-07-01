import os
import logging
from browser_use import Agent, Browser, BrowserProfile
from langchain_openai import ChatOpenAI

# Configure logging
logger = logging.getLogger(__name__)

def get_llm():
    # Configure LLM lazily to avoid error if DEEPSEEK_API_KEY is not set during import
    api_key = os.getenv("DEEPSEEK_API_KEY") or "placeholder"
    return ChatOpenAI(
        model='deepseek-chat',
        openai_api_key=api_key,
        openai_api_base='https://api.deepseek.com',
    )

def get_browser_profile():
    return BrowserProfile(
        headless=os.getenv("BROWSER_HEADLESS", "true").lower() == "true",
        disable_security=True,
    )
