## Why

The assertion-to-effect mapping pipeline achieves 78.8% accuracy (52/66 mapped). The largest remaining gap is helper parameter tracing: when a test calls `assertEqual(t, got, 12)`, the assertion inside `assertEqual` references the helper's parameter `got` which is a different `types.Object` than the test's `got` variable. Without bridging these, all assertions inside helper functions are unmapped.

## What Changes

1. Add `CallerArgs []ast.Expr` to `AssertionSite` to carry the call-site arguments for helper assertions.
2. Add `buildHelperBridge` function that maps helper parameter `types.Object` to the caller's argument `types.Object`.
3. Extend `matchAssertionToEffect` Pass 1 to check the helper bridge when direct matching fails.
4. Update ratchet floor from 76.0% to 83.0%.

## Impact

Mapping accuracy: 78.8% (52/66) → 84.8% (56/66). Four additional assertions mapped through helper parameter bridging at confidence 70.

## Constitution Alignment

All PASS. Improves Accuracy (Principle I) by reducing false negatives in assertion mapping. Testability (Principle IV) verified via ratchet test.
