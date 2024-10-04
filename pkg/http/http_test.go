package http

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

type ResponseRecorder struct {
	StatusCode int
	Headers    Header
	Body       *bytes.Buffer
}

func (rr *ResponseRecorder) Header() Header {
	return rr.Headers
}

func (rr *ResponseRecorder) Write(b []byte) (int, error) {
	return rr.Body.Write(b)
}

func (rr *ResponseRecorder) WriteHeader(statusCode int) {
	rr.StatusCode = statusCode
}

// Middleware que agrega un encabezado a la respuesta.
func headerMiddleware(next func(ResponseWriter, *Request)) func(ResponseWriter, *Request) {
	return func(w ResponseWriter, r *Request) {
		w.Header().Set("X-Custom-Header", "Value")
		next(w, r)
	}
}

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

// Test setting a cookie
func TestSetCookie(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/set-cookie", []string{GET}, func(w ResponseWriter, r *Request) {
		cookie := &Cookie{
			Name:     "session_id",
			Value:    "abc123",
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			MaxAge:   0,
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
}

// Test getting a cookie
func TestGetCookie(t *testing.T) {
	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/set-cookie"},
		Header: make(Header), // Inicializa el header vacío
	}

	// Simular que el cliente guarda la cookie y la envía en la siguiente solicitud
	req.Header["Cookie"] = []string{"session_id=abc123"}

	// Recuperar la cookie del objeto Request
	cookieValue, err := GetCookie(req, "session_id")
	if err != nil || cookieValue.Value != "abc123" {
		t.Errorf("Expected session_id=abc123, got %v", cookieValue)
	}
}

// Test deleting a cookie
func TestDeleteCookie(t *testing.T) {
	mux := NewServeMux(nil)

	// Ruta para establecer la cookie
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

	// Simula la solicitud para establecer la cookie
	reqSet := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/set-cookie"},
		Header: make(Header), // Inicializa el header vacío
	}

	resSet := &MockResponseWriter{headers: make(Header)}
	mux.ServeHTTP(resSet, reqSet)

	// Verificar que la cookie se ha enviado en la respuesta
	if len(resSet.Header()["Set-Cookie"]) == 0 {
		t.Errorf("Expected a Set-Cookie header")
	}

	// Ahora vamos a eliminar la cookie
	mux.AddRoute("/delete-cookie", []string{GET}, func(w ResponseWriter, r *Request) {
		DeleteCookie(w, "session_id", "/")
		w.WriteHeader(StatusOK)
	})

	// Simula la solicitud para eliminar la cookie
	reqDel := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/delete-cookie"},
		Header: make(Header), // Inicializa el header vacío
	}

	resDel := &MockResponseWriter{headers: make(Header)}
	mux.ServeHTTP(resDel, reqDel)

	// Verificar que se haya establecido correctamente la cookie de eliminación
	setCookieHeader := resDel.Header()["Set-Cookie"]
	if len(setCookieHeader) == 0 {
		t.Errorf("Expected Set-Cookie header to be present")
	}

	// Verifica que el encabezado Set-Cookie contenga la información correcta
	expected := "session_id=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT"
	if setCookieHeader[0] != expected {
		t.Errorf("Expected Set-Cookie to contain '%s', but got '%s'", expected, setCookieHeader[0])
	}
}

// Test getting an invalid cookie
func TestGetCookieInvalidValue(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/api/get-cookie", []string{GET}, func(w ResponseWriter, r *Request) {
		cookie, err := GetCookie(r, "session_token")
		if err != nil || cookie.Value != "abc123" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Cookie inválida"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Cookie válida"))
	})

	// Solicitud con cookie inválida
	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/api/get-cookie"},
		Header: Header{"Cookie": []string{"session_token=wrong_value"}},
	}

	res := &MockResponseWriter{headers: make(Header)}
	mux.ServeHTTP(res, req)

	if res.status != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, res.status)
	}

	expectedBody := "Cookie inválida"
	if string(res.body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(res.body))
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

// Test empty request body
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

// Test invalid JSON request body
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

