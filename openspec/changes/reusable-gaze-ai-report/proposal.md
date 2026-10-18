# Reusable Gaze AI Report Workflow

## Why

The AI quality report in `test.yml` is four GCP-gated steps inside the
`unit-and-integration` job: Workload Identity Federation auth,
setup-gcloud, opencode install, and the report itself. The chain
hard-codes the google-vertex-anthropic provider and cannot be reused by
other repositories in the organization. Unbound-force consumes a CRAP
gate from complytime/org-infra but has no AI report, and wants one on
EnMaaS (GLM 5.3) rather than GCP Vertex.

A provider-agnostic reusable workflow lets gaze own and publish the
report machinery once, dogfood it from its own CI, and let any repository
consume it with any opencode-compatible provider.

Local dry runs validated the approach: the compact AI payload for
unbound-force is 169,227 bytes (roughly 42-55K tokens, inside GLM 5.3's
262,144-token context), and an end-to-end run driven by
`OPENCODE_CONFIG_CONTENT` with EnMaaS produced a complete quality
report.

## What Changes

1. Add `.github/workflows/reusable_gaze_ai_report.yml`, a
   provider-agnostic reusable workflow (`workflow_call`) that installs
   pinned opencode and gaze, downloads a coverage artifact from the
   calling run, and runs `gaze report ./... --ai=opencode` with
   caller-supplied thresholds and a single `opencode-config` secret
   carrying the consumer's entire opencode configuration (provider,
   model, API key, permissions). Threshold breaches fail the job.
2. Dogfood the workflow from `test.yml`: upload `coverage.out` from the
   Go 1.24 matrix leg and add a `gaze-ai-report` job that calls the
   reusable workflow with thresholds 38/6/8, sourcing the provider and
   model from the org-level `OPENCODE_CONFIG_CONTENT` secret (EnMaaS
   GLM 5.3, `enmaas/rits/zai-org/glm-5-3`).
3. Remove the GCP-specific machinery from `unit-and-integration` (WIF
   auth, setup-gcloud, opencode install, in-job AI report), the GCP env
   vars, and the `id-token: write` permission. The PR-safe "Gaze
   threshold check" step is unchanged.
4. Document the reusable workflow in `docs/guides/ci-integration.md`.

## Capabilities

### New Capabilities

- `.github/workflows/reusable_gaze_ai_report.yml`: provider-agnostic
  reusable AI quality report workflow published for org-wide consumption

### Modified Capabilities

- `.github/workflows/test.yml`: the AI report moves from in-job
  GCP-gated steps to a separate job consuming the reusable workflow;
  `coverage.out` is uploaded as an artifact; GCP credentials machinery
  is removed

### Removed Capabilities

- None

## Impact

- `.github/workflows/reusable_gaze_ai_report.yml` — added
- `.github/workflows/test.yml` — modified
- `docs/guides/ci-integration.md` — modified
- Organization secret `OPENCODE_CONFIG_CONTENT` — added by the org
  maintainer; carries the full opencode config including the literal
  API key (not in files)

## Constitution Alignment

Assessed against the Gaze constitution v1.3.0.

### I. Accuracy

**Assessment**: N/A

CI workflow configuration only; no analysis behavior changes.

### II. Minimal Assumptions

**Assessment**: PASS

The reusable workflow assumes nothing about the caller's provider:
model, opencode config, thresholds, and versions are all inputs. Its
single contractual assumption — a coverage artifact in the calling run —
is explicit and documented.

### III. Actionable Output

**Assessment**: PASS

The same human-readable quality narrative is produced, now published to
the job summary of any consuming repository rather than only gaze's own
CI, and still backed by the PR-safe threshold check.

### IV. Testability

**Assessment**: PASS

The `gaze-ai-report` job runs on PRs as well as pushes: with no secret
available (fork PRs), the report steps skip but checkout, tool setup,
and the artifact handoff still execute, validating the wiring pre-merge.
The full report path runs on push to main.
