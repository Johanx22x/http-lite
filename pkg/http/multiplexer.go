package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// RouteNode represents a node in the route tree.
type RouteNode struct {
	pathSegment string
	handler     map[string]func(ResponseWriter, *Request) // Method to handler mapping
	children    sync.Map                                  // Use sync.Map for thread safety
	isDynamic   bool                                      // True if the segment represents a dynamic value like :id
}

// ServeMux is an HTTP request multiplexer with a route tree.
type ServeMux struct {
	staticDir      *string
	root           *RouteNode
	middleware     []Middleware
	defaultHandler func(ResponseWriter, *Request)
	errorHandler   func(ResponseWriter, *Request, int) // Custom error handler
}

// NewServeMux creates a new ServeMux with a root node.
func NewServeMux(staticDir *string) *ServeMux {
	return &ServeMux{
		root: &RouteNode{
			children: sync.Map{},
			handler:  make(map[string]func(ResponseWriter, *Request)),
		},
		staticDir:  staticDir,
		middleware: []Middleware{},
	}
}

// SetStaticDir establece el directorio estático para el ServeMux.
func (mux *ServeMux) SetStaticDir(staticDir string) {
	mux.staticDir = &staticDir
}

// getOrCreateChild fetches or creates a child node.
func (mux *ServeMux) getOrCreateChild(node *RouteNode, segment string) *RouteNode {
	child, exists := mux.getChild(node, segment)
	if !exists {
		child = &RouteNode{
			pathSegment: segment,
			handler:     make(map[string]func(ResponseWriter, *Request)),
			children:    sync.Map{},
		}
		node.children.Store(segment, child)
	}
	return child
}

// getChild retrieves a child node.
func (mux *ServeMux) getChild(node *RouteNode, segment string) (*RouteNode, bool) {
	if child, exists := node.children.Load(segment); exists {
		return child.(*RouteNode), true
	}
	return nil, false
}

// applyMiddleware applies all middleware in sequence.
func (mux *ServeMux) applyMiddleware(handler func(ResponseWriter, *Request)) func(ResponseWriter, *Request) {
	for _, mw := range mux.middleware {
		handler = mw(handler)
	}
	return handler
}

// traverseTree traverses the route tree to find the handler for the given path and method.
func (mux *ServeMux) traverseTree(path, method string, node *RouteNode, params map[string]string) (func(ResponseWriter, *Request), bool) {
	segments := strings.Split(path, "/")[1:] // Split the path by "/"

	for _, segment := range segments {
		child, exists := mux.getChild(node, segment)

		if !exists {
			// Handle dynamic segment
			dynamicChild, dynamicExists := mux.getDynamicChild(node)
			if dynamicExists {
				dynamicKey := strings.TrimPrefix(dynamicChild.pathSegment, ":") // Get the actual name of the dynamic param
				params[dynamicKey] = segment                                    // Store the dynamic value in params with the correct key
				node = dynamicChild
				continue
			}
			return nil, false // No match found
		}

		node = child // Traverse to the next node
	}

	// Check if the node has a handler for the given method
	if handler, exists := node.handler[method]; exists {
		return handler, true
	}

	return nil, false // No handler found for the method
}

// getDynamicChild retrieves a dynamic child node, if it exists.
func (mux *ServeMux) getDynamicChild(node *RouteNode) (*RouteNode, bool) {
	// Iterate over children to find a dynamic route (starts with ":")
	var dynamicChild *RouteNode
	node.children.Range(func(key, value interface{}) bool {
		child := value.(*RouteNode)
		if strings.HasPrefix(child.pathSegment, ":") {
			dynamicChild = child
			return false // Stop iteration
		}
		return true // Continue iteration
	})
	return dynamicChild, dynamicChild != nil
}

