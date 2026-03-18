# Tier Boost Spec

## ADDED Requirements

### Requirement: P0 effects MUST receive a tier-based confidence boost

The classifier MUST add a tier-based confidence boost to the base score before summing mechanical signals. P0 effects (ReturnValue, ErrorReturn, SentinelError, ReceiverMutation, PointerArgMutation) MUST receive +25. P1 effects MUST receive +10. P2-P4 effects MUST receive +0.

#### Scenario: P0 ReturnValue on exported function with GoDoc
- **GIVEN** an exported function with GoDoc describing its return value
- **WHEN** the classifier scores a `ReturnValue` effect
- **THEN** the confidence is at least 75 + visibility + godoc (>= 100, clamped → contractual)

#### Scenario: P0 ErrorReturn on unexported function, no signals
- **GIVEN** an unexported function with no GoDoc, no callers, no naming match
- **WHEN** the classifier scores an `ErrorReturn` effect
- **THEN** the confidence is 75 (ambiguous, but one signal pushes to contractual)

#### Scenario: P1 WriterOutput on exported function
- **GIVEN** an exported function writing to an `io.Writer`
- **WHEN** the classifier scores a `WriterOutput` effect
- **THEN** the confidence starts at 60 (not 50), trending toward contractual with signals

#### Scenario: P2 goroutine spawn unchanged
- **GIVEN** any function spawning a goroutine
- **WHEN** the classifier scores the goroutine effect
- **THEN** the confidence starts at 50 (unchanged from current behavior)

### Requirement: Tier boost MUST be additive with existing signals

The tier boost MUST be applied before mechanical signal summation, not as a replacement. All existing signal analyzers (interface, visibility, callers, godoc, naming) MUST continue to contribute their weights on top of the boosted base.

#### Scenario: Signals stack with tier boost
- **GIVEN** a P0 `ReturnValue` with visibility signal (+20) and naming signal (+10)
- **WHEN** the classifier scores the effect
- **THEN** confidence = 50 (base) + 25 (tier) + 20 (visibility) + 10 (naming) = 105, clamped to 100

### Requirement: `accumulateSignals` MUST accept effect type

The `accumulateSignals` function signature MUST be expanded to accept `effectType taxonomy.SideEffectType` as its first parameter, used to compute the tier boost.

#### Scenario: Signature change
- **GIVEN** the `accumulateSignals` function
- **WHEN** the function is called
- **THEN** it accepts `(effectType taxonomy.SideEffectType, signals []taxonomy.Signal)` and returns `(score int, hasPositive, hasNegative bool)`

## MODIFIED Requirements

None — this is additive to the existing scoring pipeline.

## REMOVED Requirements

None.
