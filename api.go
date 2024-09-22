package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

var profaneWords = []string{"kerfuffle", "sharbert", "fornax"}

type ApiConfig struct {
	fileserverHits int
	chirpMaxSize   uint
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type CleanResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

func (cfg *ApiConfig) incrementHits(next http.Handler) http.Handler {
	incr := func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(incr)
}

func (cfg *ApiConfig) getHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(getCount(cfg)))
}

func getCount(cfg *ApiConfig) string {
	return fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits)
}

func (cfg *ApiConfig) resetHits(w http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits = 0
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
}

func (cfg *ApiConfig) validateChirp(w http.ResponseWriter, r *http.Request) {
	var chirpMsg struct {
		Body string `json:"body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&chirpMsg); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if len(chirpMsg.Body) > int(cfg.chirpMaxSize) {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanResponse := checkProfane(chirpMsg.Body)
	respondWithJSON(w, http.StatusOK, cleanResponse)
}

func checkProfane(message string) CleanResponse {
	cleanedMessage := message

	words := strings.Fields(cleanedMessage)

	for _, badWord := range profaneWords {
		for _, word := range words {
			// Check if the lowercased word matches the bad word
			if strings.EqualFold(word, badWord) {
				// Replace the bad word with "****" in the cleaned message
				cleanedMessage = strings.ReplaceAll(cleanedMessage, word, "****")
			}
		}
	}

	return CleanResponse{CleanedBody: cleanedMessage}
}

func respondWithError(w http.ResponseWriter, status int, message string) {
	resp := ErrorResponse{Error: message}
	respondWithJSON(w, status, resp)
}

func respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
