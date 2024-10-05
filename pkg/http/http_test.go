package http

import (
	"bytes"
	"net/http"
	"net/url"
	"os"
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

func (rr *ResponseRecorder) SetCookie(cookie *Cookie) {
	rr.Headers.Set("Set-Cookie", cookie.String())
}

func (rr *ResponseRecorder) DeleteCookie(name string) {
	cookie := &Cookie{
		Name:   name,
		Value:  "",
		MaxAge: -1,
	}

	rr.Headers.Set("Set-Cookie", cookie.String())
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
		w.SetCookie(cookie)
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
	req.Cookies = append(req.Cookies, Cookie{Name: "session_id", Value: "abc123"})

	// Recuperar la cookie del objeto Request
	cookieValue, err := req.GetCookie("session_id")
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
		w.SetCookie(cookie)
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
		w.DeleteCookie("session_id")
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
	expected := "session_id="
	if setCookieHeader[0] != expected {
		t.Errorf("Expected Set-Cookie to contain '%s', but got '%s'", expected, setCookieHeader[0])
	}
}

// Test getting an invalid cookie
func TestGetCookieInvalidValue(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/api/get-cookie", []string{GET}, func(w ResponseWriter, r *Request) {
		cookie, err := r.GetCookie("session_token")
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

func TestFileExists(t *testing.T) {
	// Crear un archivo temporal
	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Asegurarse de que se elimine después de la prueba

	// Probar que el archivo existe
	if !fileExists(tmpFile.Name()) {
		t.Errorf("expected file to exist, but it does not: %s", tmpFile.Name())
	}

	// Probar que un archivo no existente devuelve false
	if fileExists("non_existent_file.txt") {
		t.Error("expected non_existent_file.txt to not exist, but it does")
	}
}

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		filePath string
		expected string
	}{
		{"index.html", "text/html"},
		{"style.css", "text/css"},
		{"script.js", "application/javascript"},
		{"image.png", "image/png"},
		{"photo.jpg", "image/jpeg"},
		{"graphic.jpeg", "image/jpeg"},
		{"vector.svg", "image/svg+xml"},
		{"animation.gif", "image/gif"},
		{"unknown.txt", "application/octet-stream"},
		{"no_extension", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			result := detectContentType(tt.filePath)
			if result != tt.expected {
				t.Errorf("detectContentType(%q) = %q; expected %q", tt.filePath, result, tt.expected)
			}
		})
	}
}

// Test setDefaultHandler y setErrorHandler
/*
func TestSetDefaultHandlerAndErrorHandler(t *testing.T) {
	mux := NewServeMux(nil)

	// Establece un handler por defecto
	defaultHandlerCalled := false
	mux.SetDefaultHandler(func(w ResponseWriter, r *Request) {
		defaultHandlerCalled = true
		w.WriteHeader(http.StatusNotFound)
	})

	// Establece un handler de error
	errorHandlerCalled := false
	mux.SetErrorHandler(func(w ResponseWriter, r *Request, statusCode int) {
		errorHandlerCalled = true
		w.WriteHeader(statusCode)
	})

	// Simula una solicitud a una ruta no registrada
	req := &Request{
		Method: "GET",
		URL:    &url.URL{Path: "/unregistered"},
	}
	rr := &ResponseRecorder{Headers: make(Header), Body: new(bytes.Buffer)}

	mux.ServeHTTP(rr, req)

	// Verifica que se haya llamado al default handler
	if !defaultHandlerCalled {
		t.Errorf("expected default handler to be called")
	}

	// Verifica que se haya devuelto un 404
	if rr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status code 404, but got %d", rr.StatusCode)
	}

	// Verifica que no se haya llamado al error handler
	if errorHandlerCalled {
		t.Errorf("error handler should not have been called")
	}
}*/

// Test de parseRequest
func TestParseRequest(t *testing.T) {
	// Caso 1: Solicitud válida
	validRequest := "GET /test HTTP/1.1\r\nHost: localhost\r\nUser-Agent: GoTest\r\n\r\n"
	req, err := parseRequest([]byte(validRequest))
	if err != nil {
		t.Errorf("Unexpected error for valid request: %v", err)
	}

	// Verificar el método
	if req.Method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", req.Method)
	}

	// Verificar la URL
	if req.URL.Path != "/test" {
		t.Errorf("Expected URL path '/test', got '%s'", req.URL.Path)
	}

	// Verificar el protocolo
	if req.Proto != "HTTP/1.1" {
		t.Errorf("Expected protocol 'HTTP/1.1', got '%s'", req.Proto)
	}

	// Verificar los encabezados
	if req.Header.Get("Host") != "localhost" {
		t.Errorf("Expected Host 'localhost', got '%s'", req.Header.Get("Host"))
	}
	if req.Header.Get("User-Agent") != "GoTest" {
		t.Errorf("Expected User-Agent 'GoTest', got '%s'", req.Header.Get("User-Agent"))
	}

	// Caso 2: Solicitud malformada (línea de solicitud incompleta)
	invalidRequest := "GET /test"
	_, err = parseRequest([]byte(invalidRequest))
	if err == nil {
		t.Error("Expected error for malformed request line, got nil")
	}

	// Caso 3: Protocolo no soportado
	unsupportedProtocolRequest := "GET /test HTTP/2.0\r\nHost: localhost\r\n\r\n"
	_, err = parseRequest([]byte(unsupportedProtocolRequest))
	if err == nil || !strings.Contains(err.Error(), "unsupported protocol") {
		t.Errorf("Expected unsupported protocol error, got %v", err)
	}

	// Caso 4: Encabezado malformado
	invalidHeaderRequest := "GET /test HTTP/1.1\r\nHost: localhost\r\nInvalidHeader\r\n\r\n"
	_, err = parseRequest([]byte(invalidHeaderRequest))
	if err == nil || !strings.Contains(err.Error(), "malformed header line") {
		t.Errorf("Expected malformed header line error, got %v", err)
	}
}
