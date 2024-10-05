package http

import "errors"

// ErrCookieNotFound is returned when a cookie is not found.
var ErrCookieNotFound = errors.New("cookie not found")
