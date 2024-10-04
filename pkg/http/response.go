package http

import (
	"fmt"
	"net"
)

// Response represents the structure of an HTTP response.
type Response struct {
	StatusCode  int
	Proto       string
	Headers     Header
	Body        []byte
	conn        net.Conn
	headersSent bool
}

// ResponseWriter is an interface for writing an HTTP response.
type ResponseWriter interface {
	Header() Header
	Write([]byte) (int, error)
	WriteHeader(int)
	SetCookie(*Cookie)
	DeleteCookie(string)
}

// Write writes the data to the connection as part of an HTTP reply.
func (r *Response) Write(data []byte) (int, error) {
	if !r.headersSent {
		// If headers haven't been sent yet, send the headers first
		r.WriteHeader(r.StatusCode)
	}

	// Write the body data to the connection
	return r.conn.Write(data)
}

// WriteHeader sends an HTTP response header with the provided status code.
func (r *Response) WriteHeader(statusCode int) {
	if r.headersSent {
		return
	}
	r.StatusCode = statusCode

	// Write the status line and headers
	statusText := StatusText(statusCode)
	headerStr := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, statusText)
	for k, v := range r.Headers {
		headerStr += fmt.Sprintf("%s: %s\r\n", k, v[0])
	}
	headerStr += "\r\n" // End of headers

	// Write headers to the connection
	r.conn.Write([]byte(headerStr))
	r.headersSent = true
}

// Header returns the response headers.
func (r *Response) Header() Header {
	return r.Headers
}

// SetCookie adds a cookie to the response headers.
func (r *Response) SetCookie(c *Cookie) {
	r.Headers.Set("Set-Cookie", c.String())
}

// DeleteCookie deletes a cookie from the response headers.
func (r *Response) DeleteCookie(name string) {
	c := &Cookie{Name: name, Value: "", MaxAge: -1}
	r.Headers.Set("Set-Cookie", c.String())
}

// NewResponseWriter creates a new ResponseWriter.
func NewResponseWriter(conn net.Conn) ResponseWriter {
	return &Response{
		Proto:   "HTTP/1.1",
		Headers: make(Header),
		conn:    conn,
	}
}
