package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"slices"

	"github.com/kharljhon14/starbloom-server/internal/data"
	"github.com/kharljhon14/starbloom-server/internal/validator"
)

func (app *Application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		app.Logger.PrintInfo("Request information", map[string]string{
			"method": r.Method,
			"url":    r.URL.String(),
		})

		next.ServeHTTP(w, r)
	})
}

func (app *Application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")

				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *Application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.getContextUser(r)
		if user.IsAnonymous() {
			app.authenticatedRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *Application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.String()

		if url == "/api/v1/login" ||
			url == "/api/v1/signup" ||
			url == "/api/v1/health" {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.setContextUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		v := validator.New()

		if data.ValidateTokenPlainText(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := app.Models.Tokens.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrNoRecordFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = app.setContextUser(r, user)

		next.ServeHTTP(w, r)
	})
}

func (app *Application) enableCors(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")

		origin := r.Header.Get("Origin")

		trustedOrigins := os.Getenv("trusted_origins")
		if trustedOrigins == "" {
			app.serverErrorResponse(w, r, errors.New("server encountered an issue"))
			panic("env: trusted_cors is missing")
		}

		if slices.Contains(strings.Fields(trustedOrigins), origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)

			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, PATCH, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

				w.WriteHeader(http.StatusOK)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
