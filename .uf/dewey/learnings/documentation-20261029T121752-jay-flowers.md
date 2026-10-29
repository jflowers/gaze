---
tag: documentation
author: jay-flowers
category: pattern
created_at: 2026-10-29T12:17:52Z
identity: documentation-20261029T121752-jay-flowers
tier: draft
---

When a spec plans a file rename or scaffold-set change, derive the documentation reference sweep from a full-text grep of the entire user-facing docs tree for the old filename — not from the obviously-affected tables and lists. Two recurring blind spots in the gaze repo: (1) sample-output/console-transcript blocks embedded in guides (docs/guides/opencode-integration.md prints a `gaze init` transcript naming every scaffolded file), and (2) the CLI reference page for any command whose behavior changes (docs/reference/cli/init.md documents init's scaffolded-files tree, ownership model, and example output, and is linked from the README command table). Sweeps that enumerate only file tables and ownership lists leave transcripts and reference pages stale after implementation.
