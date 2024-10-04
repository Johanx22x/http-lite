package http

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type HandlerFunc func(ResponseWriter, *Request)

type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}

// Middleware is a function that wraps an HTTP handler.
type Middleware func(func(ResponseWriter, *Request)) func(ResponseWriter, *Request)

type Server struct {
	Addr    string
	Handler Handler
}

func NewServer(addr string, handler Handler) *Server {
	return &Server{
		Addr:    addr,
		Handler: handler,
	}
}

// parseRequest parses the incoming HTTP request from a net.Conn and returns a Request object.
func parseRequest(buffer []byte) (*Request, error) {
	reader := bufio.NewReader(strings.NewReader(string(buffer)))

	// Read the request line (e.g., "GET /path HTTP/1.1")
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read request line: %w", err)
	}

	// Parse the request line
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil, fmt.Errorf("malformed request line")
	}

	method := parts[0]
	rawURL := parts[1]
	proto := parts[2]

	// XXX: Currently only support HTTP/1.1
	if proto != "HTTP/1.1" {
		return nil, fmt.Errorf("unsupported protocol: %s", proto)
	}

	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Parse headers
	headers := make(Header)
	var cookies []Cookie
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read header: %w", err)
		}

		// An empty line marks the end of headers
		if line == "\r\n" {
			break
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed header line")
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		headers[key] = append(headers[key], value)

		if key == "Cookie" {
			cookies = append(cookies, parseCookies(value)...)
		}
	}

	return &Request{
		Method:  method,
		URL:     parsedURL,
		Proto:   proto,
		Header:  headers,
		Cookies: cookies,
	}, nil
}

func parseCookies(cookieHeader string) []Cookie {
	var cookies []Cookie
	parts := strings.Split(cookieHeader, ";")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) == 2 {
			cookies = append(cookies, Cookie{Name: kv[0], Value: kv[1]})
		}
	}
	return cookies
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	// Read request
	const size = 64 << 10
	buffer := make([]byte, size)
	n, err := conn.Read(buffer)
	if err != nil && err != io.EOF {
		fmt.Println("Error reading from connection:", err)
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n\r\n", http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))))
		return
	}

	// Trim buffer to actual size
	buffer = buffer[:n]

	// Parse request
	if n > 0 {
		req, err := parseRequest(buffer)
		if err != nil {
			fmt.Println("Error parsing request:", err)
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n\r\n", http.StatusBadRequest, http.StatusText(http.StatusBadRequest))))
			return
		}

		// Create a ResponseWriter tied to the current connection
		res := NewResponseWriter(conn)

		// Pass the ResponseWriter and Request to the handler
		s.Handler.ServeHTTP(res, req)
	} else {
		// Bad request
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n\r\n", http.StatusBadRequest, http.StatusText(http.StatusBadRequest))))
	}
}

func (s *Server) listenAndServe() error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleSignals(quit chan os.Signal) {
	<-quit
	fmt.Println("Shutting down server...")
	os.Exit(0)
}

func Run(addr string, handler Handler) error {
	server := NewServer(addr, handler)

	// Set up signal catching for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go server.handleSignals(quit)

	// Start server
	fmt.Println("Server listening on", addr)
	return server.listenAndServe()
}

func Error(w ResponseWriter, m string, statusCode int) {
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, m)
}
