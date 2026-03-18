// Package main is a test fixture for verifying that gaze quality
// correctly analyzes unexported functions in main packages.
package main

import "fmt"

// add returns the sum of two integers.
func add(a, b int) int {
	return a + b
}

// greet returns a greeting string for the given name.
// Returns an error if the name is empty.
func greet(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name cannot be empty")
	}
	return "Hello, " + name + "!", nil
}

func main() {
	fmt.Println(add(1, 2))
	msg, _ := greet("world")
	fmt.Println(msg)
}
