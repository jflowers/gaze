---
tag: secret-redaction
author: jay-flowers
category: gotcha
created_at: 2026-10-30T20:12:22Z
identity: secret-redaction-20261030T201222-jay-flowers
tier: draft
---

GitHub Actions' automatic secret masking — including the second-secret registration pattern (referencing a bare API key in an unused env var of a step) — is scoped to job logs. Content a tool writes to $GITHUB_STEP_SUMMARY is a rendered file outside that log-masking guarantee, so a subprocess-output redaction choke point must treat stdout-to-step-summary as an unmasked channel when enumerating coverage. Verified during gaze's redact-env-values-from-adapter-stderr review: runSubprocess scrubs both stderr surfacing paths, but child stdout flows unscrubbed into the AI report that runTextPath writes to CLI stdout and appends to $GITHUB_STEP_SUMMARY. When a channel is deliberately left unscrubbed because verbatim value replacement would risk corrupting legitimate product content (workspace paths and long identifiers in a report), that scope decision must be recorded in the design with its rationale — an unrecorded exclusion is indistinguishable from an oversight and creates false coverage expectations for consumers reading the mitigation docs.
