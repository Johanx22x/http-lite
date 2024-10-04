package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

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

			// Set the rate in a cookie
			w.SetCookie(&http.Cookie{Name: "last-rate", Value: strconv.Itoa(rate), Expires: time.Now().Add(24 * time.Hour)})

			// Write the response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		},
	)

	// Start server
	err := http.Run(":"+port, mux)
	if err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
		os.Exit(1)
	}
}
