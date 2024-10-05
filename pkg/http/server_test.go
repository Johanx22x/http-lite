package http

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// MockConnWithReader simulates a network connection that reads from a provided buffer.
type MockConnWithReader struct {
	net.Conn
	reader *bufio.Reader
	closed bool
}

// Read simulates reading data from the connection.
func (m *MockConnWithReader) Read(b []byte) (int, error) {
	// Simulate that the connection is closed after the data is fully read.
	if m.closed {
		return 0, io.EOF
	}
	n, err := m.reader.Read(b)
	if err == io.EOF {
		m.closed = true
	}
	return n, err
}

// Write simulates writing data, but it's not used in these tests.
func (m *MockConnWithReader) Write(b []byte) (int, error) {
	return len(b), nil
}

// Close simulates closing the connection.
func (m *MockConnWithReader) Close() error {
	m.closed = true
	return nil
}

// TestParseRequest_Successful verifies that valid requests are parsed correctly.
func TestParseRequest_Successful(t *testing.T) {
	rawRequest := "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: GoTest\r\nCookie: session_id=abc123\r\n\r\n"
	conn := &MockConnWithReader{reader: bufio.NewReader(strings.NewReader(rawRequest))}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req, err := parseRequest(ctx, conn)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify that the main fields of the request are correct.
	if req.Method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", req.Method)
	}
	if req.URL.Path != "/" {
		t.Errorf("Expected path '/', got '%s'", req.URL.Path)
	}
	if req.Proto != "HTTP/1.1" {
		t.Errorf("Expected proto 'HTTP/1.1', got '%s'", req.Proto)
	}
	if req.Header.Get("Host") != "localhost" {
		t.Errorf("Expected Host 'localhost', got '%s'", req.Header.Get("Host"))
	}

	// Verify that the cookies were parsed correctly.
	if len(req.Cookies) != 1 || req.Cookies[0].Value != "abc123" {
		t.Errorf("Expected cookie session_id=abc123, got '%v'", req.Cookies)
	}
}

// TestParseRequest_SuccessfulPost verifies that a POST request is parsed correctly.
func TestParseRequest_SuccessfulPost(t *testing.T) {
	rawRequest := "POST /submit HTTP/1.1\r\nHost: localhost\r\nContent-Type: application/json\r\n\r\n{\"data\": \"test\"}"
	conn := &MockConnWithReader{reader: bufio.NewReader(strings.NewReader(rawRequest))}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req, err := parseRequest(ctx, conn)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify that the main fields of the POST request are correct.
	if req.Method != "POST" {
		t.Errorf("Expected method 'POST', got '%s'", req.Method)
	}
	if req.URL.Path != "/submit" {
		t.Errorf("Expected path '/submit', got '%s'", req.URL.Path)
	}
	if req.Proto != "HTTP/1.1" {
		t.Errorf("Expected proto 'HTTP/1.1', got '%s'", req.Proto)
	}
	if req.Header.Get("Host") != "localhost" {
		t.Errorf("Expected Host 'localhost', got '%s'", req.Header.Get("Host"))
	}
}

// TestParseRequest_MalformedRequestLine verifies that a malformed request line returns an error.
func TestParseRequest_MalformedRequestLine(t *testing.T) {
	rawRequest := "GET /malformed HTTP\r\nHost: localhost\r\n\r\n" // Incorrect request line
	conn := &MockConnWithReader{reader: bufio.NewReader(strings.NewReader(rawRequest))}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := parseRequest(ctx, conn)
	if err == nil {
		t.Fatal("Expected error for malformed request line, got none")
	}
}

// TestParseRequest_UnsupportedProtocol verifies that an unsupported HTTP protocol returns an error.
func TestParseRequest_UnsupportedProtocol(t *testing.T) {
	rawRequest := "GET / HTTP/2.0\r\nHost: localhost\r\n\r\n" // Unsupported protocol
	conn := &MockConnWithReader{reader: bufio.NewReader(strings.NewReader(rawRequest))}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := parseRequest(ctx, conn)
	if err == nil {
		t.Fatal("Expected error for unsupported protocol, got none")
	}
}

// TestParseRequest_HeaderMalformed verifies that malformed headers return an error.
func TestParseRequest_HeaderMalformed(t *testing.T) {
	rawRequest := "GET / HTTP/1.1\r\nHost localhost\r\n\r\n" // Missing ":" separator in the header
	conn := &MockConnWithReader{reader: bufio.NewReader(strings.NewReader(rawRequest))}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := parseRequest(ctx, conn)
	if err == nil {
		t.Fatal("Expected error for malformed header, got none")
	}
}

