package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	cfg := apiConfig{}
	serverMux := http.NewServeMux()

	//serverMux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	fileServer := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	serverMux.Handle("/app/", cfg.middlewareMetricsInc(fileServer))

	serverMux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	serverMux.HandleFunc("GET /api/healthz", healthHandler)
	serverMux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	serverMux.HandleFunc("POST /admin/reset", cfg.resetHandler)
	serverMux.HandleFunc("POST /api/validate_chirp", validateHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Error starting this shit: ", err)
	}

}
