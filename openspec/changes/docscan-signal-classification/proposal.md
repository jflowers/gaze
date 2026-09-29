## Why

External-analyzer projects (e.g., Python via snake-eyes) report a large share of their detected side effects as `ambiguous` — snake-eyes classifies 56.7% of detected effects (1186/2091) as `ambiguous`. The repository's own spec/design docs already encode the missing intent: the snake-eyes container-mutation design docs explicitly mark 10 `ContainerMutation` effects as observable/contractual and 14 as internal bookkeeping/incidental. `gaze docscan` already discovers and reads those documents (74 for snake-eyes), but that output is consumed only by the AI report pipeline — it never reaches classification. So documented contracts and documented internal bookkeeping stay `ambiguous`.

The scoring engine is already source-agnostic. `classify.ComputeScore` sums every `Signal.Weight` regardless of `Source`. `adapter.mergeClassifications` copies external-analyzer signals' `Source`/`Weight`/`Reasoning` verbatim into `ComputeScore`. The source vocabulary already anticipates doc-derived sources — `taxonomy.Signal.Source` lists `"readme"` and `"architecture_doc"` as examples, but no extractor produces them. The only doc-derived weight convention today is Go-native (`internal/classify/godoc.go`, `maxGodocWeight=15`), called only from the `goprovider` path.

Adding a doc → signal derivation step that maps documented public-API contracts → positive signal and documented internal bookkeeping → negative signal, keyed to `(package, function, side_effect_type)`, will resolve a large fraction of the `ambiguous` classification without any protocol changes. Note: this benefit is realized only after projects retrofit their docs with the deterministic annotation grammar (`<!-- gaze:... -->`) or the `.uf/gaze/contracts.yaml` sidecar; the grammar performs no NLP, so intent expressed only in natural-language prose is not extracted.

Closes #285.

## What Changes

1. Define a deterministic doc annotation grammar that allows spec/design docs to declare side effects as contractual or incidental, keyed by analyzer-emitted type name (e.g., `ContainerMutation`). The grammar is embedded directly in existing Markdown docs.

2. Add a machine-readable sidecar override mechanism (`.uf/gaze/contracts.yaml` or `.uf/gaze/contracts.json`) that allows explicit `(package, function, side_effect_type) → label` declarations, overriding or extending the grammar-derived signals.

3. Add a doc → signal derivation step in the adapter layer (`internal/adapter/docsignal/`) that takes `docscan.DocumentFile` output and the external analyzer's effect list, parses the annotation grammar, merges sidecar overrides, and produces `[]docsignal.DerivedSignal` — a keyed carrier type binding `(package, function, side_effect_type)` to a `taxonomy.Signal` (type-level grammar annotations fan out to per-function tuples via the effect list; `taxonomy.Signal` itself is unchanged).

4. Wire the doc-derived signals into `adapter.mergeClassifications` alongside the existing `classify_signals` protocol signals, so that doc signals flow through `classify.ComputeScore` with no protocol surgery. Derivation runs unconditionally; the `classify_signals` fetch remains capability-gated.

5. Scope strictly to external analyzers. Go-native projects already have `godoc` + native extractors; this change does not alter the godoc path.

## Capabilities

### New Capabilities

- `doc-annotation-grammar`: A deterministic Markdown annotation syntax that allows spec/design docs to declare side effects as `[contractual]` or `[incidental]`, keyed by the analyzer's own effect-type name. Parsed from docs discovered by `docscan.Scan`.

- `doc-signal-sidecar`: A machine-readable sidecar file (`.uf/gaze/contracts.yaml` or `.uf/gaze/contracts.json`) that maps `(package, function, side_effect_type)` to `contractual` or `incidental`, overriding or extending grammar-derived signals.

- `doc-signal-derivation`: A new derivation step that consumes `docscan.DocumentFile` output, extracts signals from both the annotation grammar and the sidecar file, and produces `[]docsignal.DerivedSignal` ready for `classify.ComputeScore`.

### Modified Capabilities

- `adapter.mergeClassifications`: Accepts an additional `[]docsignal.DerivedSignal` parameter for doc-derived signals, merging them with existing `classify_signals` protocol signals before computing scores. Doc signals carry a distinct `Source` value (`"architecture_doc"`/`"readme"`/`"sidecar"`) with the path in `SourceFile` for traceability.

