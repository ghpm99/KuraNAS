---
name: split-commit
description: "Split mixed changes in the working tree into logically separated commits, especially separating no-op-in-prod refactors from real behavior changes. Use when staged or unstaged changes mix multiple concerns and the user wants clean commit history."
allowed-tools: Bash(git status*) Bash(git diff*) Bash(git add*) Bash(git restore*) Bash(git log*) Bash(git commit*)
---

# /split-commit

You are helping create logically separated commits from mixed working-tree changes.

## Workflow

1. **Inventory** the changes with `git status --short` and `git diff`. Identify what concern(s) each modified file belongs to.

2. **Categorize** every change into one of:
   - **Refactor (no-op in prod)** — extract setting, add helper with default preserving behavior, rename without functional change.
   - **Real behavior change** — bugfix, feature, breaking change.
   - **Tests** — ideally co-committed with the change they cover.
   - **Pre-existing** — changes that were already in the working tree before this task started. Do not touch without explicit user approval; confirm with `git log` or by asking.
   - **Temporary patches** — workarounds applied for testing only. These should be reverted (`git restore <file>`), not committed.

3. **Propose grouping** to the user *before* committing anything:
   - Numbered list of commit batches in order.
   - For each batch: files involved + one-sentence rationale + draft commit message in Conventional Commits format.
   - Wait for approval/edit/veto before proceeding.

4. **For multi-concern single files** that can't be split by `git add <file>`, use the temporary-undo pattern:
   1. `Edit` the file to remove the second concern.
   2. `git add` everything for batch 1, `git commit`.
   3. `Edit` the file to restore the second concern.
   4. `git add` for batch 2, `git commit`.
   Never use `git add -p` or `git add -i` — interactive modes are not supported.

5. **After commits**, show `git log --oneline -N` so the user sees the final state.

## Commit message conventions

Read the project's recent style first: `git log --oneline -10`. Match it.

Defaults:
- `feat(scope): description` for new functionality
- `fix(scope): description` for bug fixes
- `refactor(scope): description` for no-behavior-change restructuring
- `test(scope): description` for tests added/updated independently
- Keep to 1-2 sentences (CLAUDE.md guidance is "concise").
- Include the project's required commit footer if any (e.g. `Co-Authored-By:` for Claude Code projects with that convention).

Pass the message via HEREDOC to preserve formatting:

```bash
git commit -m "$(cat <<'EOF'
fix(scope): one-line summary

Optional second paragraph with context.

Co-Authored-By: ...
EOF
)"
```

## Rules

- Never destructive ops (`reset --hard`, force push, `commit --amend`) unless the user explicitly requests them.
- Never commit untracked files the user said to leave alone.
- Never skip hooks (`--no-verify`).
- Always show `git diff --staged` for each batch before committing so the user can intercept.
- If a hook fails, fix the underlying issue and create a NEW commit — never amend.
- Never push. Pushing is always the user's call.
