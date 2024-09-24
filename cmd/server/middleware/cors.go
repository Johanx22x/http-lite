package middleware

import (
	"github.com/Johanx22x/http-lite/pkg/http"
)

// CORS middleware
func CORS(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header()["Access-Control-Allow-Origin"] = []string{"*"}
		w.Header()["Access-Control-Allow-Methods"] = []string{"GET, POST, PUT, DELETE, OPTIONS"}
		w.Header()["Access-Control-Allow-Headers"] = []string{"Content-Type, Authorization"}

		// Handle preflight request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}
