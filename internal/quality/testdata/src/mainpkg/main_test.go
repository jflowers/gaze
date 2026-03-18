package main

import "testing"

func TestAdd(t *testing.T) {
	got := add(2, 3)
	if got != 5 {
		t.Errorf("add(2, 3) = %d, want 5", got)
	}
}

func TestGreet(t *testing.T) {
	msg, err := greet("Alice")
	if err != nil {
		t.Fatalf("greet(Alice) returned error: %v", err)
	}
	if msg != "Hello, Alice!" {
		t.Errorf("greet(Alice) = %q, want %q", msg, "Hello, Alice!")
	}
}

func TestGreet_EmptyName(t *testing.T) {
	_, err := greet("")
	if err == nil {
		t.Fatal("greet('') expected error, got nil")
	}
}
