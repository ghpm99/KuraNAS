---
name: decide
description: "Extract actionable decisions from an analysis document. Use when the user says 'decide', 'refine', 'distill', 'what should we do', 'extract decisions', or references a completed analysis and wants to move to the next stage of planning."
---

# /decide [analysis-slug]

You are distilling an analysis into concrete, short decision documents.

## Project Mapping

| Working directory contains | Docs path |
|---|---|
| `KuraNAS` or `kuranas` | `/home/server/Documentos/docs/kuranas` |
| `kawori-backend` | `/home/server/Documentos/docs/kawori-backend` |
| `kawori-financial` | `/home/server/Documentos/docs/kawori-financial` |
| `kawori-frontend` | `/home/server/Documentos/docs/kawori-financial` |

## Steps

1. **Detect the project** from the current working directory using the project mapping above.

2. **Read the analysis** at `{docs}/analysis/{slug}.md` where `{slug}` comes from `$ARGUMENTS`. If no argument, list available analyses and ask.

3. **Identify distinct decisions** from the analysis. Each decision should be:
   - A single, actionable choice (technology, architecture, approach)
   - Under 100 lines
   - Self-contained: readable without the full analysis

4. **Write decision files** to `{docs}/decisions/{decision-slug}.md` using this template:

```markdown
# Decision: {title}

**Analysis**: ../analysis/{slug}.md
**Created**: YYYY-MM-DD
**Status**: accepted

## Decision
What was decided.

## Why
Rationale.

## Alternatives Rejected
What was considered and why not.

## Tasks Derived
- task-slug-1
- task-slug-2
```

5. **Present a summary** of all decisions created.

6. **Suggest next step**: Tell the user to run `/task-create {decision-slug}` to create executable tasks.

## Rules
- Decision files MUST be under 100 lines each
- Reference the parent analysis with a relative path
- Do NOT create tasks — that's `/task-create`'s job
- Do NOT update backlog.md or index.md
