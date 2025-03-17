package main

import (
	"encoding/json"
	"errors"
	"goserver/internal/auth"
	"goserver/internal/database"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

func ConverToChirp(dbChirp database.Chirp) Chirp {
	return Chirp{
		ID:        dbChirp.ID,
		UserID:    dbChirp.UserID,
		Body:      dbChirp.Body,
		CreatedAt: dbChirp.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: dbChirp.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func checkForProfane(msg string) (cleanMsg string, err error) {
	if len(msg) == 0 {
		return "", errors.New("empty message")
	}
	words := strings.Split(msg, " ")
	for i, word := range words {
		if strings.ToLower(word) == "kerfuffle" || strings.ToLower(word) == "sharbert" || strings.ToLower(word) == "fornax" {
			words[i] = "****"
		}
	}
	resultMsg := strings.Join(words, " ")
	return resultMsg, nil
}

func (cfg *apiConfig) chirpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method", nil)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid authorization token", nil)
	}

	userID, err := auth.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", nil)
		return
	}

	/*
		type parametrs struct {
			Body    string `json:"body"`
			User_id string `json:"user_id"`
		}
	*/
	type parametrs struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parametrs{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	/*
		userUUID, err := uuid.Parse(params.User_id)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid user_id UUID format", err)
		}
	*/

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleanMsg, err := checkForProfane(params.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't validate parameters", err)
	}

	dbChirp, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		UserID: userID,
		Body:   cleanMsg,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create chirp", err)
	}

	chirp := ConverToChirp(dbChirp)
	respondWithJSON(w, http.StatusCreated, chirp)

}

func (cfg *apiConfig) chirpsGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method", nil)
		return
	}

	authorID := uuid.Nil
	authorIDString := r.URL.Query().Get("author_id")
	if authorIDString != "" {
		parsedID, err := uuid.Parse(authorIDString)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author ID", err)
			return
		}
		authorID = parsedID
	}

	dbChirps, err := cfg.DB.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve chirps", err)
		return
	}

	var chirpsResponse []Chirp

	for _, dbChirp := range dbChirps {
		if authorID != uuid.Nil && dbChirp.UserID != authorID {
			continue
		}
		chirpsResponse = append(chirpsResponse, ConverToChirp(dbChirp))
	}

	sortOrder := strings.ToLower(r.URL.Query().Get("sort"))
	if sortOrder == "desc" {

		for i, j := 0, len(chirpsResponse)-1; i < j; i, j = i+1, j-1 {
			chirpsResponse[i], chirpsResponse[j] = chirpsResponse[j], chirpsResponse[i]
		}
	}
	respondWithJSON(w, http.StatusOK, chirpsResponse)

}

func (cfg *apiConfig) chirpGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method", nil)
		return
	}

	chirpIDStr := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp id format", err)
	}

	dbChirp, err := cfg.DB.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Could not retrieve chirp (not found)", err)
		return
	}

	chirp := ConverToChirp(dbChirp)
	respondWithJSON(w, http.StatusOK, chirp)

}

func (cfg *apiConfig) chirpDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method", nil)
		return
	}

	chirpIDStr := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID format", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid authorization token", nil)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT token", nil)
		return
	}

	dbChirp, err := cfg.DB.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusMethodNotAllowed, "Chirp not found", err)
		return
	}

	if dbChirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You are not authorized to delete this chirp", nil)
		return
	}

	err = cfg.DB.DeleteChirp(r.Context(), database.DeleteChirpParams{
		ID:     chirpID,
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete chirp", err)
	}

	w.WriteHeader(http.StatusNoContent)

}
