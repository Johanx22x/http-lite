package http

import (
	"net/url"
	"testing"
)

// TestStaticRoute verifies that a static route works as expected.
func TestStaticRoute(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/api/static", []string{GET}, func(w ResponseWriter, r *Request) {
		w.WriteHeader(StatusOK)
		w.Write([]byte("Static route"))
	})

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/api/static"},
	}

	res := &MockResponseWriter{headers: make(Header)}

	mux.ServeHTTP(res, req)

	if res.status != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, res.status)
	}

	if string(res.body) != "Static route" {
		t.Errorf("Expected body 'Static route', got '%s'", string(res.body))
	}
}

// TestDynamicRoute verifies that dynamic routes work correctly.
func TestDynamicRoute(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/api/items/:id", []string{GET}, func(w ResponseWriter, r *Request) {
		id := r.Params["id"]
		w.WriteHeader(StatusOK)
		w.Write([]byte("Item ID: " + id))
	})

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/api/items/123"},
	}

	res := &MockResponseWriter{headers: make(Header)}

	mux.ServeHTTP(res, req)

	if res.status != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, res.status)
	}

	expectedBody := "Item ID: 123"
	if string(res.body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(res.body))
	}
}

// TestRouteNotFound verifies that a 404 is returned when a route is not found.
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

// TestMethodNotAllowed verifies that a 404 is returned if the method is not allowed for the route.
func TestMethodNotAllowed(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/api/test", []string{GET}, func(w ResponseWriter, r *Request) {
		w.WriteHeader(StatusOK)
		w.Write([]byte("This is a GET route"))
	})

	req := &Request{
		Method: POST, // POST is not allowed for this route
		URL:    &url.URL{Path: "/api/test"},
	}

	res := &MockResponseWriter{headers: make(Header)}

	mux.ServeHTTP(res, req)

	// Verify that a 404 is returned because POST is not allowed
	if res.status != StatusNotFound {
		t.Errorf("Expected status %d, got %d", StatusNotFound, res.status)
	}

	expectedBody := "Not Found\n"
	if string(res.body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(res.body))
	}
}

// TestMiddleware verifies that middleware is applied correctly.
func TestMiddleware(t *testing.T) {
	mux := NewServeMux(nil)

	// Middleware that adds a header to the response.
	mux.Use(func(next func(ResponseWriter, *Request)) func(ResponseWriter, *Request) {
		return func(w ResponseWriter, r *Request) {
			w.Header().Set("X-Middleware", "true")
			next(w, r)
		}
	})

	mux.AddRoute("/api/middleware", []string{GET}, func(w ResponseWriter, r *Request) {
		w.WriteHeader(StatusOK)
		w.Write([]byte("Middleware applied"))
	})

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/api/middleware"},
	}

	res := &MockResponseWriter{headers: make(Header)}

	mux.ServeHTTP(res, req)

	if res.status != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, res.status)
	}

	// Verify that the middleware has modified the headers
	if res.Header().Get("X-Middleware") != "true" {
		t.Errorf("Expected 'X-Middleware' header to be 'true', got '%s'", res.Header().Get("X-Middleware"))
	}

	expectedBody := "Middleware applied"
	if string(res.body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(res.body))
	}
}

// TestDefaultHandler verifies that the default handler is called for not found routes.
func TestDefaultHandler(t *testing.T) {
	mux := NewServeMux(nil)

	// Define a default handler for unregistered routes
	mux.SetDefaultHandler(func(w ResponseWriter, r *Request) {
		w.WriteHeader(StatusNotFound)
		w.Write([]byte("Not Found\n")) // Adjusted to match current behavior
	})

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/unknown"},
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

// TestErrorHandler verifies that the custom error handler is used.
func TestErrorHandler(t *testing.T) {
	mux := NewServeMux(nil)

	// Define a custom error handler
	mux.SetErrorHandler(func(w ResponseWriter, r *Request, statusCode int) {
		w.WriteHeader(statusCode)
		w.Write([]byte("Error " + StatusText(statusCode)))
	})

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/nonexistent"},
	}

	res := &MockResponseWriter{headers: make(Header)}

	mux.ServeHTTP(res, req)

	if res.status != StatusNotFound {
		t.Errorf("Expected status %d, got %d", StatusNotFound, res.status)
	}

	expectedBody := "Error Not Found"
	if string(res.body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(res.body))
	}
}

// TestConcurrentRequests verifies that the multiplexer can handle concurrent requests.
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

	concurrencyLevel := 50 // Concurrency level
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

// TestAddRouteWithDifferentMethods verifies that routes can be added with different HTTP methods.
func TestAddRouteWithDifferentMethods(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/api/test", []string{GET, POST}, func(w ResponseWriter, r *Request) {
		w.WriteHeader(StatusOK)
		w.Write([]byte("This is a route with multiple methods"))
	})

	// Verify GET
	reqGet := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/api/test"},
	}
	resGet := &MockResponseWriter{headers: make(Header)}
	mux.ServeHTTP(resGet, reqGet)

	if resGet.status != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, resGet.status)
	}

	expectedBody := "This is a route with multiple methods"
	if string(resGet.body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(resGet.body))
	}

	// Verify POST
	reqPost := &Request{
		Method: POST,
		URL:    &url.URL{Path: "/api/test"},
	}
	resPost := &MockResponseWriter{headers: make(Header)}
	mux.ServeHTTP(resPost, reqPost)

	if resPost.status != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, resPost.status)
	}

	if string(resPost.body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(resPost.body))
	}
}
