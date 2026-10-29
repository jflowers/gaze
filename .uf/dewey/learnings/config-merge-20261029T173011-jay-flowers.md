---
tag: config-merge
author: jay-flowers
category: pattern
created_at: 2026-10-29T17:30:11Z
identity: config-merge-20261029T173011-jay-flowers
tier: draft
---

Merging tool-managed entries into config files owned by another tool (gaze init → .uf/reviewer-capabilities.yaml / .uf/review-matrix.yaml): decode into gopkg.in/yaml.v3 yaml.Node and round-trip rather than struct-decode/re-encode or regex splicing. yaml.Node preserves comments (including the "# scaffolded by uf" scaffold marker) and unknown fields; struct decode silently drops comments; regex breaks on formatting variance. Encode with SetIndent(2) to match the existing file style, write only when encoded bytes differ from the bytes on disk (byte-stable idempotency — no-op runs never rewrite), treat the two config files independently, and warn-and-skip malformed input so optional intelligence never fails the host operation. Guard partial registration: when any expected key (e.g. advisor.feedback) is missing, skip the whole file — in-memory yaml.Node mutations to earlier keys are safely discarded because the write never happens.
