#!/usr/bin/env python3
"""
Koola10 Automation Orchestrator v2.0
Runs all proactive automation scanners and generates reports.
"""

import argparse
import json
import os
import sys
import time
from datetime import datetime, timezone

# Add parent to path for imports
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from bounty_scanner import BountyScanner
from competitor_monitor import CompetitorMonitor
from grant_finder import GrantFinder

# ── Config ──────────────────────────────────────────────────────────────────

SCANNERS = {
    "bounty": {
        "name": "Bounty Scanner",
        "class": BountyScanner,
        "kwargs": {"min_bounty": 200, "unassigned_only": True},
    },
    "competitor": {
        "name": "Competitor Monitor",
        "class": CompetitorMonitor,
        "kwargs": {},
    },
    "grant": {
        "name": "Grant Finder",
        "class": GrantFinder,
        "kwargs": {"min_amount": 1000},
    },
}

REPORT_DIR = os.path.join(os.path.dirname(__file__), "reports")


# ── Orchestrator ────────────────────────────────────────────────────────────

class Orchestrator:
    def __init__(self, scanners=None):
        self.scanners = scanners or list(SCANNERS.keys())
        self.results = {}
        self.start_time = None

    def run_all(self):
        """Run all selected scanners."""
        self.start_time = datetime.now(timezone.utc)

        print(f"\n🤖 KOOLA10 AUTOMATION ORCHESTRATOR v2.0")
        print(f"{'='*60}")
        print(f"Scanners: {', '.join(self.scanners)}")
        print(f"Time: {self.start_time.strftime('%Y-%m-%d %H:%M UTC')}")
        print(f"{'='*60}")

        for key in self.scanners:
            config = SCANNERS.get(key)
            if not config:
                print(f"\n⚠️  Unknown scanner: {key}")
                continue

            print(f"\n{'─'*60}")
            scanner = config["class"](**config["kwargs"])

            if hasattr(scanner, "scan_all"):
                results = scanner.scan_all()
            elif hasattr(scanner, "check_all"):
                results = scanner.check_all()
            else:
                results = []

            self.results[key] = {
                "name": config["name"],
                "count": len(results),
                "data": results,
            }

            # Print results
            if hasattr(scanner, "print_results"):
                scanner.print_results()

        return self.results

    def generate_report(self, filepath=None):
        """Generate a markdown report."""
        if not filepath:
            os.makedirs(REPORT_DIR, exist_ok=True)
            timestamp = self.start_time.strftime("%Y%m%d_%H%M")
            filepath = os.path.join(REPORT_DIR, f"report_{timestamp}.md")

        lines = [
            f"# Koola10 Automation Report",
            f"",
            f"**Generated:** {self.start_time.strftime('%Y-%m-%d %H:%M UTC')}",
            f"**Scanners:** {', '.join(self.scanners)}",
            f"",
            f"## Summary",
            f"",
        ]

        total = 0
        for key, data in self.results.items():
            count = data["count"]
            total += count
            lines.append(f"- **{data['name']}:** {count} results")

        lines.append(f"- **Total:** {total} results")
        lines.append(f"")

        # Detail sections
        for key, data in self.results.items():
            lines.append(f"## {data['name']}")
            lines.append(f"")

            if not data["data"]:
                lines.append(f"No results found.")
                lines.append(f"")
                continue

            for item in data["data"][:20]:  # Top 20
                if key == "bounty":
                    amt = f"${item.get('amount', 0):,.0f}" if item.get('amount') else "TBD"
                    lines.append(f"- 💰 **{amt}** — [{item.get('title', '')[:60]}]({item.get('url', '')})")
                    lines.append(f"  - Repo: {item.get('repo', '')} | Labels: {', '.join(item.get('labels', [])[:3])}")
                elif key == "competitor":
                    status = "⚠️ Changed" if item.get("status") == "changed" else "✅ OK"
                    lines.append(f"- {status} — [{item.get('name', '')}]({item.get('url', '')})")
                elif key == "grant":
                    amt = f"${item.get('amount', 0):,.0f}" if item.get('amount') else "TBD"
                    lines.append(f"- 💰 **{amt}** — [{item.get('title', '')[:60]}]({item.get('url', '')})")

                lines.append(f"")

        # Write report
        with open(filepath, "w") as f:
            f.write("\n".join(lines))

        print(f"\n📄 Report saved to: {filepath}")
        return filepath

    def to_json(self):
        """Return all results as JSON."""
        output = {
            "timestamp": self.start_time.isoformat(),
            "scanners": self.scanners,
            "results": {},
        }
        for key, data in self.results.items():
            output["results"][key] = {
                "name": data["name"],
                "count": data["count"],
                "data": data["data"],
            }
        return json.dumps(output, indent=2)


# ── CLI ─────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(description="Koola10 Automation Orchestrator")
    parser.add_argument("--scanners", nargs="*",
                        choices=list(SCANNERS.keys()),
                        help="Specific scanners to run (default: all)")
    parser.add_argument("--report", type=str, nargs="?", const="",
                        help="Generate markdown report (optional path)")
    parser.add_argument("--claim", action="store_true",
                        help="Auto-claim top unclaimed bounty")
    parser.add_argument("--generate-pr", action="store_true",
                        help="Generate PR description for top bounty")
    parser.add_argument("--scaffold", action="store_true",
                        help="Generate starter code for top bounty")
    parser.add_argument("--dry-run", action="store_true",
                        help="Show what would be done without doing it")
    parser.add_argument("--json", action="store_true",
                        help="Output as JSON")
    args = parser.parse_args()

    orch = Orchestrator(scanners=args.scanners)
    orch.run_all()

    if args.json:
        print(orch.to_json())

    if args.report is not None:
        filepath = args.report if args.report else None
        orch.generate_report(filepath)

    # Auto-claim / PR generation / scaffold
    if (args.claim or args.generate_pr or args.scaffold) and "bounty" in (args.scanners or list(SCANNERS.keys())):
        from bounty_scanner import BountyClaimer, PRGenerator, SolutionScaffolder
        from pathlib import Path

        bounty_data = orch.results.get("bounty", {}).get("data", [])
        unclaimed = [b for b in bounty_data if not b.get("assignee") and b.get("amount", 0) >= 200]

        if not unclaimed:
            print("\n  ⚠️  No unclaimed bounties to act on.")
        else:
            target = unclaimed[0]
            amt = f"${target['amount']:,.0f}" if target.get('amount') else 'TBD'
            print(f"\n  🎯 Target: {target['repo']}#{target['number']} ({amt})")

            if args.generate_pr:
                from bounty_scanner import PRGenerator, WORK_DIR
                pr_desc = PRGenerator.generate(target)
                pr_file = WORK_DIR / f"pr-{target['number']}.md"
                WORK_DIR.mkdir(parents=True, exist_ok=True)
                pr_file.write_text(pr_desc)
                print(f"  📄 PR description saved to: {pr_file}")
                print(f"\n{pr_desc}")

            if args.claim:
                from bounty_scanner import BountyClaimer, WORK_DIR
                claimer = BountyClaimer(dry_run=args.dry_run)
                claimer.claim_issue(target)
                if args.scaffold:
                    claimer.setup_repo(target)
                    SolutionScaffolder.scaffold(target, WORK_DIR / str(target["number"]))

            if args.scaffold and not args.claim:
                from bounty_scanner import SolutionScaffolder, WORK_DIR
                scaffold_dir = WORK_DIR / str(target["number"])
                SolutionScaffolder.scaffold(target, scaffold_dir)
                print(f"  📂 Scaffolded in: {scaffold_dir}")


if __name__ == "__main__":
    main()
