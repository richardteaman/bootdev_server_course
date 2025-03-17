package main

import (
	"encoding/json"
	"goserver/internal/auth"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) polkaWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method", nil)
		return
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid API key", nil)
		return
	}

	if apiKey != cfg.PolkaKey {
		respondWithError(w, http.StatusUnauthorized, "Invalid API key", nil)
		return
	}

	type webhookPayload struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	payload := webhookPayload{}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&payload)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON request body", err)
		return
	}

	if payload.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent) //204
		return
	}

	userID, err := uuid.Parse(payload.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user_id format", err)
		return
	}

	err = cfg.DB.UpgradeUserToChirpyRed(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
