import json
import os
import subprocess

def remediate():
    analysis_file = "analysis.json"
    if not os.path.exists(analysis_file):
        return

    with open(analysis_file, 'r') as f:
        findings = json.load(f)

    if not findings:
        print("No issues found. Skipping remediation.")
        return

    # Configure git identity for GitHub Actions
    try:
        subprocess.run(["git", "config", "user.name", "github-actions[bot]"], check=True)
        subprocess.run(["git", "config", "user.email", "github-actions[bot]@users.noreply.github.com"], check=True)
    except Exception as e:
        print(f"Failed to configure git: {e}")

    # Create GitHub Issue
    issue_body = "The following issues were detected in Fly.io logs:\n\n"
    for finding in findings:
        issue_body += f"### {finding['type']} ({finding['severity']})\n"
        issue_body += f"Matches: {', '.join(finding['matches'])}\n\n"

    try:
        subprocess.run([
            "gh", "issue", "create",
            "--title", "Self-Healing Alert: Errors detected in production",
            "--body", issue_body
        ], check=True)
    except Exception as e:
        print(f"Failed to create GitHub issue: {e}")

    # Capture current branch to return to it
    base_branch = subprocess.check_output(["git", "rev-parse", "--abbrev-ref", "HEAD"], text=True).strip()

    # Attempt auto-fixes for specific issues
    for finding in findings:
        if finding["type"] == "ModuleNotFoundError":
            for module_name in finding["matches"]:
                try:
                    auto_fix_missing_dependency(module_name, base_branch)
                finally:
                    # Always return to base branch between fixes
                    subprocess.run(["git", "checkout", base_branch], check=True)

def auto_fix_missing_dependency(module_name, base_branch):
    branch_name = f"fix/missing-dependency-{module_name}"
    print(f"Attempting auto-fix for missing dependency: {module_name}")

    try:
        # 1. Create/Reset branch from base
        subprocess.run(["git", "checkout", "-B", branch_name, base_branch], check=True)

        # 2. Add dependency to requirements.txt (avoid duplicates)
        with open("requirements.txt", "r") as f:
            reqs = f.read()

        if module_name not in reqs:
            with open("requirements.txt", "a") as f:
                f.write(f"\n{module_name}")

        # 3. Commit and push
        subprocess.run(["git", "add", "requirements.txt"], check=True)
        subprocess.run(["git", "commit", "-m", f"fix: Add missing dependency {module_name}"], check=True)
        subprocess.run(["git", "push", "origin", branch_name, "--force"], check=True)

        # 4. Create PR
        subprocess.run([
            "gh", "pr", "create",
            "--title", f"fix: Add missing dependency {module_name}",
            "--body", f"Auto-generated PR to fix missing dependency `{module_name}` detected in production logs.",
            "--head", branch_name,
            "--base", base_branch
        ], check=True)
    except Exception as e:
        print(f"Auto-fix failed for {module_name}: {e}")

if __name__ == "__main__":
    remediate()
