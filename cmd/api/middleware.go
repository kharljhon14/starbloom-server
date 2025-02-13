package api

import "net/http"

func (app *Application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		app.Logger.PrintInfo("Request information", map[string]string{
			"method": r.Method,
			"url":    r.URL.String(),
		})

		next.ServeHTTP(w, r)
	})
}
