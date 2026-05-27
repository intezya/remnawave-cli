#!/usr/bin/env bash
set -euo pipefail

OPENAPI_URL="${OPENAPI_URL:-https://cdn.docs.rw/docs/openapi.json}"
GENERATOR_REPO="${GENERATOR_REPO:-https://github.com/intezya/openapi-cli-generator.git}"
GENERATOR_REF="${GENERATOR_REF:-084d277563370eae89e13c8024b3cf7290b2cb24}"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

curl --fail --location --silent --show-error "$OPENAPI_URL" --output "$ROOT_DIR/remnawave-api.json"

git clone --quiet "$GENERATOR_REPO" "$TMP_DIR/openapi-cli-generator"
git -C "$TMP_DIR/openapi-cli-generator" checkout --quiet "$GENERATOR_REF"
(
	cd "$TMP_DIR/openapi-cli-generator"
	go install .
)

(
	cd "$ROOT_DIR"
	openapi-cli-generator generate remnawave-api.json
	gofmt -w remnawave-api.go
)
