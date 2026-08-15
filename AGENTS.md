# AGENTS.md

Guidelines for automated agents (LLMs) operating in this repository.
**The agent itself decides** which commands to run and in what order. The sections below
describe *what should be done*; a competent task solver interprets and sequences those steps
based on the current state and environment (branch name, existing changes, etc.).

## 1. Git Sync on Startup
When an agent starts for the first time in a session, it should bring the repository up to
date:

- Fetch remote changes (`fetch` / `pull`).
- Keep the working branch (`master` or `main`) current.
- Inspect the current state with `git status` and `git log`.

The exact commands and ordering are up to the agent based on what it observes in the
environment.

## 2. Git Update After Every Task
After every action (code change, test run, documentation update, build, etc.) the agent
should:

- Stage all changes (`git add -A` or the relevant subset of files).
- Create a commit. The message should be **clear, concise, and descriptive** — the agent
  writes it from scratch based on what it did. It must follow these rules:
  - It should describe *what* the change does.
  - It should use conventional prefixes where appropriate (e.g. `fix:`, `feat:`, `docs:`).
  - When relevant, append the co-author trailer:

  ```
  Co-authored-by: CommandCodeBot <noreply@commandcode.ai>
  ```
- Attempt to push the commit to the remote (`git push`).
- If the push fails (conflicts, diverged remote, etc.) the agent should resolve it — e.g.
  `git pull --rebase` — and retry.

## 3. Commit Message Convention
- Start with a conventional prefix: `fix:`, `feat:`, `docs:`, `chore:`, `refactor:`, etc.
- Describe the change concisely.
- Add the `Co-authored-by: CommandCodeBot <noreply@commandcode.ai>` trailer when the change
  is authored by an automated agent.

## 4. General Rules
- Always check `git status` and `git log` before making changes.
- Build the code and run tests before committing.
- Avoid destructive operations such as `git reset --hard` or deleting files.
- Commit directly to `master` / `main` — no feature branches.

## 5. Documentation
- After changing code, update the relevant documentation (README, inline
  comments, usage examples) to keep it accurate. Do not leave docs stale.
- When adding a new feature, flag or CLI option, document it in the README.
- When removing a feature or file, remove its references from the docs too.
