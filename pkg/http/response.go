package http

type Response struct {
	StatusCode int
	Proto      string
	Header     Header
	Body       []byte
}

type ResponseWriter interface {
	// Write writes the data to the connection as part of an HTTP reply.
	Write([]byte) (int, error)

	// WriteHeader sends an HTTP response header with the provided status code.
	WriteHeader(statusCode int)
}
