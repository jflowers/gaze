## Context

External-analyzer projects (Python via snake-eyes) report 56.7% of detected effects as `ambiguous` despite having design docs that explicitly document which effects are contractual and which are incidental. `gaze docscan` already discovers and reads those docs, but its output goes only to the AI report pipeline — it never reaches classification. The scoring engine (`classify.ComputeScore`) is already source-agnostic: it sums every `Signal.Weight` regardless of `Source`. `taxonomy.Signal.Source` already anticipates doc-derived sources (`"readme"`, `"architecture_doc"`), but no extractor produces them for the non-Godoc path.

This change adds a deterministic doc → signal derivation step between `docscan.Scan` output and `classify.ComputeScore`, scoped to external analyzers only.

## Goals / Non-Goals

### Goals

1. Derive classification signals from spec/design docs already discovered by `docscan.Scan`, using a deterministic annotation grammar embedded in existing Markdown docs.
2. Provide a machine-readable sidecar override mechanism (`.uf/gaze/contracts.yaml`/`.json`) for explicit per-tuple declarations.
3. Wire doc-derived signals into `adapter.mergeClassifications` so they flow through `classify.ComputeScore` without protocol surgery.
4. Scope strictly to external analyzers — Go-native projects continue using the existing `godoc` path unchanged.

### Non-Goals

- NLP or LLM-based extraction of intent from prose. The annotation grammar is deterministic and reproducible.
- Modifying `classify.ComputeScore` or `taxonomy.Signal`. Both are already source-agnostic.
- Extending the JSON-RPC protocol. Doc signals are derived internally; analyzers don't need to know about them.
- Changing `docscan.Scan` or `docscan.Filter`. They already produce the right output.
- Modifying Go-native classification (`internal/classify/godoc.go`). This is purely for the external-analyzer path.
- Requiring projects to annotate their docs. The grammar is opt-in; unannotated docs produce no signals (graceful no-op).

## Decisions

### D1: New package `internal/adapter/docsignal/`

**Decision**: Create a new internal package `internal/adapter/docsignal/` for doc signal derivation. It lives under `adapter/` because doc signals are only relevant to the external-analyzer path. The package defines a keyed carrier type `DerivedSignal{Package, Function, SideEffectType string; Signal taxonomy.Signal}` that binds classification signals to the `(package, function, side_effect_type)` tuples they apply to, without modifying `taxonomy.Signal`.

**Rationale**: A sub-package keeps the grammar parser and sidecar loader isolated from the adapter's core orchestration logic, mirroring the existing `internal/docscan/apidoc/` sub-package precedent. The package is `internal` (not exported) and has no external dependencies beyond the standard library, `gopkg.in/yaml.v3` (already in `go.mod`), `internal/taxonomy/`, and `internal/docscan/` (for the `DocumentFile` type consumed by `ParseDocument`/`DeriveSignals`). `DeriveSignals` also accepts `[]taxonomy.AnalysisResult` (the analyzed effect list) to fan out type-level grammar annotations into per-function tuples. No import cycle is introduced (adapter → docsignal → docscan → config).

**Alternatives considered**:
- Placing derivation in `internal/classify/` — rejected because classify is language-agnostic scoring; doc-signal derivation is adapter-specific.
- Placing derivation directly in `internal/adapter/classify.go` — rejected because the grammar parser and sidecar loader add enough new surface to warrant their own package.

### D2: Annotation grammar syntax

**Decision**: Support two syntax forms: HTML comments (`<!-- gaze:contractual TypeName -->`) and fenced block callouts (`> [!gaze-contract]`). Both are invisible in rendered Markdown (comments) or render as GitHub-flavored callouts. The extractor keys on the effect type name verbatim and the `contractual`/`incidental` label token.

**Rationale**:
- HTML comments are invisible in all Markdown renderers — docs remain clean.
- Fenced block callouts render as styled blocks in GitHub/GitLab, making the contract declarations visible and reviewable in PRs.
- Both forms use analyzer-emitted type names verbatim (user decision) — no translation layer.
- The grammar is trivially parseable with regex, avoiding a full Markdown AST dependency.

**Signal weight convention** (filename-based, not priority-based): non-README documents carry ±25 (`architecture_doc`); documents whose basename is `README*` (case-insensitive) carry ±15 (`readme`); sidecar entries carry ±30 (`sidecar`). `Signal.Source` holds the bare token (`"architecture_doc"`/`"readme"`/`"sidecar"`); the path goes in `Signal.SourceFile`. These weights follow the existing `maxGodocWeight=15` convention and reflect the relative authority of each source. Grammar annotations are type-level wildcards that fan out to every `(package, function)` tuple in the analyzed effect list containing the annotated type; function-level precision is sidecar-only.

### D3: Sidecar file location and format

**Decision**: Sidecar files live at `.uf/gaze/contracts.yaml` (primary) or `.uf/gaze/contracts.json` (alternative). The `.uf/gaze/` directory is project-specific, keeping gaze configuration files together and out of the repo root.

