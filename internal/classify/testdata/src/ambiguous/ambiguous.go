// Package ambiguous provides test fixtures with contradicting
// signals for classification testing.
package ambiguous

import "fmt"

// Processor has an exported method but no interface contract
// and no callers that depend on it — contradicting signals.
type Processor struct {
	count int
}

// Process is exported (visibility signal) but has no interface
// contract and no callers use it (contradicting). Should classify
// as ambiguous.
func (p *Processor) Process(data []byte) {
	p.count++
	fmt.Printf("processed %d items\n", p.count)
}

// ExportedButUnused is an exported function that nobody calls.
// It has a return value (visibility signal) but zero callers
// (contradicting signal).
func ExportedButUnused() string {
	return "unused"
}

// MixedSignals modifies the receiver (contractual evidence from
// naming "Update*" would suggest contractual) but also logs
// (incidental evidence). Neither signal is strong enough alone.
func (p *Processor) MixedSignals(value int) {
	p.count = value
	fmt.Printf("updated to %d\n", value)
}
