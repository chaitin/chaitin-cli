#!/usr/bin/env bash

set -euo pipefail

repo_slug="chaitin/chaitin-cli"
install_name="chaitin-cli"
script_dir="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

log() {
  printf '%s\n' "$*" >&2
}

fail() {
  log "error: $*"
  exit 1
}

have_cmd() {
  command -v "$1" >/dev/null 2>&1
}

need_cmd() {
  have_cmd "$1" || fail "missing required command: $1"
}

run_windows_powershell_installer() {
  local ps_script="$script_dir/install-chaitin-cli.ps1"
  local ps_path="$ps_script"
  local ps_bin installed_path posix_path

  if have_cmd cygpath; then
    ps_path="$(cygpath -w "$ps_script")"
  fi

  if have_cmd powershell.exe; then
    ps_bin="powershell.exe"
  elif have_cmd pwsh; then
    ps_bin="pwsh"
  else
    fail "missing required command: powershell.exe or pwsh"
  fi

  installed_path="$($ps_bin -NoProfile -ExecutionPolicy Bypass -File "$ps_path")"
  [ -n "$installed_path" ] || fail "windows installer finished without an output path"

  if have_cmd cygpath; then
    posix_path="$(cygpath -u "$installed_path" 2>/dev/null || true)"
    if [ -n "$posix_path" ]; then
      export PATH="$(dirname "$posix_path"):${PATH:-}"
    fi
  fi

  printf '%s\n' "$installed_path"
}

detect_goos() {
  case "$(uname -s)" in
    Linux)
      printf 'linux\n'
      ;;
    Darwin)
      printf 'darwin\n'
      ;;
    MINGW*|MSYS*|CYGWIN*)
      printf 'windows\n'
      ;;
    *)
      fail "unsupported operating system: $(uname -s)"
      ;;
  esac
}

detect_goarch() {
  case "$(uname -m)" in
    x86_64|amd64)
      printf 'amd64\n'
      ;;
    arm64|aarch64)
      printf 'arm64\n'
      ;;
    *)
      fail "unsupported architecture: $(uname -m)"
      ;;
  esac
}

latest_tag() {
  local api_url response tag

  need_cmd curl
  api_url="https://api.github.com/repos/${repo_slug}/releases/latest"
  response="$(curl -fsSL "$api_url")" || fail "failed to query latest release from ${api_url}"
  tag="$(printf '%s' "$response" | tr -d '\n' | sed -nE 's/.*"tag_name"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/p')"
  [ -n "$tag" ] || fail "failed to parse latest release tag"
  printf '%s\n' "$tag"
}

normalize_tag() {
  local version="$1"

  [ -n "$version" ] || fail "release version is empty"
  case "$version" in
    v*)
      printf '%s\n' "$version"
      ;;
    *)
      printf 'v%s\n' "$version"
      ;;
  esac
}

append_unique() {
  local candidate="$1"
  local existing

  [ -n "$candidate" ] || return 0
  for existing in "${install_candidates[@]}"; do
    if [ "$existing" = "$candidate" ]; then
      return 0
    fi
  done
  install_candidates+=("$candidate")
}

build_install_candidates() {
  local path_dir
  local IFS=':'

  install_candidates=()

  append_unique "${CHAITIN_CLI_INSTALL_DIR:-}"
  append_unique "/usr/local/bin"
  append_unique "/opt/homebrew/bin"
  append_unique "$HOME/bin"
  append_unique "$HOME/.local/bin"

  for path_dir in ${PATH:-}; do
    [ -n "$path_dir" ] || continue
    [ "$path_dir" = "." ] && continue
    append_unique "$path_dir"
  done
}

install_file() {
  local source="$1"
  local destination_dir="$2"
  local use_sudo="$3"
  local destination_path="${destination_dir}/${install_name}"

  [ -n "$destination_dir" ] || return 1

  if [ "$use_sudo" = "true" ]; then
    sudo mkdir -p "$destination_dir"
    if have_cmd install; then
      sudo install -m 0755 "$source" "$destination_path"
    else
      sudo cp "$source" "$destination_path"
      sudo chmod 0755 "$destination_path"
    fi
  else
    mkdir -p "$destination_dir"
    if have_cmd install; then
      install -m 0755 "$source" "$destination_path"
    else
      cp "$source" "$destination_path"
      chmod 0755 "$destination_path"
    fi
  fi

  printf '%s\n' "$destination_path"
}

shell_rc_file() {
  local shell_name
  shell_name="$(basename "${SHELL:-}")"

  case "$shell_name" in
    zsh)
      printf '%s\n' "$HOME/.zshrc"
      ;;
    bash)
      if [ "$(detect_goos)" = "darwin" ]; then
        printf '%s\n' "$HOME/.bash_profile"
      else
        printf '%s\n' "$HOME/.bashrc"
      fi
      ;;
    fish)
      printf '%s\n' "$HOME/.config/fish/config.fish"
      ;;
    *)
      printf '%s\n' "$HOME/.profile"
      ;;
  esac
}

