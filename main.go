package main

import (
	"database/sql"
	"fmt"
	"goserver/internal/database"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	Platform       string
	JWTSecret      string
	PolkaKey       string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("JWT_SECRET is not set in .env file")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not set in .env file")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("platform is not set in .env file")
	}

	dbQueries := database.New(db)

	polkaKey := os.Getenv("POLKA_KEY")
	if polkaKey == "" {
		log.Fatalf("POLKA_KEY is not set in .env file")
	}

	cfg := apiConfig{
		DB:        dbQueries,
		Platform:  platform,
		JWTSecret: jwtSecret,
		PolkaKey:  polkaKey,
	}
	serverMux := http.NewServeMux()

	//serverMux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	fileServer := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	serverMux.Handle("/app/", cfg.middlewareMetricsInc(fileServer))

	serverMux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	serverMux.HandleFunc("GET /api/healthz", healthHandler)
	serverMux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	serverMux.HandleFunc("POST /admin/reset", cfg.resetHandler)
	//serverMux.HandleFunc("POST /api/validate_chirp", validateHandler)
	serverMux.HandleFunc("POST /api/users", cfg.usersHandler)
	serverMux.HandleFunc("POST /api/chirps", cfg.chirpHandler)
	serverMux.HandleFunc("GET /api/chirps", cfg.chirpsGetHandler)
	serverMux.HandleFunc("GET /api/chirps/{chirpID}", cfg.chirpGetHandler)
	serverMux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.chirpDeleteHandler)
	serverMux.HandleFunc("POST /api/login", cfg.loginHandler)
	serverMux.HandleFunc("POST /api/refresh", cfg.refreshTokenHandler)
	serverMux.HandleFunc("POST /api/revoke", cfg.revokeTokenHandler)
	serverMux.HandleFunc("PUT /api/users", cfg.updateUserHandler)
	serverMux.HandleFunc("POST /api/polka/webhooks", cfg.polkaWebhookHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}

	err = server.ListenAndServe()
	if err != nil {
		fmt.Println("Error starting this shit: ", err)
	}

}
