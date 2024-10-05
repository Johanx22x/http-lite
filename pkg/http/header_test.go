package http

import (
	"testing"
)

// TestHeaderSet verifies that the Header's Set method correctly adds and updates header values.
func TestHeaderSet(t *testing.T) {
	headers := make(Header)

	// Probar la inserción de un nuevo encabezado
	headers.Set("Content-Type", "application/json")
	if len(headers["Content-Type"]) != 1 || headers["Content-Type"][0] != "application/json" {
		t.Errorf("Expected Content-Type to be 'application/json', got %v", headers["Content-Type"])
	}

	// Probar la actualización del encabezado existente
	headers.Set("Content-Type", "text/html")
	if len(headers["Content-Type"]) != 2 {
		t.Errorf("Expected Content-Type to have 2 values, got %d", len(headers["Content-Type"]))
	}

	// Comprobar si el último valor es el correcto
	if headers["Content-Type"][1] != "text/html" {
		t.Errorf("Expected Content-Type[1] to be 'text/html', got '%s'", headers["Content-Type"][1])
	}
}

// TestHeaderGet verifies that the Header's Get method correctly retrieves the first value of a header.
func TestHeaderGet(t *testing.T) {
	headers := make(Header)

	// Añadir encabezados
	headers.Set("X-Custom-Header", "Value1")
	headers.Set("X-Custom-Header", "Value2")

	// Obtener el valor del encabezado (debe devolver el primero)
	value := headers.Get("X-Custom-Header")
	if value != "Value1" {
		t.Errorf("Expected 'Value1', got '%s'", value)
	}

	// Probar obtener un encabezado inexistente
	nonExistent := headers.Get("Non-Existent-Header")
	if nonExistent != "" {
		t.Errorf("Expected empty string for non-existent header, got '%s'", nonExistent)
	}
}

// TestMultipleHeaders verifies that headers can handle multiple values for a single key.
func TestMultipleHeaders(t *testing.T) {
	headers := make(Header)

	// Añadir varios valores a un encabezado
	headers.Set("Accept", "text/html")
	headers.Set("Accept", "application/json")

	// Verificar que ambos valores están presentes
	if len(headers["Accept"]) != 2 {
		t.Errorf("Expected 2 values for 'Accept', got %d", len(headers["Accept"]))
	}

	// Verificar el orden de los valores
	if headers["Accept"][0] != "text/html" || headers["Accept"][1] != "application/json" {
		t.Errorf("Expected 'Accept' to contain 'text/html' and 'application/json', got %v", headers["Accept"])
	}
}

// TestHeaderOverwrite verifies that the last value is correctly appended and does not overwrite previous values.
func TestHeaderOverwrite(t *testing.T) {
	headers := make(Header)

	// Añadir un encabezado y luego otro con el mismo nombre
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Cache-Control", "max-age=3600")

	// Verificar que ambos valores están presentes
	if len(headers["Cache-Control"]) != 2 {
		t.Errorf("Expected 2 values for 'Cache-Control', got %d", len(headers["Cache-Control"]))
	}

	// Verificar que los valores son correctos
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

	// Probar obtener un encabezado inexistente
	value := headers.Get("Non-Existent-Header")
	if value != "" {
		t.Errorf("Expected empty string for non-existent header, got '%s'", value)
	}

	// Verificar que un encabezado no existe
	if len(headers) != 0 {
		t.Errorf("Expected empty header map, got %v", headers)
	}
}
