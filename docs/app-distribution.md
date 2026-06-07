# App distribution (downloads)

KuraNAS distributes its companion apps from the server itself: the web UI has a
**Downloads** page (`/downloads`) that lists the installable clients, and the
backend serves the artifacts. This keeps distribution and updates in one place —
a user on the LAN installs everything straight from their own NAS, offline.

## What is distributed

| Artifact | Source app | How the user installs |
|---|---|---|
| `kuranas-android.apk` | `android/` (Kotlin/Compose, minSdk 33) | Download + open the APK |
| `kuranas-android-legacy.apk` | `mobile/` (legacy Java, minSdk 16) | Download + open the APK |
| `kuranas-extension.zip` | `plugin/` (MV3 extension) | Unzip + "Load unpacked" (steps shown on the page) |

The backend itself and the frontend are not distributed here — they ship as the
server build and are updated by the in-app `updater`.

## How it works

```
scripts/build-downloads.sh  ->  downloads/ (+ manifest.json)  ->  make move bundles it into build/
                                      |
                          backend serves it            web UI lists it
                          GET /api/v1/downloads         /downloads page
                          GET /api/v1/downloads/:id     (cards + download buttons)
```

- **The server never builds the apps.** Building an APK needs the Android SDK and
  packaging the extension needs the plugin toolchain — neither lives on the NAS.
  CI (or a maintainer) runs `scripts/build-downloads.sh`, which builds the
  artifacts, drops them in `downloads/`, and writes `downloads/manifest.json`.
- **Offline by default.** Artifacts live next to the binary in `downloads/` and
  are served by the NAS — no internet needed at download time.
- **Updates ride the existing updater.** The `updater` feature already replaces
  asset directories in place from a GitHub Release; `downloads` was added to that
  set, so publishing a new release refreshes the apps too.
- **Graceful when empty.** No `downloads/` directory ⇒ `GET /api/v1/downloads`
  returns `[]` and the page shows an empty state (e.g. the dev server).

## Manifest contract (`downloads/manifest.json`)

```json
{
  "artifacts": [
    {
      "id": "android",
      "platform": "android",
      "name_key": "DOWNLOAD_APP_ANDROID_NAME",
      "description_key": "DOWNLOAD_APP_ANDROID_DESC",
      "file": "kuranas-android.apk",
      "version": "1.0.0",
      "min_os": "Android 13",
      "sha256": "…"
    }
  ]
}
```

- `id` — stable identifier; the download URL is `/api/v1/downloads/<id>`.
- `platform` — `android` or `browser` (a `browser` entry triggers the "load
  unpacked" install instructions on the page).
- `name_key` / `description_key` — i18n keys resolved server-side from
  `backend/translations/*.json` (text is translated at the source). Omit them and
  the name falls back to the id.
- `file` — filename inside `downloads/`. The server ignores the URL `:id` for the
  filesystem path (it uses this field), and refuses anything outside `downloads/`.
- `size_bytes` is **not** in the manifest — the server fills it from the file on
  disk, so the catalog always reflects what is actually served. An artifact whose
  `file` is missing on disk is silently skipped.

## Building the bundle

```bash
# Build everything (needs Android SDK + ./gradlew in android/ and mobile/):
scripts/build-downloads.sh

# Only re-zip the extension and regenerate the manifest:
SKIP_GRADLE=1 scripts/build-downloads.sh
```

Versions/min-OS are overridable via env (`ANDROID_VERSION`, `ANDROID_MIN_OS`,
`MOBILE_VERSION`, `MOBILE_MIN_OS`, `PLUGIN_VERSION`). `downloads/` is gitignored;
`make all` copies it into `build/` when present.
