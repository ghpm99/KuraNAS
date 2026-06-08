#!/usr/bin/env bash
#
# build-downloads.sh — assemble the offline app-distribution bundle.
#
# Produces a ./downloads/ directory containing the pre-built client apps and a
# manifest.json describing them. The backend's `distribution` feature serves
# this directory (GET /api/v1/downloads) and the web UI lists it on /downloads.
# `make all` bundles ./downloads/ into build/ when it exists, and the in-app
# updater syncs it from a GitHub Release on update.
#
# The server never builds these artifacts; this script (CI or a maintainer) does.
#
# Usage:
#   scripts/build-downloads.sh                 # build everything it can find
#   SKIP_GRADLE=1 scripts/build-downloads.sh   # only (re)zip the plugin + manifest
#   GRADLE_VARIANT=release scripts/build-downloads.sh  # once a keystore exists
#
# Requires: bash, zip, sha256sum (coreutils). Android builds also require the
# Android SDK + a working ./gradlew in android/ and mobile/.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="${ROOT_DIR}/downloads"
SKIP_GRADLE="${SKIP_GRADLE:-0}"

ANDROID_VERSION="${ANDROID_VERSION:-1.0.0}"
ANDROID_MIN_OS="${ANDROID_MIN_OS:-Android 13}"
MOBILE_VERSION="${MOBILE_VERSION:-1.0.0}"
MOBILE_MIN_OS="${MOBILE_MIN_OS:-Android 4.1}"
PLUGIN_VERSION="${PLUGIN_VERSION:-1.0.0}"

# GRADLE_VARIANT selects the build type. Default `debug` because it is signed
# with the Android debug key and installs on a device out of the box; a
# `release` APK is unsigned without a keystore and won't install. Switch to
# `release` only once a signing config exists.
GRADLE_VARIANT="${GRADLE_VARIANT:-debug}"

rm -rf "${OUT_DIR}"
mkdir -p "${OUT_DIR}"

# entries accumulates one JSON object per artifact actually produced.
entries=()

# add_entry id platform name_key description_key file version min_os
add_entry() {
    local id="$1" platform="$2" name_key="$3" desc_key="$4" file="$5" version="$6" min_os="$7"
    local abs="${OUT_DIR}/${file}"
    [ -f "${abs}" ] || { echo "  ! ${file} not found, skipping ${id}"; return; }
    local sha json
    sha="$(sha256sum "${abs}" | awk '{print $1}')"
    json="$(printf '{"id":"%s","platform":"%s","name_key":"%s","description_key":"%s","file":"%s","version":"%s","min_os":"%s","sha256":"%s"}' \
        "${id}" "${platform}" "${name_key}" "${desc_key}" "${file}" "${version}" "${min_os}" "${sha}")"
    entries+=("${json}")
    echo "  + ${id} -> ${file} (${sha:0:12}...)"
}

build_gradle_apk() {
    local module_dir="$1" out_name="$2"
    if [ "${SKIP_GRADLE}" = "1" ]; then
        echo "  (SKIP_GRADLE=1) reusing existing ${out_name} if present"
        return
    fi
    local task variant_dir
    case "${GRADLE_VARIANT}" in
        release) task="assembleRelease"; variant_dir="release" ;;
        *)       task="assembleDebug";   variant_dir="debug" ;;
    esac
    echo "  building ${module_dir} ${GRADLE_VARIANT} APK..."
    ( cd "${ROOT_DIR}/${module_dir}" && ./gradlew --no-daemon "${task}" )
    local apk
    apk="$(find "${ROOT_DIR}/${module_dir}/app/build/outputs/apk/${variant_dir}" -name '*.apk' | head -n1 || true)"
    [ -n "${apk}" ] && cp "${apk}" "${OUT_DIR}/${out_name}"
}

echo "Assembling ${OUT_DIR}"

echo "[android] modern app"
build_gradle_apk "android" "kuranas-android.apk"
add_entry "android" "android" "DOWNLOAD_APP_ANDROID_NAME" "DOWNLOAD_APP_ANDROID_DESC" \
    "kuranas-android.apk" "${ANDROID_VERSION}" "${ANDROID_MIN_OS}"

echo "[mobile] legacy app"
build_gradle_apk "mobile" "kuranas-android-legacy.apk"
add_entry "android-legacy" "android" "DOWNLOAD_APP_ANDROID_LEGACY_NAME" "DOWNLOAD_APP_ANDROID_LEGACY_DESC" \
    "kuranas-android-legacy.apk" "${MOBILE_VERSION}" "${MOBILE_MIN_OS}"

echo "[plugin] browser extension"
( cd "${ROOT_DIR}/plugin" && \
    zip -r -q "${OUT_DIR}/kuranas-extension.zip" . \
        -x 'node_modules/*' 'tests/*' 'package-lock.json' '*.map' )
add_entry "plugin" "browser" "DOWNLOAD_APP_PLUGIN_NAME" "DOWNLOAD_APP_PLUGIN_DESC" \
    "kuranas-extension.zip" "${PLUGIN_VERSION}" ""

# Join the entries into downloads/manifest.json.
{
    printf '{\n  "artifacts": [\n'
    for i in "${!entries[@]}"; do
        sep=","
        [ "${i}" -eq $(( ${#entries[@]} - 1 )) ] && sep=""
        printf '    %s%s\n' "${entries[$i]}" "${sep}"
    done
    printf '  ]\n}\n'
} > "${OUT_DIR}/manifest.json"

echo "Wrote ${OUT_DIR}/manifest.json with ${#entries[@]} artifact(s)."
