# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

"KuraNAS Stream Grabber" — a Manifest V3 browser extension that detects/captures media streams on web pages and uploads them to a KuraNAS server. Plain JavaScript ES modules, no bundler.

## Commands (`cd plugin`)

- `npm run lint` — ESLint over the extension.
- `npm test` — `node --test tests/**/*.test.js` (Node's built-in test runner).

Load unpacked in a Chromium browser from this directory to run it; `manifest.json` is the entry manifest.

## Architecture

MV3 has three execution worlds, reflected in the layout:

- **Service worker** (`background.js`, `"type": "module"`) — the orchestrator. The real logic is in `src/background/`: `media-detection.js`, `message-router.js`, `hybrid-state.js`, `downloader.js`, `uploader.js`. Has broad permissions: `webRequest`, `tabs`, `tabCapture`, `offscreen`, `storage`, plus `host_permissions: <all_urls>`.
- **Content scripts** (injected on `<all_urls>` at `document_start`):
  - `content/bridge.js` runs in the `ISOLATED` world (talks to the extension).
  - `content/blob-interceptor.js` + `content/title-detector.js` run in the `MAIN` world (patch page APIs / read page state).
- **Offscreen document** (`offscreen/recorder.html` + `recorder.js`) — used for `tabCapture`-based recording, since a service worker can't access media APIs directly.
- **Popup** (`popup/popup.html` + `popup.js` + `popup.css`) — the toolbar UI.
- **Shared** (`src/shared/`): `constants.js`, `utils.js`.

## Talking to the backend

The extension uploads to the KuraNAS captures API. `src/background/uploader.js` posts to `${apiUrl}/captures/upload/init`, `/captures/upload/chunk`, `/captures/upload/complete` (chunked upload). The default base is `DEFAULT_KURANAS_API_BASE = "http://localhost:8000/api/v1"` (`src/shared/constants.js`). These endpoints are defined backend-side in `RegisterCapturesRoutes` — keep the two in sync when changing the upload contract.
</content>
