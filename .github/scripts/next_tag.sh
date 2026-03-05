#!/usr/bin/env bash
set -euo pipefail

LAST_TAG=$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-version:refname | head -n 1)

if [ -z "${LAST_TAG}" ]; then
  NEXT_TAG="v0.1.0"
else
  VERSION=${LAST_TAG#v}
  MAJOR=$(echo "$VERSION" | cut -d. -f1)
  MINOR=$(echo "$VERSION" | cut -d. -f2)
  PATCH=$(echo "$VERSION" | cut -d. -f3)
  NEXT_TAG="v${MAJOR}.${MINOR}.$((PATCH + 1))"
fi

echo "LAST_TAG=${LAST_TAG}"
echo "NEXT_TAG=${NEXT_TAG}"
