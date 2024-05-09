package main

import (
	"fmt"
	"net/http"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a deferred function. It will always be run in the event of a panic as Go
		// unwinds the stack.
		defer func() {
			// Use the builtin recover function to check if there was a panic
			if err := recover(); err != nil {
				// Set the "Connection" header to "close" to make sure the client
				// does not expect anything else after the response has been written.
				// And will make Go's HTTP server automatically close the connection after
				// the response has been sent.
				w.Header().Set("Connection", "close")
				// Call the serverErrorResponse helper method to handle the error.
				// This will log the error and return a 500 Internal Server Error
				// response to the client.
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
