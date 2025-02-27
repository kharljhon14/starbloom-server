package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrAlreadyLiked = errors.New("already liked")

type Like struct {
	PostID    int64     `json:"post_id"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type LikeModel struct {
	DB *pgx.Conn
}

func (l LikeModel) Insert(like *Like) error {
	query := `
		INSERT INTO likes (post_id, user_id)
		VALUES ($1, $2)
		RETURNING created_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{like.PostID, like.UserID}

	err := l.DB.QueryRow(ctx, query, args...).Scan(&like.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case "unique_like":
				return ErrAlreadyLiked
			default:
				return pgErr
			}
		}

		return err
	}

	return nil
}

func (l LikeModel) Delete(postID, userID int64) error {
	query := `
		DELETE FROM likes WHERE post_id = $1 AND user_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row, err := l.DB.Exec(ctx, query, postID, userID)
	if err != nil {
		return err
	}

	if row.RowsAffected() == 0 {
		return ErrNoRecordFound
	}

	return nil
}
