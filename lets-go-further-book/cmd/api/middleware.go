package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
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
