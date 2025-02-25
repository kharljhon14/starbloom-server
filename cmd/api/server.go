package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kharljhon14/starbloom-server/internal/data"
	"github.com/kharljhon14/starbloom-server/internal/jsonlog"
)

const version = "1.0.0"

type Config struct {
	Port int
	Env  string
	Db   struct {
		Dsn string
	}
}

type Application struct {
	Config Config
	Logger *jsonlog.Logger
	Models data.Models
}

func (app *Application) Mount() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", app.healthCheckHandler)

	mux.HandleFunc("POST /api/v1/signup", app.createUserHandler)
	mux.HandleFunc("POST /api/v1/login", app.createAuthenticationTokenHandler)
	mux.HandleFunc("GET /api/v1/users/{username}", app.requireAuthenticatedUser(app.getUserhandler))

	mux.HandleFunc("POST /api/v1/follow", app.requireAuthenticatedUser(app.followUserHandler))
	mux.HandleFunc("POST /api/v1/unfollow", app.requireAuthenticatedUser(app.unFollowUserHandler))

	mux.HandleFunc("GET /api/v1/followers", app.requireAuthenticatedUser(app.getFollowersHandler))

	mux.HandleFunc("GET /api/v1/posts", app.requireAuthenticatedUser(app.getPostsHandler))
	mux.HandleFunc("POST /api/v1/posts", app.requireAuthenticatedUser(app.createPostHandler))
	mux.HandleFunc("GET /api/v1/posts/{id}", app.requireAuthenticatedUser(app.getPostHandler))
	mux.HandleFunc("PATCH /api/v1/posts/{id}", app.requireAuthenticatedUser(app.updatePostHandler))
	mux.HandleFunc("DELETE /api/v1/posts/{id}", app.requireAuthenticatedUser(app.deletePostHandler))
	mux.HandleFunc("GET /api/v1/posts/following", app.requireAuthenticatedUser(app.getFollowingPostsHandler))

	return app.recoverPanic(app.logRequest(app.authenticate(mux)))
}

func (app *Application) Serve(router http.Handler) error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.Config.Port),
		Handler:      router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	app.Logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.Config.Env,
	})

	return srv.ListenAndServe()
}
