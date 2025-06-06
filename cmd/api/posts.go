package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/kharljhon14/starbloom-server/internal/data"
	"github.com/kharljhon14/starbloom-server/internal/validator"
)

func (app *Application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Content string `json:"content"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(input.Content != "", "content", "content is required")
	v.Check(len(input.Content) <= 255, "content", "content must not exceed 255 characters")

	if !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	user := app.getContextUser(r)

	post := data.Post{
		UserID:  user.ID,
		Content: input.Content,
	}

	err = app.Models.Posts.Insert(&post)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"post": post}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		app.notFoundErrorResponse(w, r)
		return
	}

	post, err := app.Models.Posts.Get(ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecordFound):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"post": post}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) getPostsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID int64 `json:"id"`
		data.Filter
	}

	v := validator.New()

	qs := r.URL.Query()

	id := int64(app.readInt(qs, "id", 0, v))
	page := app.readInt(qs, "page", 1, v)
	pageSize := app.readInt(qs, "pageSize", 10, v)

	input.ID = id
	input.Filter.Page = page
	input.Filter.PageSize = pageSize

	v.Check(input.ID > 0, "id", "must be a valid id")
	if data.ValidateFilters(v, input.Filter); !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	posts, metadata, err := app.Models.Posts.GetAll(input.ID, input.Filter)
	if err != nil {

		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"_metadata": metadata, "posts": posts}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getContextUser(r)
	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		app.notFoundErrorResponse(w, r)
		return
	}

	post, err := app.Models.Posts.Get(ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecordFound):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	if user.ID != post.UserID {
		app.authorizationRequiredResponse(w, r)
		return
	}

	var input struct {
		Content string `json:"content"`
	}
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(input.Content != "", "content", "content is required")
	v.Check(len(input.Content) <= 255, "content", "content must not exceed 255 characters")

	if !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	post.Content = input.Content

	err = app.Models.Posts.Update(post)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"post": post}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getContextUser(r)

	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		app.notFoundErrorResponse(w, r)
		return
	}

	post, err := app.Models.Posts.Get(ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecordFound):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if user.ID != post.UserID {
		app.authorizationRequiredResponse(w, r)
		return
	}

	err = app.Models.Posts.Delete(post.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecordFound):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "cost successfuly deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) getFollowingPostsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID int64 `json:"id"`
		data.Filter
	}
	v := validator.New()

	qs := r.URL.Query()

	ID := app.readInt(qs, "id", 0, v)
	page := app.readInt(qs, "page", 1, v)
	pageSize := app.readInt(qs, "pageSize", 50, v)

	input.ID = int64(ID)
	input.Filter.Page = page
	input.Filter.PageSize = pageSize

	v.Check(input.ID > 0, "id", "ID must be valid")
	if data.ValidateFilters(v, input.Filter); !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	posts, metadata, err := app.Models.Follows.GetFollowingPosts(input.ID, input.Filter)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"_metadata": metadata, "posts": posts}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
