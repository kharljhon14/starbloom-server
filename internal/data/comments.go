package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrInvalidPostID = errors.New("invalid post_id")

type Comment struct {
	ID        int64     `json:"id"`
	PostID    int64     `json:"post_id"`
	UserID    int64     `json:"user_id"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CommentModel struct {
	DB *pgx.Conn
}

func (c CommentModel) Insert(comment *Comment) error {
	query := `
		INSERT INTO comments (post_id, user_id, comment)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		comment.PostID,
		comment.UserID,
		comment.Comment,
	}

	err := c.DB.QueryRow(ctx, query, args...).Scan(
		&comment.ID,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case "comments_post_id_fkey":
				return ErrInvalidPostID
			default:
				return err
			}

		}
		return err
	}

	return nil
}
