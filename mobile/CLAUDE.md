# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**Legacy Android client, superseded by `../android`.** Java + Android Views (no Compose, no Kotlin), `applicationId com.kuranas.mobile`. Targets very old devices: `minSdk 16`, `targetSdk 35`, `compileSdk 35`. Standalone Gradle project. Prefer working in `../android`; only touch this for maintenance of the legacy app.

## Commands (`cd mobile`)

- `./gradlew :app:assembleDebug` — build the debug APK.
- `./gradlew :app:testDebugUnitTest` — JVM unit tests (`unitTests.returnDefaultValues = true`).

## Notes

- Dependencies are deliberately minimal: `appcompat`, `recyclerview`, `swiperefreshlayout`; tests use JUnit, Mockito (`core` + `inline`), `org.json`.
- **Backend URL is compile-time**, unlike `../android`: `API_BASE_URL` is a `buildConfigField` baked into every build type (default `http://192.168.1.100:8000`). Change it in `app/build.gradle` to point at a different server.

## User-facing text goes through i18n (mandatory)

User-visible strings must come from `TranslationManager` (`i18n/TranslationManager.java`) reading `app/src/main/assets/translations/*.json` by `SCREAMING_SNAKE` key — no literals in Views/adapters/activities. Backend `error` messages arrive already translated; show them as-is. Full cross-app rule in the root `CLAUDE.md` → "No user-facing literal strings".
</content>
