package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/diomonogatari/Chirpy/internal/database"
)

var profaneWords = []string{"kerfuffle", "sharbert", "fornax"}

type ApiConfig struct {
	chirpMaxSize   uint
	fileserverHits int
	db             *database.Queries
}

func NewApiConfig(maxChirpSize uint, queries *database.Queries) (*ApiConfig, error) {

	cfg := &ApiConfig{chirpMaxSize: maxChirpSize, fileserverHits: 0, db: queries}
	return cfg, nil
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (cfg *ApiConfig) IncrementHits(next http.Handler) http.Handler {
	incr := func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(incr)
}

func (cfg *ApiConfig) GetHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(getCount(cfg)))
}

func getCount(cfg *ApiConfig) string {
	return fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits)
}

func (cfg *ApiConfig) ResetHits(w http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits = 0
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
}

func (cfg *ApiConfig) PostChirp(w http.ResponseWriter, r *http.Request) {
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

	// savedChirp, err := cfg.db.CreateChirp(checkProfane(chirpMsg.Body))
	// if err != nil {
	// 	respondWithError(w, http.StatusInternalServerError, "Chirp is too long")
	// }

	// respondWithJSON(w, http.StatusCreated, savedChirp)
}

func (cfg *ApiConfig) GetChirp(w http.ResponseWriter, r *http.Request) {

	// requestedId, atoiErr := strconv.Atoi(r.PathValue("chirpID"))
	// if atoiErr != nil {
	// 	respondWithError(w, http.StatusInternalServerError, atoiErr.Error())
	// 	return
	// }

	// chirp, err := cfg.db.GetChirp(requestedId)
	// if err != nil {
	// 	respondWithError(w, http.StatusNotFound, err.Error())
	// 	return
	// }

	// respondWithJSON(w, http.StatusOK, chirp)
}

func (cfg *ApiConfig) GetChirps(w http.ResponseWriter, _ *http.Request) {
	// chirps, err := cfg.db.GetChirps()
	// if err != nil {
	// 	respondWithError(w, http.StatusInternalServerError, err.Error())
	// }

	// respondWithJSON(w, http.StatusOK, chirps)
}

func checkProfane(message string) string {
	cleanedMessage := message

	// Regular expression to match words, ignoring punctuation
	re := regexp.MustCompile(`[^\w]+`) // Matches one or more non-word characters
	words := re.Split(cleanedMessage, -1)

	for _, badWord := range profaneWords {
		for _, word := range words {
			// Check if the lowercased word matches the bad word
			if strings.EqualFold(word, badWord) {
				// Replace the bad word with "****" in the cleaned message
				cleanedMessage = strings.ReplaceAll(cleanedMessage, word, "****")
				break
			}
		}
	}

	return cleanedMessage
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
