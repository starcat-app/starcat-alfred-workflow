#!/usr/bin/env bash
#
# 构建 Alfred Workflow 使用的 macOS universal helper，并生成高对比度 fallback 图标。
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
bin_dir="${repo_dir}/bin"
assets_dir="${repo_dir}/assets"

# Go module resolution is based on cwd even when the package path is absolute.
# Always enter the repository so CI and package.sh can invoke this script from anywhere.
cd "${repo_dir}"
mkdir -p "${bin_dir}" "${assets_dir}"

CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 \
  go build -trimpath -ldflags="-s -w" -o "${bin_dir}/starcat-alfred-arm64" \
  ./cmd/starcat-alfred
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 \
  go build -trimpath -ldflags="-s -w" -o "${bin_dir}/starcat-alfred-amd64" \
  ./cmd/starcat-alfred

/usr/bin/lipo -create \
  "${bin_dir}/starcat-alfred-arm64" \
  "${bin_dir}/starcat-alfred-amd64" \
  -output "${bin_dir}/starcat-alfred"
chmod 755 "${bin_dir}/starcat-alfred"

"${bin_dir}/starcat-alfred" generate-assets "${assets_dir}/repo-fallback.png"
cp "${assets_dir}/repo-fallback.png" "${assets_dir}/workflow-icon.png"
