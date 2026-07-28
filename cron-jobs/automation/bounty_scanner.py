#!/usr/bin/env python3
"""
Koola10 Bounty Scanner v2.0
Scans GitHub for bounty opportunities across repositories.
"""

import argparse
import json
import os
import re
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from datetime import datetime, timezone


# ── Config ──────────────────────────────────────────────────────────────────

GITHUB_API = "https://api.github.com"
GITHUB_TOKEN = os.environ.get("GITHUB_TOKEN", "")

# Repos known to have bounty programs
BOUNTY_REPOS = [
    "open-webui/open-webui",
    "langchain-ai/langchain",
    "run-llama/llama_index",
    "microsoft/autogen",
    "crewAIInc/crewAI",
    "Significant-Gravitas/AutoGPT",
    "Pythagora-io/gpt-pilot",
    "lobehub/lobe-chat",
    "danny-avila/LibreChat",
    "chatanywhere/GPT_API_free",
    "friyiaan/Discord-ChatGPT-Bot",
    "transitive-bullshit/chatgpt-api",
    "nicepkg/Aide",
    "open-webui/open-webui",
    "huggingface/transformers",
    "vllm-project/vllm",
    "ollama/ollama",
    "ggerganov/llama.cpp",
    "ggerganov/whisper.cpp",
    "letta-ai/letta",
    "mem0ai/mem0",
    "embedchain/embedchain",
    "chatchat-space/Langchain-Chatchat",
    "instructor-ai/instructor",
    "BerriAI/litellm",
    "Aider-AI/aider",
    "All-Hands-AI/OpenHands",
    "camel-ai/camel",
    "ScrapeGraphAI/Scrapegraph-ai",
    "ComposioHQ/composio",
    "phidatahq/phidata",
    "jina-ai/reader",
    "jina-ai/node-DeepResearch",
]

# Search queries to find bounties
BOUNTY_QUERIES = [
    "label:bounty state:open",
    '"bounty" in:title state:open',
    '"$500" in:body state:open',
    '"$1000" in:body state:open',
    '"$2000" in:body state:open',
    '"bounty:" in:body state:open is:issue',
]


# ── Helpers ─────────────────────────────────────────────────────────────────

def api_get(path, params=None):
    """Make a GitHub API GET request."""
    url = f"{GITHUB_API}{path}"
    if params:
        url += "?" + urllib.parse.urlencode(params)

    headers = {
        "Accept": "application/vnd.github.v3+json",
        "User-Agent": "Koola10-BountyScanner/2.0",
    }
    if GITHUB_TOKEN:
        headers["Authorization"] = f"token {GITHUB_TOKEN}"

    req = urllib.request.Request(url, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            return json.loads(resp.read().decode())
    except urllib.error.HTTPError as e:
        if e.code == 403:
            print(f"  ⚠️  Rate limited — pausing 60s...")
            time.sleep(60)
            return api_get(path, params)
        print(f"  ❌ API error {e.code}: {e.reason}")
        return None
    except Exception as e:
        print(f"  ❌ Request failed: {e}")
        return None


def extract_bounty_amount(text):
    """Extract dollar amount from issue text."""
    patterns = [
        r"\$(\d[\d,]*(?:\.\d{2})?)\s*(?:USD|bounty|reward|prize)",
        r"bounty[:\s]*\$(\d[\d,]*(?:\.\d{2})?)",
        r"reward[:\s]*\$(\d[\d,]*(?:\.\d{2})?)",
        r"\$(\d[\d,]*(?:\.\d{2})?)",
    ]
    amounts = []
    for pattern in patterns:
        matches = re.findall(pattern, text, re.IGNORECASE)
        for m in matches:
            try:
                amounts.append(float(m.replace(",", "")))
            except ValueError:
                pass
    return max(amounts) if amounts else 0


def format_bounty(issue, repo):
    """Format a bounty for display."""
    amount = extract_bounty_amount(issue.get("body", "") + issue.get("title", ""))
    labels = [l["name"] for l in issue.get("labels", [])]
    return {
        "repo": repo,
        "number": issue["number"],
        "title": issue["title"],
        "url": issue["html_url"],
        "amount": amount,
        "labels": labels,
        "created": issue["created_at"],
        "comments": issue.get("comments", 0),
    }


# ── Scanner ─────────────────────────────────────────────────────────────────

class BountyScanner:
    def __init__(self, min_bounty=0):
        self.min_bounty = min_bounty
        self.found = []
        self.scanned = 0

    def scan_repo(self, repo):
        """Scan a single repo for bounty issues."""
        issues = []
        # Search for bounty label
        result = api_get(f"/repos/{repo}/issues", {
            "labels": "bounty",
            "state": "open",
            "per_page": "20",
            "sort": "created",
            "direction": "desc",
        })
        if result and isinstance(result, list):
            issues.extend(result)

        # Search for bounty keyword in title
        result = api_get(f"/search/issues", {
            "q": f"repo:{repo} bounty in:title state:open",
            "per_page": "10",
        })
        if result and "items" in result:
            seen = {i["id"] for i in issues}
            for item in result["items"]:
                if item["id"] not in seen:
                    issues.append(item)

        return issues

    def scan_all(self):
        """Scan all target repos."""
        print(f"\n🔍 Koola10 Bounty Scanner v2.0")
        print(f"   Scanning {len(BOUNTY_REPOS)} targets...")
        print(f"   Min bounty: ${self.min_bounty}")
        print()

        for i, repo in enumerate(BOUNTY_REPOS):
            self.scanned += 1
            print(f"  📋 [{i+1}/{len(BOUNTY_REPOS)}] Scanning: {repo}", end=" ")

            issues = self.scan_repo(repo)
            bounties = [format_bounty(i, repo) for i in issues]
            bounties = [b for b in bounties if b["amount"] >= self.min_bounty]
            self.found.extend(bounties)

            print(f"→ {len(bounties)} bounties")
            time.sleep(0.5)  # Rate limit respect

        # Sort by amount descending
        self.found.sort(key=lambda x: x["amount"], reverse=True)
        return self.found

    def print_results(self):
        """Print formatted results."""
        print(f"\n{'='*60}")
        print(f"✅ Bounty Scan Complete — {len(self.found)} opportunities")
        print(f"{'='*60}")

        if not self.found:
            print("  No bounties found matching criteria.")
            return

        for b in self.found[:20]:  # Top 20
            amt = f"${b['amount']:,.0f}" if b['amount'] else "TBD"
            print(f"\n  💰 {amt} — {b['repo']}#{b['number']}")
            print(f"     {b['title'][:80]}")
            print(f"     {b['url']}")
            if b['labels']:
                print(f"     Labels: {', '.join(b['labels'][:5])}")

        if len(self.found) > 20:
            print(f"\n  ... and {len(self.found) - 20} more")

    def to_json(self):
        """Return results as JSON."""
        return json.dumps(self.found, indent=2)


# ── CLI ─────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(description="Koola10 Bounty Scanner")
    parser.add_argument("--min-bounty", type=float, default=0,
                        help="Minimum bounty amount (default: 0)")
    parser.add_argument("--json", action="store_true",
                        help="Output as JSON")
    args = parser.parse_args()

    scanner = BountyScanner(min_bounty=args.min_bounty)
    scanner.scan_all()

    if args.json:
        print(scanner.to_json())
    else:
        scanner.print_results()


if __name__ == "__main__":
    main()
