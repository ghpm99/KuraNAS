# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

The current Android client: Kotlin + Jetpack Compose (Material 3), `applicationId com.kuranas.android`. Standalone Gradle project (wrapper: Gradle 8.13). `minSdk 33`, `targetSdk 36`, `compileSdk 36`, JVM target 11. Supersedes the Java client in `../mobile`.

## Commands (`cd android`)

- `./gradlew :app:assembleDebug` — build the debug APK (debug `applicationId` is suffixed `.debug`).
- `./gradlew :app:testDebugUnitTest` — JVM unit tests.
- Single test: `./gradlew :app:testDebugUnitTest --tests "com.kuranas.android.*"`.

## Stack (`app/build.gradle.kts`, version catalog)

Hilt (DI, KSP), Retrofit + OkHttp + `kotlinx.serialization` (JSON), Coroutines, Media3 ExoPlayer (`exoplayer`, `ui`, `session`, `hls` — audio/video/HLS), DataStore (preferences + `security.crypto`), Coil, WorkManager, Navigation Compose. Tests: JUnit, MockK, Turbine, `okhttp mockwebserver`, coroutines-test.

## Architecture

- `core/{network,server,ui,discovery}` — infrastructure.
- `feature/<name>/{data,ui}` — one ViewModel + Compose screen per screen. Features: `files`, `music` (also `playback`), `video`, `home`, `connection`, `search`, `diary`, `notifications`, `jobs`, `settings`, `images`.
- `navigation/`, `ui/theme/`.

## Server selection is dynamic — there is no compile-time backend URL

This is the key thing to understand before touching networking (`core/network/NetworkModule.kt`):
- `core/server/ServerStore` persists the chosen server in DataStore (`ServerState`).
- `NetworkModule` installs an OkHttp **interceptor** (`provideServerInterceptor`) that, on every call, reads `ServerStore.serverUrl` and rewrites the request's `scheme`/`host`/`port`. Retrofit's `baseUrl` is a throwaway `http://localhost:8000/` placeholder.
- If no server is set (or the value is invalid), the interceptor **proceeds without rewriting and never throws** — throwing from an interceptor would crash the app.
- `core/discovery/NsdDiscovery` finds servers on the LAN.

OkHttp timeouts: connect 30s, read/write 60s. `HttpLoggingInterceptor` logs bodies only when `BuildConfig.DEBUG`.

The shared API plumbing (`AppResult`, `SafeApiCall`, `PageDto`, `MimeType`) lives in `core/network/`; feature `data` layers build on it.

## User-facing text goes through i18n (mandatory)

No literal string may reach a Composable the user sees. Use `stringResource(R.string.key)` (and `pluralStringResource` / formatted args where needed); add the term to `app/src/main/res/values/strings.xml` (plus `values-en`, `values-pt-rBR` as locales are added — Android picks by device locale). The existing literal `Text("…")` call sites are debt to migrate as screens are touched. **Backend messages** (`AppResult` errors surfaced from `SafeApiCall`) are already translated server-side — display them as received; do **not** add a parallel `strings.xml` entry for them. So each screen mixes app-owned text (`strings.xml`) with server text (verbatim) — this "mix simples" is intentional. Full cross-app rule in the root `CLAUDE.md` → "No user-facing literal strings".
</content>
