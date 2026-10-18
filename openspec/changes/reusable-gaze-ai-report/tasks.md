# Tasks: Reusable Gaze AI Report Workflow

## 1. Reusable workflow

- [x] 1.1 Create `.github/workflows/reusable_gaze_ai_report.yml` with a
  `workflow_call` trigger exposing inputs: `model` (string, required),
  `opencode-config` (string, required), `coverage-artifact` (string,
  required), `coverprofile` (string, default `coverage.out`),
  `max-crapload`, `max-gaze-crapload`, `min-contract-coverage` (number,
  optional), `opencode-version` (string, default `1.19.0`),
  `gaze-version` (string, default `v2.2.0`), `go-version` (string,
  default `1.25`), and secret `provider-key` (optional).
- [x] 1.2 Add the `gaze-ai-report` job (ubuntu-latest, `permissions:
  contents: read`, `timeout-minutes: 30`, job env `HAS_PROVIDER_KEY:
  ${{ secrets.provider-key != '' }}`) with steps: Checkout
  (actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v4.2.2),
  Setup Go (actions/setup-go@40f1582b2485089dde7abd97c1529aa768e1baff #
  v5.6.0, `go-version` from the input), Download coverage
  (actions/download-artifact@9000827ccba6bdab643e8b6fd33ac0654aef8333 #
  v8.0.2, `name` from the `coverage-artifact` input, configured so
  `coverage.out` extracts to the repository root), and Install gaze
  (`go install github.com/unbound-force/gaze/v2/cmd/gaze` at the
  `gaze-version` input).
- [x] 1.3 Add the report steps gated on `env.HAS_PROVIDER_KEY ==
  'true'`: Install OpenCode (`npm install -g opencode-ai` at the
  `opencode-version` input) and Run gaze AI report, which unsets
  `GITHUB_TOKEN`/`GH_TOKEN`, exports `OPENCODE_CONFIG_CONTENT` (from
  the `opencode-config` input) and `OPENCODE_PROVIDER_KEY` (from the
  `provider-key` secret), assembles threshold flags from env vars only
  when the inputs are set, and runs `gaze report ./... --ai=opencode
  --model=<model> --coverprofile=<coverprofile> [thresholds]`.
- [x] 1.4 Add a skip-notice step (`if: env.HAS_PROVIDER_KEY != 'true'`)
  stating that the AI report is skipped because no provider key is
  available.

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
  ./.github/workflows/reusable_gaze_ai_report.yml`, `with`: `model:
  enmaas/rits/zai-org/glm-5-3`, `opencode-config` set to the EnMaaS
  JSON (options-nested baseURL/apiKey referencing
  `{env:OPENCODE_PROVIDER_KEY}`, model limits context 262144 / output
  65536, permission denials for edit/bash/webfetch/websearch),
  `coverage-artifact: coverage-out`, `max-crapload: 38`,
  `max-gaze-crapload: 6`, `min-contract-coverage: 8`; `secrets`:
  `provider-key: ENMAAS_KEY`.

## 3. Documentation

- [x] 3.1 Add a "Reusable workflow" section to
  `docs/guides/ci-integration.md` documenting the inputs, the
  `provider-key` secret, the coverage artifact contract, and a consumer
  example.

## 4. Verification

- [x] 4.1 `openspec validate reusable-gaze-ai-report` passes.
- [x] 4.2 The changed workflow YAML passes yamllint (MegaLinter
  parity).
- [x] 4.3 `go build ./...` still passes (no Go changes; sanity).
- [x] 4.4 Spec artifacts committed and pushed before implementation.
- [ ] 4.5 After PR: verify the PR run executes the artifact handoff
  and skips the report steps (fork PR, no secret); after merge, verify
  the push run produces the EnMaaS quality report.
