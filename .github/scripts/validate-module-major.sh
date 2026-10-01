#!/usr/bin/env bash

set -euo pipefail

if [[ "${1:-}" != "--" || $# -lt 2 || $# -gt 3 ]]; then
  printf 'usage: %s -- <release-tag> [go.mod path]\n' "$0" >&2
  exit 2
fi

release_tag="$2"
go_mod_path="${3:-go.mod}"
canonical_base="github.com/unbound-force/gaze"

if [[ ! "$release_tag" =~ ^v(0|[1-9][0-9]*)\. ]]; then
  printf 'invalid release tag: %s\n' "$release_tag" >&2
  exit 2
fi
major="${BASH_REMATCH[1]}"

mapfile -t module_paths < <(awk '$1 == "module" { print $2 }' "$go_mod_path")
if [[ ${#module_paths[@]} -ne 1 ]]; then
  printf 'expected exactly one module directive in %s\n' "$go_mod_path" >&2
  exit 2
fi

expected_module="$canonical_base"
if (( major >= 2 )); then
  expected_module="$canonical_base/v$major"
fi

if [[ "${module_paths[0]}" != "$expected_module" ]]; then
  printf 'module path mismatch: tag %s requires %s, found %s\n' \
    "$release_tag" "$expected_module" "${module_paths[0]}" >&2
  exit 1
fi

printf 'module path matches release tag: %s\n' "$expected_module"
