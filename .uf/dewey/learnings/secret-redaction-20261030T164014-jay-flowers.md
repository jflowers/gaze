---
tag: secret-redaction
author: jay-flowers
category: pattern
created_at: 2026-10-30T16:40:14Z
identity: secret-redaction-20261030T164014-jay-flowers
tier: draft
---

Defense-in-depth secret scrubbing of inherited env values layers two mitigations because verbatim-value replacement alone cannot catch fragment echoes. Tool-side: replace sensitive inherited env-var VALUES verbatim at a single choke point, before any size truncation (so a value straddling the truncation boundary cannot leak a prefix), classifying values by name markers (KEY/TOKEN/SECRET/CONFIG_CONTENT) plus a length threshold, with a garble guard skipping sub-minimum-length values. This catches full-value echoes but NOT fragments of a secret nested inside a larger env value (e.g., a bare apiKey echoed from inside a strict-JSON config secret such as OPENCODE_CONFIG_CONTENT) nor transformed values (base64-encoded, rewrapped). Close the fragment gap at the platform layer instead: document registering the bare key as a second CI secret referenced in an unused env var, so the platform masks verbatim occurrences anywhere in logs. Document residual under-redaction (sub-threshold, sub-minimum-length, transformed values) as explicit limitations rather than attempting brittle content-shape scrubbing. Pattern realized in gaze's redact-env-values-from-adapter-stderr change (unbound-force/gaze issue #301, follow-up to the PR #300 config-as-single-secret workflow).
