# Nova Bounty Solver — Clone, Fix, Test, PR

You are solving a specific GitHub bounty issue. Follow these steps precisely:

## Input
You will be given:
- `REPO`: The GitHub repository (e.g., `owner/repo`)
- `ISSUE`: The issue number to solve
- `TITLE`: The issue title
- `APPROACH`: The suggested approach
- `COMMENT`: The comment already posted on the issue

## Steps

### 1. Read the issue
Use the `github` tool to read the issue body and all comments. Understand:
- What exactly needs to be fixed/implemented
- Any acceptance criteria or test requirements
- Any metadata files required (check for `metadata.json` or similar)
- The repo's contribution guidelines (check README, CONTRIBUTING.md)

### 2. Clone the repo
```bash
cd /tmp
git clone https://github.com/{REPO}.git bounty-work
cd bounty-work
```

### 3. Read the codebase
- Read the relevant source files mentioned in the issue
- Understand the existing code structure, tests, and conventions
- Check for linting configs, test frameworks, build scripts

### 4. Implement the fix
- Create a new branch: `git checkout -b fix/{issue-slug}`
- Make the minimal, focused changes required
- Follow the repo's existing code style and conventions
- Add or update tests as specified in the issue

### 5. Run tests
- Run the existing test suite to make sure nothing breaks
- Run any new tests you added
- If tests fail, fix them before proceeding

### 6. Commit and push
```bash
git add -A
git commit -m "fix: {description of the fix}\n\nCloses #{ISSUE}"
git push origin fix/{issue-slug}
```

### 7. Create a PR
Use the `github` tool to create a pull request:
- Title: Clear, descriptive title referencing the issue
- Body: Explain what was done, how it was tested, and link to the issue
- Reference the issue with `Closes #{ISSUE}` in the body

### 8. Report results
Write a summary to `output/bounty-solve-{REPO}-{ISSUE}.md` with:
- What was done
- Files changed
- Tests run and results
- PR URL
- Any caveats or follow-up items

## Rules
- **Never merge the PR** — that's for the repo maintainer
- **Never force-push** — keep the history clean
- **Never modify files outside the scope** of the issue
- **Always run tests** before pushing
- **Always follow the repo's contribution guidelines**
- If the issue is unclear or the codebase is too complex, **report back** instead of guessing
- If the fix requires changes to infrastructure, dependencies, or architecture beyond the issue scope, **report back** with what you found
- Keep the PR focused — one fix per PR
