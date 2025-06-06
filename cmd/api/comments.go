package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/kharljhon14/starbloom-server/internal/data"
	"github.com/kharljhon14/starbloom-server/internal/validator"
)

func (app *Application) addCommentHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getContextUser(r)

	var input struct {
		PostID  int64  `json:"post_id"`
		Comment string `json:"comment"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(input.PostID > 0, "post_id", "must be a valid post_id")
	v.Check(input.Comment != "", "comment", "comment is required")
	v.Check(len(input.Comment) <= 255, "comment", "comment must not exceed 255 characters")

	if !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	comment := data.Comment{
		PostID:  input.PostID,
		UserID:  user.ID,
		Comment: input.Comment,
	}

	err = app.Models.Comments.Insert(&comment)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrInvalidPostID):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"comment": comment}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) getCommentByIDHandler(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		app.notFoundErrorResponse(w, r)
		return
	}

	v := validator.New()

	v.Check(ID > 0, "id", "must be a valid id")

	if !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	comment, err := app.Models.Comments.Get(ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecordFound):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"comment": comment}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) getCommentsByPostHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PostID int64 `json:"post_id"`
		data.Filter
	}

	v := validator.New()

	qs := r.URL.Query()

	postID := int64(app.readInt(qs, "postID", 0, v))

	pageSize := app.readInt(qs, "pageSize", 10, v)
	page := app.readInt(qs, "page", 1, v)
	sort := qs.Get("sort")

	if sort == "" {
		input.Filter.Sort = "DESC"
	} else {
		input.Filter.Sort = sort
	}

	input.PostID = postID
	input.Filter.Page = page
	input.Filter.PageSize = pageSize

	v.Check(postID > 0, "post_id", "postID must be valid")
	if data.ValidateFilters(v, input.Filter); !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	_, err := app.Models.Posts.Get(postID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecordFound):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	comments, metadata, err := app.Models.Comments.GetCommentsByPost(input.PostID, input.Filter)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"_metadata": metadata, "comments": comments}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) updateCommentHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getContextUser(r)

	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		app.notFoundErrorResponse(w, r)
		return
	}

	var input struct {
		Comment string `json:"comment"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(ID > 0, "id", "ID must be valid")
	v.Check(input.Comment != "", "comment", "comment is required")
	v.Check(len(input.Comment) <= 255, "comment", "comment must not exceed 255 characters")

	if !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	commentWithUser, err := app.Models.Comments.Get(ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecordFound):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if commentWithUser.UserID != user.ID {
		app.authorizationRequiredResponse(w, r)
		return
	}

	comment := data.Comment{
		ID:        commentWithUser.ID,
		PostID:    commentWithUser.PostID,
		UserID:    commentWithUser.UserID,
		Comment:   input.Comment,
		CreatedAt: commentWithUser.CreatedAt,
		UpdatedAt: commentWithUser.UpdatedAt,
	}

	err = app.Models.Comments.Update(&comment)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecordFound):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"comment": comment}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getContextUser(r)

	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		app.notFoundErrorResponse(w, r)
		return
	}

	v := validator.New()

	v.Check(ID > 0, "id", "ID must be valid")

	if !v.Valid() {
		app.validationErrorResponse(w, r, v.Errors)
		return
	}

	comment, err := app.Models.Comments.Get(ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecordFound):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if user.ID != comment.UserID {
		app.authorizationRequiredResponse(w, r)
		return
	}

	err = app.Models.Comments.Delete(comment.ID, comment.UserID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecordFound):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "comment successfuly deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
