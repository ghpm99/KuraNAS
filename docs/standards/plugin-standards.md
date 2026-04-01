# Plugin Standards (Canonical)

This document is the canonical source for browser extension implementation patterns in this repository.

## Agent Enforcement
- Before any plugin change, read this file first.
- If a rule here conflicts with local file style, prefer this file and then normalize surrounding code when safe.
- If a requested change conflicts with these rules, explicitly call out the conflict before implementing.

## 1) Architecture and Scope
- Plugin stack is Chrome Extension Manifest V3.
- Keep extension behavior split by runtime context:
- background service worker (`background.js` or `src/background/*` during migration),
- content scripts (`content/*`),
- popup UI (`popup/*`),
- offscreen runtime (`offscreen/*`).
- Do not mix UI behavior, network orchestration, and message routing into a single giant module when refactoring.
- Migration target is modular ownership by domain (`background`, `content`, `popup`, `offscreen`, `shared`).

## 2) Modularization Rules
- Extract pure constants/utilities first when breaking large files.
- Keep each module focused on one responsibility (detection, routing, upload, download, state, parsing).
- Message contract keys and payload shapes must be centralized and versioned.
- Avoid hidden cross-module state; explicit state containers only.
- Preserve functional behavior during structural batches (no feature changes mixed with reorg).

## 3) Manifest and Runtime Safety
- Keep `manifest.json` coherent with actual file paths and runtime contexts.
- Any path update in code must include manifest updates when required.
- Use least-privilege permissions when adding new capabilities.
- Keep host permissions explicit and justified.
- Do not add remote code execution patterns or dynamic code injection.

## 4) i18n and User-Facing Text
- Avoid hardcoded user-facing text for new or changed popup flows.
- Add text keys to the plugin i18n layer as it is introduced/expanded.
- Keep naming semantic and stable across popup/content/background boundaries.
- Do not change visible wording in structural PRs unless required for consistency fixes.

## 5) Lint and Test Baseline (Batch 0.x)
- Plugin must expose local commands:
- `npm run lint`
- `npm test`
- Initial lint gate may focus on syntax/scope safety; rule strictness can increase in Batch 4.x.
- Tests should prioritize pure logic and manifest/runtime contract validation.
- Structural refactor PRs must keep plugin lint/tests passing.

## 6) Pull Request Checklist (Plugin)
- Manifest, scripts, and runtime contexts stay consistent.
- Structural changes do not alter expected capture/upload behavior.
- Lint and unit tests executed locally with evidence.
- Message routing and shared constants remain explicit and testable.
- If a new plugin pattern is introduced, update this standards file in the same PR.

## Change Log
- 2026-04-01: Initial canonical plugin standards file created for reorganization governance (Batch 0.x).
