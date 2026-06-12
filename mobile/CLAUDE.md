# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**Legacy Android client, superseded by `../android` — now a wall-panel (kiosk) app.** Java + Android Views (no Compose, no Kotlin), `applicationId com.kuranas.mobile`. Targets very old devices: `minSdk 16`, `targetSdk 35`, `compileSdk 35`. Standalone Gradle project.

Since task 17 of `docs/melhorias/` (2026-06-12) every media-browsing screen (files/images/music/video/search/settings) was **removed for good** — media navigation belongs to `../android`. What remains: `ConnectionActivity` (LAN discovery: mDNS/UDP/scan + cache) → `MainActivity` hosting a single `HomeFragment` (clock + date), `KioskManager` (fullscreen pinning), `TranslationManager` + `HttpClient`, and a trimmed `ServiceLocator`. Task 18 builds the kiosk panel (notifications + e-mail digest) on top of `HomeFragment` — keep DTO consumption tiny, no WebView (2012 tablet).

## Commands (`cd mobile`)

- `./gradlew :app:assembleDebug` — build the debug APK.
- `./gradlew :app:testDebugUnitTest` — JVM unit tests (`unitTests.returnDefaultValues = true`).

## Notes

- Dependencies are deliberately minimal: `appcompat`, `recyclerview`, `swiperefreshlayout`; tests use JUnit, Mockito (`core` + `inline`), `org.json`.
- **Backend URL is compile-time**, unlike `../android`: `API_BASE_URL` is a `buildConfigField` baked into every build type (default `http://192.168.1.100:8000`). Change it in `app/build.gradle` to point at a different server.

## User-facing text goes through i18n (mandatory)

User-visible strings must come from `TranslationManager` (`i18n/TranslationManager.java`) reading `app/src/main/assets/translations/*.json` by `SCREAMING_SNAKE` key — no literals in Views/adapters/activities. Backend `error` messages arrive already translated; show them as-is. Full cross-app rule in the root `CLAUDE.md` → "No user-facing literal strings".
</content>
