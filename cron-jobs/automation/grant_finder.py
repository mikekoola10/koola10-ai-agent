#!/usr/bin/env python3
"""
Koola10 Grant Opportunity Finder v2.0
Scans for AI/SaaS business grants and funding opportunities.
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

# Grant sources to scan
GRANT_SOURCES = {
    "developer_bounties": {
        "name": "Developer Bounties",
        "repos": [
            "TheSCInitiative/bounties",
            "projectdiscovery/oss-bounty-program",
            "zama-ai/bounty-program",
            "stacksgov/critical-bounties",
            "Concordium/Concordium-Free-Open-Grants-Program",
        ],
    },
    "crypto_grants": {
        "name": "Crypto/Web3 Grants",
        "repos": [
            "filecoin-project/devgrants",
            "scrtlabs/Grants",
            "PolymeshAssociation/Grants-Program",
            "stacksgov/decentralized-grants",
            "informalsystems/Open-Grants-Program",
        ],
    },
    "ai_funding": {
        "name": "AI/SaaS Funding",
        "queries": [
            '"grant" in:body state:open is:issue',
            '"funding" in:body state:open is:issue',
            '"bounty" in:title state:open is:issue',
        ],
        "repos": [
            "microsoft/autogen",
            "langchain-ai/langchain",
            "huggingface/transformers",
        ],
    },
}

# Keywords that indicate grant opportunities
GRANT_KEYWORDS = [
    "grant", "funding", "sponsor", "awards", "prize",
    "fellowship", "scholarship", "accelerator", "incubator",
    "seed fund", "micro-grant", "development fund",
]


# ── Helpers ─────────────────────────────────────────────────────────────────

def api_get(path, params=None):
    """Make a GitHub API GET request."""
    url = f"{GITHUB_API}{path}"
    if params:
        url += "?" + urllib.parse.urlencode(params)

    headers = {
        "Accept": "application/vnd.github.v3+json",
        "User-Agent": "Koola10-GrantFinder/2.0",
    }
    if GITHUB_TOKEN:
        headers["Authorization"] = f"token {GITHUB_TOKEN}"

    req = urllib.request.Request(url, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            return json.loads(resp.read().decode())
    except urllib.error.HTTPError as e:
        if e.code == 403:
            print(f"  ⚠️  Rate limited — waiting 60s...")
            time.sleep(60)
            return api_get(path, params)
        print(f"  ❌ API error {e.code}")
        return None
    except Exception as e:
        print(f"  ❌ Request failed: {e}")
        return None


def extract_amount(text):
    """Extract funding amount from text."""
    patterns = [
        r"\$(\d[\d,]*(?:\.\d{2})?)\s*(?:USD|grant|award|prize|funding)",
        r"(\d[\d,]*)\s*(?:USD|grant|award)",
        r"grant[:\s]*\$(\d[\d,]*(?:\.\d{2})?)",
        r"up to \$(\d[\d,]*(?:\.\d{2})?)",
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


def format_opportunity(issue, source):
    """Format a grant opportunity."""
    text = issue.get("title", "") + " " + issue.get("body", "")
    amount = extract_amount(text)
    return {
        "source": source,
        "title": issue.get("title", ""),
        "url": issue.get("html_url", ""),
        "amount": amount,
        "created": issue.get("created_at", ""),
        "labels": [l["name"] for l in issue.get("labels", [])],
    }


# ── Scanner ─────────────────────────────────────────────────────────────────

class GrantFinder:
    def __init__(self, min_amount=0):
        self.min_amount = min_amount
        self.found = []

    def scan_source(self, key, source):
        """Scan a single grant source."""
        print(f"\n  🔍 Scanning: {source['name']}")

        # Search across repos
        for repo in source.get("repos", []):
            print(f"     📋 Repository: {repo}", end=" ")
            issues = []
            result = api_get(f"/repos/{repo}/issues", {
                "state": "open",
                "per_page": "30",
                "sort": "created",
                "direction": "desc",
            })
            if result and isinstance(result, list):
                issues.extend(result)
            print(f"→ {len(issues)} issues")
            time.sleep(0.5)

            for issue in issues:
                opp = format_opportunity(issue, source["name"])
                if opp["amount"] >= self.min_amount:
                    self.found.append(opp)

        # Run queries
        for query in source.get("queries", []):
            print(f"     🔎 Query: {query}", end=" ")
            result = api_get("/search/issues", {
                "q": f"{query}",
                "per_page": "20",
            })
            if result and "items" in result:
                for item in result["items"]:
                    opp = format_opportunity(item, source["name"])
                    if opp["amount"] >= self.min_amount:
                        self.found.append(opp)
                print(f"→ {len(result['items'])} results")
            else:
                print("→ 0 results")
            time.sleep(0.5)

    def scan_all(self):
        """Scan all grant sources."""
        print(f"\n💰 Koola10 Grant Opportunity Finder v2.0")
        print(f"   Scanning {len(GRANT_SOURCES)} sources...")
        print(f"   Min amount: ${self.min_amount}")

        for key, source in GRANT_SOURCES.items():
            self.scan_source(key, source)

        # Deduplicate by URL
        seen = set()
        deduped = []
        for opp in self.found:
            if opp["url"] not in seen:
                seen.add(opp["url"])
                deduped.append(opp)
        self.found = deduped

        # Sort by amount
        self.found.sort(key=lambda x: x["amount"], reverse=True)
        return self.found

    def print_results(self):
        """Print formatted results."""
        print(f"\n{'='*60}")
        print(f"✅ Grant Scan Complete — {len(self.found)} opportunities")
        print(f"{'='*60}")

        if not self.found:
            print("  No grant opportunities found.")
            return

        for g in self.found[:15]:
            amt = f"${g['amount']:,.0f}" if g['amount'] else "TBD"
            print(f"\n  💰 {amt} — {g['source']}")
            print(f"     {g['title'][:80]}")
            print(f"     {g['url']}")

    def to_json(self):
        """Return results as JSON."""
        return json.dumps(self.found, indent=2)


# ── CLI ─────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(description="Koola10 Grant Finder")
    parser.add_argument("--min-amount", type=float, default=0,
                        help="Minimum grant amount (default: 0)")
    parser.add_argument("--json", action="store_true",
                        help="Output as JSON")
    args = parser.parse_args()

    finder = GrantFinder(min_amount=args.min_amount)
    finder.scan_all()

    if args.json:
        print(finder.to_json())
    else:
        finder.print_results()


if __name__ == "__main__":
    main()
