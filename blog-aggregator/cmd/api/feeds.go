package main

import (
	"net/http"
	"time"

	"github.com/chhoumann/blogaggregator/internal/database"
	"github.com/google/uuid"
)

func (app *application) createFeed(w http.ResponseWriter, r *http.Request, user database.User) {
	var input struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	feed := database.CreateFeedParams{
		ID:        uuid.New(),
		Name:      input.Name,
		Url:       input.URL,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
	}

	createdFeed, err := app.DB.CreateFeed(r.Context(), feed)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Automatically follow the feed
	feedFollow := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    createdFeed.ID,
	}

	_, err = app.DB.CreateFeedFollow(r.Context(), feedFollow)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"feed": createdFeed, "feed_follow": feedFollow}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getAllFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := app.DB.GetAllFeeds(r.Context())
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"feeds": feeds}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