- `ExternalSideEffectAnalyzer`: After `analyze`/`analyze_stream` and `classify_signals` (when available), calls the doc-signal derivation step and passes the result into `mergeClassifications`. The derivation runs unconditionally after the cache is populated; it is skipped only when the session `rootDir` is empty or `docscan.Scan` fails (graceful degradation, D5).

- `gaze report` and `gaze quality`: Benefit from resolved classifications without changes to their own logic — the doc signals flow through the existing classification pipeline.

### Removed Capabilities

- None.

## Impact

### Files Added

- `internal/adapter/docsignal/` — New package for doc signal derivation:
  - `parse.go` — Annotation grammar parser (Markdown → `[]taxonomy.Signal`)
  - `derive.go` — Orchestration: runs parser + sidecar merge, produces final signal set
  - `sidecar.go` — Sidecar file loader (`.uf/gaze/contracts.yaml`/`.json`)
  - `grammar.go` — Annotation grammar specification (constants, regex patterns)
  - `*_test.go` — Tests for parser, derivation, and sidecar loading

### Files Modified

- `internal/adapter/classify.go` — Extend `mergeClassifications` to accept optional `[]docsignal.DerivedSignal` doc signals
- `internal/adapter/sideeffect.go` — Add doc-signal derivation call (unconditional) after `classify_signals`
- `internal/protocol/testdata/fake_analyzer/main.go` — Add a `--no-classify-signals` flag (mirroring `--no-doc-coverage`) so the doc-signal path can be tested independently without breaking existing `classify_signals: true` assertions

### No Changes

- `internal/classify/godoc.go` — Go-native godoc path unchanged
- `internal/classify/score.go` — `ComputeScore` unchanged (already source-agnostic)
- `internal/taxonomy/types.go` — `Signal.Source` already anticipates `"architecture_doc"`
- `docs/protocol.md` — No protocol changes needed

## Constitution Alignment

Assessed against the Gaze project constitution (`.specify/memory/constitution.md` v1.3.0).

### I. Accuracy

**Assessment**: PASS

This change resolves a systematic accuracy gap where 56.7% of effects from external analyzers are classified as `ambiguous` despite design docs explicitly documenting their contractual or incidental nature. By deriving signals from documentation the tool already reads, classification accuracy improves without introducing new sources of error. The annotation grammar is deterministic and reproducible — no NLP ambiguity. False negatives in classification erode trust (Constitution Principle I) — this change drives the `ambiguous` fraction toward zero for well-documented projects.

### II. Minimal Assumptions

**Assessment**: PASS

The annotation grammar is opt-in — projects without annotated docs continue to work exactly as before, with effects remaining `ambiguous` or classified via other signals. The sidecar file is optional and layered on top of grammar annotations. No mandatory restructuring or annotation is required. The derivation step is scoped to external analyzers only; Go-native projects are unaffected. The capability degrades gracefully: if `docscan.Scan` fails or finds no docs, no doc-derived signals are produced and classification falls through to existing mechanisms.

### III. Actionable Output

**Assessment**: PASS

Doc-derived signals carry explicit `Source` values (`"architecture_doc"`, `"sidecar"`) and `Reasoning` strings (e.g., "design doc declares ContainerMutation as contractual") so users can trace why an effect was classified as contractual or incidental. Classification labels are machine-parseable (JSON output unchanged). Resolved classifications reduce false-positive gap counts, guiding users to real test gaps rather than phantom `ambiguous` effects.

### IV. Testability

**Assessment**: PASS

Each component in the new `internal/adapter/docsignal/` package is independently testable: the annotation grammar parser operates on plain Markdown strings, the sidecar loader operates on filesystem paths (testable with temp dirs), and the derivation orchestrator can be tested with synthetic `docscan.DocumentFile` inputs. The fake analyzer binary in `internal/protocol/testdata/fake_analyzer/` provides integration-level verification of the end-to-end doc-signal → classification flow. Coverage strategy: unit tests for all new exported functions (100% branch coverage target), integration test via fake analyzer.
