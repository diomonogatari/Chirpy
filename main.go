package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/diomonogatari/Chirpy/internal/api"
)

func main() {
	absPath, _ := filepath.Abs("./internal/database/database.json")
	apiConfig, err := api.NewApiConfig(140, absPath)
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
	mux.HandleFunc("GET /api/reset", apiConfig.ResetHits)   // Reset hits counter
	mux.HandleFunc("POST /api/chirps", apiConfig.PostChirp)
	mux.HandleFunc("GET /api/chirps", apiConfig.GetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiConfig.GetChirp)
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
