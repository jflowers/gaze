## Context

Go requires modules at major version two or later to include `/vN` in the
module path. Gaze v2.1.0 was tagged with an unsuffixed `go.mod`, so that tag
cannot be installed through either the v1 or v2 module path.

## Goals / Non-Goals

### Goals

- Adopt the canonical `/v2` module identity consistently.
- Correct active installation guidance.
- Prevent future release-tag/module-major mismatches.

### Non-Goals

- Repairing or replacing the existing v2.1.0 tag.
- Adding a v1 compatibility shim.
- Automating post-publication proxy verification.
- Changing Homebrew, signing, notarization, or GoReleaser behavior.

## Decisions

### D1: Mechanical Module Migration

Change the `module` directive and exact module-local import/package prefixes.
Repository URLs, schema IDs, and Homebrew references remain unchanged and are
reviewed through the normal diff and code-review process.

### D2: Small Release Guard

A standalone shell script extracts the requested tag major, derives the exact
canonical module path, and compares it with the sole `go.mod` module directive.
The release preflight depends on this read-only check. Five offline cases cover
matching v1/v2, unsuffixed v2, a wrong higher-major suffix, and malformed input.

### D3: Durable Documentation

Current documentation uses the `/v2` install path and explains that a corrected
patch release is required. Historical specs remain unchanged unless they are
active executable guidance.

## Risks / Trade-offs

- The import migration is broad but mechanical; existing build, race, E2E, and
  lint gates detect missed paths.
- v2.1.0 remains unavailable through `go install`; users must use Homebrew,
  GitHub Release binaries, or the next corrected patch.
- The release guard intentionally validates only module-major consistency;
  reusable release preflight remains responsible for full SemVer validation.

## Coverage Strategy

- Five offline release-guard cases MUST pass without network access.
- Existing build, `-race -count=1` tests, E2E, lint, and security gates MUST
  remain green.
- No live public-proxy test is required before merge or release.
