#!/usr/bin/env python3
"""Find truly open, unassigned bounty issues on GitHub."""

import json
import re
import time
import urllib.parse
import urllib.request

QUERIES = [
    'label:bounty state:open no:assignee',
    '"bounty" in:title state:open is:issue no:assignee',
]

all_results = []
seen = set()

for q in QUERIES:
    url = f'https://api.github.com/search/issues?q={urllib.parse.quote(q)}&per_page=30&sort=created&direction=desc'
    req = urllib.request.Request(url, headers={
        'Accept': 'application/vnd.github.v3+json',
        'User-Agent': 'Koola10'
    })
    try:
        with urllib.request.urlopen(req, timeout=15) as r:
            data = json.loads(r.read())
            for item in data.get('items', []):
                repo_url = item['repository_url']
                repo = '/'.join(repo_url.split('/')[-2:])
                key = f"{repo}#{item['number']}"
                if key in seen:
                    continue
                seen.add(key)
                labels = [l['name'] for l in item.get('labels', [])]
                text = item.get('title', '') + ' ' + (item.get('body', '') or '')
                amounts = [float(m.replace(',', '')) for m in re.findall(r'\$(\d[\d,]*)', text)]
                amt = max(amounts) if amounts else 0
                all_results.append({
                    'repo': repo,
                    'number': item['number'],
                    'title': item['title'][:120],
                    'amount': amt,
                    'url': item['html_url'],
                    'labels': labels[:5],
                    'created': item['created_at'][:10],
                    'comments': item.get('comments', 0),
                })
    except Exception as e:
        print(f'Error on query: {e}')
    time.sleep(2)

all_results.sort(key=lambda x: x['amount'], reverse=True)

print(f"Found {len(all_results)} unassigned bounty issues\n")
for i, r in enumerate(all_results[:20]):
    amt = f"${r['amount']:,.0f}" if r['amount'] else 'TBD'
    print(f"{i+1}. {amt} — {r['repo']}#{r['number']} ({r['created']})")
    print(f"   {r['title']}")
    print(f"   {r['url']}")
    if r['labels']:
        print(f"   Labels: {', '.join(r['labels'])}")
    print()
