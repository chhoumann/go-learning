package main

import (
	"net/http"
	"time"

	"github.com/chhoumann/blogaggregator/internal/database"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

func (app *application) createFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
	var input struct {
		FeedID uuid.UUID `json:"feed_id"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	feedFollow := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    input.FeedID,
	}

	_, err := app.DB.CreateFeedFollow(r.Context(), feedFollow)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"feed_follow": feedFollow}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.DB.DeleteFeedFollow(r.Context(), database.DeleteFeedFollowParams{
		ID:     id,
		UserID: user.ID,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "feed follow successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getFollowedFeeds(w http.ResponseWriter, r *http.Request, user database.User) {
	feeds, err := app.DB.GetFollowedFeeds(r.Context(), user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"feeds": feeds}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
