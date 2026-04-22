from tools.registry import ToolRegistry
from tools.web_browser import web_search_tool
from tools.github_tool import github_tool
from tools.file_tool import file_tool

registry = ToolRegistry()

registry.register("web_search", web_search_tool)
registry.register("github", github_tool)
registry.register("file", file_tool)
