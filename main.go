package main

import (
	"encoding/json"
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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	// ...

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	count := cfg.fileserverHits.Load()
	html := fmt.Sprintf(`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>`, count)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits counter reset to 0\n"))
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	type parametrs struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parametrs{}
	err := decoder.Decode(&params)
	if err != nil {
		type errReturnVals struct {
			// the key will be the name of struct field unless you give it an explicit JSON tag

			Err string `json:"error"`
		}

		respBody := errReturnVals{
			Err: "Something went wrong",
		}
		dat, _ := json.Marshal(respBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(dat)
		return
	}
	if len(params.Body) > 140 {
		type errReturnVals struct {
			Err string `json:"error"`
		}

		respBody := errReturnVals{
			Err: "Chirp is too long",
		}
		dat, _ := json.Marshal(respBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(dat)
		return
	}

	type returnVals struct {
		Valid bool `json:"valid"`
	}
	respBody := returnVals{
		Valid: true,
	}
	dat, _ := json.Marshal(respBody)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}
