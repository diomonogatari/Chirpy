package main

import (
	"log"
	"net/http"
)

func main() {
	apiConfig := new(ApiConfig)
	apiConfig.chirpMaxSize = 140
	mux := http.NewServeMux()

	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiConfig.incrementHits(fileServerHandler))

	registerRoutes(mux, apiConfig)

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	log.Fatal(server.ListenAndServe())
}

func registerRoutes(mux *http.ServeMux, apiConfig *ApiConfig) {
	mux.HandleFunc("GET /admin/metrics", apiConfig.getHits)             // Metrics endpoint
	mux.HandleFunc("GET /api/healthz", healthzHandler)                  // Health check
	mux.HandleFunc("GET /api/reset", apiConfig.resetHits)               // Reset hits counter
	mux.HandleFunc("POST /api/validate_chirp", apiConfig.validateChirp) // Check if chirp is valid
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
