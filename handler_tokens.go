package main

import (
	"net/http"
	"time"

	"github.com/trungdoanle1101/chirp/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "a valid bearer token is required", err)
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid refresh token", err)
		return
	}

	if refreshToken.ExpiresAt.UTC().Before(time.Now().UTC()) {
		respondWithError(w, http.StatusUnauthorized, "refresh token already expired", nil)
		return
	}

	if refreshToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "refresh token already revoked", nil)
		return
	}

	type response struct {
		Token string `json:"token"`
	}

	accessToken, err := auth.MakeJWT(refreshToken.UserID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create a new access token", err)
		return
	}

	resp := response{
		Token: accessToken,
	}
	respondWithJSON(w, http.StatusOK, resp)

}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "a valid bearer token is required", err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke the refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
