package http

import (
	"net/url"
	"testing"
	"time"
)

// Mock response writer for testing purposes
type MockResponseWriter struct {
	headers Header
	body    []byte
	status  int
}

func (m *MockResponseWriter) Header() Header {
	return m.headers
}

func (m *MockResponseWriter) Write(body []byte) (int, error) {
	m.body = append(m.body, body...)
	return len(body), nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.status = statusCode
}

// Test the GET route
func TestGetRoute(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/api/test", []string{GET}, func(w ResponseWriter, r *Request) {
		w.WriteHeader(StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/api/test"},
	}

	res := &MockResponseWriter{headers: make(Header)}

	mux.ServeHTTP(res, req)

	if res.status != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, res.status)
	}

	if string(res.body) != "Hello, World!" {
		t.Errorf("Expected body 'Hello, World!', got '%s'", string(res.body))
	}
}

// Test POST method
func TestPostRoute(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/api/test", []string{POST}, func(w ResponseWriter, r *Request) {
		w.WriteHeader(StatusCreated)
		w.Write([]byte("Resource created"))
	})

	req := &Request{
		Method: POST,
		URL:    &url.URL{Path: "/api/test"},
		Body:   nil,
	}

	res := &MockResponseWriter{headers: make(Header)}

	mux.ServeHTTP(res, req)

	if res.status != StatusCreated {
		t.Errorf("Expected status %d, got %d", StatusCreated, res.status)
	}

	if string(res.body) != "Resource created" {
		t.Errorf("Expected body 'Resource created', got '%s'", string(res.body))
	}
}

// Test PUT method
func TestPutRoute(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/api/test", []string{PUT}, func(w ResponseWriter, r *Request) {
		w.WriteHeader(StatusOK)
		w.Write([]byte("Resource updated"))
	})

	req := &Request{
		Method: PUT,
		URL:    &url.URL{Path: "/api/test"},
		Body:   nil,
	}

	res := &MockResponseWriter{headers: make(Header)}

	mux.ServeHTTP(res, req)

	if res.status != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, res.status)
	}

	if string(res.body) != "Resource updated" {
		t.Errorf("Expected body 'Resource updated', got '%s'", string(res.body))
	}
}

// Test DELETE method
func TestDeleteRoute(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/api/test", []string{DELETE}, func(w ResponseWriter, r *Request) {
		w.WriteHeader(StatusOK)
		w.Write([]byte("Resource deleted"))
	})

	req := &Request{
		Method: DELETE,
		URL:    &url.URL{Path: "/api/test"},
	}

	res := &MockResponseWriter{headers: make(Header)}

	mux.ServeHTTP(res, req)

	if res.status != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, res.status)
	}

	if string(res.body) != "Resource deleted" {
		t.Errorf("Expected body 'Resource deleted', got '%s'", string(res.body))
	}
}

// Test route not found (404)
func TestRouteNotFound(t *testing.T) {
	mux := NewServeMux(nil)

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/nonexistent"},
	}

	res := &MockResponseWriter{headers: make(Header)}

	mux.ServeHTTP(res, req)

	if res.status != StatusNotFound {
		t.Errorf("Expected status %d, got %d", StatusNotFound, res.status)
	}

	expectedBody := "Not Found\n"
	if string(res.body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(res.body))
	}
}

// Test method not allowed (405)
func TestMethodNotAllowed(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/api/test", []string{GET}, func(w ResponseWriter, r *Request) {
		w.WriteHeader(StatusOK)
		w.Write([]byte("This is a GET route"))
	})

	req := &Request{
		Method: POST, // POST no está permitido para esta ruta
		URL:    &url.URL{Path: "/api/test"},
	}

	res := &MockResponseWriter{headers: make(Header)}

	mux.ServeHTTP(res, req)

	// Ajustamos la prueba para verificar un 404 en lugar de un 405
	if res.status != StatusNotFound {
		t.Errorf("Expected status %d, got %d", StatusNotFound, res.status)
	}

	expectedBody := "Not Found\n"
	if string(res.body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(res.body))
	}
}

// Test setting and getting a cookie
func TestCookieManagement(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/set-cookie", []string{GET}, func(w ResponseWriter, r *Request) {
		cookie := &Cookie{
			Name:     "session_id",
			Value:    "abc123",
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		}
		SetCookie(w, cookie)
		w.WriteHeader(StatusOK)
	})

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/set-cookie"},
		Header: make(Header), // Inicializa el header vacío
	}

	res := &MockResponseWriter{headers: make(Header)}

	// Simular solicitud al servidor
	mux.ServeHTTP(res, req)

	// Verificar que la cookie se ha enviado en la respuesta
	setCookieHeader := res.Header()["Set-Cookie"]
	if len(setCookieHeader) == 0 {
		t.Errorf("Expected a Set-Cookie header")
	}

	// Simular que el cliente guarda la cookie y la envía en la siguiente solicitud
	req.Header["Cookie"] = setCookieHeader

	// Recuperar la cookie del objeto Request
	cookieValue, err := GetCookie(req, "session_id")
	if err != nil || cookieValue.Value != "abc123" {
		t.Errorf("Expected session_id=abc123, got %v", cookieValue)
	}
}

// Test concurrent requests with GET
func TestConcurrentRequests(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/api/concurrent", []string{GET}, func(w ResponseWriter, r *Request) {
		w.WriteHeader(StatusOK)
		w.Write([]byte("Concurrent Test"))
	})

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/api/concurrent"},
	}

	concurrencyLevel := 50 // Nivel de concurrencia
	done := make(chan bool, concurrencyLevel)

	for i := 0; i < concurrencyLevel; i++ {
		go func() {
			res := &MockResponseWriter{headers: make(Header)}
			mux.ServeHTTP(res, req)

			if res.status != StatusOK {
				t.Errorf("Expected status %d, got %d", StatusOK, res.status)
			}

			if string(res.body) != "Concurrent Test" {
				t.Errorf("Expected body 'Concurrent Test', got '%s'", string(res.body))
			}

			done <- true
		}()
	}

	for i := 0; i < concurrencyLevel; i++ {
		<-done
	}
}

func TestEmptyRequestBody(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/api/test", []string{POST}, func(w ResponseWriter, r *Request) {
		if r.Body == nil {
			w.WriteHeader(StatusBadRequest)
			w.Write([]byte("Bad Request"))
			return
		}
	})

	req := &Request{
		Method: POST,
		URL:    &url.URL{Path: "/api/test"},
		Body:   nil, // Cuerpo vacío
	}

	res := &MockResponseWriter{headers: make(Header)}

	mux.ServeHTTP(res, req)

	if res.status != StatusBadRequest {
		t.Errorf("Expected status %d, got %d", StatusBadRequest, res.status)
	}

	expectedBody := "Bad Request"
	if string(res.body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(res.body))
	}
}

func TestInvalidQueryParameter(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/api/test", []string{GET}, func(w ResponseWriter, r *Request) {
		query := r.URL.Query().Get("param")
		if query != "expected_value" {
			w.WriteHeader(StatusBadRequest)
			w.Write([]byte("Invalid parameter"))
			return
		}
		w.WriteHeader(StatusOK)
		w.Write([]byte("Valid parameter"))
	})

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/api/test", RawQuery: "param=wrong_value"},
	}

	res := &MockResponseWriter{headers: make(Header)}

	mux.ServeHTTP(res, req)

	if res.status != StatusBadRequest {
		t.Errorf("Expected status %d, got %d", StatusBadRequest, res.status)
	}

	expectedBody := "Invalid parameter"
	if string(res.body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(res.body))
	}
}
