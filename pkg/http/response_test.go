package http

import (
	"testing"
)

// TestWriteHeader verifica que WriteHeader escriba correctamente los encabezados de la respuesta.
func TestWriteHeader(t *testing.T) {
	conn := &MockConn{}
	writer := NewResponseWriter(conn)

	writer.WriteHeader(StatusOK)

	expectedHeader := "HTTP/1.1 200 OK\r\n\r\n"
	actual := conn.writeBuffer.String()

	if actual != expectedHeader {
		t.Errorf("Expected header '%s', got '%s'", expectedHeader, actual)
	}
}

// TestWrite verifica que Write escriba los datos en la conexión.
func TestWrite(t *testing.T) {
	conn := &MockConn{}
	writer := NewResponseWriter(conn)
	writer.WriteHeader(StatusOK)

	body := []byte("Hello, World!")
	n, err := writer.Write(body)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if n != len(body) {
		t.Errorf("Expected %d bytes written, got %d", len(body), n)
	}

	expectedOutput := "HTTP/1.1 200 OK\r\n\r\nHello, World!"
	actualOutput := conn.writeBuffer.String()

	if actualOutput != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, actualOutput)
	}
}

// TestWriteHeaderAlreadySent verifica que no se escriba el encabezado de la respuesta más de una vez.
func TestWriteHeaderAlreadySent(t *testing.T) {
	conn := &MockConn{}
	writer := NewResponseWriter(conn)

	writer.WriteHeader(StatusOK)
	writer.WriteHeader(StatusBadRequest) // No debería sobrescribir el encabezado ya enviado

	expectedOutput := "HTTP/1.1 200 OK\r\n\r\n"
	actualOutput := conn.writeBuffer.String()

	if actualOutput != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, actualOutput)
	}
}
