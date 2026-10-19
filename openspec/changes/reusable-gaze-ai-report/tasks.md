# Tasks: Reusable Gaze AI Report Workflow

## 1. Reusable workflow

- [x] 1.1 Create `.github/workflows/reusable_gaze_ai_report.yml` with a
  `workflow_call` trigger exposing inputs: `model` (string, optional),
  `coverage-artifact` (string, required), `coverprofile` (string,
  default `coverage.out`), `max-crapload`, `max-gaze-crapload`,
  `min-contract-coverage` (number, optional), `opencode-version`
  (string, default `1.19.0`), `gaze-version` (string, default
  `v2.2.0`), `go-version` (string, default `1.25`), and secret
  `opencode-config` (optional) carrying the consumer's full opencode
  config as strict JSON, including the literal API key.
- [x] 1.2 Add the `gaze-ai-report` job (ubuntu-latest, `permissions:
  contents: read`, `timeout-minutes: 30`, job env `HAS_OPENCODE_CONFIG:
  ${{ secrets.opencode-config != '' }}`) with steps: Checkout
  (actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v4.2.2),
  Setup Go (actions/setup-go@40f1582b2485089dde7abd97c1529aa768e1baff #
  v5.6.0, `go-version` from the input), Download coverage
  (actions/download-artifact@9000827ccba6bdab643e8b6fd33ac0654aef8333 #
  v8.0.2, `name` from the `coverage-artifact` input, configured so
  `coverage.out` extracts to the repository root), and Install gaze
  (`go install github.com/unbound-force/gaze/v2/cmd/gaze` at the
  `gaze-version` input).
- [x] 1.3 Add the report steps gated on `env.HAS_OPENCODE_CONFIG ==
  'true'`: Install OpenCode (`npm install -g opencode-ai` at the
  `opencode-version` input) and Run gaze AI report, which unsets
  `GITHUB_TOKEN`/`GH_TOKEN`, exports `OPENCODE_CONFIG_CONTENT` (from
  the `opencode-config` secret) and `GAZE_MODEL` (from the optional
  `model` input), assembles the `--model` arg and threshold flags from
  env vars only when set, and runs `gaze report ./... --ai=opencode
  [--model=<model>] --coverprofile=<coverprofile> [thresholds]`.
- [x] 1.4 Add a skip-notice step (`if: env.HAS_OPENCODE_CONFIG !=
  'true'`) stating that the AI report is skipped because no opencode
  config secret is available.

## 2. Dogfood in test.yml

- [x] 2.1 In `unit-and-integration`, after the "Gaze threshold check"
  step, add an "Upload coverage" step (`if: matrix.go-version ==
  '1.24'`) uploading `coverage.out` via
  actions/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02 #
  v4.6.2 with `name: coverage-out`.
- [x] 2.2 Remove from `unit-and-integration`: "Authenticate to Google
  Cloud", "Setup GCloud", "Install OpenCode", and "Gaze quality report"
  steps; the GCP/Vertex/Anthropic env vars and `HAS_GCP_SECRETS`; and
  `id-token: write` from the job permissions.
- [x] 2.3 Add the `gaze-ai-report` job: `needs: unit-and-integration`,
  `permissions: contents: read`, `uses:
  ./.github/workflows/reusable_gaze_ai_report.yml`, `with`:
  `coverage-artifact: coverage-out`, `max-crapload: 38`,
  `max-gaze-crapload: 6`, `min-contract-coverage: 8`; `secrets`:
  `opencode-config: ${{ secrets.OPENCODE_CONFIG_CONTENT }}` (org
  secret carrying the EnMaaS/GLM 5.3 config with permission denials
  and the literal API key).

Amendment (PR review feedback): a workflow-level `concurrency:` group
(`group: ${{ github.ref }}-${{ github.workflow }}`,
`cancel-in-progress: ${{ github.ref != 'refs/heads/main' }}`) was added
to `test.yml` per CI-012 — the `gaze-ai-report` job spends LLM API
budget, so superseded PR runs cancel while main runs queue. The
reusable workflow itself defines no group; a called workflow joins the
caller's run, so the policy is caller-owned.

## 3. Documentation

- [x] 3.1 Add a "Reusable workflow" section to
  `docs/guides/ci-integration.md` documenting the inputs, the
  `opencode-config` secret (canonical config shape, strict-JSON and
  permission-denial and literal-key notes), the coverage artifact
  contract, and a consumer example.

## 4. Verification

- [x] 4.1 `openspec validate reusable-gaze-ai-report` passes.
- [x] 4.2 The changed workflow YAML passes yamllint (MegaLinter
  parity).
- [x] 4.3 `go build ./...` still passes (no Go changes; sanity).
- [x] 4.4 Spec artifacts committed and pushed before implementation.
- [ ] 4.5 After PR: verify the PR run executes the artifact handoff
  and skips the report steps (fork PR, no secret); after merge, verify
  the push run produces the EnMaaS quality report. Also verify once on
  a same-repo branch (org secret resolves): (a) omitting all threshold
  inputs runs the report in report-only mode — unset `type: number`
  inputs render as empty env vars, so the flags are omitted — and
  (b) a threshold breach (e.g. `max-crapload: 0`) fails the
  `gaze-ai-report` job.

## 5. Amendment: config as the single org secret

The contract was amended before merge: the consumer's entire opencode
config — provider, model, permissions, and the literal API key —
travels as one `opencode-config` secret exported verbatim as
`OPENCODE_CONFIG_CONTENT`. The `opencode-config` input, the
`provider-key` secret, and the `OPENCODE_PROVIDER_KEY` env var are
eliminated; `model` becomes an optional input. Tasks 1.1-1.4, 2.3,
and 3.1 above describe the final contract.

- [x] 5.1 Dry run the amended contract end-to-end: strict-JSON config
  with the literal API key and permission denials, fresh `HOME`, only
  `OPENCODE_CONFIG_CONTENT` exported (no `OPENCODE_PROVIDER_KEY`, no
  `ENMAAS_KEY`, no `--model` flag); gaze produced the full report
  using the config's top-level model.
- [x] 5.2 Amend `.github/workflows/reusable_gaze_ai_report.yml`,
  `test.yml`, and `docs/guides/ci-integration.md` to the amended
  contract.
- [x] 5.3 `openspec validate reusable-gaze-ai-report --strict`,
  yamllint on both changed workflows, and a JSON parse of the
  canonical config example pass after the amendment.
