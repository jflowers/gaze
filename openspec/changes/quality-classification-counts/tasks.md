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

## 1. Add classification counts to taxonomy

- [x] 1.1 Add `ClassificationCounts` struct (`Contractual`/`Incidental`/`Ambiguous` integers with json tags `contractual`/`incidental`/`ambiguous`) to `internal/taxonomy/types.go`, and add `ClassificationCounts *ClassificationCounts \`json:"classification_counts,omitempty"\`` to `PackageSummary` (after `SkippedTestNames`), with a comment noting the field is nil for Go-native analysis.

## 2. Tally and populate classification counts in the adapter

- [x] 2.1 Add unexported `countClassifications(results []taxonomy.AnalysisResult) *taxonomy.ClassificationCounts` to `internal/adapter/quality.go` — iterates `results.SideEffects`, skips nil `Classification`, dedups by `e.ID` via a `seen` map, switches on `e.Classification.Label` (Contractual/Incidental/Ambiguous), returns nil when nothing was classified.
- [x] 2.2 In `BuildQualityFromMappings` (`internal/adapter/quality.go`), set `summary.ClassificationCounts = countClassifications(results)` after `buildQualitySummary(reports)`.

## 3. Test the classification tally

- [x] 3.1 Add `TestCountClassifications` to `internal/adapter/quality_internal_test.go` — subtests for dedup-by-ID (1/1/1), unclassified-ignored (1/0/0), and nil-when-nothing-classified (incl. nil results).
- [x] 3.2 Extend `TestBuildQualityFromMappings` in `internal/adapter/quality_internal_test.go` to assert `ClassificationCounts` is non-nil with `Contractual=2`/`Incidental=1`/`Ambiguous=0`.

## 4. Wire analyzer/language into the gaze-reporter and /gaze prompts

- [x] 4.1 [P] In `internal/scaffold/assets/agents/gaze-reporter.md`, add an "Analyzer & Language Detection" section (`.gaze.yaml` `analyzers:` map then project-marker heuristic), add `[--analyzer <cmd>] [--language <lang>]` to the crap/quality/docscan commands, and switch Full Mode classification sourcing to `classification_counts` for external projects.
- [x] 4.2 [P] In `internal/scaffold/assets/commands/gaze.md`, update the frontmatter description and add the analyzer/language detection note to the instructions.
- [x] 4.3 Update the dogfood copies `.opencode/agents/gaze-reporter.md` and `.opencode/commands/gaze.md` to stay byte-identical to their scaffold assets (embed.FS sync invariant #262/#263).
- [x] 4.4 Update the embedded copy `internal/aireport/assets/agents/gaze-reporter.md` to stay byte-identical to the scaffold asset (`//go:embed` as `defaultPromptRaw`; `TestEmbeddedPromptMatchesScaffold` invariant).
- [x] 4.5 Make `runCrap` and `runQuality` dispatch on `p.analyzerFlag != "" || p.languageFlag != ""` (matching `runDocscan`), and add a language-specific `initExternalSession` error message. Add `TestCrapWithLanguageOnly_TriggersExternalPath` to `cmd/gaze/external_analyzer_test.go`.

## 5. Mirror classification_counts into the quality schema

- [x] 5.1 In `internal/report/schema.go`, add a `classification_counts` property to `QualitySchema` `$defs.PackageSummary.properties` — an inline object (`type: object`) with integer properties `contractual`/`incidental`/`ambiguous` and a description noting external analyzers only. Do NOT add it to the `required` array.
- [x] 5.2 Add `TestQualitySchema_ValidatesClassificationCounts` to `internal/report/report_test.go` — validates quality JSON that includes `classification_counts` AND quality JSON that omits it.
- [x] 5.3 Assert in `internal/quality/quality_test.go` (`TestWriteJSON_Structure`) that Go-native JSON output omits `classification_counts`.
- [x] 5.4 Document `classification_counts` in the `PackageSummary` field table in `docs/reference/json-schemas.md`.

## 6. Verification

- [x] 6.1 Run `gofmt -l` on the changed Go files (expect no output).
- [x] 6.2 Run `go vet ./internal/report/... ./internal/adapter/... ./internal/taxonomy/...`.
- [x] 6.3 Run `go test -race -count=1 ./internal/report/... ./internal/adapter/... ./internal/taxonomy/...`.
- [x] 6.4 Verify constitution alignment: Accuracy (counts derive from real classify_signals labels, deduped), Minimal Assumptions (omitempty pointer field, best-effort layered detection, no new dependencies), Actionable Output (machine-parseable counts declared in the schema), Testability (pure helper + synthetic tests + schema validation test).
<!-- scaffolded by uf vdev -->
