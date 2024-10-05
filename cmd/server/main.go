package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
	"encoding/json"

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

	// Login endpoint (showing how to get a parameter from the URL)
	mux.AddRoute("/api/login/:id", []string{http.POST},
		func(w http.ResponseWriter, r *http.Request) {
			// Get the ID from the URL
			id := r.Params["id"]

			// Write the response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": "` + id + `"}`))
		},
	)

	// Delete method endpoint
	mux.AddRoute("/api/delete", []string{http.DELETE},
		func(w http.ResponseWriter, r *http.Request) {
			body := r.Body
			defer body.Close()

			// Check if content type is application/json
			if r.Header.Get("Content-Type") != "application/json" {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"error": "Invalid content type"}`))
				return
			}

			// Read the body
			buf := make([]byte, 1024)
			n, err := body.Read(buf)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"error": "Internal server error"}`))
				return
			}

			// Check if the body is empty
			if n == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"error": "Empty body"}`))
				return
			}

			// Unmarshal the JSON
			var data map[string]interface{}
			err = json.Unmarshal(buf[:n], &data)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"error": "Invalid JSON"}`))
				return
			}

			// Check if the ID is present
			if _, ok := data["id"]; !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"error": "ID not found"}`))
				return
			}

			id, ok := data["id"].(string)	
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"error": "Invalid ID"}`))
				return
			}

			// Write the response
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`Deleted ID: ` + id))
		},
	)

	// Put method endpoint
	mux.AddRoute("/api/update/:id", []string{http.PUT},
		func(w http.ResponseWriter, r *http.Request) {
			// Get the ID from the URL
			id, err := strconv.Atoi(r.Params["id"])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "Invalid ID"}`))
				return
			}
	
			newID := strconv.Itoa(rand.Intn(1000) + id)

			// Write the response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": "` + newID + `"}`))
		},
	)

	// Start server
	err := http.Run(":"+port, mux)
	if err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
		os.Exit(1)
	}
}
