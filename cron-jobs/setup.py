#!/usr/bin/env python3
"""
Bulk-create the cron-jobs.json manifest in cron-job.org via their REST API.

Setup:
  1. Sign in at https://cron-job.org
  2. Account -> API -> enable & copy the API key
  3. export CRON_JOB_ORG_API_KEY='<key>'   (or CRON_JOB_API)
  4. export ADMIN_KEY='<render admin api key>'
  5. python3 cron-jobs/setup.py                       # dry-run
     python3 cron-jobs/setup.py --apply               # PUT every job
     python3 cron-jobs/setup.py --apply --skip-existing   # only PUT missing

Docs: https://docs.cron-job.org/rest-api.html
"""

import argparse
import json
import os
import sys
import time
import urllib.error
import urllib.request

API_BASE = "https://api.cron-job.org"
MANIFEST_PATH = os.path.join(os.path.dirname(__file__), "cron-jobs.json")
TIMEZONE = "America/New_York"

BOUNDS = {
    "minutes": (0, 59),
    "hours":   (0, 23),
    "mdays":   (1, 31),
    "months":  (1, 12),
    "wdays":   (1, 7),
}
CRON_FIELDS = list(BOUNDS.keys())  # order: min hour dom mon dow


def replace_tokens(s, admin_key):
    if not isinstance(s, str):
        return s
    return (
        s.replace("${ADMIN_KEY}", admin_key)
        .replace("${EPOCH}", str(int(time.time())))
    )


def token_to_field(token, lo, hi):
    """Convert a single cron field token to a list of ints.
    '*'       -> [-1]      (any)
    'N'       -> [N]
    'N,M,...' -> sorted unique [N,M,...]
    '*/N'     -> step from lo..hi by N
    """
    token = token.strip()
    if token in ("*", ""):
        return [-1]
    if token.startswith("*/"):
        step = int(token[2:])
        if step <= 0:
            raise ValueError(f"Invalid step in cron token {token!r}")
        vals = list(range(lo, hi + 1, step))
        return vals if vals else [-1]
    if "," in token:
        out = []
        for part in token.split(","):
            out.extend(token_to_field(part, lo, hi))
        seen = []
        for v in out:
            if lo <= v <= hi and v not in seen:
                seen.append(v)
        return sorted(seen)
    n = int(token)
    if n < lo or n > hi:
        raise ValueError(f"Value {n} out of range [{lo},{hi}]")
    return [n]


def cron_to_schedule(expr, tz=TIMEZONE):
    parts = expr.split()
    if len(parts) != 5:
        raise ValueError(f"Expected 5 cron fields, got: {expr!r}")
    schedule = {"timezone": tz, "expiresAt": 0}
    for i, field in enumerate(CRON_FIELDS):
        lo, hi = BOUNDS[field]
        schedule[field] = token_to_field(parts[i], lo, hi)
    return schedule


def job_to_payload(job, admin_key):
    url = replace_tokens(job["url"], admin_key)
    raw_body = job.get("body")
    body = replace_tokens(raw_body, admin_key) if raw_body else ""
    headers = {h: replace_tokens(v, admin_key) for h, v in (job.get("headers") or {}).items()}
    method = job.get("method", "GET").upper()
    request_method = 1 if method == "POST" else 0
    return {
        "job": {
            "title": job["name"],
            "url": url,
            "enabled": True,
            "saveResponses": True,
            "requestMethod": request_method,
            "schedule": cron_to_schedule(job["cron_expression"]),
            "notification": {
                "onFailure": True,
                "onSuccess": False,
                "onDisable": False,
            },
            "extendedData": {
                "headers": headers,
                "body": body,
            },
        }
    }


def list_existing_jobs(api_key):
    req = urllib.request.Request(
        f"{API_BASE}/jobs",
        method="GET",
        headers={"Authorization": f"Bearer {api_key}"},
    )
    with urllib.request.urlopen(req, timeout=20) as resp:
        data = json.loads(resp.read().decode("utf-8"))
    return [j for j in data.get("jobs", []) if j.get("title")]


