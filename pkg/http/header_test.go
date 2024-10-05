package http

import (
	"testing"
)

// TestHeaderSet verifies that the Header's Set method correctly adds and updates header values.
func TestHeaderSet(t *testing.T) {
	headers := make(Header)

	// Test inserting a new header
	headers.Set("Content-Type", "application/json")
	if len(headers["Content-Type"]) != 1 || headers["Content-Type"][0] != "application/json" {
		t.Errorf("Expected Content-Type to be 'application/json', got %v", headers["Content-Type"])
	}

	// Test updating the existing header
	headers.Set("Content-Type", "text/html")
	if len(headers["Content-Type"]) != 2 {
		t.Errorf("Expected Content-Type to have 2 values, got %d", len(headers["Content-Type"]))
	}

	// Check if the last value is correct
	if headers["Content-Type"][1] != "text/html" {
		t.Errorf("Expected Content-Type[1] to be 'text/html', got '%s'", headers["Content-Type"][1])
	}
}

// TestHeaderGet verifies that the Header's Get method correctly retrieves the first value of a header.
func TestHeaderGet(t *testing.T) {
	headers := make(Header)

	// Add headers
	headers.Set("X-Custom-Header", "Value1")
	headers.Set("X-Custom-Header", "Value2")

	// Get the header value (should return the first one)
	value := headers.Get("X-Custom-Header")
	if value != "Value1" {
		t.Errorf("Expected 'Value1', got '%s'", value)
	}

	// Test getting a non-existent header
	nonExistent := headers.Get("Non-Existent-Header")
	if nonExistent != "" {
		t.Errorf("Expected empty string for non-existent header, got '%s'", nonExistent)
	}
}

// TestMultipleHeaders verifies that headers can handle multiple values for a single key.
func TestMultipleHeaders(t *testing.T) {
	headers := make(Header)

	// Add multiple values to a header
	headers.Set("Accept", "text/html")
	headers.Set("Accept", "application/json")

	// Verify that both values are present
	if len(headers["Accept"]) != 2 {
		t.Errorf("Expected 2 values for 'Accept', got %d", len(headers["Accept"]))
	}

	// Verify the order of the values
	if headers["Accept"][0] != "text/html" || headers["Accept"][1] != "application/json" {
		t.Errorf("Expected 'Accept' to contain 'text/html' and 'application/json', got %v", headers["Accept"])
	}
}

// TestHeaderOverwrite verifies that the last value is correctly appended and does not overwrite previous values.
func TestHeaderOverwrite(t *testing.T) {
	headers := make(Header)

	// Add a header and then another with the same name
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Cache-Control", "max-age=3600")

	// Verify that both values are present
	if len(headers["Cache-Control"]) != 2 {
		t.Errorf("Expected 2 values for 'Cache-Control', got %d", len(headers["Cache-Control"]))
	}

	// Verify that the values are correct
	expectedValues := []string{"no-cache", "max-age=3600"}
	for i, v := range expectedValues {
		if headers["Cache-Control"][i] != v {
			t.Errorf("Expected 'Cache-Control[%d]' to be '%s', got '%s'", i, v, headers["Cache-Control"][i])
		}
	}
}

// TestEmptyHeader verifies behavior when no headers are set.
func TestEmptyHeader(t *testing.T) {
	headers := make(Header)

	// Test getting a non-existent header
	value := headers.Get("Non-Existent-Header")
	if value != "" {
		t.Errorf("Expected empty string for non-existent header, got '%s'", value)
	}

	// Verify that a header does not exist
	if len(headers) != 0 {
		t.Errorf("Expected empty header map, got %v", headers)
	}
}
