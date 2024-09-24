package main

import (
	"flag"

	"github.com/Johanx22x/http-lite/cmd/server/middleware"
	"github.com/Johanx22x/http-lite/pkg/http"
)

var port string

func init() {
	flag.StringVar(&port, "port", "8080", "Port to listen on")
}

func main() {
	// Parse flags
	flag.Parse()

	dir := "./cmd/server/website"
	mux := http.NewServeMux(&dir)

	mux.Use(http.LoggingMiddleware)
	mux.Use(middleware.CORS)

	// Routes
	mux.AddRoute("/api/test", []string{http.GET},
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello, World!"))
		},
	)

	// Start server
	err := http.Run(":"+port, mux)
	if err != nil {
		panic(err)
	}
}
