package main

import (
	"net/http"
	"strconv"

	"github.com/chhoumann/blogaggregator/internal/database"
)

func (app *application) getPostsByUser(w http.ResponseWriter, r *http.Request, user database.User) {
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}
		limit = parsedLimit
	}

	// Get posts for the user
	posts, err := app.DB.GetPostsByUser(r.Context(), database.GetPostsByUserParams{
		ID:    user.ID,
		Limit: int32(limit),
	})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"posts": posts}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
