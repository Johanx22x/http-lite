package http

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// HandlerFunc is a function that handles an HTTP request.
type HandlerFunc func(ResponseWriter, *Request)

// ServeHTTP calls f(w, r).
// It's used to satisfy the Handler interface.
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}

// Middleware is a function that wraps an HTTP handler.
type Middleware func(func(ResponseWriter, *Request)) func(ResponseWriter, *Request)

type Server struct {
	Addr    string
	Handler Handler
	mu      sync.Mutex
	wg      sync.WaitGroup
}

// NewServer creates a new HTTP server with the given address and handler.
func NewServer(addr string, handler Handler) *Server {
	return &Server{
		Addr:    addr,
		Handler: handler,
	}
}

// parseRequest reads and parses an HTTP request from a connection.
func parseRequest(ctx context.Context, conn net.Conn) (*Request, error) {
	reader := bufio.NewReader(conn)

	// Create a channel to signal when the request parsing is done
	done := make(chan struct{})
	var req *Request
	var err error

	go func() {
		defer close(done)
		req, err = parseRequestWithTimeout(reader)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
		return req, err
	}
}

// parseRequestWithTimeout reads and parses an HTTP request from a connection with a timeout.
func parseRequestWithTimeout(reader *bufio.Reader) (*Request, error) {
	// Read the request line (e.g., "GET /path HTTP/1.1")
	line, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return nil, err
		}

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

	// The request body is the remaining data in the reader
	// Convert the reader to an io.ReadCloser
	body := io.NopCloser(reader)

	return &Request{
		Method:  method,
		URL:     parsedURL,
		Proto:   proto,
		Header:  headers,
		Cookies: cookies,
		Body:    body,
	}, nil
}

// parseCookies parses a cookie header string and returns a slice of cookies.
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

// handleConn reads and parses an HTTP request from a connection and calls the handler.
func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	s.wg.Add(1)
	defer s.wg.Done()

	req, err := parseRequest(ctx, conn)
	if err != nil {
		if err == io.EOF {
			return
		}

		fmt.Println("Error parsing request:", err)
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n\r\n", http.StatusBadRequest, http.StatusText(http.StatusBadRequest))))
		return
	}

	// Create a ResponseWriter tied to the current connection
	res := NewResponseWriter(conn)

	// Pass the ResponseWriter and Request to the handler
	s.Handler.ServeHTTP(res, req)
}

// listenAndServe listens on the TCP network address and handles incoming connections.
func (s *Server) listenAndServe() error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		go func() {
			defer cancel()
			s.handleConn(ctx, conn)
		}()
	}
}

// Shutdown gracefully closes the server and waits for ongoing connections to finish
func (s *Server) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Println("Shutting down server...")
	s.wg.Wait() // Wait for all connections to finish
}

// handleSignals listens for SIGINT and SIGTERM signals to gracefully shutdown the server
func (s *Server) handleSignals(quit chan os.Signal) {
	<-quit
	s.Shutdown()
	os.Exit(0)
}

// Run starts an HTTP server with the given address and handler.
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

// Error writes an HTTP error response with the given message and status code.
func Error(w ResponseWriter, m string, statusCode int) {
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, m)
}
