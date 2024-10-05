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

// Write simula escribir datos en la conexión.
func (mc *MockConn) Write(b []byte) (n int, err error) {
	return mc.writeBuffer.Write(b)
}

// Read no se utiliza en estas pruebas, pero se incluye para cumplir la interfaz de net.Conn.
func (mc *MockConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

// Close no hace nada en la simulación, pero cumple con la interfaz de net.Conn.
func (mc *MockConn) Close() error {
	return nil
}

// LocalAddr devuelve una dirección simulada.
func (mc *MockConn) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr devuelve una dirección simulada.
func (mc *MockConn) RemoteAddr() net.Addr {
	return nil
}

// SetDeadline simula establecer un tiempo límite para las operaciones de lectura y escritura.
func (mc *MockConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline simula establecer un tiempo límite para las operaciones de lectura.
func (mc *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline simula establecer un tiempo límite para las operaciones de escritura.
func (mc *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// MockResponseWriter es una implementación de ResponseWriter que captura respuestas para pruebas.
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
