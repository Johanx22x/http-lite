package http

import (
	"bytes"
	"net"
	"time"
)

// MockHandler simula un handler para pruebas. Este implementa la interfaz Handler.
type MockHandler struct{}

// ServeHTTP simula la respuesta HTTP que sería enviada al cliente.
func (m *MockHandler) ServeHTTP(w ResponseWriter, r *Request) {
	w.WriteHeader(StatusOK)          // Simula que el handler siempre devuelve un estado 200 OK.
	w.Write([]byte("Mock response")) // Simula una respuesta con un cuerpo "Mock response".
}

// MockConn simula una conexión de red para pruebas.
type MockConn struct {
	writeBuffer bytes.Buffer
}

// Write simulates writing data to the connection.
func (mc *MockConn) Write(b []byte) (n int, err error) {
	return mc.writeBuffer.Write(b)
}

// Read is not used in these tests, but is included to fulfill the net.Conn interface.
func (mc *MockConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

// Close does nothing in the simulation, but fulfills the net.Conn interface.
func (mc *MockConn) Close() error {
	return nil
}

// LocalAddr returns a simulated address.
func (mc *MockConn) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr returns a simulated address.
func (mc *MockConn) RemoteAddr() net.Addr {
	return nil
}

// SetDeadline simulates setting a deadline for read and write operations.
func (mc *MockConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline simulates setting a deadline for read operations.
func (mc *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline simulates setting a deadline for write operations.
func (mc *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// MockResponseWriter is an implementation of ResponseWriter that captures responses for testing.
type MockResponseWriter struct {
	headers Header
	body    []byte
	status  int
}

func (m *MockResponseWriter) Header() Header {
	if m.headers == nil {
		m.headers = make(Header)
	}
	return m.headers
}

func (m *MockResponseWriter) Write(body []byte) (int, error) {
	m.body = append(m.body, body...)
	return len(body), nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.status = statusCode
}

func (m *MockResponseWriter) SetCookie(cookie *Cookie) {
	m.headers.Set("Set-Cookie", cookie.String())
}

func (m *MockResponseWriter) DeleteCookie(name string) {
	cookie := &Cookie{
		Name:   name,
		Value:  "",
		MaxAge: -1,
	}
	m.headers.Set("Set-Cookie", cookie.String())
}