ensure_path_visible() {
  local dir="$1"
  local rc_file line

  case ":${PATH:-}:" in
    *":${dir}:"*)
      return 0
      ;;
  esac

  rc_file="$(shell_rc_file)"
  mkdir -p "$(dirname "$rc_file")"

  case "$rc_file" in
    */config.fish)
      line="fish_add_path ${dir}"
      ;;
    *)
      line="export PATH=\"${dir}:\\$PATH\""
      ;;
  esac

  if [ ! -f "$rc_file" ] || ! grep -F "$dir" "$rc_file" >/dev/null 2>&1; then
    printf '\n%s\n' "$line" >> "$rc_file"
    log "updated PATH in ${rc_file}"
  fi

  export PATH="${dir}:${PATH:-}"
}

extract_archive() {
  local archive_path="$1"
  local output_dir="$2"
  local goos="$3"

  case "$goos" in
    windows)
      need_cmd unzip
      unzip -q "$archive_path" -d "$output_dir"
      ;;
    *)
      need_cmd tar
      tar -xzf "$archive_path" -C "$output_dir"
      ;;
  esac
}

download_release_binary() {
  local goos="$1"
  local goarch="$2"
  local workdir="$3"
  local version tag archive_ext asset_name archive_path download_url extracted_dir found_binary

  need_cmd curl
  version="${CHAITIN_CLI_VERSION:-}"
  if [ -z "$version" ]; then
    version="$(latest_tag)"
  fi
  tag="$(normalize_tag "$version")"

  case "$goos" in
    windows)
      archive_ext="zip"
      ;;
    *)
      archive_ext="tar.gz"
      ;;
  esac

  asset_name="chaitin-cli_${tag}_${goos}_${goarch}.${archive_ext}"
  archive_path="${workdir}/${asset_name}"
  download_url="https://github.com/${repo_slug}/releases/download/${tag}/${asset_name}"

  log "downloading ${download_url}"
  curl -fsSL "$download_url" -o "$archive_path" || fail "failed to download ${download_url}"

  extracted_dir="${workdir}/extract"
  mkdir -p "$extracted_dir"
  extract_archive "$archive_path" "$extracted_dir" "$goos"

  found_binary="$(find "$extracted_dir" -type f \( -name "${install_name}" -o -name "${install_name}.exe" \) | head -n 1)"
  [ -n "$found_binary" ] || fail "failed to locate ${install_name} in downloaded archive"
  printf '%s\n' "$found_binary"
}

install_from_candidates() {
  local source_binary="$1"
  local destination_dir destination_path

  for destination_dir in "${install_candidates[@]}"; do
    [ -n "$destination_dir" ] || continue
    case "$destination_dir" in
      "$HOME"/*)
        continue
        ;;
    esac

    if [ -d "$destination_dir" ] && [ -w "$destination_dir" ]; then
      install_file "$source_binary" "$destination_dir" false
      return 0
    fi
  done

  if have_cmd sudo; then
    for destination_dir in "${install_candidates[@]}"; do
      [ -n "$destination_dir" ] || continue
      case "$destination_dir" in
        "$HOME"/*)
          continue
          ;;
      esac

      if destination_path="$(install_file "$source_binary" "$destination_dir" true 2>/dev/null)"; then
        printf '%s\n' "$destination_path"
        return 0
      fi
    done
  fi

  for destination_dir in "${install_candidates[@]}"; do
    [ -n "$destination_dir" ] || continue

    case "$destination_dir" in
      "$HOME"/*)
        if [ -d "$destination_dir" ] && [ -w "$destination_dir" ]; then
          install_file "$source_binary" "$destination_dir" false
          ensure_path_visible "$destination_dir"
          return 0
        fi

        if [ -d "$(dirname "$destination_dir")" ] && [ -w "$(dirname "$destination_dir")" ]; then
          destination_path="$(install_file "$source_binary" "$destination_dir" false)"
          ensure_path_visible "$destination_dir"
          printf '%s\n' "$destination_path"
          return 0
        fi
        ;;
    esac
  done

  destination_dir="$HOME/.local/bin"
  destination_path="$(install_file "$source_binary" "$destination_dir" false)"
  ensure_path_visible "$destination_dir"
  printf '%s\n' "$destination_path"
}

main() {
  local goos goarch tmpdir source_binary installed_path

  if have_cmd "$install_name"; then
    command -v "$install_name"
    return 0
  fi

  goos="$(detect_goos)"
  if [ "$goos" = "windows" ]; then
    run_windows_powershell_installer
    return 0
  fi

  goarch="$(detect_goarch)"
  build_install_candidates
  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT INT TERM
  source_binary="$(download_release_binary "$goos" "$goarch" "$tmpdir")"

  installed_path="$(install_from_candidates "$source_binary")"
  [ -n "$installed_path" ] || fail "installation finished without an output path"

  log "installed ${install_name} to ${installed_path}"
  printf '%s\n' "$installed_path"
}

install_candidates=()
main "$@"
