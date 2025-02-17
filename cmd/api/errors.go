package api

import (
	"net/http"
)

func (app *Application) logError(r *http.Request, err error) {
	app.Logger.PrintError(err.Error(), map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
		"request_host":   r.Host,
	})
}

func (app *Application) errorResponse(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	message interface{},
) {
	env := envelope{"error": message}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(status)
	}
}

func (app *Application) serverErrorResponse(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	app.logError(r, err)

	message := "the server encountered an issue and could not process the request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (app *Application) badRequestErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err)
}

func (app *Application) validationErrorResponse(w http.ResponseWriter, r *http.Request, err map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, err)
}
