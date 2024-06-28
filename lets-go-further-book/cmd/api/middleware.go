package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"greenlight.bagerbach.com/internal/data"
	"greenlight.bagerbach.com/internal/validator"
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

func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		// Maps are not thread-safe. We need to lock the mutex before reading from the map.
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Background goroutine to clean up old clients.
	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.config.limiter.enabled {
			next.ServeHTTP(w, r)
			return
		}

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{
				limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
			}
		}

		clients[ip].lastSeen = time.Now()

		// Check if the request is allowed by the rate limiter.
		// `limiter.Allow()` returns `true` if the request is allowed, and `false` if the request is not allowed.
		// It consumes a token from the bucket.
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}
		// We aren't using defer here because we want to unlock the mutex as soon as possible.
		// If we deferred the unlock, it only would be executed after all the downstream handlers have returned.
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This header indicates that the response may vary based on the value of the Authorization header in the request.
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// We expect the value in the Authorization header to be in the format of "Bearer <token>".
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		// Validate token
		v := validator.New()
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Get user associated with authentication token
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// Set user to request context
		app.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}
