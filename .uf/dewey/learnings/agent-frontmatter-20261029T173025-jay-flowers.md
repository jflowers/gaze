---
tag: agent-frontmatter
author: jay-flowers
category: gotcha
created_at: 2026-10-29T17:30:25Z
identity: agent-frontmatter-20261029T173025-jay-flowers
tier: draft
---

Migrating an agent persona from the legacy `tools:` frontmatter dialect (write/edit/bash: false) to the divisor `permission:` dialect changes the restricted set: the canonical divisor deny block is edit/bash/webfetch, and the `write` key is not carried over. Spec scenarios that enumerate the legacy restricted set ("MUST deny write/edit/bash") read as unimplemented against the migrated frontmatter — enumerate the actual divisor deny set in the spec scenario, or the archived requirement becomes unverifiable as written (observed in openspec/changes/divisor-gaze-testing: the spec scenario said write/edit/bash while design D3, task 1.2, and the shipped file implement edit/bash/webfetch).
