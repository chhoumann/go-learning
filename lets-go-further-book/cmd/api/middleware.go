package main

import (
	"errors"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strconv"
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

		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// accepts and returns a HandlerFunc so we can wrap handler functions directly
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	return app.requireAuthenticatedUser(fn)
}

func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		if !permissions.Include(code) {
			app.nonPermittedResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}

	return app.requireActivatedUser(fn)
}

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")
		if origin != "" && slices.Contains(app.config.cors.trustedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)

			// If the request is a preflight OPTIONS request, we need to set the
			// Access-Control-Allow-Methods and Access-Control-Allow-Headers headers.
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
				// Since we're allowing Authorization, Allow-Origin should be checked against a
				// list of trusted origins. Never use `*` in this case.
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

				// Write headers along with 200 OK status and return from the middleware with no further action
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

type metricsResponseWriter struct {
	wrapped       http.ResponseWriter
	statusCode    int
	headerWritten bool
}

func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{wrapped: w, statusCode: http.StatusOK}
}

func (mrw *metricsResponseWriter) Header() http.Header {
	return mrw.wrapped.Header()
}

func (mrw *metricsResponseWriter) WriteHeader(statusCode int) {
	mrw.wrapped.WriteHeader(statusCode)

	if !mrw.headerWritten {
		mrw.statusCode = statusCode
		mrw.headerWritten = true
	}
}

func (mrw *metricsResponseWriter) Write(b []byte) (int, error) {
	mrw.headerWritten = true
	return mrw.wrapped.Write(b)
}

func (mrw *metricsResponseWriter) Unwrap() http.ResponseWriter {
	return mrw.wrapped
}

func (app *application) metrics(next http.Handler) http.Handler {
	var (
		totalRequestsReceived           = expvar.NewInt("total_requests_received")
		totalResponsesSent              = expvar.NewInt("total_responses_sent")
		totalProcessingTimeMicroseconds = expvar.NewInt("total_processing_time_μs")
		totalResponsesSentByStatus      = expvar.NewMap("total_responses_sent_by_status")
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		totalRequestsReceived.Add(1)

		mrw := newMetricsResponseWriter(w)

		next.ServeHTTP(mrw, r)

		totalResponsesSent.Add(1)
		totalResponsesSentByStatus.Add(strconv.Itoa(mrw.statusCode), 1)

		duration := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(duration)
	})
}
