# Mobile Standards (Canonical)

This document is the canonical source for Android mobile implementation patterns in this repository.

## Agent Enforcement
- Before any mobile change, read this file first.
- If a rule here conflicts with local file style, prefer this file and then normalize surrounding code when safe.
- If a requested change conflicts with these rules, explicitly call out the conflict before implementing.

## 1) Platform and Compatibility (Mandatory)
- Target device baseline is Samsung Galaxy Tab 2 7.0 (GT-P3110) on Android 4.1.2 (API level 16).
- Mandatory stack is `Java + XML Views + AppCompat`.
- Kotlin is forbidden for mobile code in this repository.
- Jetpack Compose is forbidden for mobile code in this repository.
- All APIs, dependencies, and runtime behavior must remain compatible with API 16.
- Keep `minSdk 16` intact unless an explicit architecture decision updates the baseline first.

## 2) Architecture and Ownership
- Keep ownership explicit by feature and layer, with migration target:
- `feature/<domain>/{presentation,domain,data}`.
- During transition, preserve current layers (`app`, `domain`, `data`, `infra`, `presentation`) and extract by feature incrementally.
- Keep UI orchestration in presentation classes (`Activity`, `Fragment`, adapters, view binders).
- Keep business rules in domain/use-case classes.
- Keep IO/network/cache in data/infra classes.
- Keep dependency wiring centralized in `ServiceLocator` until replacement is planned and documented.

## 3) UI and XML Rules
- Build screens with XML layouts and AppCompat widgets only.
- Do not introduce Compose artifacts, Compose dependencies, or Kotlin UI wrappers.
- Keep `Activity` and `Fragment` classes focused and small; extract collaborators for state handling and flow coordination.
- Preserve tablet-focused usability for 1024x600 layout constraints.
- Avoid expensive runtime layout operations on the main thread.

## 4) Data, Networking, and Error Handling
- Route HTTP through infra clients (`infra/http`) and repository abstractions.
- Parse/mapping logic must be explicit in dedicated mapper classes.
- Do not perform networking directly from view classes when repository/service boundaries exist.
- Error mapping should use domain-friendly models (`domain/error`) and avoid leaking transport details to UI.

## 5) i18n (Mandatory)
- Never hardcode user-facing text in Java code for new or changed behavior.
- Use `TranslationManager` and translation keys as the source of runtime text.
- Keep local fallback translations in `mobile/app/src/main/assets/translations`.
- Any new visible text key must be synchronized with backend translation keys/content.

## 6) Performance and Stability on API 16
- Keep memory usage conservative (bitmap/cache handling, list rendering, media playback state).
- Avoid reflection-heavy, modern-only, or API-21+ libraries.
- Keep background work out of UI thread.
- Use clear lifecycle cleanup in `Activity`/`Fragment` for observers, callbacks, and media resources.

## 7) Testing and Quality Gates
- Mandatory local quality commands for mobile changes:
- `cd mobile && ./gradlew test`
- `cd mobile && ./gradlew assembleDebug`
- Use unit tests for domain logic and mapper behavior when changed.
- Use instrumented tests only when behavior cannot be validated with unit-level coverage.

## 8) Pull Request Checklist (Mobile)
- API 16 compatibility preserved.
- Only Java/XML/AppCompat used.
- No Kotlin/Compose introduced.
- i18n respected for user-visible text.
- Architecture boundaries preserved (presentation/domain/data/infra).
- `test` and `assembleDebug` executed with evidence.
- If a new mobile pattern is introduced, update this standards file in the same PR.

## Change Log
- 2026-04-01: Initial canonical mobile standards file created for reorganization governance (Batch 0.x).
