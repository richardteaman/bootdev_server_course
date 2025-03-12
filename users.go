package main

import (
	"encoding/json"
	"goserver/internal/auth"
	"goserver/internal/database"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func ConverToUser(dbUser database.User) User {
	return User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
}

func (cfg *apiConfig) usersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method", nil)
		return
	}

	type parametrs struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parametrs{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid JSON request Body", nil)
		return
	}
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password", err)
		return
	}

	dbUser, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create user", err)
	}

	user := ConverToUser(dbUser)
	respondWithJSON(w, http.StatusCreated, user)

}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "invalid request method", nil)
	}

	type loginParams struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := loginParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not decode login params (invalid json)", err)
		return
	}

	dbUser, err := cfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorect email or password", nil)
		return
	}

	err = auth.CheckPasswordHash(params.Password, dbUser.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorect email or password", nil)
		return
	}

	expirationTime := time.Hour
	if params.ExpiresInSeconds > 0 {
		expiration := time.Duration(params.ExpiresInSeconds) * time.Second
		if expiration < expirationTime {
			expirationTime = expiration
		}
	}

	accessToken, err := auth.MakeJWT(dbUser.ID, cfg.JWTSecret, expirationTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token", err)
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate refresh token", err)
	}

	err = cfg.DB.CreateRefreshTokens(r.Context(), database.CreateRefreshTokensParams{
		Token:     refreshToken,
		UserID:    dbUser.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour), //30 days
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to store refresh token in the DB", err)
	}

	userWithTokens := map[string]interface{}{
		"id":            dbUser.ID,
		"email":         dbUser.Email,
		"created_at":    dbUser.CreatedAt,
		"updated_at":    dbUser.UpdatedAt,
		"token":         accessToken,
		"refresh_token": refreshToken,
	}

	respondWithJSON(w, http.StatusOK, userWithTokens)
	//user := ConverToUser(dbUser)
	//respondWithJSON(w, http.StatusOK, user)
}
