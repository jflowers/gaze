## 1. Module Migration

- [x] 1.1 Change `go.mod` and module-local production imports to `/v2`.
- [x] 1.2 Update test imports, package fixtures, and coverage fixtures.
- [x] 1.3 Confirm repository URLs, schema IDs, and Homebrew references remain unchanged.

## 2. Release Guard

- [x] 2.1 Add the module-major validation script.
- [x] 2.2 Add five offline matching/mismatch cases.
- [x] 2.3 Gate reusable release preflight on module-major validation.

## 3. Documentation

- [x] 3.1 Update README and active installation/package-path documentation.
- [x] 3.2 Synchronize Gaze-owned reporter prompt copies.
- [x] 3.3 Record website and UF-owned guidance follow-ups.

## 4. Verification

- [x] 4.1 Run build, race-enabled tests, E2E, lint, and security checks.
- [x] 4.2 Run the review council and resolve all blocking findings.

<!-- spec-review: passed -->
<!-- code-review: passed -->
