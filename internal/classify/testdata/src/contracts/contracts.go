// Package contracts provides test fixtures with known contractual
// side effects for classification testing.
package contracts

import (
	"fmt"
	"io"
)

// Reader is an interface that defines a contractual Read method.
type Reader interface {
	// Read reads data into p and returns the number of bytes read.
	Read(p []byte) (n int, err error)
}

// Writer is an interface that defines a contractual Write method.
type Writer interface {
	// Write writes data from p and returns the number of bytes written.
	Write(p []byte) (n int, err error)
}

// Store is an interface for persistence operations.
type Store interface {
	// Save persists the given data and returns an error if it fails.
	Save(data []byte) error
	// Delete removes the item with the given ID.
	Delete(id string) error
}

// FileStore implements Store and Writer.
type FileStore struct {
	Path string
	data []byte
}

// Save persists data to the file store. This is a contractual side
// effect because FileStore implements Store.
func (fs *FileStore) Save(data []byte) error {
	fs.data = data
	return nil
}

// Delete removes the file. This is a contractual side effect
// because FileStore implements Store.
func (fs *FileStore) Delete(id string) error {
	fs.data = nil
	return nil
}

// Write implements io.Writer. This is a contractual side effect
// because FileStore implements Writer.
func (fs *FileStore) Write(p []byte) (n int, err error) {
	fs.data = append(fs.data, p...)
	return len(p), nil
}

// GetData returns the stored data. The return value is contractual
// because the function name follows the Get* convention and callers
// depend on it.
func GetData(fs *FileStore) []byte {
	return fs.data
}

// FetchConfig loads configuration. The return value is contractual
// because the function name follows the Fetch* convention.
func FetchConfig(path string) (map[string]string, error) {
	return map[string]string{"path": path}, nil
}

// SaveRecord writes a record. The mutation is contractual because
// the function name follows the Save* convention.
func SaveRecord(w io.Writer, record string) error {
	_, err := fmt.Fprintln(w, record)
	return err
}

// ErrNotFound is a sentinel error. Its existence as a named error
// is a contractual signal.
var ErrNotFound = fmt.Errorf("not found")

// ErrPermission is a sentinel error for permission failures.
var ErrPermission = fmt.Errorf("permission denied")
