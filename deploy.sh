#!/usr/bin/env bash
set -euo pipefail

require_path() {
	local path="$1"
	if [ ! -e "$path" ]; then
		echo "missing required path: $path" >&2
		exit 1
	fi
}

require_cmd() {
	local cmd="$1"
	if ! command -v "$cmd" >/dev/null 2>&1; then
		echo "missing required command: $cmd" >&2
		exit 1
	fi
}

require_cmd go
require_cmd gofmt
require_path cmd/ff15
require_path internal/ff15
require_path openspec/changes

if ! find openspec/changes -mindepth 1 -print -quit | grep -q .; then
	echo "openspec/changes must contain at least one change artifact" >&2
	exit 1
fi

fmt_out="$(gofmt -l cmd/ff15 internal/ff15)"
if [ -n "$fmt_out" ]; then
	echo "gofmt reported unformatted files:" >&2
	echo "$fmt_out" >&2
	exit 1
fi

go test ./cmd/ff15 ./internal/ff15

mkdir -p dist
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false -o ff15 ./cmd/ff15
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -buildvcs=false -o ff15.exe ./cmd/ff15

echo "built ff15 and ff15.exe"
