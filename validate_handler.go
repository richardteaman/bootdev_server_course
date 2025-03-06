package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

func validateHandler(w http.ResponseWriter, r *http.Request) {
	type parametrs struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		Cleaned_body string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parametrs{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleanMsg, err := checkForProfane(params.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't validate parameters", err)
	}

	respondWithJSON(w, http.StatusOK, returnVals{
		Cleaned_body: cleanMsg,
	})

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