def put_job(payload, api_key):
    req = urllib.request.Request(
        f"{API_BASE}/jobs",
        data=json.dumps(payload).encode("utf-8"),
        method="PUT",
        headers={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {api_key}",
        },
    )
    with urllib.request.urlopen(req, timeout=20) as resp:
        return resp.status, json.loads(resp.read().decode("utf-8"))


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--apply", action="store_true",
                        help="Actually PUT jobs (default is dry-run)")
    parser.add_argument("--delay", type=float, default=13.0,
                        help="Seconds between PUTs (cron-job.org: 1/s, 5/min)")
    parser.add_argument("--skip-existing", action="store_true",
                        help="GET /jobs first; skip titles already present (safe re-run)")
    args = parser.parse_args()

    api_key = (
        os.environ.get("CRON_JOB_ORG_API_KEY")
        or os.environ.get("CRON_JOB_API")
        or ""
    )
    if args.apply and not api_key:
        print("Error: set CRON_JOB_ORG_API_KEY or CRON_JOB_API before --apply",
              file=sys.stderr)
        sys.exit(2)

    with open(MANIFEST_PATH) as f:
        manifest = json.load(f)

    admin_key = os.environ.get("ADMIN_KEY", "")
    if args.apply and not admin_key:
        print("Warning: ADMIN_KEY not set — protected jobs will get 401 on first run.",
              file=sys.stderr)

    mode = "APPLY (PUT to cron-job.org)" if args.apply else "DRY-RUN"
    print(f"{mode} {len(manifest['jobs'])} jobs (delay={args.delay}s)\n")

    existing_titles = set()
    if args.apply and args.skip_existing:
        try:
            existing = list_existing_jobs(api_key)
            existing_titles = {j["title"] for j in existing}
            print(f"Found {len(existing_titles)} existing jobs in account; will skip on re-run.\n")
        except urllib.error.HTTPError as e:
            print(f"Warning: could not list existing jobs (HTTP {e.code}); proceeding without skip.\n",
                  file=sys.stderr)
        except urllib.error.URLError as e:
            print(f"Warning: could not list existing jobs ({e.reason}); proceeding without skip.\n",
                  file=sys.stderr)

    if args.skip_existing and existing_titles:
        pending = [j for j in manifest["jobs"] if j["name"] not in existing_titles]
        skipped_count = len(manifest["jobs"]) - len(pending)
        if skipped_count:
            print(f"Skipping {skipped_count} already-present jobs.\n")
    else:
        pending = list(manifest["jobs"])

    ok = err = skipped = 0
    job_ids = []

    for i, job in enumerate(pending):
        try:
            payload = job_to_payload(job, admin_key)
        except Exception as ex:
            print(f"[{job['type']:<11}] {job['name']:<40} \u2717 payload error: {ex}")
            err += 1
            continue

        sched = payload["job"]["schedule"]
        sched_str = (
            f"min={sched['minutes']} hr={sched['hours']} "
            f"dom={sched['mdays']} mon={sched['months']} dow={sched['wdays']}"
        )
        print(f"[{job['type']:<11}] {job['name']:<40} \u2192 {job['method']} {job['url']}")
        print(f"           cron: {job['cron_expression']!r} -> {sched_str}")

        if not args.apply:
            continue

        try:
            status, body = put_job(payload, api_key)
            jid = body.get("jobId", "?")
            ok += 1
            job_ids.append(jid)
            print(f"           \u2713 status={status} jobId={jid}")
        except urllib.error.HTTPError as e:
            err += 1
            err_body = e.read().decode("utf-8", errors="ignore")[:240]
            print(f"           \u2717 HTTP {e.code}: {err_body}")
        except urllib.error.URLError as e:
            err += 1
            print(f"           \u2717 URL error: {e.reason}")

        if i < len(pending) - 1:
            time.sleep(args.delay)

    if args.apply:
        print(f"\nDone. {ok} created, {err} failed, {skipped} skipped.")
        if job_ids:
            print(f"jobIds: {job_ids}")


if __name__ == "__main__":
    main()
