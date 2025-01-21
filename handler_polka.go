package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/trungdoanle1101/chirp/internal/auth"
)

type Event string

const EventUserUpgraded Event = "user.upgraded"

func (cfg *apiConfig) handlerPolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	providedKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to extract api key", err)
		return
	}

	if providedKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "invalid api key", nil)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if params.Event != string(EventUserUpgraded) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse uuid", err)
		return
	}

	_, err = cfg.db.SetChirpyRed(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Couldn't find user with the provided id", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Couldn't set chirp to red", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
