## 1. Module Migration

- [ ] 1.1 Change `go.mod` and module-local production imports to `/v2`.
- [ ] 1.2 Update test imports, package fixtures, and coverage fixtures.
- [ ] 1.3 Confirm repository URLs, schema IDs, and Homebrew references remain unchanged.

## 2. Release Guard

- [ ] 2.1 Add the module-major validation script.
- [ ] 2.2 Add five offline matching/mismatch cases.
- [ ] 2.3 Gate reusable release preflight on module-major validation.

## 3. Documentation

- [ ] 3.1 Update README and active installation/package-path documentation.
- [ ] 3.2 Synchronize Gaze-owned reporter prompt copies.
- [ ] 3.3 Record website and UF-owned guidance follow-ups.

## 4. Verification

- [ ] 4.1 Run build, race-enabled tests, E2E, lint, and security checks.
- [ ] 4.2 Run the review council and resolve all blocking findings.

<!-- spec-review: passed -->
