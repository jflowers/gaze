## 1. Data Model

- [x] 1.1 Add `AIMapperContext` struct to `internal/quality/ai_mapper.go` with fields: `AssertionSource`, `AssertionKind`, `TestFuncSource`, `TargetFunc`, `SideEffects`.
- [x] 1.2 Add `AIMapperFunc func(AIMapperContext) (string, error)` type definition.
- [x] 1.3 Add `AIMapperFunc AIMapperFunc` field to `quality.Options`.

## 2. Pipeline Integration

- [x] 2.1 Add `aiMapper AIMapperFunc` parameter to `MapAssertionsToEffects` signature (variadic for backward compat).
- [x] 2.2 Add `tryAIMapping` function that constructs `AIMapperContext` from the assertion site, target function, and effects, calls the AI mapper, and returns a mapping at confidence 50 or nil.
- [x] 2.3 Insert AI fallback call after all mechanical passes fail and before unmapped classification in `MapAssertionsToEffects`.
- [x] 2.4 Update `Assess` to pass `opts.AIMapperFunc` through to `MapAssertionsToEffects`.
- [x] 2.5 All existing callers use variadic (no arg = no AI mapper) — backward compatible.

## 3. Source Extraction

- [x] 3.1 Add `extractExprSource(expr ast.Expr, fset *token.FileSet) string` helper.
- [x] 3.2 Add `extractFuncSource(decl *ast.FuncDecl, fset *token.FileSet) string` helper.

## 4. Tests

- [x] 4.1 `TestAIMapper_Match` — mock AI mapper returns valid effect ID, assertion mapped at confidence 50.
- [x] 4.2 `TestAIMapper_NoMatch` — mock returns empty string, assertion remains unmapped.
- [x] 4.3 `TestAIMapper_Error` — mock returns error, assertion remains unmapped, no panic.
- [x] 4.4 `TestAIMapper_Nil` — nil mapper produces same results as no-match mapper.
- [x] 4.5 `TestAIMapper_ContextPopulated` — verifies all AIMapperContext fields populated for indirectmatch fixture.

## 5. Prompt & Response Helpers

- [x] 5.1 Add `BuildAIMapperPrompt(ctx AIMapperContext) string` — structured prompt for any AI backend.
- [x] 5.2 Add `ParseAIMapperResponse(response string, validIDs map[string]bool) string` — extract effect ID from AI response.
- [x] 5.3 Add `TestBuildAIMapperPrompt` and `TestParseAIMapperResponse` tests.

## 6. Agent-Level Mapping (gaze-reporter prompt)

- [x] 6.1 Add "Unmapped Assertion Evaluation" section to the gaze-reporter agent prompt. When quality JSON contains unmapped assertions, the agent should read source files, evaluate semantic relationships, and report AI-mapped assertions with confidence indicators.
- [x] 6.2 Sync all 3 copies of the gaze-reporter prompt (scaffold, .opencode, aireport).
- [x] 6.3 Run scaffold drift tests to verify sync.

## 7. Documentation & Verification

- [x] 7.1 GoDoc on `AIMapperContext`, `AIMapperFunc`, `tryAIMapping`, `extractExprSource`, `extractFuncSource`, `BuildAIMapperPrompt`, `ParseAIMapperResponse`.
- [x] 7.2 Update `AGENTS.md` Recent Changes.
- [x] 7.3 `go build ./...` and `go vet ./...` — clean.
- [x] 7.4 Affected package tests pass (quality, scaffold, aireport).
