# Side Effects

A **side effect** is any observable change that a function produces beyond its internal computation. When you call a function, the side effects are everything a caller can detect: the values it returns, the fields it mutates, the files it writes, the channels it sends on.

Gaze detects side effects through static analysis of Go source code. Every detected side effect is assigned a [type](../reference/glossary.md#side-effect), a [tier](../reference/glossary.md#tier) (P0 through P4), and a stable ID for tracking across runs. Understanding the side effect taxonomy is the foundation for everything else Gaze does — [classification](classification.md), [scoring](scoring.md), and [quality assessment](quality.md) all build on this layer.

## Why Side Effects Matter

Traditional code coverage answers "was this line executed?" but says nothing about whether the test *verified* anything meaningful. A test that calls a function and ignores its return value achieves 100% line coverage but 0% [contract coverage](../reference/glossary.md#contract-coverage).

Side effects are the bridge between "code was executed" and "behavior was verified." By enumerating every observable change a function can produce, Gaze can measure whether your tests actually assert on the things that matter.

## The Taxonomy: 37 Effect Types Across 5 Tiers

Gaze defines 37 side effect types organized into five priority tiers. The tier determines how critical the effect is to detect and how it influences [classification scoring](classification.md).

### P0 — Must Detect

P0 effects are a function's direct observable outputs. These are definitionally [contractual](../reference/glossary.md#contractual) — they are the reason callers invoke the function. Gaze targets zero false negatives and zero false positives for P0 effects.

| Effect Type | Description | Detection |
|---|---|---|
| `ReturnValue` | A non-error value returned to the caller | Implemented (AST) |
| `ErrorReturn` | An error-typed value returned to the caller | Implemented (AST) |
| `SentinelError` | A package-level `var Err* = errors.New(...)` sentinel | Implemented (AST) |
| `ReceiverMutation` | Mutation of a pointer receiver's fields (e.g., `s.count++`) | Implemented (SSA, AST fallback) |
| `PointerArgMutation` | Mutation through a pointer parameter (e.g., `*out = value`) | Implemented (SSA, AST fallback) |

P0 effects are detected using a combination of AST analysis (for returns and sentinels) and SSA analysis (for mutations). When SSA construction fails, Gaze falls back to AST-based mutation detection with lower fidelity. See [Analysis Pipeline](analysis-pipeline.md) for details.

### P1 — High Value

P1 effects are significant observable changes that go beyond direct return values. They modify shared state, write to I/O interfaces, or communicate through channels.

| Effect Type | Description | Detection |
|---|---|---|
| `SliceMutation` | Direct index assignment on a slice parameter (e.g., `s[i] = v`) | Implemented (AST) |
| `MapMutation` | Map index assignment on a map parameter (e.g., `m[key] = v`) | Implemented (AST) |
| `GlobalMutation` | Assignment to a package-level variable | Implemented (AST) |
| `WriterOutput` | Calls to `io.Writer.Write` or `fmt.Fprint*` with a writer parameter | Implemented (AST) |
| `HTTPResponseWrite` | Calls to `http.ResponseWriter` methods (`Write`, `WriteHeader`, `Header`) | Implemented (AST) |
| `ChannelSend` | Send statement (`ch <- value`) | Implemented (AST) |
| `ChannelClose` | Call to `close(ch)` | Implemented (AST) |
| `DeferredReturnMutation` | Named return variable modified inside a `defer` statement | Implemented (AST) |

### P2 — Important

P2 effects represent interactions with external systems (filesystem, database), concurrency primitives, and control flow changes. Detection uses AST pattern matching against known standard library APIs.

| Effect Type | Description | Detection |
|---|---|---|
| `FileSystemWrite` | File creation or write operations (`os.WriteFile`, `os.Create`, `os.Mkdir`, etc.) | Implemented (AST) |
| `FileSystemDelete` | File or directory removal (`os.Remove`, `os.RemoveAll`) | Implemented (AST) |
| `FileSystemMeta` | File metadata changes (`os.Chmod`, `os.Chown`, `os.Symlink`, etc.) | Implemented (AST) |
| `DatabaseWrite` | Database write operations (`db.Exec`, `db.ExecContext` on `*sql.DB`/`*sql.Tx`/`*sql.Stmt`) | Implemented (AST) |
| `DatabaseTransaction` | Database transaction initiation (`db.Begin`, `db.BeginTx` on `*sql.DB`) | Implemented (AST) |
| `GoroutineSpawn` | Goroutine creation via `go` statement | Implemented (AST) |
| `Panic` | Call to the builtin `panic()` function | Implemented (AST) |
| `CallbackInvocation` | Invocation of a function-typed parameter | Implemented (AST) |
| `LogWrite` | Logging calls (`log.Print*`, `log.Fatal*`, `slog.Debug/Info/Warn/Error`) | Implemented (AST) |
| `ContextCancellation` | Context cancellation setup (`context.WithCancel`, `WithTimeout`, `WithDeadline`) | Implemented (AST) |

### P3 — Nice to Have

P3 effects cover standard I/O, environment manipulation, synchronization primitives, and other observable behaviors. These types are defined in the taxonomy but detection is not yet implemented.

| Effect Type | Description | Detection |
|---|---|---|
| `StdoutWrite` | Writing to standard output | Defined — detection not yet implemented |
| `StderrWrite` | Writing to standard error | Defined — detection not yet implemented |
| `EnvVarMutation` | Modification of environment variables | Defined — detection not yet implemented |
| `MutexOp` | Mutex lock/unlock operations | Defined — detection not yet implemented |
| `WaitGroupOp` | WaitGroup Add/Done/Wait operations | Defined — detection not yet implemented |
| `AtomicOp` | Atomic load/store/swap operations | Defined — detection not yet implemented |
| `TimeDependency` | Dependency on current time (`time.Now()`, `time.Since()`) | Defined — detection not yet implemented |
| `ProcessExit` | Process termination (`os.Exit()`) | Defined — detection not yet implemented |
| `RecoverBehavior` | Use of `recover()` to handle panics | Defined — detection not yet implemented |

### P4 — Exotic

P4 effects involve reflection, unsafe operations, and other advanced Go features. These are the lowest priority for detection and are defined for taxonomy completeness.

| Effect Type | Description | Detection |
|---|---|---|
| `ReflectionMutation` | Mutation via the `reflect` package | Defined — detection not yet implemented |
| `UnsafeMutation` | Mutation via the `unsafe` package | Defined — detection not yet implemented |
| `CgoCall` | Calls to C code via cgo | Defined — detection not yet implemented |
| `FinalizerRegistration` | Registration of finalizers via `runtime.SetFinalizer` | Defined — detection not yet implemented |
| `SyncPoolOp` | Operations on `sync.Pool` (Get/Put) | Defined — detection not yet implemented |
| `ClosureCaptureMutation` | Mutation of variables captured by a closure | Defined — detection not yet implemented |

## Stable IDs

Every detected side effect receives a stable, deterministic ID generated from a SHA-256 hash of the package path, function name, effect type, and source location. IDs are formatted as `se-` followed by 8 hex characters (e.g., `se-a1b2c3d4`). This enables diffing side effects across runs — you can track when effects appear, disappear, or change classification over time.

## What's Next

- [Classification](classification.md) — how Gaze determines whether each side effect is contractual, ambiguous, or incidental
- [Scoring](scoring.md) — how side effects feed into CRAP and GazeCRAP scores
- [Analysis Pipeline](analysis-pipeline.md) — how AST and SSA analysis work together to detect effects
