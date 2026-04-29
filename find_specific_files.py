import os
import subprocess

def run(cmd):
    try:
        return subprocess.check_output(cmd, shell=True, stderr=subprocess.STDOUT).decode()
    except Exception as e:
        return ""

branches = [b.strip() for b in run("git branch -a").split('\n') if b.strip() and '->' not in b]
for b in branches:
    if b.startswith('*'): b = b[2:].strip()
    res = run(f"git ls-tree -r {b} --name-only | grep -E 'hn_launch.txt|product_hunt.txt|reddit_saas.txt'")
    if res:
        print(f"Branch {b} contains:\n{res}")
