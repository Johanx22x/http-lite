package main

import (
	"fmt"
	"net/http"
	"sync"
)

func sendRequest(wg *sync.WaitGroup, url string) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error al hacer la solicitud a %s: %s\n", url, err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Respuesta de %s: %s\n", url, resp.Status)
}

func main() {
	var wg sync.WaitGroup
	routes := []string{
		"http://localhost:8080/api/exchange",
	}

	// Definir la cantidad de solicitudes por ruta
	const requestsPerRoute = 1000

	for _, route := range routes {
		for i := 0; i < requestsPerRoute; i++ {
			wg.Add(1)
			go sendRequest(&wg, route)
		}
	}

	wg.Wait()
	fmt.Println("Todas las solicitudes han sido enviadas.")
}
