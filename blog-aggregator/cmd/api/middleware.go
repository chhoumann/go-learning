package main

import (
	"database/sql"
	"net/http"

	"github.com/chhoumann/blogaggregator/internal/database"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (app *application) authMiddleware(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("Authorization")
		if apiKey == "" {
			app.writeError(w, http.StatusUnauthorized, "missing authorization header", nil)
			return
		}

		apiKey = extractAPIKey(apiKey)

		user, err := app.DB.GetUserByAPIKey(r.Context(), sql.NullString{String: apiKey, Valid: true})
		if err != nil {
			app.writeError(w, http.StatusUnauthorized, "invalid API key", nil)
			return
		}

		handler(w, r, user)
	}
}
