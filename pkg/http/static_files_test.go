package http

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

// TestServeStaticFile verifica que el servidor puede servir un archivo estático.
func TestServeStaticFile(t *testing.T) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile.html")
	content := []byte("<html><body>Hello, Static World!</body></html>")

	err := ioutil.WriteFile(tmpFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary static file: %v", err)
	}
	defer os.Remove(tmpFile) // Limpia el archivo después de la prueba

	mux := NewServeMux(&tmpDir)

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/testfile.html"},
	}

	res := &MockResponseWriter{headers: make(Header)}

	mux.ServeHTTP(res, req)

	if res.status != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, res.status)
	}

	if string(res.body) != string(content) {
		t.Errorf("Expected body '%s', got '%s'", string(content), string(res.body))
	}

	expectedContentType := "text/html"
	actualContentType := res.Header().Get("Content-Type")
	if actualContentType != expectedContentType {
		t.Errorf("Expected Content-Type '%s', got '%s'", expectedContentType, actualContentType)
	}
}

// TestServeStaticFileNotFound verifica que se devuelve un 404 si el archivo no existe.
func TestServeStaticFileNotFound(t *testing.T) {
	tmpDir := os.TempDir()

	mux := NewServeMux(&tmpDir)

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/nonexistentfile.html"},
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

// TestServeIndexFile verifica que el servidor sirve index.html cuando se accede a una ruta raíz.
func TestServeIndexFile(t *testing.T) {
	tmpDir := os.TempDir()
	indexFile := filepath.Join(tmpDir, "index.html")
	content := []byte("<html><body>Welcome to the index!</body></html>")

	err := ioutil.WriteFile(indexFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary index file: %v", err)
	}
	defer os.Remove(indexFile) // Limpia el archivo después de la prueba

	mux := NewServeMux(&tmpDir)

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/"},
	}

	res := &MockResponseWriter{headers: make(Header)}

	mux.ServeHTTP(res, req)

	if res.status != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, res.status)
	}

	if string(res.body) != string(content) {
		t.Errorf("Expected body '%s', got '%s'", string(content), string(res.body))
	}
}

// TestServeStaticFileWithCustomExtension verifica que se sirvan archivos con extensiones personalizadas.
func TestServeStaticFileWithCustomExtension(t *testing.T) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "customfile.xyz")
	content := []byte("This is a custom file")

	err := ioutil.WriteFile(tmpFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary custom file: %v", err)
	}
	defer os.Remove(tmpFile)

	mux := NewServeMux(&tmpDir)

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/customfile.xyz"},
	}

	res := &MockResponseWriter{headers: make(Header)}

	mux.ServeHTTP(res, req)

	if res.status != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, res.status)
	}

	if string(res.body) != string(content) {
		t.Errorf("Expected body '%s', got '%s'", string(content), string(res.body))
	}

	expectedContentType := "application/octet-stream"
	actualContentType := res.Header().Get("Content-Type")
	if actualContentType != expectedContentType {
		t.Errorf("Expected Content-Type '%s', got '%s'", expectedContentType, actualContentType)
	}
}

// TestServeStaticFileWithError verifica el manejo de errores al intentar leer un archivo inexistente.
func TestServeStaticFileWithError(t *testing.T) {
	mux := NewServeMux(nil) // No se configura ningún directorio estático

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/testfile.html"},
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

// TestServeEmptyStaticDir verifica que el servidor maneje correctamente un directorio vacío.
func TestServeEmptyStaticDir(t *testing.T) {
	tmpDir := os.TempDir() // Usamos el directorio temporal vacío

	mux := NewServeMux(&tmpDir)

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/emptyfile.html"},
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