// MockConnWithSlowRead simulates a slow-reading connection to test timeouts.
type MockConnWithSlowRead struct {
	reader      *bufio.Reader
	writeBuffer bytes.Buffer
}

// Read simulates a slow data read.
func (mc *MockConnWithSlowRead) Read(b []byte) (n int, err error) {
	time.Sleep(200 * time.Millisecond) // Simulate slow read
	return mc.reader.Read(b)
}

// Write writes to the simulated response buffer.
func (mc *MockConnWithSlowRead) Write(b []byte) (n int, err error) {
	return mc.writeBuffer.Write(b)
}

// Close simulates closing the connection.
func (mc *MockConnWithSlowRead) Close() error {
	return nil
}

// LocalAddr returns a simulated address.
func (mc *MockConnWithSlowRead) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr returns a simulated remote address.
func (mc *MockConnWithSlowRead) RemoteAddr() net.Addr {
	return nil
}

// SetDeadline simulates setting a deadline for operations.
func (mc *MockConnWithSlowRead) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline simulates setting a read deadline for operations.
func (mc *MockConnWithSlowRead) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline simulates setting a write deadline for operations.
func (mc *MockConnWithSlowRead) SetWriteDeadline(t time.Time) error {
	return nil
}

// TestParseRequest_Timeout verifica que una solicitud que tarda demasiado en leerse retorna un error por timeout.
func TestParseRequest_Timeout(t *testing.T) {
	rawRequest := "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"
	conn := &MockConnWithSlowRead{reader: bufio.NewReader(strings.NewReader(rawRequest))}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond) // Timeout muy corto
	defer cancel()

	_, err := parseRequest(ctx, conn)
	if err == nil || err != context.DeadlineExceeded {
		t.Fatalf("Expected deadline exceeded error, got %v", err)
	}
}

// TestParseCookies_Success verifica que las cookies se parseen correctamente desde el header "Cookie".
func TestParseCookies_Success(t *testing.T) {
	cookieHeader := "session_id=abc123; user=JohnDoe; theme=dark"
	cookies := parseCookies(cookieHeader)

	if len(cookies) != 3 {
		t.Fatalf("Expected 3 cookies, got %d", len(cookies))
	}

	expected := []Cookie{
		{Name: "session_id", Value: "abc123"},
		{Name: "user", Value: "JohnDoe"},
		{Name: "theme", Value: "dark"},
	}

	for i, cookie := range cookies {
		if cookie != expected[i] {
			t.Errorf("Expected cookie %v, got %v", expected[i], cookie)
		}
	}
}

// TestParseCookies_Malformed verifica que las cookies malformadas se manejen correctamente.
func TestParseCookies_Malformed(t *testing.T) {
	cookieHeader := "session_id=abc123; malformed_cookie; theme=dark"
	cookies := parseCookies(cookieHeader)

	if len(cookies) != 2 {
		t.Fatalf("Expected 2 valid cookies, got %d", len(cookies))
	}

	expected := []Cookie{
		{Name: "session_id", Value: "abc123"},
		{Name: "theme", Value: "dark"},
	}

	for i, cookie := range cookies {
		if cookie != expected[i] {
			t.Errorf("Expected cookie %v, got %v", expected[i], cookie)
		}
	}
}

// TestParseCookies_EmptyValue verifica que las cookies sin valor se manejen correctamente.
func TestParseCookies_EmptyValue(t *testing.T) {
	cookieHeader := "session_id=abc123; user=; theme=dark"
	cookies := parseCookies(cookieHeader)

	if len(cookies) != 3 {
		t.Fatalf("Expected 3 cookies, got %d", len(cookies))
	}

	expected := []Cookie{
		{Name: "session_id", Value: "abc123"},
		{Name: "user", Value: ""},
		{Name: "theme", Value: "dark"},
	}

	for i, cookie := range cookies {
		if cookie != expected[i] {
			t.Errorf("Expected cookie %v, got %v", expected[i], cookie)
		}
	}
}

// TestHandleConn_Success verifica que una conexión válida lea correctamente una solicitud y la maneje con el handler asignado.
func TestHandleConn_Success(t *testing.T) {
	mockHandler := &MockHandler{}
	server := NewServer(":8080", mockHandler)

	mockConn := &MockConn{}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Simula la solicitud HTTP
	mockConn.writeBuffer.WriteString("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n")

	go server.handleConn(ctx, mockConn)

	time.Sleep(100 * time.Millisecond)

	if len(mockConn.writeBuffer.String()) == 0 {
		t.Errorf("Expected response, got empty output")
	}
}

