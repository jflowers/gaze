---
tag: schema-sentinel-pattern
author: jay-flowers
created_at: 2026-09-28T22:32:15Z
identity: schema-sentinel-pattern-20260928T223215-jay-flowers
tier: draft
---

Using an out-of-band numeric sentinel like -1 for "no contract expected" on ContractCoverage.Percentage is a bug: the embedded JSON Schema declares percentage with minimum 0 / maximum 100, so -1 serializes as schema-invalid JSON and breaks the project's JSON-schema-validation convention. The correct representation is an additive boolean NoContractExpected (json tag no_contract_expected,omitempty) plus a Reason string (reason,omitempty), keeping Percentage at a legal 0. The omitempty tags ensure the fields are absent (not false/"") for production functions. This preserves schema-validity while still giving machine-readable signals.
