---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-10-30T20:12:33Z
identity: test-quality-20261030T201233-jay-flowers
tier: draft
---

Scrub-before-truncation regression tests should be self-verifying at every step: (1) compute the straddle offset from the live constants, never hardcoded byte counts, so a constant change re-derives the geometry; (2) t.Fatalf on setup if the straddle invariant fails (the secret must start before the cap and extend past it); (3) assert the truncation path actually engaged (e.g. the "... (truncated)" marker) BEFORE asserting the secret's absence — otherwise the test passes vacuously if the scrubbed buffer happens to land under the cap; (4) start leaked-prefix scans at the redaction minimum length so short prefixes that collide with unrelated suffix text (e.g. "te" inside "truncated") do not produce false failures, and document why the scan floor exists. Pattern realized in gaze's TestRunSubprocess_RedactsSecretStraddlingTruncationBoundary (internal/aireport/subprocess_test.go), guarding the redactEnvValues-before-maxAdapterStderrBytes-truncation invariant.
