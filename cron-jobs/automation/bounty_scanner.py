#!/usr/bin/env python3
"""
Koola10 Bounty Scanner v3.0
Scans GitHub for bounty opportunities, auto-generates PR descriptions,
claims bounties, and scaffolds starter solutions.
"""

import argparse
import json
import os
import smtplib
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
import re
import subprocess
import sys
import textwrap
import time
import urllib.error
import urllib.parse
import urllib.request
from datetime import datetime, timezone
from pathlib import Path


# ── Config ──────────────────────────────────────────────────────────────────

GITHUB_API = "https://api.github.com"
GITHUB_TOKEN = os.environ.get("GITHUB_TOKEN", "")
BOT_USERNAME = os.environ.get("GITHUB_BOT_USER", "koola10-ai")
WORK_DIR = Path(os.environ.get("BOUNTY_WORK_DIR", "/tmp/koola10-bounties"))

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
    "huggingface/transformers",
    "vllm-project/vllm",
    "ollama/ollama",
    "ggerganov/llama.cpp",
    "letta-ai/letta",
    "mem0ai/mem0",
    "BerriAI/litellm",
    "Aider-AI/aider",
    "All-Hands-AI/OpenHands",
    "camel-ai/camel",
    "ScrapeGraphAI/Scrapegraph-ai",
    "ComposioHQ/composio",
    "phidatahq/phidata",
    "jina-ai/reader",
    "TheSCInitiative/bounties",
    "projectdiscovery/oss-bounty-program",
    "zama-ai/bounty-program",
    "stacksgov/critical-bounties",
    "Concordium/Concordium-Free-Open-Grants-Program",
]

# Search queries to find bounties (unassigned preferred)
BOUNTY_QUERIES = [
    'label:bounty state:open no:assignee',
    '"bounty" in:title state:open is:issue no:assignee',
]


# ── Helpers ─────────────────────────────────────────────────────────────────

