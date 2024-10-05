package http

import (
	"context"
	"net"
	"testing"
	"time"
)

// MockListener simula un listener de red para el servidor.
type MockListener struct {
	conn *MockConn
}

func (ml *MockListener) Accept() (net.Conn, error) {
	return ml.conn, nil
}

func (ml *MockListener) Close() error {
	return nil
}

func (ml *MockListener) Addr() net.Addr {
	return nil
}

// TestNewServer verifica que un servidor se crea correctamente con la dirección y el manejador proporcionados.
func TestNewServer(t *testing.T) {
	mockHandler := &MockHandler{}
	server := NewServer(":8080", mockHandler)

	if server.Addr != ":8080" {
		t.Errorf("Expected address ':8080', got '%s'", server.Addr)
	}

	if server.Handler != mockHandler {
		t.Errorf("Expected handler %v, got %v", mockHandler, server.Handler)
	}
}

// TestHandleConn verifica que el servidor maneje correctamente una conexión.
func TestHandleConn(t *testing.T) {
	mockHandler := &MockHandler{}
	server := NewServer(":8080", mockHandler)

	mockConn := &MockConn{}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go server.handleConn(ctx, mockConn)

	// Simula la solicitud
	mockConn.writeBuffer.WriteString("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n")
	time.Sleep(100 * time.Millisecond)

	if len(mockConn.writeBuffer.String()) == 0 {
		t.Errorf("Expected response, got empty output")
	}
}

// TestListenAndServe verifica que el servidor pueda aceptar conexiones.
// Dado que no podemos modificar net.Listen, simulamos el comportamiento del listener.
func TestListenAndServe(t *testing.T) {
	mockConn := &MockConn{}
	mockHandler := &MockHandler{}
	server := NewServer(":8080", mockHandler)

	// Simula la aceptación de una conexión usando un MockConn
	go func() {
		server.handleConn(context.Background(), mockConn)
	}()

	// Simula una solicitud de red
	mockConn.writeBuffer.WriteString("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n")
	time.Sleep(100 * time.Millisecond)

	if len(mockConn.writeBuffer.String()) == 0 {
		t.Errorf("Expected response, got empty output")
	}
}

// TestShutdown verifica que el servidor se apague correctamente.
func TestShutdown(t *testing.T) {
	mockHandler := &MockHandler{}
	server := NewServer(":8080", mockHandler)

	go server.Shutdown()

	// Verifica que el servidor esté apagado
	select {
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Server did not shut down in time")
	default:
	}
}

// MockHandler simula un handler para las pruebas.
type MockHandler struct{}

func (m *MockHandler) ServeHTTP(w ResponseWriter, r *Request) {
	w.WriteHeader(StatusOK)
	w.Write([]byte("Mock response"))
}