// AddRoute adds a route and method(s) to the tree.
func (mux *ServeMux) AddRoute(pattern string, methods []string, handler func(ResponseWriter, *Request)) {
	segments := strings.Split(pattern, "/")[1:] // Split the pattern by "/" and ignore the first empty segment
	currentNode := mux.root

	for _, segment := range segments {
		isDynamic := strings.HasPrefix(segment, ":")
		var childNode *RouteNode

		// Retrieve existing or create new node
		if isDynamic {
			childNode = mux.getOrCreateChild(currentNode, segment)
			childNode.isDynamic = true
		} else {
			childNode = mux.getOrCreateChild(currentNode, segment)
		}
		currentNode = childNode
	}

	// Add the handler for each specified HTTP method
	for _, method := range methods {
		currentNode.handler[method] = handler
	}
}

// Handle asigna un manejador a la ruta especificada para todos los métodos HTTP.
func (mux *ServeMux) Handle(pattern string, handler func(ResponseWriter, *Request)) {
	// Aplicar middleware al manejador
	for _, mw := range mux.middleware {
		handler = mw(handler)
	}

	// Asignar la ruta utilizando todos los métodos HTTP
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	mux.AddRoute(pattern, methods, handler)
}

// ServeHTTP dispatches the request to the appropriate handler by traversing the route tree.
func (mux *ServeMux) ServeHTTP(w ResponseWriter, r *Request) {
	if mux.staticDir != nil && mux.serveStaticFile(w, r) {
		return
	}

	params := make(map[string]string)
	handler, found := mux.traverseTree(r.URL.Path, r.Method, mux.root, params)

	if !found {
		if mux.errorHandler != nil {
			mux.errorHandler(w, r, http.StatusNotFound)
		} else {
			mux.defaultErrorHandler(w, r, http.StatusNotFound)
		}
		return
	}

	// Set the params in the request
	r.Params = params

	// Apply middleware
	handler = mux.applyMiddleware(handler)

	handler(w, r)
}

// SetDefaultHandler sets a default handler for unregistered routes.
func (mux *ServeMux) SetDefaultHandler(handler func(ResponseWriter, *Request)) {
	mux.defaultHandler = handler
}

// SetErrorHandler sets a custom error handler.
func (mux *ServeMux) SetErrorHandler(handler func(ResponseWriter, *Request, int)) {
	mux.errorHandler = handler
}

// Use registers middleware to be applied to all routes.
func (mux *ServeMux) Use(mw Middleware) {
	mux.middleware = append(mux.middleware, mw)
}

// LoggingMiddleware is a simple middleware that logs the request.
func LoggingMiddleware(next func(ResponseWriter, *Request)) func(ResponseWriter, *Request) {
	return func(w ResponseWriter, r *Request) {
		// Log the request
		fmt.Printf("Received request: %s %s\n", r.Method, r.URL.Path)
		next(w, r) // Call the next handler
	}
}

// defaultErrorHandler is the default error response for 404 Not Found.
func (mux *ServeMux) defaultErrorHandler(w ResponseWriter, _ *Request, statusCode int) {
	w.WriteHeader(statusCode)
	switch statusCode {
	case http.StatusNotFound:
		fmt.Fprintln(w, StatusText(http.StatusNotFound))
	default:
		fmt.Fprintln(w, "Error:", statusCode)
	}
}

// FileExists checks if a file or directory exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err) // Return true if no error (file exists)
}

func (mux *ServeMux) serveStaticFile(w ResponseWriter, r *Request) bool {
	// Check if a static directory is set
	if mux.staticDir == nil {
		return false
	}

	// Get the file path from the URL
	filePath := (*mux.staticDir) + r.URL.Path

	// When the URL ends with a "/", serve the index.html file
	if strings.HasSuffix(r.URL.Path, "/") {
		filePath += "index.html"
	}

	// Check if the file exists
	if !fileExists(filePath) {
		return false
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false
	}

	w.Header()["Content-Type"] = []string{detectContentType(filePath)}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return true
}

// detectContentType returns the content type based on the file data.
func detectContentType(filePath string) string {
	// Map of file extensions to content types
	contentTypes := map[string]string{
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".svg":  "image/svg+xml",
		".gif":  "image/gif",
	}

	// Get the file extension
	ext := strings.ToLower(filepath.Ext(filePath))

	// Lookup the content type
	if contentType, exists := contentTypes[ext]; exists {
		return contentType
	}

	// Default to binary data
	return "application/octet-stream"
}
