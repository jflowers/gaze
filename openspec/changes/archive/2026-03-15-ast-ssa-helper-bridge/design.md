## Context

Issue #6 requests bridging the AST-to-SSA name gap to reach 90% mapping accuracy. The timeboxed exploration identified helper parameter tracing as the largest single gap (~8-10 unmapped assertions). The fix is an AST-level parameter-to-argument bridge, not SSA-level.

## Approach

When `detectHelperAssertions` recurses into a helper function, it now attaches `CallerArgs` (the call-site argument expressions) to each resulting `AssertionSite`. In `matchAssertionToEffect`, when an identifier matches a helper parameter, `buildHelperBridge` resolves it back to the caller's argument `types.Object` and checks `objToEffectID`.

## Remaining gap to 90%

After this change, 10/66 assertions remain unmapped. The remaining gaps are:
- Inline call returns (`if f() != x`) — ~3-4
- Complex helper patterns (multi-level, variadic) — ~3
- Cross-target assertions — ~3

Issue #6 remains open for these remaining gaps.
