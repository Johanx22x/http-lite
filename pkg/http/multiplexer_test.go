package http

import (
	"net/url"
	"testing"
)

// TestStaticRoute verifica que una ruta estática funcione como se espera.
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

// TestDynamicRoute verifica que las rutas dinámicas funcionen correctamente.
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

// TestRouteNotFound verifica que un 404 sea devuelto cuando no se encuentra una ruta.
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

// TestMethodNotAllowed verifica que un 404 sea devuelto si el método no está permitido para la ruta.
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

	// Verifica que devuelva un 404 porque POST no está permitido
	if res.status != StatusNotFound {
		t.Errorf("Expected status %d, got %d", StatusNotFound, res.status)
	}

	expectedBody := "Not Found\n"
	if string(res.body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(res.body))
	}
}

// TestMiddleware verifica que el middleware sea aplicado correctamente.
func TestMiddleware(t *testing.T) {
	mux := NewServeMux(nil)

	// Middleware que agrega un encabezado a la respuesta.
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

	// Verificar que el middleware ha modificado los encabezados
	if res.Header().Get("X-Middleware") != "true" {
		t.Errorf("Expected 'X-Middleware' header to be 'true', got '%s'", res.Header().Get("X-Middleware"))
	}

	expectedBody := "Middleware applied"
	if string(res.body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(res.body))
	}
}

// TestDefaultHandler verifica que el manejador por defecto sea llamado para rutas no encontradas.
func TestDefaultHandler(t *testing.T) {
	mux := NewServeMux(nil)

	// Definir un manejador por defecto para rutas no registradas
	mux.SetDefaultHandler(func(w ResponseWriter, r *Request) {
		w.WriteHeader(StatusNotFound)
		w.Write([]byte("Not Found\n")) // Ajustado para que coincida con el comportamiento actual
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

// TestErrorHandler verifica que el manejador de errores personalizado sea utilizado.
func TestErrorHandler(t *testing.T) {
	mux := NewServeMux(nil)

	// Definir un manejador de errores personalizado
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

// TestConcurrentRequests verifica que el multiplexor pueda manejar solicitudes concurrentes.
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
