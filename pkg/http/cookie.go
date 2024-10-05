package http

import (
	"strconv"
	"time"
)

// Cookie represents an HTTP cookie.
type Cookie struct {
	Name     string
	Value    string
	Path     string
	Domain   string
	Expires  time.Time
	MaxAge   int
	Secure   bool
	HttpOnly bool
}

// String returns a string representation of the cookie.
func (c *Cookie) String() string {
	cookieStr := c.Name + "=" + c.Value
	if c.Path != "" {
		cookieStr += "; Path=" + c.Path
	}
	if c.Domain != "" {
		cookieStr += "; Domain=" + c.Domain
	}
	if !c.Expires.IsZero() {
		cookieStr += "; Expires=" + c.Expires.Format(time.RFC1123)
	}
	if c.MaxAge > 0 {
		cookieStr += "; Max-Age=" + strconv.Itoa(c.MaxAge)
	}
	if c.Secure {
		cookieStr += "; Secure"
	}
	if c.HttpOnly {
		cookieStr += "; HttpOnly"
	}
	return cookieStr
}
