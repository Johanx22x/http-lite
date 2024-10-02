package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
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

// SetCookie writes the Set-Cookie header in the response to set a cookie.
func SetCookie(w ResponseWriter, cookie *Cookie) error {
	if cookie.Name == "" {
		return errors.New("cookie name cannot be empty")
	}

	cookieString := cookieToString(cookie)
	headers := w.Header()
	headers["Set-Cookie"] = append(headers["Set-Cookie"], cookieString)
	return nil
}

// cookieToString converts a Cookie struct to a string following the cookie format.
func cookieToString(cookie *Cookie) string {
	cookieString := cookie.Name + "=" + cookie.Value

	if cookie.Path != "" {
		cookieString += "; Path=" + cookie.Path
	}

	if cookie.Domain != "" {
		cookieString += "; Domain=" + cookie.Domain
	}

	if !cookie.Expires.IsZero() {
		cookieString += "; Expires=" + cookie.Expires.UTC().Format(http.TimeFormat)
	}

	if cookie.MaxAge > 0 {
		cookieString += "; Max-Age=" + fmt.Sprint(cookie.MaxAge)
	}

	if cookie.Secure {
		cookieString += "; Secure"
	}

	if cookie.HttpOnly {
		cookieString += "; HttpOnly"
	}

	return cookieString
}

// GetCookie retrieves the value of a cookie from the request.
func GetCookie(r *Request, name string) (*Cookie, error) {
	for _, cookie := range r.Header["Cookie"] {
		if cookieValue, err := parseCookie(cookie, name); err == nil {
			return cookieValue, nil
		} else if err == http.ErrNoCookie {
			continue
		} else {
			return nil, err
		}
	}
	return nil, http.ErrNoCookie
}

// parseCookie parses the raw cookie header to find the cookie by name.
func parseCookie(rawCookie string, name string) (*Cookie, error) {
	cookieParts := strings.Split(rawCookie, "; ")
	for _, part := range cookieParts {
		if strings.HasPrefix(part, name+"=") {
			value := strings.TrimPrefix(part, name+"=")
			return &Cookie{Name: name, Value: value}, nil
		}
	}
	return nil, http.ErrNoCookie
}

// DeleteCookie sets a cookie with an expiration in the past, effectively deleting it.
func DeleteCookie(w ResponseWriter, name string, path string) error {
	expiredCookie := &Cookie{
		Name:    name,
		Value:   "",
		Path:    path,
		Expires: time.Unix(0, 0),
		MaxAge:  -1, // Indicates to delete the cookie
	}
	return SetCookie(w, expiredCookie)
}