def api_get(path, params=None):
    """Make a GitHub API GET request."""
    url = f"{GITHUB_API}{path}"
    if params:
        url += "?" + urllib.parse.urlencode(params)

    headers = {
        "Accept": "application/vnd.github.v3+json",
        "User-Agent": "Koola10-BountyScanner/3.0",
    }
    if GITHUB_TOKEN:
        headers["Authorization"] = f"token {GITHUB_TOKEN}"

    req = urllib.request.Request(url, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            return json.loads(resp.read().decode())
    except urllib.error.HTTPError as e:
        if e.code == 403:
            print("  ⚠️  Rate limited — pausing 60s...")
            time.sleep(60)
            return api_get(path, params)
        print(f"  ❌ API error {e.code}: {e.reason}")
        return None
    except Exception as e:
        print(f"  ❌ Request failed: {e}")
        return None


def api_post(path, body):
    """Make a GitHub API POST request."""
    url = f"{GITHUB_API}{path}"
    headers = {
        "Accept": "application/vnd.github.v3+json",
        "Content-Type": "application/json",
        "User-Agent": "Koola10-BountyScanner/3.0",
    }
    if GITHUB_TOKEN:
        headers["Authorization"] = f"token {GITHUB_TOKEN}"

    data = json.dumps(body).encode("utf-8")
    req = urllib.request.Request(url, method="POST", headers=headers, data=data)
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            return json.loads(resp.read().decode())
    except urllib.error.HTTPError as e:
        body_text = ""
        try:
            body_text = e.read().decode()[:200]
        except Exception:
            pass
        print(f"  ❌ POST error {e.code}: {e.reason} — {body_text}")
        return None
    except Exception as e:
        print(f"  ❌ POST failed: {e}")
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
    assignee = issue.get("assignee")
    return {
        "repo": repo,
        "number": issue["number"],
        "title": issue["title"],
        "url": issue["html_url"],
        "amount": amount,
        "labels": labels,
        "created": issue["created_at"],
        "comments": issue.get("comments", 0),
        "assignee": assignee["login"] if assignee else None,
        "body": (issue.get("body") or "")[:2000],
    }


# ── PR Description Generator ───────────────────────────────────────────────

class PRGenerator:
    """Generates structured PR descriptions for bounty issues."""

    @staticmethod
    def generate(bounty):
        """Generate a full PR description for a bounty issue."""
        repo = bounty["repo"]
        number = bounty["number"]
        title = bounty["title"]
        amount = bounty.get("amount", 0)
        labels = bounty.get("labels", [])

        amt_str = f"${amount:,.0f}" if amount else "TBD"
        labels_str = ", ".join(labels) if labels else "none"

        pr_desc = (
            f"## Bounty Claim — {amt_str}\n"
            f"\n"
            f"**Issue:** [{repo}#{number}]({bounty['url']})\n"
            f"**Bounty:** {amt_str} USD\n"
            f"**Labels:** {labels_str}\n"
            f"\n"
            f"---\n"
            f"\n"
            f"### Summary\n"
            f"\n"
            f"This PR addresses [{repo}#{number}]({bounty['url']}) — \"{title}\"\n"
            f"\n"
            f"### What This PR Does\n"
            f"\n"
            f"> **Note:** This is an auto-generated starter PR. The implementation\n"
            f"> details should be filled in based on the specific issue requirements.\n"
            f"\n"
            f"- [ ] Reads and understands the bounty requirements\n"
            f"- [ ] Implements the requested feature/fix\n"
            f"- [ ] Adds tests if applicable\n"
            f"- [ ] Updates documentation if needed\n"
            f"- [ ] Verifies the solution works end-to-end\n"
            f"\n"
            f"### Technical Approach\n"
            f"\n"
            f"*To be filled in after reviewing the issue details and codebase.*\n"
            f"\n"
            f"### Testing\n"
            f"\n"
            f"- [ ] Unit tests added/updated\n"
            f"- [ ] Integration tests pass\n"
            f"- [ ] Manual testing completed\n"
            f"\n"
            f"### Checklist\n"
            f"\n"
            f"- [ ] Code follows the project's style guidelines\n"
            f"- [ ] Self-review completed\n"
            f"- [ ] Comments added for complex logic\n"
            f"- [ ] Documentation updated\n"
            f"- [ ] No new warnings introduced\n"
            f"\n"
            f"---\n"
            f"\n"
            f"**Bounty:** {amt_str} | **Auto-generated by:** [Koola10 AI Agent](https://koola10aiagent.freebuff.app)\n"
        )
        return pr_desc

    @staticmethod
    def generate_branch_name(bounty):
        """Generate a git branch name for the bounty."""
        number = bounty["number"]
        slug = re.sub(r'[^a-z0-9]+', '-', bounty["title"].lower())[:40].strip('-')
        return f"bounty/{number}-{slug}"


# ── Bounty Claimer ──────────────────────────────────────────────────────────

class BountyClaimer:
    """Claims bounties by commenting on issues and setting up local repos."""

    CLAIM_COMMENT = (
        "🤖 **Bounty Claimed by Koola10 AI Agent**\n"
        "\n"
        "I've claimed this bounty and will work on a solution.\n"
        "\n"
        "**My approach:**\n"
        "- Analyzing the requirements and codebase\n"
        "- Implementing the requested feature/fix\n"
        "- Adding tests and documentation\n"
        "- Submitting for review\n"
        "\n"
        "**ETA:** Within 24 hours\n"
        "\n"
        "---\n"
        "*Auto-claimed via [Koola10 Automation](https://koola10aiagent.freebuff.app)*\n"
    )

    def __init__(self, dry_run=False):
        self.dry_run = dry_run
        self.claims = []

    def claim_issue(self, bounty):
        """Comment on a GitHub issue to claim the bounty."""
        repo = bounty["repo"]
        number = bounty["number"]

        print(f"\n  🎯 Claiming: {repo}#{number}")

        if self.dry_run:
            print("     [DRY RUN] Would post claim comment")
            self.claims.append({"bounty": bounty, "status": "dry_run"})
            return True

        if not GITHUB_TOKEN:
            print("     ⚠️  No GITHUB_TOKEN — cannot claim (dry run only)")
            self.claims.append({"bounty": bounty, "status": "no_token"})
            return False

        result = api_post(f"/repos/{repo}/issues/{number}/comments", {
            "body": self.CLAIM_COMMENT
        })

        if result:
            comment_url = result.get("html_url", "")
            print(f"     ✅ Claimed — comment posted: {comment_url}")
            self.claims.append({"bounty": bounty, "status": "claimed", "comment_url": comment_url})
            return True
        else:
            print("     ❌ Failed to claim")
            self.claims.append({"bounty": bounty, "status": "failed"})
            return False

    def setup_repo(self, bounty):
        """Clone the bounty repo and create a working branch."""
        repo = bounty["repo"]
        branch = PRGenerator.generate_branch_name(bounty)

        repo_dir = WORK_DIR / repo.replace("/", "_")
        clone_url = f"https://github.com/{repo}.git"

        print(f"\n  📦 Setting up repo: {repo}")

        if self.dry_run:
            print(f"     [DRY RUN] Would clone {clone_url}")
            print(f"     [DRY RUN] Would create branch: {branch}")
            return str(repo_dir)

        WORK_DIR.mkdir(parents=True, exist_ok=True)

        if repo_dir.exists():
            print("     📂 Repo exists, fetching...")
            subprocess.run(["git", "fetch", "origin"], cwd=str(repo_dir), capture_output=True)
        else:
            print(f"     ⬇️  Cloning {clone_url}...")
            result = subprocess.run(
                ["git", "clone", "--depth=1", clone_url, str(repo_dir)],
                capture_output=True, text=True
            )
            if result.returncode != 0:
                print(f"     ❌ Clone failed: {result.stderr[:100]}")
                return None

        # Create and checkout branch
        subprocess.run(
            ["git", "checkout", "-b", branch],
            cwd=str(repo_dir), capture_output=True
        )
        print(f"     ✅ Branch created: {branch}")

        return str(repo_dir)

    def get_claims_summary(self):
        """Return a summary of all claims made."""
        return {
            "total": len(self.claims),
            "claimed": len([c for c in self.claims if c["status"] == "claimed"]),
            "dry_run": len([c for c in self.claims if c["status"] == "dry_run"]),
            "failed": len([c for c in self.claims if c["status"] == "failed"]),
            "details": self.claims,
        }


# ── Solution Scaffolder ─────────────────────────────────────────────────────

class SolutionScaffolder:
    """Generates starter code files for bounty issues."""

    LANG_PATTERNS = {
        "python": [r"python", r"\.py", r"django", r"flask", r"fastapi"],
        "typescript": [r"typescript", r"\.ts", r"next\.js", r"react"],
        "javascript": [r"javascript", r"\.js", r"node"],
        "rust": [r"rust", r"\.rs", r"cargo"],
        "go": [r"golang", r"\.go"],
        "solidity": [r"solidity", r"smart.?contract", r"web3"],
        "clarity": [r"clarity", r"stacks", r"sbtc"],
    }

    @staticmethod
    def detect_language(bounty):
        """Detect the primary language from issue context."""
        text = (bounty.get("title", "") + " " + bounty.get("body", "") + " " +
                " ".join(bounty.get("labels", []))).lower()

        for lang, patterns in SolutionScaffolder.LANG_PATTERNS.items():
            for pat in patterns:
                if re.search(pat, text):
                    return lang
        return "python"

    @staticmethod
    def scaffold(bounty, output_dir):
        """Generate starter files for the bounty."""
        lang = SolutionScaffolder.detect_language(bounty)
        repo = bounty["repo"]
        number = bounty["number"]
        title = bounty["title"]
        amount = bounty.get("amount", 0)
        body = bounty.get("body", "See issue for details.")[:500]
        amt_str = f"${amount:,.0f}" if amount else "TBD"

        output_path = Path(output_dir)
        output_path.mkdir(parents=True, exist_ok=True)

        files_created = []

        # README with bounty details
        readme_content = (
            f"# Bounty #{number}: {title}\n"
            f"\n"
            f"**Repository:** {repo}\n"
            f"**Amount:** {amt_str}\n"
            f"**URL:** {bounty['url']}\n"
            f"\n"
            f"## Requirements\n"
            f"\n"
            f"{body}\n"
            f"\n"
            f"## Solution\n"
            f"\n"
            f"<!-- Describe your solution here -->\n"
            f"\n"
            f"## Testing\n"
            f"\n"
            f"<!-- Describe how to test -->\n"
        )
        (output_path / "BOUNTY_README.md").write_text(readme_content)
        files_created.append("BOUNTY_README.md")

        # Language-specific starter
        if lang == "python":
            py_code = (
                "#!/usr/bin/env python3\n"
                '"""Solution for Bounty #' + str(number) + ': ' + title + '"""\n'
                "\n"
                "import json\n"
                "import sys\n"
                "\n"
                "\n"
                "def solve():\n"
                '    """Main solution entry point."""\n'
                "    # TODO: Implement the solution\n"
                '    print("Solution placeholder for Bounty #' + str(number) + '")\n'
                "    pass\n"
                "\n"
                "\n"
                'if __name__ == "__main__":\n'
                "    solve()\n"
            )
            (output_path / "solution.py").write_text(py_code)
            files_created.append("solution.py")

            test_code = (
                "#!/usr/bin/env python3\n"
                '"""Tests for Bounty #' + str(number) + ': ' + title + '"""\n'
                "\n"
                "from solution import solve\n"
                "\n"
                "\n"
                "def test_basic():\n"
                '    """Basic smoke test."""\n'
                "    # TODO: Add real tests\n"
                "    assert True\n"
                "\n"
                "\n"
                "def test_edge_cases():\n"
                '    """Edge case tests."""\n'
                "    # TODO: Add edge case tests\n"
                "    pass\n"
            )
            (output_path / "test_solution.py").write_text(test_code)
            files_created.append("test_solution.py")

        elif lang in ("typescript", "javascript"):
            ts_code = (
                "/**\n"
                " * Solution for Bounty #" + str(number) + ": " + title + "\n"
                " * Repository: " + repo + "\n"
                " */\n"
                "\n"
                "export function solve(): void {\n"
                "  // TODO: Implement the solution\n"
                '  console.log("Solution placeholder for Bounty #' + str(number) + '");\n'
                "}\n"
                "\n"
                "if (require.main === module) {\n"
                "  solve();\n"
                "}\n"
            )
            (output_path / "solution.ts").write_text(ts_code)
            files_created.append("solution.ts")

        elif lang == "rust":
            rs_code = (
                "//! Solution for Bounty #" + str(number) + ": " + title + "\n"
                "//! Repository: " + repo + "\n"
                "\n"
                "fn main() {\n"
                "    // TODO: Implement the solution\n"
                '    println!("Solution placeholder for Bounty #{}", ' + str(number) + ');\n'
                "}\n"
                "\n"
                "#[cfg(test)]\n"
                "mod tests {\n"
                "    use super::*;\n"
                "\n"
                "    #[test]\n"
                "    fn test_basic() {\n"
                "        // TODO: Add real tests\n"
                "        assert!(true);\n"
                "    }\n"
                "}\n"
            )
            (output_path / "solution.rs").write_text(rs_code)
            files_created.append("solution.rs")

        elif lang == "go":
            go_code = (
                "// Solution for Bounty #" + str(number) + ": " + title + "\n"
                "// Repository: " + repo + "\n"
                "\n"
                "package main\n"
                "\n"
                'import "fmt"\n'
                "\n"
                "func main() {\n"
                "    // TODO: Implement the solution\n"
                '    fmt.Println("Solution placeholder for Bounty #' + str(number) + '")\n'
                "}\n"
            )
            (output_path / "solution.go").write_text(go_code)
            files_created.append("solution.go")

        elif lang == "clarity":
            clar_code = (
                ";; Solution for Bounty #" + str(number) + ": " + title + "\n"
                ";; Repository: " + repo + "\n"
                "\n"
                "(define-public (solve)\n"
                "  (ok true))\n"
            )
            (output_path / "solution.clar").write_text(clar_code)
            files_created.append("solution.clar")

        print(f"     📝 Scaffolded {len(files_created)} files ({lang}): {', '.join(files_created)}")
        return files_created


# ── Scanner ─────────────────────────────────────────────────────────────────

class EmailAlerter:
    """Sends email alerts when high-value bounties are found."""

    def __init__(self):
        self.smtp_server = os.environ.get('SMTP_SERVER', 'smtp.gmail.com')
        self.smtp_port = int(os.environ.get('SMTP_PORT', '587'))
        self.smtp_user = os.environ.get('SMTP_USER', '')
        self.smtp_pass = os.environ.get('SMTP_PASS', '')
        self.alert_to = os.environ.get('ALERT_EMAIL', '')
        self.alert_from = os.environ.get('ALERT_FROM', self.smtp_user)

    def is_configured(self):
        return bool(self.smtp_user and self.smtp_pass and self.alert_to)

    def send_alert(self, subject, body, bounties=None):
        if not self.is_configured():
            print('[EMAIL] Not configured - skipping alert')
            return False
        try:
            msg = MIMEMultipart()
            msg['From'] = self.alert_from
            msg['To'] = self.alert_to
            msg['Subject'] = subject
            html = '<html><body style="font-family:monospace;background:#0a0a0a;color:#00f0ff;padding:20px">'
            html += '<h2 style="color:#39ff14">🤖 Koola10 Bounty Alert</h2>'
            html += '<div style="border:1px solid #00f0ff;padding:15px;border-radius:8px">' + body + '</div>'
            if bounties:
                html += '<h3 style="color:#8b00ff">💰 High-Value Bounties Found:</h3>'
                html += '<table style="border-collapse:collapse;width:100%">'
                for b in bounties:
                    amt = b.get('amount', 0)
                    html += '<tr style="border-bottom:1px solid #333">'
                    html += f'<td style="padding:8px;color:#39ff14">${amt:,.0f}</td>'
                    html += f'<td style="padding:8px">{b.get("repo", "")}</td>'
                    html += f'<td style="padding:8px"><a href="{b.get("url", "#")}" style="color:#00f0ff">{b.get("title", "")}</a></td>'
                    html += '</tr>'
                html += '</table>'
            html += '<hr style="border-color:#333"><p style="color:#666;font-size:11px">Auto-generated by Koola10 Bounty Scanner</p>'
            html += '</body></html>'
            msg.attach(MIMEText(html, 'html'))
            with smtplib.SMTP(self.smtp_server, self.smtp_port) as server:
                server.starttls()
                server.login(self.smtp_user, self.smtp_pass)
                server.send_message(msg)
            print(f'[EMAIL] Alert sent to {self.alert_to}')
            return True
        except Exception as e:
            print(f'[EMAIL] Error: {e}')
            return False

    def alert_high_value_bounties(self, bounties, threshold=500):
        high_value = [b for b in bounties if b.get('amount', 0) >= threshold]
        if not high_value:
            return
        subject = f'🤖 {len(high_value)} High-Value Bounties Found (>=${threshold})'
        body = f'<p>Found <strong style="color:#39ff14">{len(high_value)}</strong> bounties worth <strong style="color:#39ff14">${sum(b.get("amount", 0) for b in high_value):,.0f}</strong> total!</p>'
        self.send_alert(subject, body, high_value)


class BountyScanner:
    def __init__(self, min_bounty=0, unassigned_only=False):
        self.min_bounty = min_bounty
        self.unassigned_only = unassigned_only
        self.found = []
        self.scanned = 0

    def scan_repo(self, repo):
        """Scan a single repo for bounty issues."""
        issues = []
        params = {
            "labels": "bounty",
            "state": "open",
            "per_page": "20",
            "sort": "created",
            "direction": "desc",
        }
        result = api_get(f"/repos/{repo}/issues", params)
        if result and isinstance(result, list):
            issues.extend(result)

        result = api_get("/search/issues", {
            "q": f"repo:{repo} bounty in:title state:open",
            "per_page": "10",
        })
        if result and "items" in result:
            seen = {i["id"] for i in issues}
            for item in result["items"]:
                if item["id"] not in seen:
                    issues.append(item)

        return issues

    def scan_github_search(self):
        """Scan GitHub search for unassigned bounty issues."""
        results = []
        seen = set()

        for query in BOUNTY_QUERIES:
            result = api_get("/search/issues", {
                "q": query,
                "per_page": "30",
                "sort": "created",
                "direction": "desc",
            })
            if result and "items" in result:
                for item in result["items"]:
                    repo_url = item["repository_url"]
                    repo = "/".join(repo_url.split("/")[-2:])
                    key = f"{repo}#{item['number']}"
                    if key not in seen:
                        seen.add(key)
                        bounty = format_bounty(item, repo)
                        if bounty["amount"] >= self.min_bounty:
                            results.append(bounty)
            time.sleep(2)

        return results

    def scan_all(self):
        """Scan all target repos + GitHub search."""
        print(f"\n🔍 Koola10 Bounty Scanner v3.0")
        print(f"   Scanning {len(BOUNTY_REPOS)} repos + GitHub search...")
        print(f"   Min bounty: ${self.min_bounty}")
        print(f"   Unassigned only: {self.unassigned_only}")
        print()

        for i, repo in enumerate(BOUNTY_REPOS):
            self.scanned += 1
            print(f"  📋 [{i+1}/{len(BOUNTY_REPOS)}] Scanning: {repo}", end=" ")

            issues = self.scan_repo(repo)
            bounties = [format_bounty(iss, repo) for iss in issues]
            bounties = [b for b in bounties if b["amount"] >= self.min_bounty]
            if self.unassigned_only:
                bounties = [b for b in bounties if b["assignee"] is None]
            self.found.extend(bounties)

            print(f"→ {len(bounties)} bounties")
            time.sleep(0.5)

        print("\n  🔎 Scanning GitHub search (unassigned bounties)...")
        search_results = self.scan_github_search()
        existing_keys = {f"{b['repo']}#{b['number']}" for b in self.found}
        new_from_search = 0
        for b in search_results:
            key = f"{b['repo']}#{b['number']}"
            if key not in existing_keys:
                self.found.append(b)
                new_from_search += 1
        print(f"     → {len(search_results)} from search ({new_from_search} new)")

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

        for b in self.found[:20]:
            amt = f"${b['amount']:,.0f}" if b["amount"] else "TBD"
            claimed = " [CLAIMED]" if b.get("assignee") else ""
            print(f"\n  💰 {amt}{claimed} — {b['repo']}#{b['number']}")
            print(f"     {b['title'][:80]}")
            print(f"     {b['url']}")
            if b["labels"]:
                print(f"     Labels: {', '.join(b['labels'][:5])}")

        if len(self.found) > 20:
            print(f"\n  ... and {len(self.found) - 20} more")

    def to_json(self):
        """Return results as JSON."""
        return json.dumps(self.found, indent=2)


# ── CLI ─────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(
        description="Koola10 Bounty Scanner v3.0 — scan, claim, and generate PRs"
    )
    parser.add_argument("--min-bounty", type=float, default=0,
                        help="Minimum bounty amount (default: 0)")
    parser.add_argument("--unassigned", action="store_true",
                        help="Only show unassigned bounties")
    parser.add_argument("--claim", action="store_true",
                        help="Auto-claim the top unclaimed bounty")
    parser.add_argument("--claim-all", action="store_true",
                        help="Auto-claim ALL unclaimed bounties matching criteria")
    parser.add_argument("--generate-pr", action="store_true",
                        help="Generate PR description for top bounty")
    parser.add_argument("--scaffold", action="store_true",
                        help="Generate starter code for top bounty")
    parser.add_argument("--email", action="store_true",
                        help="Send email alerts for high-value bounties")
    parser.add_argument("--threshold", type=float, default=500,
                        help="Minimum bounty amount for email alerts (default: 500)")
    parser.add_argument("--dry-run", action="store_true",
                        help="Show what would be done without doing it")
    parser.add_argument("--json", action="store_true",
                        help="Output results as JSON")
    args = parser.parse_args()

    # Scan
    scanner = BountyScanner(min_bounty=args.min_bounty, unassigned_only=args.unassigned)
    scanner.scan_all()

    if args.json:
        print(scanner.to_json())
    else:
        scanner.print_results()

    # Send email alerts if requested
    if args.email and scanner.found:
        alerter = EmailAlerter()
        if alerter.is_configured():
            alerter.alert_high_value_bounties(scanner.found, args.threshold)
        else:
            print('[EMAIL] Not configured. Set SMTP_USER, SMTP_PASS, ALERT_EMAIL env vars.')

    if not scanner.found:
        print("\n  No bounties to act on.")
        return

    # Filter to unclaimed for actions
    unclaimed = [b for b in scanner.found if not b.get("assignee")]

    # Generate PR description
    if args.generate_pr:
        target = unclaimed[0] if unclaimed else scanner.found[0]
        print(f"\n{'='*60}")
        print(f"📄 PR Description for: {target['repo']}#{target['number']}")
        print(f"{'='*60}")
        pr_desc = PRGenerator.generate(target)
        print(pr_desc)

        pr_file = WORK_DIR / f"pr-{target['number']}.md"
        WORK_DIR.mkdir(parents=True, exist_ok=True)
        pr_file.write_text(pr_desc)
        print(f"  💾 Saved to: {pr_file}")

    # Claim bounty
    if args.claim:
        if not unclaimed:
            print("\n  ⚠️  No unclaimed bounties found.")
            return
        target = unclaimed[0]
        claimer = BountyClaimer(dry_run=args.dry_run)
        claimer.claim_issue(target)

        if args.scaffold:
            claimer.setup_repo(target)
            SolutionScaffolder.scaffold(target, WORK_DIR / str(target["number"]))

    # Claim all
    if args.claim_all:
        claimer = BountyClaimer(dry_run=args.dry_run)
        claimed_count = 0
        for b in unclaimed[:5]:
            success = claimer.claim_issue(b)
            if success:
                claimed_count += 1
                time.sleep(2)
        print(f"\n  📊 Claimed {claimed_count}/{min(len(unclaimed), 5)} bounties")

    # Scaffold
    if args.scaffold and not args.claim:
        target = unclaimed[0] if unclaimed else scanner.found[0]
        scaffold_dir = WORK_DIR / str(target["number"])
        SolutionScaffolder.scaffold(target, scaffold_dir)
        print(f"\n  📂 Scaffolded solution in: {scaffold_dir}")


if __name__ == "__main__":
    main()
