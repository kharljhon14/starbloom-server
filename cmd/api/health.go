package api

import (
	"net/http"
)

func (app *Application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"enviroment": app.Config.Env,
			"version":    version,
		},
	}

	if err := app.writeJSON(w, http.StatusOK, env, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
