# Design: Reusable Gaze AI Report Workflow

## Context

The current AI report in `test.yml` is four steps inside
`unit-and-integration`, gated on GCP Workload Identity Federation
secrets and hard-coded to google-vertex/claude-opus-4-6. Other org
repositories want the same report with different providers
(unbound-force wants EnMaaS GLM 5.3) but cannot reuse steps buried in
gaze's job. The org already has a reusable-workflow precedent:
unbound-force consumes
complytime/org-infra/.github/workflows/reusable_crapload_analysis.yml.

## Goals / Non-Goals

### Goals

- Publish the AI report machinery as a provider-agnostic reusable
  workflow
- Dogfood it from gaze's own CI on push to main with EnMaaS/GLM 5.3
- Keep the PR-safe JSON threshold check exactly as it is
- Let any org repository consume the workflow with any opencode
  provider
- Validate the wiring on PRs without requiring secrets

### Non-Goals

- Running the AI report on fork PRs (secrets remain unavailable there)
- Changing threshold values (38/6/8)
- Changing gaze report or adapter Go code
- Supporting callers that cannot produce a coverage artifact

## Decisions

### D1: Provider-agnostic contract

`reusable_gaze_ai_report.yml` exposes `workflow_call` with inputs
`model` (optional), `coverage-artifact` (required), `coverprofile`
(default `coverage.out`), `max-crapload`, `max-gaze-crapload`,
`min-contract-coverage` (optional), `opencode-version` (default
`1.19.0`), `gaze-version` (default `v2.2.0`), `go-version` (default
`1.25`), and one optional secret `opencode-config`.

The `opencode-config` secret carries the consumer's entire opencode
configuration as strict JSON — provider, model, permissions, and the
literal API key. The workflow exports it verbatim as
`OPENCODE_CONFIG_CONTENT` for the report step, so the consumer
controls everything the report needs and the workflow dictates as
little as possible. Because the key lives inside the config, no
second secret, no fixed env-var name (`OPENCODE_PROVIDER_KEY`), and
no repo-level key setup exists for any consumer.

The `model` input stays optional: when set it is passed as `--model`;
when unset gaze omits the flag and opencode resolves the model from
the config's top-level `model` field (validated end-to-end in the
dry run). The recommended config shape embeds permission denials
(edit/bash/webfetch/websearch) — with key material inside the config,
denying tool access is more valuable — and the denials were verified
harmless in the dry run.

Secrets are not available in job-level `if:` conditions, so
secret-presence gating uses env indirection at step level
(`HAS_OPENCODE_CONFIG`), the same pattern `test.yml` already uses for
`HAS_GCP_SECRETS`.

Trade-off: the config leaves code review (a write-only org secret).
Accepted — the secret is org-admin-controlled and
`docs/guides/ci-integration.md` documents the canonical shape with a
placeholder key.

### D2: Coverage via artifact handoff

A reusable workflow runs on its own runner; the calling job cannot
share a filesystem with it. The workflow therefore downloads the
coverprofile from the caller-named artifact in the calling workflow
run. Same-run artifact downloads use the runtime token and need no
extra permissions. Callers upload `coverage.out` from whichever job ran
the tests; the workflow never runs the test suite itself.

### D3: Dogfood job replaces the GCP steps

- `unit-and-integration` uploads `coverage.out` from the Go 1.24 leg
  (the same leg that runs the threshold check).
- A new `gaze-ai-report` job (`needs: unit-and-integration`,
  `permissions: contents: read`) calls the reusable workflow with
  thresholds 38/6/8, mapping the org secret `OPENCODE_CONFIG_CONTENT`
  to `opencode-config`; the model and provider (EnMaaS GLM 5.3) come
  from that config.
- The four GCP-gated steps, the GCP env vars, and `id-token: write` are
  removed from `unit-and-integration`.
- The job runs on PRs and pushes. On fork PRs the secret is absent and
  the report steps skip with a notice, but checkout, setup-go, gaze
  install, and the artifact download still execute, validating the
  handoff pre-merge. The full report runs on push to main where
  secrets exist.

### D4: Version pins

opencode-ai@1.19.0 (validated end-to-end with EnMaaS in the dry run),
gaze v2.2.0 (latest tag; main is seven docs-only commits ahead), and
SHA-pinned actions matching repo precedent: checkout v4.2.2, setup-go
v5.6.0 (current test.yml), upload-artifact v4.6.2 (mega-linter.yml),
download-artifact v8.0.2 (proven in unbound-force). Both artifact
actions use the post-v4 artifact API, so the v4.6.2/v8.0.2 pair is
compatible.

### D5: Threshold enforcement

The report command carries the same thresholds as the JSON check; gaze
exits non-zero on breach and fails the push-to-main job. This is a
backstop to the PR-side threshold check; PR gating is unchanged because
the report does not run on PRs.

## Risks / Trade-offs

- **First real EnMaaS run is post-merge**: fork PRs cannot access the
  org secret, so the report only executes after merge to main. Local
  dry runs (169 KB payload; end-to-end report with the literal-key
  org-secret config and no `--model` flag verified) mitigate.
- **Config leaves code review**: the org secret is write-only, so
  provider/model/key changes are not PR-reviewable. Accepted — the
  secret is org-admin-controlled and the canonical shape is
  documented in `docs/guides/ci-integration.md`.
- **PR overhead**: the `gaze-ai-report` job adds roughly two minutes to
  PR runs (checkout, setup-go, gaze install, artifact download) to
  validate the wiring. Accepted for pre-merge handoff coverage.
- **Artifact action version pair**: upload v4.6.2 with download v8.0.2
  is exercised on the first PR run of this change; the job runs on PRs
  precisely to validate this pair before merge.
