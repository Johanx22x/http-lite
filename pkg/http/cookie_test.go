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

	// Simular la solicitud
	mux.ServeHTTP(res, req)

	// Verificar que la cookie se ha enviado en la respuesta
	setCookieHeader := res.Header()["Set-Cookie"]
	if len(setCookieHeader) == 0 {
		t.Errorf("Expected a Set-Cookie header")
	}

	// Verificar el valor de la cookie
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

	// Simular que el cliente guarda la cookie y la envía en la siguiente solicitud
	req.Cookies = append(req.Cookies, Cookie{Name: "session_id", Value: "abc123"})

	// Recuperar la cookie del objeto Request
	cookieValue, err := req.GetCookie("session_id")
	if err != nil || cookieValue.Value != "abc123" {
		t.Errorf("Expected session_id=abc123, got %v", cookieValue)
	}
}

// Test deleting a cookie
func TestDeleteCookie(t *testing.T) {
	mux := NewServeMux(nil)

	// Ruta para establecer la cookie
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

	// Simula la solicitud para establecer la cookie
	reqSet := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/set-cookie"},
		Header: make(Header),
	}

	resSet := &MockResponseWriter{headers: make(Header)}
	mux.ServeHTTP(resSet, reqSet)

	// Verificar que la cookie se ha enviado en la respuesta
	if len(resSet.Header()["Set-Cookie"]) == 0 {
		t.Errorf("Expected a Set-Cookie header")
	}

	// Ahora vamos a eliminar la cookie
	mux.AddRoute("/delete-cookie", []string{GET}, func(w ResponseWriter, r *Request) {
		w.DeleteCookie("session_id")
		w.WriteHeader(StatusOK)
	})

	// Simula la solicitud para eliminar la cookie
	reqDel := &Request{
		Method: GET,
		URL:    &url.URL{Path: "/delete-cookie"},
		Header: make(Header),
	}

	resDel := &MockResponseWriter{headers: make(Header)}
	mux.ServeHTTP(resDel, reqDel)

	// Verificar que se haya establecido correctamente la cookie de eliminación
	setCookieHeader := resDel.Header()["Set-Cookie"]
	if len(setCookieHeader) == 0 {
		t.Errorf("Expected Set-Cookie header to be present")
	}

	// Verifica que el encabezado Set-Cookie contenga la información correcta
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

	// Intentar recuperar una cookie que no existe
	_, err := req.GetCookie("non_existent_cookie")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if err != ErrCookieNotFound {
		t.Errorf("Expected ErrCookieNotFound, got %v", err)
	}
}
