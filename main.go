package main

import (
	"log"
	"net/http"
)

func main() {
	apiConfig := new(ApiConfig)
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
	mux.HandleFunc("/healthz", healthzHandler)    // Health check
	mux.HandleFunc("/metrics", apiConfig.getHits) // Metrics endpoint
	mux.HandleFunc("/reset", apiConfig.resetHits) // Reset hits counter
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
