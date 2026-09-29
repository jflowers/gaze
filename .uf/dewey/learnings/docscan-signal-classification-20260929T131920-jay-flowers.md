---
tag: docscan-signal-classification
author: jay-flowers
category: pattern
created_at: 2026-09-29T13:19:20Z
identity: docscan-signal-classification-20260929T131920-jay-flowers
tier: draft
---

When deriving classification signals from documentation that must be keyed to a (package, function, side_effect_type) tuple, the existing taxonomy.Signal type cannot carry that key because it only has Source, Weight, SourceFile, Excerpt, and Reasoning fields. The correct approach is to introduce a keyed carrier type in the deriving package (e.g. docsignal.DerivedSignal with Package, Function, SideEffectType fields plus an embedded taxonomy.Signal) rather than modifying taxonomy.Signal or reusing the protocol.ClassifySignalData type. This preserves the taxonomy type's stability while giving the merge step the tuple key it needs for grouping. This was a CRITICAL spec-review finding in the docscan-signal-classification change (#285) that blocked implementation until resolved.
