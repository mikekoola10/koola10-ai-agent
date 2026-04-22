def web_search_tool(payload):
    query = payload.get("query")
    return {
        "results": [
            {"title": f"Result for {query}", "url": f"https://example.com/search?q={query}"},
            {"title": "MetaClaw Agent Framework", "url": "https://metaclaw.ai"}
        ]
    }
