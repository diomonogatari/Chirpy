package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type ApiConfig struct {
	fileserverHits int
	chirpMaxSize   uint
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ValidResponse struct {
	Valid bool `json:"valid"`
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
	type chirp struct {
		Body string `json:"body"`
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	chirpDecoder := json.NewDecoder(r.Body)
	chirpMsg := chirp{}

	err := chirpDecoder.Decode(&chirpMsg)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		errMessage := ErrorResponse{Error: "Something went wrong"}

		resp, errMarshal := json.Marshal(errMessage)
		if errMarshal != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte{})
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resp)
		return
	}

	// Too long
	if len(chirpMsg.Body) > int(cfg.chirpMaxSize) {
		tooLongError := ErrorResponse{Error: "Chirp is too long"}

		resp, errMarshal := json.Marshal(tooLongError)
		if errMarshal != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte{})
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	} else {
		validResponse := ValidResponse{Valid: true}
		resp, errMarshal := json.Marshal(validResponse)
		if errMarshal != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte{})
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}

}
