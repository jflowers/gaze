---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-10-30T20:12:35Z
identity: test-quality-20261030T201235-jay-flowers
tier: draft
---

To integration-test code whose behavior derives from the process environment (e.g. secret scrubbing keyed off os.Environ()), plant an obviously fake value with t.Setenv in the parent test process and have the spawned child echo it: the child inherits the var through the real exec path (the wrapper never sets cmd.Env) while the code under test reads the same var from os.Environ(), exercising env inheritance and the post-Wait() snapshot read end-to-end with no mocks. Use unique GAZE_TEST_*-style names and random-looking fake values so the developer's real environment cannot collide with assertions, and assert both directions — the redaction marker appears AND surrounding diagnostic text survives — so the test pins surgical replacement rather than wholesale nuking. Pattern from gaze's TestRunSubprocess_NonZeroExitRedactsSensitiveEnvValue / _EmptyStdoutSurfacesRedactedStderr (internal/aireport/subprocess_test.go).
