package main

import (
	"flag"
	"math/rand"
	"strconv"

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

	// US Dollar to CRC exchange rate endpoint
	mux.AddRoute("/api/exchange", []string{http.GET},
		func(w http.ResponseWriter, r *http.Request) {
			// Random rate
			rate := 550 + rand.Intn(100) - 50
			response := `{"rate": ` + strconv.Itoa(rate) + `}`

			w.WriteHeader(http.StatusOK)
			w.Header()["Content-Type"] = []string{"application/json"}
			w.Write([]byte(response))
		},
	)

	// Start server
	err := http.Run(":"+port, mux)
	if err != nil {
		panic(err)
	}
}
