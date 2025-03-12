package main

import (
	"goserver/internal/auth"
	"net/http"
	"time"
)

func (cfg *apiConfig) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method", nil)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid authorization token", nil)
		return
	}

	userID, err := auth.ValidateRefreshToken(token, cfg.DB)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or expired refresh token", nil)
		return
	}

	accessToken, err := auth.MakeJWT(userID, cfg.JWTSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate access token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"token": accessToken,
	})

}

func (cfg *apiConfig) revokeTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method", nil)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid authorization token", nil)
		return
	}

	err = cfg.DB.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to revoke token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent) // code 204

}
