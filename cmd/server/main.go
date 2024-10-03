package main

import (
	"flag"
	"log"
	"os"
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

	// Routes
	mux.AddRoute("/api/test", []string{http.GET},
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello, World!"))
		},
	)

	mux.AddRoute("/api/set-cookie", []string{http.POST},
		func(w http.ResponseWriter, r *http.Request) {
			token := "abc123" // Aquí deberías generar tu token de sesión
			cookie := &http.Cookie{
				Name:     "session_token",
				Value:    token,
				Path:     "/",
				Expires:  time.Now().Add(1 * time.Hour),
				HttpOnly: true,
				Secure:   false, // Cambia a true si usas HTTPS
			}
			err := http.SetCookie(w, cookie)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error setting cookie"))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Cookie set"))
		},
	)

	mux.AddRoute("/api/delete-cookie", []string{http.DELETE},
		func(w http.ResponseWriter, r *http.Request) {
			http.DeleteCookie(w, "session_token", "/")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Cookie deleted"))
		},
	)

	// Start server
	address := ":" + port
	err := http.Run(address, mux)
	if err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
		os.Exit(1)
	}
}
