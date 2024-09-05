package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	mux := httprouter.New()

	mux.HandlerFunc(http.MethodGet, "/v1/healthz", func(w http.ResponseWriter, r *http.Request) {
		app.writeJSON(w, http.StatusOK, envelope{"status": "ok"}, nil)
	})

	mux.HandlerFunc(http.MethodGet, "/v1/err", func(w http.ResponseWriter, r *http.Request) {
		app.writeError(w, http.StatusInternalServerError, "internal server error", nil)
	})

	mux.HandlerFunc(http.MethodPost, "/v1/users", app.createUser)
	mux.HandlerFunc(http.MethodGet, "/v1/users", app.getUserByAPIKey)

	mux.HandlerFunc(http.MethodPost, "/v1/feeds", app.authMiddleware(app.createFeed))
	mux.HandlerFunc(http.MethodGet, "/v1/feeds", app.getAllFeeds)

	// @feed_follows.go curl -X POST -H "Authorization: ApiKey <your-api-key>" -H "Content-Type: application/json" -d '{"feed_id":"<feed-uuid>"}' http://localhost:8080/v1/feed_follows
	mux.HandlerFunc(http.MethodPost, "/v1/feed_follows", app.authMiddleware(app.createFeedFollow))

	// @feed_follows.go curl -X DELETE -H "Authorization: ApiKey <your-api-key>" http://localhost:8080/v1/feed_follows/<feed-follow-uuid>
	mux.HandlerFunc(http.MethodDelete, "/v1/feed_follows/:id", app.authMiddleware(app.deleteFeedFollow))

	// @feed_follows.go curl -H "Authorization: ApiKey <your-api-key>" http://localhost:8080/v1/feed_follows
	mux.HandlerFunc(http.MethodGet, "/v1/feed_follows", app.authMiddleware(app.getFollowedFeeds))

	// curl -H "Authorization: ApiKey <your-api-key>" "http://localhost:8080/v1/posts?limit=20"
	mux.HandlerFunc(http.MethodGet, "/v1/posts", app.authMiddleware(app.getPostsByUser))

	return mux
}
