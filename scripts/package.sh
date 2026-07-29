#!/usr/bin/env bash
#
# 只组装本地 .alfredworkflow；发布、上传与 GitHub Release 由独立发版流程负责。
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dist_dir="${repo_dir}/dist"
stage_dir="$(mktemp -d "${TMPDIR:-/tmp}/starcat-alfred.XXXXXX")"

cleanup() {
  rm -rf "${stage_dir}"
}
trap cleanup EXIT

"${repo_dir}/scripts/build.sh"
mkdir -p "${dist_dir}" "${stage_dir}/bin" "${stage_dir}/assets"
cp "${repo_dir}/info.plist" "${stage_dir}/info.plist"
cp "${repo_dir}/bin/starcat-alfred" "${stage_dir}/bin/starcat-alfred"
cp "${repo_dir}/assets/repo-fallback.png" "${stage_dir}/assets/repo-fallback.png"
cp "${repo_dir}/assets/workflow-icon.png" "${stage_dir}/icon.png"

(
  cd "${stage_dir}"
  /usr/bin/zip -qry "${dist_dir}/Starcat.alfredworkflow" .
)

shasum -a 256 "${dist_dir}/Starcat.alfredworkflow"
