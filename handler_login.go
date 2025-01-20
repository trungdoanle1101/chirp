package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/trungdoanle1101/chirp/internal/auth"
	"github.com/trungdoanle1101/chirp/internal/database"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	result, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password", err)
		return
	}

	err = auth.CheckPasswordHash(params.Password, result.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password", err)
		return
	}

	expirationTime := time.Hour

	accessToken, err := auth.MakeJWT(result.ID, cfg.jwtSecret, expirationTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to generate token", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to generate refresh token", err)
		return
	}

	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    result.ID,
		ExpiresAt: time.Now().UTC().Add(60 * 24 * time.Hour),
	}

	_, err = cfg.db.CreateRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to save refresh token", err)
	}

	resp := response{

		User: User{
			ID:        result.ID,
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
			Email:     result.Email,
		},
		Token:        accessToken,
		RefreshToken: refreshToken,
	}
	respondWithJSON(w, http.StatusOK, resp)
}
