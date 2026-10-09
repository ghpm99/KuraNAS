---
name: analyze
description: "Start a deep analysis of a codebase problem, feature, or refactoring need. Use when the user says 'analyze', 'investigate', 'research', 'deep dive', 'understand the problem', 'explore', or wants to start a new feature/refactoring exploration. Also use when the user provides a PRD or problem statement to break down."
---

# /analyze [topic]

You are performing a deep codebase analysis. The user wants a thorough investigation written to a persistent file.

## Project Mapping

| Working directory contains | Docs path |
|---|---|
| `KuraNAS` or `kuranas` | `/home/server/Documentos/docs/kuranas` |
| `kawori-backend` | `/home/server/Documentos/docs/kawori-backend` |
| `kawori-financial` | `/home/server/Documentos/docs/kawori-financial` |
| `kawori-frontend` | `/home/server/Documentos/docs/kawori-financial` |

## Steps

1. **Detect the project** from the current working directory using the project mapping above.

2. **Determine the slug** from the topic argument (`$ARGUMENTS`). Convert to lowercase-kebab-case. If no argument, ask the user what to analyze.

3. **Check if analysis already exists** at `{docs}/analysis/{slug}.md`. If it does, read it and ask the user if they want to continue/update or start fresh.

4. **Explore the codebase thoroughly**:
   - Use Glob to find relevant files
   - Use Grep to search for patterns, function names, imports
   - Use Read to understand key files
   - Use Agent with subagent_type=Explore for broad searches
   - Be exhaustive — this is the ONE time we load heavy context

5. **Write the analysis** to `{docs}/analysis/{slug}.md` using this template:

```markdown
# Analysis: {title}

**Created**: YYYY-MM-DD
**Project**: {project}

## Context
Why this analysis was needed.

## Findings
Detailed exploration results with file paths and line numbers.

## Recommendations
What should be done based on findings.
```

Include: code structure maps, dependency graphs, data flows, problems found, risks identified, recommendations.

6. **Suggest next step**: Tell the user to run `/decide {slug}` to extract actionable decisions from this analysis.

## Rules
- This skill produces LARGE documents — that's by design
- Always include file paths and line numbers in findings
- Do NOT update backlog.md or index.md — this is analysis only
- Do NOT read existing analysis/decisions unless the user explicitly asks to build on them
