// Package incidental provides test fixtures with known incidental
// side effects for classification testing.
package incidental

import (
	"log"
)

// logger is a package-level logger for internal use.
var logger = log.New(log.Writer(), "incidental: ", log.LstdFlags)

// ProcessItem does work and logs progress. The log writes are
// incidental — no caller depends on the log output.
func ProcessItem(id string) error {
	logger.Printf("processing item %s", id)
	// do some work
	logger.Printf("item %s processed", id)
	return nil
}

// debugTrace writes debug output. The function name signals
// incidental behavior.
func debugTrace(msg string) {
	logger.Printf("TRACE: %s", msg)
}

// logError writes an error to the log. Logging is incidental.
func logError(err error) {
	logger.Printf("ERROR: %v", err)
}

// cache is a simple in-memory cache. Caching is an incidental
// optimization, not a contractual behavior. Unexported to reflect
// that no external caller depends on it.
type cache struct {
	items map[string]string
}

// set stores a value in the cache. This mutation is incidental
// because it is an internal optimization — no interface or
// caller contract depends on caching behavior.
func (c *cache) set(key, value string) {
	if c.items == nil {
		c.items = make(map[string]string)
	}
	c.items[key] = value
}

// InternalHelper is an unexported helper that no external caller
// depends on. Its side effects are incidental by nature.
func internalHelper(data []byte) []byte {
	logger.Printf("helper called with %d bytes", len(data))
	return data
}
