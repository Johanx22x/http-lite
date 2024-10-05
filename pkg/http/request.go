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

// GetCookie returns a cookie by name.
func (r *Request) GetCookie(name string) (*Cookie, error) {
	for _, cookie := range r.Cookies {
		if cookie.Name == name {
			return &cookie, nil
		}
	}
	return nil, ErrCookieNotFound
}
