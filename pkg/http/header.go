package http

// Header represents an HTTP header.
type Header map[string][]string

// Set sets a header field.
func (h Header) Set(key, value string) {
	h[key] = append(h[key], value)
}

// Get returns a header field.
func (h Header) Get(key string) string {
	if values, ok := h[key]; ok {
		return values[0]
	}
	return ""
}
