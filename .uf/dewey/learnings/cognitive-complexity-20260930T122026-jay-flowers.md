---
tag: cognitive-complexity
author: jay-flowers
category: pattern
created_at: 2026-09-30T12:20:26Z
identity: cognitive-complexity-20260930T122026-jay-flowers
tier: draft
---

When implementing Sonar cognitive complexity for Go AST, the else clause DOES increment complexity (+1, no nesting penalty). This differs from what some spec documents claim. The correct values are: if=+1+nesting, else if=+1, else=+1. Logical operator sequences count as +1 per distinct consecutive group of the same operator: a && b && c = 1, a && b || c = 2. The key implementation pattern is recursive descent through AST statements with explicit nesting level tracking, rather than using ast.Walk which doesn't provide fine-grained nesting control.
