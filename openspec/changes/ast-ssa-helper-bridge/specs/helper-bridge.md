## ADDED Requirements

### Requirement: CallerArgs on AssertionSite
AssertionSite MUST include CallerArgs populated with the call-site argument expressions when Depth > 0.

### Requirement: Helper parameter bridge in matching
matchAssertionToEffect MUST check helper parameter bridge when direct matching fails, mapping at confidence 70.

## MODIFIED Requirements

### Requirement: Mapping accuracy ratchet
Baseline floor MUST be raised from 76.0% to 83.0%.
