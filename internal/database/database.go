package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type DB struct {
	fullPath    string
	mux         *sync.RWMutex
	lastChirpId int
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

var dbStructure = DBStructure{Chirps: make(map[int]Chirp)}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(fullPath string) (*DB, error) {
	logger := log.Default()
	logger.Printf("The path prompted was %s", fullPath)
	db := &DB{
		fullPath: fullPath,
		mux:      &sync.RWMutex{},
	}

	if err := db.ensureDB(); err != nil {
		return nil, err
	}

	return db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.lastChirpId++

	newChirp := Chirp{
		Body: body,
		Id:   db.lastChirpId,
	}

	dbStructure.Chirps[newChirp.Id] = newChirp

	return newChirp, db.writeDB(dbStructure)
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	var chirps []Chirp
	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirp(chirpId int) (Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	badChirp := Chirp{}
	foundChirp := dbStructure.Chirps[chirpId]

	if foundChirp == badChirp {
		return foundChirp, fmt.Errorf("Chirp with Id %d does not exist", chirpId)
	}
	return foundChirp, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	if _, err := os.Stat(db.fullPath); os.IsNotExist(err) {
		file, err := os.Create(db.fullPath)
		if err != nil {
			return err
		}
		defer file.Close()

		// Initialize an empty database structure
		return db.writeDB(dbStructure)
	}
	loaded, err := db.loadDB()
	dbStructure = loaded
	return err
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	fileDb, err := os.ReadFile(db.fullPath)
	if err != nil {
		return DBStructure{}, err
	}

	result := &DBStructure{}
	if err := json.Unmarshal(fileDb, result); err != nil {
		return DBStructure{}, err
	}

	for _, chirp := range result.Chirps {
		if chirp.Id > db.lastChirpId {
			db.lastChirpId = chirp.Id
		}
	}

	return *result, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	encoded, err := json.MarshalIndent(dbStructure, "", " ")
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return err
	}

	return os.WriteFile(db.fullPath, encoded, os.ModePerm)
}
