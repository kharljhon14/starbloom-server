package api

import (
	"errors"
	"net/http"

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
