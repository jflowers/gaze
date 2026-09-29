<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Doc-Signal Derivation Package

- [x] 1.1 Create `internal/adapter/docsignal/grammar.go` — annotation grammar constants, regex patterns for the two syntax forms (HTML comment `<!-- gaze:label TypeName -->` and fenced block `> [!gaze-contract]`), the `DerivedSignal` keyed carrier type (`{Package, Function, SideEffectType string; Signal taxonomy.Signal}`), and the filename-based weight constants (architecture_doc ±25, readme ±15, sidecar ±30)
- [x] 1.2 Create `internal/adapter/docsignal/parse.go` — parse grammar annotations from `docscan.DocumentFile` into type-level declarations (type name + label + source token + source file + weight + reasoning), with filename-based source/weight assignment and priority-based conflict precedence
- [x] 1.3 Create `internal/adapter/docsignal/sidecar.go` — `LoadSidecar(moduleRoot string) []docsignal.DerivedSignal` supporting `.uf/gaze/contracts.yaml` and `.uf/gaze/contracts.json`, with semantic validation (skip + stderr warning on unknown label, empty package/function, or unknown side_effect_type), returning empty when absent
- [x] 1.4 Create `internal/adapter/docsignal/derive.go` — `DeriveSignals(docs []docscan.DocumentFile, cached []taxonomy.AnalysisResult, moduleRoot string) []docsignal.DerivedSignal` orchestrating parse + fan-out (type-level annotations → per-function tuples via `cached`) + sidecar merge with sidecar precedence
- [x] 1.5 [P] Create `internal/adapter/docsignal/parse_test.go` — table-driven tests for both grammar forms, malformed annotations, conflict precedence, empty documents
- [x] 1.6 [P] Create `internal/adapter/docsignal/sidecar_test.go` — tests for YAML/JSON loading, missing file, invalid file graceful degradation, semantic validation skip, sidecar precedence
- [x] 1.7 Create `internal/adapter/docsignal/derive_test.go` — end-to-end derivation tests with synthetic `docscan.DocumentFile` inputs, synthetic `AnalysisResult` fixtures for fan-out, and temp-dir sidecar fixtures

## 2. Adapter Integration

- [x] 2.1 Modify `internal/adapter/classify.go` — extend `mergeClassifications` to accept `docSignals []docsignal.DerivedSignal` (preserving the existing `cached`/`signals`/`cfg` parameters), reduce each `DerivedSignal` to its `Signal` field, and group them by `(Package, Function, SideEffectType)` alongside protocol signals before `classify.ComputeScore`, and extend the empty-signal early return (`len(signals) == 0`) to also account for `docSignals` so doc-signal-only classification proceeds when `classify_signals` is unavailable
- [x] 2.2 Modify `internal/adapter/sideeffect.go` — restructure `classifyAndMerge` to run doc-signal derivation (scan + `DeriveSignals`) unconditionally after the cache is populated, fetching `classify_signals` only when `caps.ClassifySignals` is true, and pass both signal sets into `mergeClassifications`
- [x] 2.3 Verify `internal/adapter/session.go` — confirm `rootDir` is already threaded to `NewExternalSideEffectAnalyzer`; no session change is required (remove any assumption of new threading)
- [x] 2.4 Modify `internal/protocol/testdata/fake_analyzer/main.go` — add a `--no-classify-signals` flag (mirroring `--no-doc-coverage`) that disables the `classify_signals` capability in the initialize response; do NOT change the default (`classify_signals: true`)
- [x] 2.5 [P] Update `internal/adapter/classify_test.go` — add cases covering doc-signal + protocol-signal additive merging, doc-signal-only (no classify_signals), and empty doc-signal no-op
- [x] 2.6 [P] Add integration test for end-to-end doc-signal → classification flow via the fake analyzer binary (using `--no-classify-signals` and an analyze payload that emits at least one effect with no inline `classification`, so doc signals are the sole classification source)

## 3. Documentation & Validation

- [x] 3.1 Add `docs/guides/doc-annotations.md` documenting the annotation grammar and sidecar format (`.uf/gaze/contracts.yaml`)
- [x] 3.2 Update `docs/guides/improving-scores.md` to reference the new doc-annotation capability for external-analyzer projects
- [x] 3.3 Update `AGENTS.md` "Recent Changes" with the new capability and package
- [x] 3.4 Verify constitution alignment: confirm Accuracy (deterministic, reproducible derivation), Minimal Assumptions (opt-in grammar, graceful degradation), Actionable Output (Source/Reasoning traceability), and Testability (isolated unit tests + fake-analyzer integration) hold for the final implementation
- [x] 3.5 Run `go build ./cmd/gaze`, `go vet ./...`, `golangci-lint run`, and `go test -race -count=1 -short ./...` locally to confirm CI parity before marking complete
- [x] 3.6 Update `docs/concepts/classification.md` to document the new doc-derived signal source (`architecture_doc`/`readme`/`sidecar`) and its weights for the external-analyzer path
- [x] 3.7 Update `docs/reference/configuration.md` to document the `.uf/gaze/contracts.yaml` sidecar schema, precedence, and graceful-degradation behavior
- [x] 3.8 Add `docs/guides/doc-annotations.md` to the Guides TOC in `docs/index.md`
- [x] 3.9 Create a `unbound-force/website` documentation issue for the new user-facing annotation grammar and sidecar capability (Website Documentation Gate)
- [x] 3.10 Add a coverage ratchet test for `internal/adapter/docsignal/` (mirroring the `TestSC003_MappingAccuracy` pattern) to enforce the 100% branch-coverage target (Constitution Principle IV)

<!-- spec-review: passed -->
<!-- code-review: passed -->
