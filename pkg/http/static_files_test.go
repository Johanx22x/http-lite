package http

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

// TestServeStaticFile verifies that the server can serve a static file.
func TestServeStaticFile(t *testing.T) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile.html")
	content := []byte("<html><body>Hello, Static World!</body></html>")

	err := ioutil.WriteFile(tmpFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary static file: %v", err)
	}
	defer os.Remove(tmpFile) // Clean up the file after the test

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

// TestServeStaticFileNotFound verifies that a 404 is returned if the file does not exist.
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

// TestServeIndexFile verifies that the server serves index.html when accessing a root path.
func TestServeIndexFile(t *testing.T) {
	tmpDir := os.TempDir()
	indexFile := filepath.Join(tmpDir, "index.html")
	content := []byte("<html><body>Welcome to the index!</body></html>")

	err := ioutil.WriteFile(indexFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary index file: %v", err)
	}
	defer os.Remove(indexFile) // Clean up the file after the test

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

// TestServeStaticFileWithCustomExtension verifies that files with custom extensions are served.
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

// TestServeStaticFileWithError verifies error handling when trying to read a non-existent file.
func TestServeStaticFileWithError(t *testing.T) {
	mux := NewServeMux(nil) // No static directory configured

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

// TestServeEmptyStaticDir verifies that the server correctly handles an empty directory.
func TestServeEmptyStaticDir(t *testing.T) {
	tmpDir := os.TempDir() // Use the empty temporary directory

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
