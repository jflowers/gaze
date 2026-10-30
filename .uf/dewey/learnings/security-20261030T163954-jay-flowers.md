---
tag: security
author: jay-flowers
category: gotcha
created_at: 2026-10-30T16:39:54Z
identity: security-20261030T163954-jay-flowers
tier: draft
---

Env-value redaction built on os.Environ()-style "NAME=VALUE" entries must pin the parse semantics in the spec: split at the FIRST '=' (strings.Cut / SplitN(entry, "=", 2)) and include a test case whose secret value itself contains '=' (JSON configs and base64 padding routinely do). A naive strings.Split(entry, "=")[1] silently truncates the redaction set, passes every ordinary test case, and under-redacts exactly the JSON-config secrets the control exists to catch.
