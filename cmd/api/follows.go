package api

import (
	"net/http"

	"github.com/kharljhon14/starbloom-server/internal/data"
	"github.com/kharljhon14/starbloom-server/internal/validator"
)

func (app *Application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	var input data.Follow

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(input.UserID != 0, "user_id", "user_id is required")
	v.Check(input.FollowerID != 0, "follower_id", "follower_id is required")
	v.Check(input.UserID != input.FollowerID, "follower_id", "invalid follower_id")

	if !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	follow, err := app.Models.Follows.Insert(input.UserID, input.FollowerID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"follow": follow}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) unFollowUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID     int64 `json:"user_id"`
		FollowerID int64 `json:"follower_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(input.UserID != 0, "user_id", "user_id is required")
	v.Check(input.FollowerID != 0, "follower_id", "follower_id is required")
	v.Check(input.UserID != input.FollowerID, "follower_id", "invalid follower_id")

	if !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	err = app.Models.Follows.Delete(input.UserID, input.FollowerID)
	if err != nil {
		// Update error response to 404
		app.serverErrorResponse(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app *Application) getFollowersHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID int64 `json:"user_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(input.UserID != 0, "user_id", "user_id is required")

	if !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	users, err := app.Models.Follows.GetFollowers(input.UserID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"users": users}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
