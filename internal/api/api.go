package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/diomonogatari/Chirpy/internal/database"
	"github.com/google/uuid"
)

type ApiConfig struct {
	chirpMaxSize   uint
	fileserverHits int
	db             *database.Queries
	platform       string
}

func NewApiConfig(maxChirpSize uint, queries *database.Queries) (*ApiConfig, error) {
	runningEnv := os.Getenv("PLATFORM")
	cfg := &ApiConfig{chirpMaxSize: maxChirpSize, fileserverHits: 0, db: queries, platform: runningEnv}
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

func (cfg *ApiConfig) Reset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform == "env" {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	cfg.fileserverHits = 0
	if err := cfg.db.DeleteUsers(r.Context()); err != nil {
		log.Printf("Error truncate the users table: %s", err)
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// CreateUser handles the creation of a new user. It expects a JSON payload
// containing an email address in the request body. If the payload is successfully
// decoded, it attempts to create a new user in the database using the provided email.
// On success, it responds with the user ID in JSON format. If any error occurs during
// the process, it responds with an appropriate error message and status code.
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response.
//   - r: *http.Request containing the HTTP request data.
//
// Possible responses:
//   - HTTP 200: User created successfully, returns the user ID in JSON format.
//   - HTTP 500: Internal server error, returns an error message.
func (cfg *ApiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), requestBody.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	myUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, http.StatusCreated, myUser)
}

func (cfg *ApiConfig) PostChirp(w http.ResponseWriter, r *http.Request) {
	var chirpMsg struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&chirpMsg); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if len(chirpMsg.Body) > int(cfg.chirpMaxSize) {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	chirpParams := database.CreateChirpParams{
		Body:   chirpMsg.Body,
		UserID: chirpMsg.UserId,
	}
	savedChirp, err := cfg.db.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Chirp is too long")
	}

	respondWithJSON(w, http.StatusCreated, mapChirp(&savedChirp))
}

func (cfg *ApiConfig) GetChirp(w http.ResponseWriter, r *http.Request) {

	uuidRequestedId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), uuidRequestedId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, mapChirp(&chirp))
}

func (cfg *ApiConfig) GetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	responseChirps := make([]Chirp, 0)
	for _, chirp := range chirps {
		responseChirps = append(responseChirps, mapChirp(&chirp))
	}

	respondWithJSON(w, http.StatusOK, responseChirps)
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

// Map chirp to response object
func mapChirp(savedChirp *database.Chirp) Chirp {
	return Chirp{
		ID:        savedChirp.ID,
		CreatedAt: savedChirp.CreatedAt,
		UpdatedAt: savedChirp.UpdatedAt,
		Body:      savedChirp.Body,
		UserID:    savedChirp.UserID,
	}
}
