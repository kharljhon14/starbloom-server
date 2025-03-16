package api

import (
	"errors"
	"net/http"

	"github.com/kharljhon14/starbloom-server/internal/data"
	"github.com/kharljhon14/starbloom-server/internal/validator"
)

func (app *Application) likePostHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getContextUser(r)

	var input struct {
		PostID int64 `json:"post_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(input.PostID > 0, "post_id", "post_id must be valid")

	if !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	like := data.Like{
		PostID: input.PostID,
		UserID: user.ID,
	}

	err = app.Models.Likes.Insert(&like)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrAlreadyLiked):
			app.badRequestErrorResponse(w, r, err)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"like": like}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) unlikePostHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getContextUser(r)

	var input struct {
		PostID int64 `json:"post_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(input.PostID > 0, "post_id", "post_id must be valid")

	if !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	err = app.Models.Likes.Delete(input.PostID, user.ID)
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

func (app *Application) getLikeCountHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PostID int64 `json:"post_id"`
	}

	v := validator.New()

	qs := r.URL.Query()

	postID := app.readInt(qs, "postID", 0, v)

	input.PostID = int64(postID)

	v.Check(input.PostID > 0, "postID", "post_id must be valid")
	if !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	_, err := app.Models.Posts.Get(input.PostID)
	if err != nil {
		app.notFoundErrorResponse(w, r)
		return
	}

	count, err := app.Models.Likes.Get(input.PostID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"likes": count}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
