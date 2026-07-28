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
        "kwargs": {"min_bounty": 200},
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


if __name__ == "__main__":
    main()