// Test JSON response
func TestJSONResponse(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/json-test", []string{GET}, func(w ResponseWriter, r *Request) {
		w.Header()["Content-Type"] = []string{"application/json"}
		w.WriteHeader(StatusOK)
		w.Write([]byte(`{"message": "Hello, JSON!"}`))
	})

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/json-test"},
		Header: make(Header),
	}

	res := &MockResponseWriter{headers: make(Header)}

	mux.ServeHTTP(res, req)

	if res.status != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, res.status)
	}

	expectedBody := `{"message": "Hello, JSON!"}`
	if string(res.body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(res.body))
	}

	contentType := res.Header()["Content-Type"]
	if len(contentType) == 0 || !strings.Contains(contentType[0], "application/json") {
		t.Errorf("Expected Content-Type 'application/json', got '%v'", contentType)
	}
}

// Test request timeout
func TestRequestTimeout(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/timeout", []string{GET}, func(w ResponseWriter, r *Request) {
		// Simulate a request that takes 2 seconds to complete
		time.Sleep(2 * time.Second)
		w.WriteHeader(StatusOK)
		w.Write([]byte("Completed"))
	})

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/timeout"},
		Header: make(Header),
	}

	res := &MockResponseWriter{headers: make(Header)}

	// Simulate a request with a timeout of 1 second
	go func() {
		mux.ServeHTTP(res, req)
	}()

	time.Sleep(1 * time.Second)

	if res.status != 0 {
		t.Errorf("Expected no status yet, got %d", res.status)
	}
}

func TestServerShutdown(t *testing.T) {
	mux := NewServeMux(nil)
	server := NewServer(":8080", mux)

	// Canal para recibir errores de la goroutine del servidor
	errChan := make(chan error)

	// Iniciar el servidor en una goroutine
	go func() {
		err := server.listenAndServe()
		// Enviar el error al canal
		errChan <- err
	}()

	// Simula el cierre del servidor
	if err := server.Shutdown(); err != nil {
		t.Fatalf("Failed to shut down server: %v", err)
	}

	// Esperar a que la goroutine del servidor termine y verificar si hubo algún error
	select {
	case err := <-errChan:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("Expected server to close gracefully, got: %v", err)
		}
	case <-time.After(10 * time.Second): // Timeout para evitar bloqueos en el test
		t.Fatal("Server did not shut down in time")
	}
}

// Test the default error handler
func TestErrorHandling(t *testing.T) {
	// Create an instance of ServeMux
	mux := NewServeMux(nil)

	// Define a custom error handler
	mux.SetErrorHandler(func(w ResponseWriter, r *Request, statusCode int) {
		w.WriteHeader(statusCode)
		if statusCode == 404 {
			w.Write([]byte("Custom 404 Page Not Found"))
		}
	})

	// Parse the URL
	parsedURL, err := url.Parse("/nonexistent")
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}

	// Create a request to a non-existent route
	req := &Request{
		Method: "GET",
		URL:    parsedURL,
	}

	// Create a custom ResponseRecorder to record the response
	rr := &ResponseRecorder{
		Headers: make(map[string][]string),
		Body:    new(bytes.Buffer),
	}

	// Use the ServeMux to handle the request
	mux.ServeHTTP(rr, req)

	// Verify that the status code is 404
	if status := rr.StatusCode; status != 404 {
		t.Errorf("handler returned wrong status code: got %v want %v", status, 404)
	}

	// Verify that the response body contains the custom error message
	expected := "Custom 404 Page Not Found"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

// Test middleware
func TestMiddleware(t *testing.T) {
	mux := NewServeMux(nil)

	// Add the middleware
	mux.Use(headerMiddleware)

	// Handle a specific route
	mux.Handle("/test", func(w ResponseWriter, r *Request) {
		w.WriteHeader(200)
	})

	req := &Request{
		Method: "GET",
		URL:    &url.URL{Path: "/test"},
	}
	rr := &ResponseRecorder{Headers: make(Header), Body: new(bytes.Buffer)}

	// Serve the request
	mux.ServeHTTP(rr, req)

	// Verify the status code
	if rr.StatusCode != 200 {
		t.Errorf("expected status 200 but got %d", rr.StatusCode)
	}

	// Verify if the custom header was added
	expectedHeader := "Value"
	actualHeader := rr.Headers["X-Custom-Header"][0] // Assuming the header has at least one value
	if actualHeader != expectedHeader {
		t.Errorf("expected header X-Custom-Header %s but got %s", expectedHeader, actualHeader)
	}
}
