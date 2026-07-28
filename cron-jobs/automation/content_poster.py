#!/usr/bin/env python3
"""
Koola10 Content Auto-Poster v2.0
Posts articles to Medium, Dev.to, and LinkedIn.
"""

import argparse
import json
import os
import sys
import urllib.error
import urllib.request
from datetime import datetime


# ── Config ──────────────────────────────────────────────────────────────────

MEDIUM_TOKEN = os.environ.get("MEDIUM_TOKEN", "")
DEVTO_API_KEY = os.environ.get("DEVTO_API_KEY", "")
LINKEDIN_TOKEN = os.environ.get("LINKEDIN_TOKEN", "")
LINKEDIN_PERSON_ID = os.environ.get("LINKEDIN_PERSON_ID", "")


# ── Platforms ───────────────────────────────────────────────────────────────

class MediumPoster:
    """Post to Medium via their API."""

    def post(self, title, body, tags=None, dry_run=False):
        if not MEDIUM_TOKEN:
            return {"platform": "medium", "status": "skipped", "reason": "No MEDIUM_TOKEN"}

        if dry_run:
            return {"platform": "medium", "status": "dry-run", "title": title}

        payload = {
            "title": title,
            "contentFormat": "markdown",
            "content": body,
            "publishStatus": "public",
        }
        if tags:
            payload["tags"] = tags[:5]

        try:
            req = urllib.request.Request(
                "https://api.medium.com/v1/posts",
                data=json.dumps(payload).encode(),
                headers={
                    "Content-Type": "application/json",
                    "Authorization": f"Bearer {MEDIUM_TOKEN}",
                },
                method="POST",
            )
            with urllib.request.urlopen(req, timeout=30) as resp:
                data = json.loads(resp.read().decode())
                return {
                    "platform": "medium",
                    "status": "published",
                    "url": data.get("data", {}).get("url", ""),
                }
        except Exception as e:
            return {"platform": "medium", "status": "error", "error": str(e)}


class DevtoPoster:
    """Post to Dev.to via their API."""

    def post(self, title, body, tags=None, dry_run=False):
        if not DEVTO_API_KEY:
            return {"platform": "devto", "status": "skipped", "reason": "No DEVTO_API_KEY"}

        if dry_run:
            return {"platform": "devto", "status": "dry-run", "title": title}

        payload = {
            "article": {
                "title": title,
                "published": True,
                "body_markdown": body,
            }
        }
        if tags:
            payload["article"]["tags"] = tags[:4]

        try:
            req = urllib.request.Request(
                "https://dev.to/api/articles",
                data=json.dumps(payload).encode(),
                headers={
                    "Content-Type": "application/json",
                    "api-key": DEVTO_API_KEY,
                },
                method="POST",
            )
            with urllib.request.urlopen(req, timeout=30) as resp:
                data = json.loads(resp.read().decode())
                return {
                    "platform": "devto",
                    "status": "published",
                    "url": data.get("url", ""),
                }
        except Exception as e:
            return {"platform": "devto", "status": "error", "error": str(e)}


class LinkedInPoster:
    """Post to LinkedIn via their API."""

    def post(self, title, body, dry_run=False):
        if not LINKEDIN_TOKEN:
            return {"platform": "linkedin", "status": "skipped", "reason": "No LINKEDIN_TOKEN"}

        if dry_run:
            return {"platform": "linkedin", "status": "dry-run", "title": title}

        payload = {
            "author": LINKEDIN_PERSON_ID,
            "lifecycleState": "PUBLISHED",
            "specificContent": {
                "com.linkedin.ugc.ShareContent": {
                    "shareCommentary": {"text": f"{title}\n\n{body[:500]}..."},
                    "shareMediaCategory": "NONE",
                }
            },
            "visibility": {"com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC"},
        }

        try:
            req = urllib.request.Request(
                "https://api.linkedin.com/v2/ugcPosts",
                data=json.dumps(payload).encode(),
                headers={
                    "Content-Type": "application/json",
                    "Authorization": f"Bearer {LINKEDIN_TOKEN}",
                },
                method="POST",
            )
            with urllib.request.urlopen(req, timeout=30) as resp:
                return {
                    "platform": "linkedin",
                    "status": "published",
                }
        except Exception as e:
            return {"platform": "linkedin", "status": "error", "error": str(e)}


# ── Main ────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(description="Koola10 Content Auto-Poster")
    parser.add_argument("--title", required=True, help="Article title")
    parser.add_argument("--body", help="Article body text")
    parser.add_argument("--body-file", help="Path to markdown file")
    parser.add_argument("--tags", nargs="*", help="Tags for the article")
    parser.add_argument("--dry-run", action="store_true", help="Preview only")
    parser.add_argument("--platform", choices=["medium", "devto", "linkedin", "all"],
                        default="all", help="Target platform")
    args = parser.parse_args()

    # Read body
    body = args.body or ""
    if args.body_file:
        with open(args.body_file, "r") as f:
            body = f.read()

    if not body:
        print("❌ Error: --body or --body-file required")
        sys.exit(1)

    print(f"\n📤 Koola10 Content Auto-Poster v2.0")
    print(f"   Title: {args.title}")
    print(f"   Tags: {args.tags or 'none'}")
    print(f"   Dry run: {args.dry_run}")
    print()

    results = []

    # Post to platforms
    platforms = {
        "medium": MediumPoster(),
        "devto": DevtoPoster(),
        "linkedin": LinkedInPoster(),
    }

    targets = [args.platform] if args.platform != "all" else platforms.keys()

    for name in targets:
        poster = platforms.get(name)
        if poster:
            print(f"  📤 Posting to {name}...", end=" ")
            result = poster.post(args.title, body, args.tags, args.dry_run)
            results.append(result)
            print(f"→ {result['status']}")

    # Summary
    print(f"\n{'='*60}")
    print(f"✅ Content Auto-Poster Complete")
    print(f"{'='*60}")
    for r in results:
        status = r["status"]
        icon = "✅" if status == "published" else "⏭️" if status == "skipped" else "❌"
        print(f"  {icon} {r['platform']}: {status}")
        if r.get("url"):
            print(f"     {r['url']}")

    return results


if __name__ == "__main__":
    main()