**Rationale**:
- `.uf/gaze/` is a new namespaced directory (distinct from the convention packs at `.opencode/uf/packs/`) that co-locates gaze-specific project files with the existing `.gaze/` artifacts.
- YAML as primary format matches `.gaze.yaml` conventions; JSON as alternative for tool-generated sidecars.
- Sidecar overrides grammar (higher authority), reflecting the explicit, intentional nature of sidecar declarations.
- Each sidecar entry is semantically validated: unknown label, empty package/function, or an unknown `side_effect_type` skips the entry with a stderr warning.

### D4: Adapter integration point

**Decision**: Doc-signal derivation runs from `ExternalSideEffectAnalyzer.classifyAndMerge` **unconditionally** after `analyze`/`analyze_stream` populates the cache, regardless of whether the analyzer advertises `classify_signals`. Protocol `classify_signals` is fetched only when the capability is present. The doc-derived signals are passed as an additional `[]docsignal.DerivedSignal` parameter to `mergeClassifications`, which combines them with protocol-derived signals before calling `classify.ComputeScore`.

**Rationale**:
- The adapter already orchestrates the classification pipeline (`analyze` → `classify_signals` → `mergeClassifications`). Adding a doc-signal step after protocol signals is a natural extension.
- The current `classifyAndMerge` early-returns when `!caps.ClassifySignals`; this gate is restructured so doc-signal derivation runs regardless, and only the `classify_signals` fetch is capability-gated.
- `mergeClassifications` already groups signals by `(Package, Function, SideEffectType)`. Adding doc signals (keyed by the same tuple via `DerivedSignal`) to the same grouping logic requires minimal change.
- `classify.ComputeScore` naturally handles multiple signals for the same tuple via weight summation — no conflict resolution needed.

### D5: Graceful degradation

**Decision**: Every step in the derivation pipeline degrades gracefully:
- `docscan.Scan` failure → empty doc list, no doc signals (no error).
- No annotated docs → empty signal list (no error).
- Malformed grammar annotation → skip that annotation, continue (no error).
- Missing sidecar file → proceed with grammar signals only (no error).
- Invalid sidecar file → warn to stderr, proceed with grammar signals only (no error).

**Rationale**: Doc signals are a quality improvement, not a correctness requirement. Classification works without them (just with more `ambiguous` results). A failing derivation step should never block analysis.

### D6: Module root discovery for docscan

**Decision**: The project root for `docscan.Scan` is the session's `rootDir`, which is **already threaded** through `Session` to `NewExternalSideEffectAnalyzer` (for external analyzers this is the user's project directory, i.e. `cwd` as set by the CLI for `--analyzer` invocations). No new root-discovery or `Session` change is required. `docscan.Scan` runs once per session with `PackageDir: ""` — the external-analyzer path has no single package directory, so `PrioritySamePackage` is never produced; priorities are `PriorityModuleRoot` (root-level docs) and `PriorityOther` (nested docs).

**Rationale**: `docscan.Scan` requires a filesystem root to walk, and for external analyzers that is the user's project directory. The session already has this path. Because the weight convention is filename-based (D2), the absence of a `PackageDir` only affects conflict precedence (`PriorityModuleRoot` > `PriorityOther`), which remains deterministic.

## Risks / Trade-offs

### Risk 1: Annotation grammar fragmentation

**Risk**: Different projects may invent their own annotation conventions if the grammar is not well-documented.

**Mitigation**: The grammar is defined by this spec and documented in `docs/guides/improving-scores.md` and a new `docs/guides/doc-annotations.md`. The sidecar file provides an escape hatch for projects that need more flexibility.

### Risk 2: Module root not always discoverable

**Risk**: In some invocation paths (e.g., `gaze quality --analyzer ./snake-eyes mypackage`), the module root may not be trivially derivable from the package pattern.

**Mitigation**: `rootDir` is already threaded through the session as the user's project directory (the CLI passes `cwd` for `--analyzer`). No `go.mod` discovery is attempted for external analyzers. If `rootDir` is empty or `docscan.Scan` fails, doc-signal derivation is skipped (graceful degradation, D5).

### Risk 3: Sidecar file proliferation

**Risk**: `.uf/gaze/` is a new directory; projects may accumulate multiple gaze-specific files there over time.

**Mitigation**: The directory is a new namespaced location distinct from the convention packs (`.opencode/uf/packs/`). A single `contracts.yaml` covers all packages in the project. Future gaze-specific files can use the same directory.

### Risk 4: Signal weight calibration

**Risk**: The chosen weights (+15/+25/+30) may not produce the right classification outcomes for all projects.

**Mitigation**: Weights follow existing conventions (`maxGodocWeight=15` for godoc). The `classify.ComputeScore` threshold is configurable via `.gaze.yaml`. Sidecar entries can carry custom reasoning for auditability. Weight tuning can be adjusted in a future change without breaking the grammar or protocol.
