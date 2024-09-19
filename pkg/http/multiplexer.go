package http

// ServeMux is an HTTP request multiplexer.
type ServeMux struct {
	handlers map[string]func(ResponseWriter, *Request)
}

// NewServeMux returns a new ServeMux.
func NewServeMux() *ServeMux {
	return &ServeMux{
		handlers: make(map[string]func(ResponseWriter, *Request)),
	}
}

// Handle registers the handler for the given pattern.
func (mux *ServeMux) HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
	mux.handlers[pattern] = handler
}

// ServeHTTP dispatches the request to the handler whose pattern most
// closely matches the request URL.
func (mux *ServeMux) ServeHTTP(w ResponseWriter, r *Request) {
	handler, ok := mux.handlers[r.URL.Path]
	if !ok {
		w.WriteHeader(404)
		return
	}

	handler(w, r)
}
