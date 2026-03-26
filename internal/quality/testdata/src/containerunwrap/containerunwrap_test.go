package containerunwrap

import (
	"encoding/json"
	"testing"
)

// TestWrapJSON_BasicUnmarshal exercises the core container-unwrap-assert
// pattern: assign return value, access .Body field, unmarshal into a
// map, assert on map keys.
func TestWrapJSON_BasicUnmarshal(t *testing.T) {
	container := WrapJSON("status", "ok")

	// Direct nil check — maps via existing direct pass.
	if container == nil {
		t.Fatal("expected non-nil container")
	}

	// Direct field assertion — maps via existing indirect pass.
	if container.Body == "" {
		t.Fatal("expected non-empty Body")
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(container.Body), &data); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Container unwrap assertion — maps via container unwrap pass.
	if data["status"] != "ok" {
		t.Errorf("data[\"status\"] = %v, want \"ok\"", data["status"])
	}
}

// TestWrapJSON_StructUnmarshal exercises the container-unwrap-assert
// pattern with unmarshal into a typed struct (FR-006).
func TestWrapJSON_StructUnmarshal(t *testing.T) {
	container := WrapJSON("name", "gaze")

	// Direct nil check — maps via existing direct pass.
	if container == nil {
		t.Fatal("expected non-nil container")
	}

	// Direct field assertion — maps via existing indirect pass.
	if container.Body == "" {
		t.Fatal("expected non-empty Body")
	}

	type payload struct {
		Name string `json:"name"`
	}
	var p payload
	if err := json.Unmarshal([]byte(container.Body), &p); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Container unwrap assertion — maps via container unwrap pass.
	if p.Name != "gaze" {
		t.Errorf("p.Name = %q, want \"gaze\"", p.Name)
	}
}

// TestWrapMultiField_MultipleAssertions exercises multiple assertions
// on different map keys from a single unmarshal operation.
func TestWrapMultiField_MultipleAssertions(t *testing.T) {
	container := WrapMultiField(map[string]string{
		"a": "alpha",
		"b": "bravo",
		"c": "charlie",
		"d": "delta",
		"e": "echo",
		"f": "foxtrot",
	})

	// Direct nil check — maps via existing direct pass.
	if container == nil {
		t.Fatal("expected non-nil container")
	}

	// Direct field assertion — maps via existing indirect pass.
	if container.Body == "" {
		t.Fatal("expected non-empty Body")
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(container.Body), &data); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Container unwrap assertions — all map via container unwrap pass.
	if len(data) != 6 {
		t.Errorf("len(data) = %d, want 6", len(data))
	}
	if data["a"] != "alpha" {
		t.Errorf("data[\"a\"] = %v, want \"alpha\"", data["a"])
	}
	if data["b"] != "bravo" {
		t.Errorf("data[\"b\"] = %v, want \"bravo\"", data["b"])
	}
	if data["c"] != "charlie" {
		t.Errorf("data[\"c\"] = %v, want \"charlie\"", data["c"])
	}
	if data["d"] != "delta" {
		t.Errorf("data[\"d\"] = %v, want \"delta\"", data["d"])
	}
	if data["e"] != "echo" {
		t.Errorf("data[\"e\"] = %v, want \"echo\"", data["e"])
	}
	if data["f"] != "foxtrot" {
		t.Errorf("data[\"f\"] = %v, want \"foxtrot\"", data["f"])
	}
}

// TestWrapMCPStyle_DeepChain exercises the full MCP test pattern with
// 4+ intermediate steps: result.Content[0].Text -> []byte() ->
// json.Unmarshal -> data["key"] (FR-004).
func TestWrapMCPStyle_DeepChain(t *testing.T) {
	result := WrapMCPStyle("tool", "hammer")

	// Direct nil check — maps via existing direct pass.
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Direct field length check — maps via existing indirect pass.
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty Content")
	}

	text := result.Content[0].Text
	var data map[string]any
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Container unwrap assertions — map via container unwrap pass.
	if len(data) != 1 {
		t.Errorf("len(data) = %d, want 1", len(data))
	}
	if data["tool"] != "hammer" {
		t.Errorf("data[\"tool\"] = %v, want \"hammer\"", data["tool"])
	}
}

// TestWrapWithInterface_TypeAssertChain exercises the type assertion
// container unwrap pattern: result.Content[0].(TextContent).Text ->
// []byte() -> json.Unmarshal -> data["key"]. This mirrors the real
// MCP SDK pattern where Content is an interface slice and tests must
// type-assert before accessing fields.
func TestWrapWithInterface_TypeAssertChain(t *testing.T) {
	result := WrapWithInterface("key", "value")

	// Direct nil check — maps via existing direct pass.
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Direct field length check — maps via existing indirect pass.
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty Content")
	}

	// Type assertion + field access — exercises TypeAssertExpr in resolveExprRoot.
	tc := result.Content[0].(TextContent)
	text := tc.Text

	var data map[string]any
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Container unwrap assertion — maps via container unwrap pass.
	if data["key"] != "value" {
		t.Errorf("data[\"key\"] = %v, want \"value\"", data["key"])
	}
}

// TestWrapMCPStyle_ErrorExclusion validates FR-009: the error assertion
// from json.Unmarshal should NOT be mapped to ReturnValue, but the
// data field assertions SHOULD be.
func TestWrapMCPStyle_ErrorExclusion(t *testing.T) {
	result := WrapMCPStyle("color", "blue")

	// Direct nil check — maps via existing direct pass.
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Direct field length check — maps via existing indirect pass.
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty Content")
	}

	text := result.Content[0].Text
	var data map[string]any
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatal(err)
	}

	// Container unwrap assertions — map via container unwrap pass.
	if len(data) != 1 {
		t.Errorf("len(data) = %d, want 1", len(data))
	}
	if data["color"] != "blue" {
		t.Errorf("data[\"color\"] = %v, want \"blue\"", data["color"])
	}
}
