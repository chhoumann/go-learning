package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/chhoumann/blogaggregator/internal/database"
	"github.com/google/uuid"
)

func (app *application) createUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := database.CreateUserParams{
		ID:        uuid.New(),
		Name:      input.Name,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	_, err := app.DB.CreateUser(r.Context(), user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getUserByAPIKey(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("Authorization")
	if apiKey == "" {
		app.badRequestResponse(w, r, errors.New("missing authorization header"))
		return
	}

	apiKey = extractAPIKey(apiKey)

	user, err := app.DB.GetUserByAPIKey(r.Context(), sql.NullString{String: apiKey, Valid: true})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	userResponse := struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Name      string    `json:"name"`
		ApiKey    string    `json:"api_key"`
	}{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Name:      user.Name,
		ApiKey:    user.ApiKey.String,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": userResponse}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
