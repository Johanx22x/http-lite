package http

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}

type Server struct {
	Addr    string
	Handler Handler
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewServer(addr string, handler Handler) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		Addr:    addr,
		Handler: handler,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	s.wg.Add(1)
	defer s.wg.Done()

	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	const size = 64 << 10
	buffer := make([]byte, size)

	done := make(chan error, 1)
	go func() {
		n, err := conn.Read(buffer)
		if err != nil && err != io.EOF {
			done <- fmt.Errorf("error reading from connection: %w", err)
			return
		}
		buffer = buffer[:n]
		done <- nil
	}()

	select {
	case <-ctx.Done():
		if s.ctx.Err() != nil {
			conn.Write([]byte("HTTP/1.1 503 Service Unavailable\r\n\r\n"))
			fmt.Println("Server unavailable")
		} else {
			conn.Write([]byte("HTTP/1.1 408 Request Timeout\r\n\r\n"))
			fmt.Println("Request timeout")
		}
		return
	case err := <-done:
		if err != nil {
			fmt.Println(err)
			conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
			return
		}

		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	}
}

func (s *Server) serve(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				s.wg.Wait()
				return nil
			default:
				return err
			}
		}
		go s.handleConn(conn)
	}
}

func (s *Server) listenAndServe() error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	return s.serve(ln)
}

func (s *Server) handleSignals() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("Shutting down server...")
		s.cancel()
	}()
}

func Run(addr string, handler Handler) error {
	s := NewServer(addr, handler)
	s.handleSignals()

	errChan := make(chan error, 1)
	go func() {
		errChan <- s.listenAndServe()
	}()

	select {
	case <-s.ctx.Done():
		fmt.Println("Server stopped")
		return nil
	case err := <-errChan:
		return err
	}
}