// TestHandleConn_Timeout verifica que una solicitud que demora demasiado en procesarse termine en un timeout.
func TestHandleConn_Timeout(t *testing.T) {
	mockHandler := &MockHandler{}
	server := NewServer(":8080", mockHandler)

	mockConn := &MockConnWithSlowRead{reader: bufio.NewReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"))}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	go server.handleConn(ctx, mockConn)

	time.Sleep(100 * time.Millisecond)

	if !strings.Contains(mockConn.writeBuffer.String(), "400 Bad Request") {
		t.Errorf("Expected timeout and bad request response, got '%s'", mockConn.writeBuffer.String())
	}
}

// TestHandleConn_MalformedRequest verifica que una solicitud malformada se maneje con un error.
func TestHandleConn_MalformedRequest(t *testing.T) {
	mockHandler := &MockHandler{}
	server := NewServer(":8080", mockHandler)

	mockConn := &MockConn{}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Simula una solicitud malformada
	mockConn.writeBuffer.WriteString("BADREQUEST\r\n\r\n")

	go server.handleConn(ctx, mockConn)

	time.Sleep(100 * time.Millisecond)

	if !strings.Contains(mockConn.writeBuffer.String(), "400 Bad Request") {
		t.Errorf("Expected '400 Bad Request', got '%s'", mockConn.writeBuffer.String())
	}
}

// TestHandleConn_ClosedBeforeComplete verifica que cuando la conexión se cierra antes de completar
// la lectura de la solicitud, no se envíe una respuesta o se maneje el error correctamente.
func TestHandleConn_ClosedBeforeComplete(t *testing.T) {
	mockHandler := &MockHandler{}
	server := NewServer(":8080", mockHandler)

	// Crear una conexión simulada que se cierra antes de que la solicitud se complete
	mockConn := &MockConnWithCloseBeforeComplete{
		reader: bufio.NewReader(strings.NewReader("GET / HTTP/1.1\r\n")),
	}

	// Simula el contexto con un timeout para la lectura
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go server.handleConn(ctx, mockConn)

	// Espera un poco para que la conexión se cierre
	time.Sleep(200 * time.Millisecond)

	// Verifica que no se haya enviado una respuesta después de que la conexión se cerró
	if len(mockConn.writeBuffer.String()) > 0 {
		t.Errorf("Expected no response after connection closed, but got '%s'", mockConn.writeBuffer.String())
	}
}

// MockConnWithCloseBeforeComplete simula una conexión que se cierra antes de completar la solicitud.
type MockConnWithCloseBeforeComplete struct {
	reader      *bufio.Reader
	writeBuffer bytes.Buffer
	closed      bool
}

// Read simula la lectura de datos, pero cierra la conexión antes de que se complete.
func (mc *MockConnWithCloseBeforeComplete) Read(b []byte) (n int, err error) {
	if mc.closed {
		return 0, io.EOF
	}
	mc.closed = true
	return mc.reader.Read(b)
}

// Write escribe en el buffer de respuesta simulado.
func (mc *MockConnWithCloseBeforeComplete) Write(b []byte) (n int, err error) {
	return mc.writeBuffer.Write(b)
}

// Close simula el cierre de la conexión.
func (mc *MockConnWithCloseBeforeComplete) Close() error {
	mc.closed = true
	return nil
}

// LocalAddr devuelve una dirección simulada.
func (mc *MockConnWithCloseBeforeComplete) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr devuelve una dirección simulada.
func (mc *MockConnWithCloseBeforeComplete) RemoteAddr() net.Addr {
	return nil
}

// SetDeadline simula la configuración de un tiempo límite para las operaciones.
func (mc *MockConnWithCloseBeforeComplete) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline simula la configuración de un tiempo límite para las operaciones de lectura.
func (mc *MockConnWithCloseBeforeComplete) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline simula la configuración de un tiempo límite para las operaciones de escritura.
func (mc *MockConnWithCloseBeforeComplete) SetWriteDeadline(t time.Time) error {
	return nil
}

// TestHandleConn_OverloadedServer verifica qué sucede cuando el servidor está sobrecargado y no puede procesar nuevas conexiones.
func TestHandleConn_OverloadedServer(t *testing.T) {
	mockHandler := &MockHandler{}
	server := NewServer(":8080", mockHandler)

	mockConn := &MockConn{}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	// Simular múltiples conexiones para sobrecargar el servidor
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			server.handleConn(ctx, mockConn)
		}()
	}

	wg.Wait()

	if len(mockConn.writeBuffer.String()) == 0 {
		t.Errorf("Expected some responses even under load, but got empty output")
	}
}
