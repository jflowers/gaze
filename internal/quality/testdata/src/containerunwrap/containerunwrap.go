// Package containerunwrap is a test fixture for exercising container
// unwrap assertion mapping patterns. It contains functions that return
// generic container/wrapper types wrapping JSON bodies, mirroring
// real-world patterns like MCP tool results and HTTP response wrappers.
//
// Tests in this package assign the return value, extract a field,
// pass it through a transformation (JSON unmarshal), and assert on
// the resulting values. The assertion mapping pipeline must trace
// these assertions back to the original ReturnValue side effect.
package containerunwrap

import "encoding/json"

// Container wraps a raw JSON body, representing a generic response
// wrapper type (e.g., an SDK result, HTTP response body, or message
// envelope).
type Container struct {
	Body string
}

// WrapJSON returns a Container whose Body field contains a JSON-encoded
// key-value pair. Tests unmarshal the Body and assert on the decoded
// values.
func WrapJSON(key, value string) *Container {
	data := map[string]string{key: value}
	b, _ := json.Marshal(data)
	return &Container{Body: string(b)}
}

// WrapMultiField returns a Container whose Body field contains
// multiple JSON key-value pairs. Tests unmarshal and assert on
// individual keys.
func WrapMultiField(fields map[string]string) *Container {
	b, _ := json.Marshal(fields)
	return &Container{Body: string(b)}
}

// WrapNestedJSON returns a Container whose Body field contains
// nested JSON (an object within an object). Tests unmarshal and
// assert on inner fields.
func WrapNestedJSON(key, innerKey, value string) *Container {
	inner := map[string]string{innerKey: value}
	outer := map[string]any{key: inner}
	b, _ := json.Marshal(outer)
	return &Container{Body: string(b)}
}

// TextContent holds a text payload, mirroring the MCP TextContent
// type used in tool result responses.
type TextContent struct {
	Text string
}

// Result holds a slice of TextContent entries, mirroring the MCP
// CallToolResult pattern where the result wraps content items.
type Result struct {
	Content []TextContent
}

// WrapMCPStyle returns a Result whose Content[0].Text contains a
// JSON-encoded key-value pair. This mirrors the real-world MCP test
// pattern: result.Content[0].Text -> type conversion -> unmarshal ->
// assert on map keys.
func WrapMCPStyle(key, value string) *Result {
	data := map[string]string{key: value}
	b, _ := json.Marshal(data)
	return &Result{
		Content: []TextContent{
			{Text: string(b)},
		},
	}
}
