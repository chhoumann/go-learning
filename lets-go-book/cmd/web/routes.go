package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	dynamic := alice.New(app.sessionManager.LoadAndSave)

	router.HandlerFunc(http.MethodGet, "/", dynamic.ThenFunc(app.home).ServeHTTP)
	router.HandlerFunc(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.snippetView).ServeHTTP)
	router.HandlerFunc(http.MethodGet, "/snippet/create", dynamic.ThenFunc(app.snippetCreate).ServeHTTP)
	router.HandlerFunc(http.MethodPost, "/snippet/create", dynamic.ThenFunc(app.snippetCreatePost).ServeHTTP)

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}
