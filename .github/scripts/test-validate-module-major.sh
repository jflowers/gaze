#!/usr/bin/env bash

set -euo pipefail

repository_root="$(git rev-parse --show-toplevel)"
validator="$repository_root/.github/scripts/validate-module-major.sh"
temporary_directory="$(mktemp -d)"
trap 'rm -rf "$temporary_directory"' EXIT

expected_module="github.com/unbound-force/gaze/v2"
actual_module="$(cd "$repository_root" && go list -m)"
if [[ "$actual_module" != "$expected_module" ]]; then
  printf 'repository module mismatch: expected %s, got %s\n' \
    "$expected_module" "$actual_module" >&2
  exit 1
fi

if git -C "$repository_root" grep -n -E \
  'github\.com/unbound-force/gaze/(cmd|internal)(/|$)' \
  -- '*.go' '*.coverprofile'; then
  printf 'unsuffixed module-local path found; expected %s/...\n' \
    "$expected_module" >&2
  exit 1
fi

run_case() {
  local name="$1"
  local tag="$2"
  local module_path="$3"
  local expected_status="$4"
  local go_mod="$temporary_directory/$name.go.mod"

  printf 'module %s\n\ngo 1.25.0\n' "$module_path" >"$go_mod"

  set +e
  "$validator" -- "$tag" "$go_mod" >/dev/null 2>&1
  actual_status=$?
  set -e

  if [[ "$actual_status" -ne "$expected_status" ]]; then
    printf '%s: expected status %s, got %s\n' \
      "$name" "$expected_status" "$actual_status" >&2
    exit 1
  fi
}

run_case v2-match v2.2.0 github.com/unbound-force/gaze/v2 0
run_case v1-match v1.9.1 github.com/unbound-force/gaze 0
run_case v2-unsuffixed v2.2.0 github.com/unbound-force/gaze 1
run_case v3-wrong-suffix v3.0.0 github.com/unbound-force/gaze/v2 1
run_case malformed-tag release-2 github.com/unbound-force/gaze/v2 2

printf 'module-major and repository-state validation passed: 5 cases\n'
