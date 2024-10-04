package http

import (
	"io"
	"net/url"
)

// Request represents an HTTP request.
type Request struct {
	Method  string
	URL     *url.URL
	Params  map[string]string
	Proto   string
	Header  Header
	Body    io.ReadCloser
	Cookies []Cookie
}
