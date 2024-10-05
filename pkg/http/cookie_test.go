package http

import (
	"net/url"
	"testing"
	"time"
)

// Test setting a cookie
func TestSetCookie(t *testing.T) {
	mux := NewServeMux(nil)

	mux.AddRoute("/set-cookie", []string{GET}, func(w ResponseWriter, r *Request) {
		cookie := &Cookie{
			Name:     "session_id",
			Value:    "abc123",
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		}
		w.SetCookie(cookie)
		w.WriteHeader(StatusOK)
	})

	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/set-cookie"},
		Header: make(Header),
	}

	res := &MockResponseWriter{headers: make(Header)}

	// Simulate the request
	mux.ServeHTTP(res, req)

	// Verify that the cookie was set in the response
	setCookieHeader := res.Header()["Set-Cookie"]
	if len(setCookieHeader) == 0 {
		t.Errorf("Expected a Set-Cookie header")
	}

	// Verify the value of the cookie
	expected := "session_id=abc123"
	if setCookieHeader[0][:len(expected)] != expected {
		t.Errorf("Expected Set-Cookie to contain '%s', but got '%s'", expected, setCookieHeader[0])
	}
}

// Test getting a cookie
func TestGetCookie(t *testing.T) {
	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/get-cookie"},
		Header: make(Header),
	}

	// Simulate that the client saves the cookie and sends it in the next request
	req.Cookies = append(req.Cookies, Cookie{Name: "session_id", Value: "abc123"})

	// Retrieve the cookie from the Request object
	cookieValue, err := req.GetCookie("session_id")
	if err != nil || cookieValue.Value != "abc123" {
		t.Errorf("Expected session_id=abc123, got %v", cookieValue)
	}
}

// Test deleting a cookie
func TestDeleteCookie(t *testing.T) {
	mux := NewServeMux(nil)

	// Route to set the cookie
	mux.AddRoute("/set-cookie", []string{GET}, func(w ResponseWriter, r *Request) {
		cookie := &Cookie{
			Name:     "session_id",
			Value:    "abc123",
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		}
		w.SetCookie(cookie)
		w.WriteHeader(StatusOK)
	})

	// Simulate the request to set the cookie
	reqSet := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/set-cookie"},
		Header: make(Header),
	}

	resSet := &MockResponseWriter{headers: make(Header)}
	mux.ServeHTTP(resSet, reqSet)

	// Verify that the cookie was sent in the response
	if len(resSet.Header()["Set-Cookie"]) == 0 {
		t.Errorf("Expected a Set-Cookie header")
	}

	// Now let's delete the cookie
	mux.AddRoute("/delete-cookie", []string{GET}, func(w ResponseWriter, r *Request) {
		w.DeleteCookie("session_id")
		w.WriteHeader(StatusOK)
	})

	// Simulate the request to delete the cookie
	reqDel := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/delete-cookie"},
		Header: make(Header),
	}

	resDel := &MockResponseWriter{headers: make(Header)}
	mux.ServeHTTP(resDel, reqDel)

	// Verify that the delete cookie header was set correctly
	setCookieHeader := resDel.Header()["Set-Cookie"]
	if len(setCookieHeader) == 0 {
		t.Errorf("Expected Set-Cookie header to be present")
	}

	// Verify that the Set-Cookie header contains the correct information
	expected := "session_id="
	if setCookieHeader[0][:len(expected)] != expected {
		t.Errorf("Expected Set-Cookie to contain '%s', but got '%s'", expected, setCookieHeader[0])
	}
}

// Test trying to get a non-existent cookie
func TestGetNonExistentCookie(t *testing.T) {
	req := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/non-existent-cookie"},
		Header: make(Header),
	}

	// Attempt to retrieve a cookie that does not exist
	_, err := req.GetCookie("non_existent_cookie")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if err != ErrCookieNotFound {
		t.Errorf("Expected ErrCookieNotFound, got %v", err)
	}
}

func TestCookieStringNoOptionalFields(t *testing.T) {
	cookie := &Cookie{
		Name:  "test",
		Value: "123",
	}

	expected := "test=123"
	result := cookie.String()

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestCookieStringSomeOptionalFields(t *testing.T) {
	tests := []struct {
		name     string
		cookie   Cookie
		expected string
	}{
		{
			name: "Cookie with Path and Secure",
			cookie: Cookie{
				Name:   "test",
				Value:  "123",
				Path:   "/",
				Secure: true,
			},
			expected: "test=123; Path=/; Secure",
		},
		{
			name: "Cookie with Domain and HttpOnly",
			cookie: Cookie{
				Name:     "test",
				Value:    "123",
				Domain:   "example.com",
				HttpOnly: true,
			},
			expected: "test=123; Domain=example.com; HttpOnly",
		},
		{
			name: "Cookie with MaxAge and Expires",
			cookie: Cookie{
				Name:    "test",
				Value:   "123",
				MaxAge:  3600,
				Expires: time.Date(2024, 10, 4, 0, 0, 0, 0, time.UTC),
			},
			expected: "test=123; Expires=Fri, 04 Oct 2024 00:00:00 UTC; Max-Age=3600",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cookie.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
