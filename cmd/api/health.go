package api

import "net/http"

func (app *Application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("ok"))
}
