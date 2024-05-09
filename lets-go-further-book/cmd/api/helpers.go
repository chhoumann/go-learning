package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Not strictly necessary, but there are some benefits to using envelopes.
// Enveloping a response means to wrap the response in a JSON object that contains
// the response data and any metadata. This is useful for returning a response to the
// client with a consistent structure.
// Other benefits follow:
// 1. the response is more self-documenting.
// 2. the client has to explicitly access the response via e.g. the `movie` key, so it's clear
//    what the response is for.
// 3. mitigation of a potential security issue (https://haacked.com/archive/2008/11/20/anatomy-of-a-subtle-json-vulnerability.aspx/)
type envelope map[string]interface{}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}
