#!/bin/bash
set -e

platforms=(
  darwin-amd64
  darwin-arm64
  linux-386
  linux-amd64
  linux-arm64
  windows-386
  windows-amd64
)

IFS=$'\n' read -d '' -r -a supported_platforms < <(go tool dist list) || true

for p in "${platforms[@]}"; do
	goos="${p%-*}"
	goarch="${p#*-}"
	if [[ " ${supported_platforms[*]} " != *" ${goos}/${goarch} "* ]]; then
		echo "warning: skipping unsupported platform $p" >&2
		continue
	fi
	ext=""
	if [ "$goos" = "windows" ]; then
		ext=".exe"
	fi
	GOOS="$goos" GOARCH="$goarch" go build -trimpath -ldflags="-s -w" -o "dist/${p}${ext}"
done
