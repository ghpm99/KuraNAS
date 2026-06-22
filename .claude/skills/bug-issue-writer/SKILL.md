---
name: bug-issue-writer
description: "Write a structured Markdown bug issue with summary, reproduction, root cause (with file:line), evidence (queries/logs), decision among options, and implementation notes. Use when documenting a bug after diagnosis is complete, before opening an issue in a tracker."
argument-hint: [title]
---

# /bug-issue-writer [title]

You are writing a bug issue as a `.md` file. The filename is the title; the body follows a fixed template.

## Inputs you need

Confirm before writing:
- **Title** — becomes the filename. Match the user's preferred language and casing.
- **Where to save** — current working directory by default. If the project has a `docs/issues/` or similar convention, prefer that. Ask if unclear.
- **Diagnosis source** — what investigation produced these findings? (recent chat context, an analysis file, etc.)
- **Decision** — if multiple fixes were considered, which one is chosen and why?

If the diagnosis is incomplete, push back. This skill assumes the bug is understood; it does not perform discovery.

## Template

Write to `{title}.md`:

```markdown
# {title}

## Resumo / Summary

One paragraph stating the bug clearly: what breaks, in what conditions, what the visible symptom is.

## Reprodução / Reproduction

Pre-conditions + exact steps to reproduce. If the bug surfaced through a real incident, describe the data state + the operation that exposed it.

## Comportamento esperado / Expected behavior

The invariant or contract that should hold.

## Comportamento observado / Observed behavior

What actually happens, with enough detail to distinguish from expected.

## Causa raiz / Root cause

Reference specific `file:line` ranges. Show the offending code in a fenced block. Explain why the code produces the wrong outcome.

## Evidência / Evidence

Real queries, logs, or audit entries that confirmed the diagnosis. Include the actual output, not paraphrased descriptions.

## Decisão de correção / Fix decision

The chosen approach, named explicitly.

### Opções consideradas / Options considered (and rejected)

- **Option A**: brief description. Rejected because: ...
- **Option B**: brief description. Rejected because: ...

## Notas para implementação / Notes for implementation

- Exact functions or files to change.
- Test cases to add (golden path + edge cases).
- Follow-ups that are out of scope but worth tracking.
- Cross-links to related issues: `[[other-issue-slug]]`.
```

## Rules

- **Language match**: write in the same language the user has been using in chat (PT or EN).
- **Real evidence only**: include actual query output / log line / error message — never paraphrase.
- **`file:line` is mandatory** in the root cause section.
- **"Options considered" is mandatory** even when only one fix was obvious — state that explicitly so the rationale is captured for future readers.
- Do not include things you didn't verify. If production impact is unknown, say so.
- This is engineering documentation, not a status update. Keep it factual and dense.
