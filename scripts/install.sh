#!/usr/bin/env bash

set -euo pipefail

readonly default_repository="edram/qc"
repository="${QC_REPOSITORY:-$default_repository}"
version="${QC_VERSION:-latest}"
install_dir="${QC_INSTALL_DIR:-$HOME/.local/bin}"

die() {
  printf 'qc install: %s\n' "$*" >&2
  exit 1
}

usage() {
  cat <<'EOF'
Install qc from GitHub Releases.

Usage:
  install.sh [--version VERSION] [--install-dir DIRECTORY]

Options:
  --version VERSION       Install VERSION (with or without the leading v).
  --install-dir DIRECTORY Install into DIRECTORY (default: ~/.local/bin).
  -h, --help              Show this help.

Environment:
  QC_VERSION              Same as --version.
  QC_INSTALL_DIR          Same as --install-dir.
  QC_REPOSITORY           GitHub repository in OWNER/REPO form (default: edram/qc).
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version)
      [ "$#" -ge 2 ] || die "--version requires a value"
      version="$2"
      shift 2
      ;;
    --install-dir)
      [ "$#" -ge 2 ] || die "--install-dir requires a value"
      install_dir="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown option: $1"
      ;;
  esac
done

[ -n "$version" ] || die "version cannot be empty"

command -v curl >/dev/null 2>&1 || die "curl is required"
command -v tar >/dev/null 2>&1 || die "tar is required"

case "$repository" in
  */*) ;;
  *) die "QC_REPOSITORY must use OWNER/REPO form" ;;
esac
case "$repository" in
  *\?*|*\#*) die "QC_REPOSITORY must use OWNER/REPO form" ;;
esac

case "$(uname -s)" in
  Linux) os="linux" ;;
  Darwin) os="darwin" ;;
  *) die "unsupported operating system: $(uname -s)" ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) die "unsupported architecture: $(uname -m)" ;;
esac

case "$version" in
  latest)
    release_json="$(curl -fsSL --retry 3 --proto '=https' --tlsv1.2 \
      -H 'Accept: application/vnd.github+json' \
      "https://api.github.com/repos/${repository}/releases/latest")" \
      || die "could not find the latest GitHub release"
    tag="$(printf '%s\n' "$release_json" | sed -nE 's/^[[:space:]]*"tag_name":[[:space:]]*"([^"]+)".*$/\1/p' | head -n 1)"
    [ -n "$tag" ] || die "GitHub did not return a release tag"
    ;;
  v*)
    tag="$version"
    ;;
  *)
    tag="v$version"
    ;;
esac

case "$tag" in
  *[!A-Za-z0-9._-]*) die "unsupported release tag: $tag" ;;
esac

release_version="${tag#v}"
archive="qc_${release_version}_${os}_${arch}.tar.gz"
base_url="https://github.com/${repository}/releases/download/${tag}"

tmp_dir="$(mktemp -d 2>/dev/null || mktemp -d -t qc-install)"
cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT

archive_path="$tmp_dir/$archive"
checksums_path="$tmp_dir/checksums.txt"
extract_dir="$tmp_dir/extracted"

curl -fsSL --retry 3 --proto '=https' --tlsv1.2 \
  "$base_url/$archive" -o "$archive_path" \
  || die "could not download $archive"
curl -fsSL --retry 3 --proto '=https' --tlsv1.2 \
  "$base_url/checksums.txt" -o "$checksums_path" \
  || die "could not download checksums.txt"

expected_checksum="$(awk -v archive="$archive" '$2 == archive { print $1; exit }' "$checksums_path")"
[ -n "$expected_checksum" ] || die "checksums.txt does not contain $archive"

if command -v sha256sum >/dev/null 2>&1; then
  actual_checksum="$(sha256sum "$archive_path" | awk '{ print $1 }')"
elif command -v shasum >/dev/null 2>&1; then
  actual_checksum="$(shasum -a 256 "$archive_path" | awk '{ print $1 }')"
else
  die "sha256sum or shasum is required to verify the release"
fi

[ "$actual_checksum" = "$expected_checksum" ] \
  || die "checksum verification failed for $archive"

mkdir -p "$extract_dir"
tar -xzf "$archive_path" -C "$extract_dir"
[ -f "$extract_dir/qc" ] || die "release archive does not contain qc"

mkdir -p "$install_dir"
temporary_binary="$install_dir/.qc.$$"
cp "$extract_dir/qc" "$temporary_binary"
chmod 0755 "$temporary_binary"
mv -f "$temporary_binary" "$install_dir/qc"

printf 'Installed qc %s to %s/qc\n' "$tag" "$install_dir"
case ":${PATH:-}:" in
  *":${install_dir}:"*) ;;
  *) printf 'Add %s to PATH to run qc from any shell.\n' "$install_dir" ;;
esac
