#!/usr/bin/env python3
"""
Koola10 Competitor Monitor v2.0
Tracks pricing and feature changes on competitor websites.
"""

import argparse
import hashlib
import json
import os
import sys
import time
import urllib.error
import urllib.request
from datetime import datetime, timezone


# ── Config ──────────────────────────────────────────────────────────────────

STATE_FILE = os.path.join(os.path.dirname(__file__), ".competitor_state.json")

# Default competitors to monitor
DEFAULT_URLS = [
    ("Vercel Pricing", "https://vercel.com/pricing"),
    ("Netlify Pricing", "https://www.netlify.com/pricing/"),
    ("Railway Pricing", "https://railway.com/pricing"),
    ("Fly.io Pricing", "https://fly.io/docs/about/pricing/"),
    ("Render Pricing", "https://render.com/pricing"),
    ("Supabase Pricing", "https://supabase.com/pricing"),
    ("PlanetScale Pricing", "https://planetscale.com/pricing"),
    ("Firebase Pricing", "https://firebase.google.com/pricing"),
    ("Cloudflare Pages", "https://pages.cloudflare.com/"),
    ("Replit Pricing", "https://replit.com/pricing"),
]


# ── Helpers ─────────────────────────────────────────────────────────────────

def load_state():
    """Load previous state for change detection."""
    try:
        with open(STATE_FILE, "r") as f:
            return json.load(f)
    except (FileNotFoundError, json.JSONDecodeError):
        return {}


def save_state(state):
    """Save current state."""
    with open(STATE_FILE, "w") as f:
        json.dump(state, f, indent=2)


def fetch_page(url, timeout=15):
    """Fetch a page and return its content hash + snippet."""
    headers = {
        "User-Agent": "Mozilla/5.0 (compatible; Koola10-Monitor/2.0)",
        "Accept": "text/html",
    }
    req = urllib.request.Request(url, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            content = resp.read().decode(errors="replace")
            # Extract pricing-relevant text (rough)
            text = content[:50000]  # Limit size
            content_hash = hashlib.md5(text.encode()).hexdigest()
            # Extract text content (strip HTML tags roughly)
            import re
            text = re.sub(r"<[^>]+>", " ", text)
            text = re.sub(r"\s+", " ", text).strip()
            return {
                "status": "ok",
                "hash": content_hash,
                "length": len(content),
                "snippet": text[:500],
            }
    except Exception as e:
        return {"status": "error", "error": str(e)}


# ── Monitor ─────────────────────────────────────────────────────────────────

class CompetitorMonitor:
    def __init__(self, urls=None):
        self.urls = urls or DEFAULT_URLS
        self.results = []
        self.changes = []
        self.state = load_state()

    def check_all(self):
        """Check all competitor URLs."""
        print(f"\n🔍 Koola10 Competitor Monitor v2.0")
        print(f"   Monitoring {len(self.urls)} targets...")
        print()

        for name, url in self.urls:
            print(f"  📋 Checking: {name}", end=" ")
            result = fetch_page(url)
            result["name"] = name
            result["url"] = url
            self.results.append(result)

            if result["status"] == "ok":
                # Check for changes
                prev = self.state.get(url, {})
                if prev.get("hash") != result["hash"]:
                    if prev.get("hash"):  # Had a previous state
                        self.changes.append({
                            "name": name,
                            "url": url,
                            "old_hash": prev.get("hash", ""),
                            "new_hash": result["hash"],
                            "change_type": "content_changed",
                        })
                        print(f"⚠️  CHANGED!")
                    else:
                        print(f"✅ (new baseline)")
                else:
                    print(f"✅ (no change)")

                # Update state
                self.state[url] = {
                    "hash": result["hash"],
                    "length": result["length"],
                    "last_check": datetime.now(timezone.utc).isoformat(),
                }

            time.sleep(0.5)

        save_state(self.state)
        return self.results

    def print_results(self):
        """Print formatted results."""
        print(f"\n{'='*60}")
        print(f"✅ Competitor Monitor Complete")
        print(f"{'='*60}")

        # Changes
        if self.changes:
            print(f"\n⚠️  {len(self.changes)} CHANGES DETECTED:")
            for c in self.changes:
                print(f"\n  🔔 {c['name']}")
                print(f"     {c['url']}")
                print(f"     Change: {c['change_type']}")
        else:
            print(f"\n  ✅ No changes detected across {len(self.urls)} competitors")

        # Status summary
        ok = sum(1 for r in self.results if r["status"] == "ok")
        err = len(self.results) - ok
        print(f"\n  Status: {ok} checked, {err} errors")

    def to_json(self):
        """Return results as JSON."""
        return json.dumps({
            "results": self.results,
            "changes": self.changes,
            "timestamp": datetime.now(timezone.utc).isoformat(),
        }, indent=2)


# ── CLI ─────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(description="Koola10 Competitor Monitor")
    parser.add_argument("--urls", nargs="*", help="Custom URLs to monitor")
    parser.add_argument("--json", action="store_true", help="Output as JSON")
    args = parser.parse_args()

    urls = DEFAULT_URLS
    if args.urls:
        urls = [("Custom", u) for u in args.urls]

    monitor = CompetitorMonitor(urls=urls)
    monitor.check_all()

    if args.json:
        print(monitor.to_json())
    else:
        monitor.print_results()


if __name__ == "__main__":
    main()
