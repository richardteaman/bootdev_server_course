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
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func ConvertToUser(dbUser database.User) User {
	return User{
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.IsChirpyRed,
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

	user := ConvertToUser(dbUser)
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
		"is_chirpy_red": dbUser.IsChirpyRed,
	}

	respondWithJSON(w, http.StatusOK, userWithTokens)
	//user := ConvertToUser(dbUser)
	//respondWithJSON(w, http.StatusOK, user)
}

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method", nil)
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
	}

	type updateParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params := updateParams{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON request body", err)
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password", err)
	}

	err = cfg.DB.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update user", err)
		return
	}

	updatedUser, err := cfg.DB.GetUserById(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch updated user", err)
		return
	}

	respondWithJSON(w, http.StatusOK, ConvertToUser(updatedUser))

}
