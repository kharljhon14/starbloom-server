package api

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Config struct {
	Port int
	Env  string
}

type Application struct {
	Config Config
	Logger *log.Logger
}

func (app *Application) Mount() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", app.healthCheckHandler)

	return mux
}

func (app *Application) Serve(router *http.ServeMux) error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.Config.Port),
		Handler:      router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	app.Logger.Printf("starting server: %s\n", srv.Addr)
	return srv.ListenAndServe()
}
