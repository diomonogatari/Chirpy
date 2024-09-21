package main

import (
	"fmt"
	"net/http"
)

func main() {
	var apiConfig ApiConfig
	mux := http.NewServeMux()

	wrappedHandler := apiConfig.incrementHits(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.Handle("/app/", wrappedHandler)
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/metrics", apiConfig.getHits)
	mux.HandleFunc("/reset", apiConfig.resetHits)

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	server.ListenAndServe()
}

func healthzHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("Content-Type", "text/plain; charset=utf-8")
	response.WriteHeader(200)
	response.Write([]byte("OK"))
}

type ApiConfig struct {
	fileserverHits int
}

func (cfg *ApiConfig) incrementHits(next http.Handler) http.Handler {
	incr := func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(incr)
}

func (cfg *ApiConfig) getHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("Hits: " + fmt.Sprint(cfg.fileserverHits)))
}

func (cfg *ApiConfig) resetHits(w http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits = 0
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
}
