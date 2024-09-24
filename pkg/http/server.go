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
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read header: %w", err)
		}

		// An empty line marks the end of headers
		if line == "\r\n" {
			break
		}

		// Parse header line (e.g., "Content-Type: text/plain")
		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			return nil, fmt.Errorf("malformed header line")
		}
		name := strings.TrimSpace(line[:colonIndex])
		value := strings.TrimSpace(line[colonIndex+1:])
		headers[name] = append(headers[name], value)
	}

	// Parse body (if there is one)
	var body io.ReadCloser = nil
	if method == "POST" || method == "PUT" {
		// Pass the reader itself as the body to be read later
		body = io.NopCloser(reader)
	}

	// Construct and return the request
	return &Request{
		Method: method,
		URL:    parsedURL,
		Proto:  proto,
		Header: headers,
		Body:   body,
	}, nil
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
