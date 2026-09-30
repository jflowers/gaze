## Why

Gaze v2 tags were published while `go.mod` still declared the unsuffixed v1
module path. Go therefore resolves the documented unsuffixed `@latest` command
to v1.9.0 and rejects v2 installation. Existing v2.1.0 tags cannot be repaired;
a corrected patch release must use the canonical `/v2` module identity.

## What Changes

- Change the module declaration and module-local imports to
  `github.com/unbound-force/gaze/v2`.
- Update active installation and package-path documentation.
- Add a small release guard that rejects a tag whose major version does not
  match the canonical module path.
- Publish a new v2 patch release after merge.

## Capabilities

### New Capabilities

- `v2-module-identity`: Defines the canonical v2 module and import path.
- `release-major-module-validation`: Prevents another mismatched release tag.
- `v2-installation-documentation`: Documents the corrected install command and
  v1-to-v2 import boundary.

### Modified Capabilities

- None.

### Removed Capabilities

- None.

## Impact

- `go.mod`, Go imports, package fixtures, and active package-path examples move
  to `/v2`.
- The release workflow gains one read-only validation job before tag creation.
- Existing v1 consumers remain on the unsuffixed module path.
- Homebrew and GitHub Release installation remain unchanged.

## Constitution Alignment

- **Accuracy - PASS**: documentation identifies the version users install.
- **Minimal Assumptions - PASS**: follows Go semantic import versioning.
- **Actionable Output - PASS**: release mismatches report expected and observed paths.
- **Testability - PASS**: the release guard has five offline contract cases.
