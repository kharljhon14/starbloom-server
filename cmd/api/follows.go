package api

import (
	"errors"
	"net/http"

	"github.com/kharljhon14/starbloom-server/internal/data"
	"github.com/kharljhon14/starbloom-server/internal/validator"
)

func (app *Application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getContextUser(r)
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
	v.Check(input.UserID != user.ID, "user_id", "must not be own user_id")

	if !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	follow, err := app.Models.Follows.Insert(input.UserID, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrAlreadyFollowing):
			app.badRequestErrorResponse(w, r, err)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"follow": follow}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) unFollowUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getContextUser(r)

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
	v.Check(input.UserID != user.ID, "user_id", "must not be own user_id")

	if !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	err = app.Models.Follows.Delete(input.UserID, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecordFound):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app *Application) getFollowersHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID int64 `json:"user_id"`
		data.Filter
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(input.UserID != 0, "user_id", "user_id is required")

	qs := r.URL.Query()

	page := app.readInt(qs, "page", 1, v)
	pageSize := app.readInt(qs, "pageSize", 50, v)

	input.Filter.Page = page
	input.Filter.PageSize = pageSize

	if data.ValidateFilters(v, input.Filter); !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	users, metadata, err := app.Models.Follows.GetFollowers(input.UserID, input.Filter)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"_metadata": metadata, "users": users}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
