package main

import (
	"github.com/Johanx22x/http-lite/pkg/http"
)

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", helloWorld)

	err := http.Run(":8080", mux)
	if err != nil {
		panic(err)
	}
}
