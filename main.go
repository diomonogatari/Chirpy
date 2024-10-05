package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/diomonogatari/Chirpy/internal/api"
	"github.com/diomonogatari/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	apiConfig, err := api.NewApiConfig(140, database.New(db))
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()

	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiConfig.IncrementHits(fileServerHandler))

	registerRoutes(mux, apiConfig)

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	log.Fatal(server.ListenAndServe())
}

func registerRoutes(mux *http.ServeMux, apiConfig *api.ApiConfig) {
	mux.HandleFunc("GET /admin/metrics", apiConfig.GetHits) // Metrics endpoint
	mux.HandleFunc("GET /api/healthz", healthzHandler)      // Health check
	mux.HandleFunc("POST /admin/reset", apiConfig.Reset)    // Reset hits counter
	mux.HandleFunc("POST /api/chirps", apiConfig.PostChirp)
	mux.HandleFunc("POST /api/users", apiConfig.CreateUser)
	mux.HandleFunc("GET /api/chirps", apiConfig.GetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiConfig.GetChirp)
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
