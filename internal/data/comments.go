package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
	DB *pgxpool.Pool
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

type CommentWithUser struct {
	ID        int64     `json:"id"`
	PostID    int64     `json:"post_id"`
	UserID    int64     `json:"user_id"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
}

func (c CommentModel) Get(commentID int64) (*CommentWithUser, error) {
	query := `
		SELECT c.id, c.post_id, c.user_id, c.comment, c.created_at, c.updated_at, u.username, u.first_name, u.last_name FROM comments c
		INNER JOIN users u ON c.user_id = u.id
		WHERE c.id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var comment CommentWithUser
	err := c.DB.QueryRow(ctx, query, commentID).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.Comment,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&comment.Username,
		&comment.FirstName,
		&comment.LastName,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecordFound
		default:
			return nil, err
		}
	}

	return &comment, nil
}

func (c CommentModel) GetCommentsByPost(postID int64, filters Filter) ([]*CommentWithUser, Metadata, error) {
	query := `
		WITH total AS (
			SELECT COUNT(*) AS total_count FROM comments WHERE post_id = $1
		)
		SELECT 
			total.total_count,
			c.id, c.post_id, c.user_id, c.comment, c.created_at, c.updated_at,
			u.username, u.first_name, u.last_name
		FROM comments c
		INNER JOIN users u ON c.user_id = u.id
		CROSS JOIN total
		WHERE c.post_id = $1
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.Query(ctx, query, postID, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	commentsWithUser := []*CommentWithUser{}

	// comments, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[CommentWithUser])
	// if err != nil {
	// 	return nil, Metadata{}, err
	// }

	for rows.Next() {
		var comment CommentWithUser

		err := rows.Scan(
			&totalRecords,
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Comment,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.Username,
			&comment.FirstName,
			&comment.LastName,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		commentsWithUser = append(commentsWithUser, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return commentsWithUser, metadata, nil
}

func (c CommentModel) Update(comment *Comment) error {
	query := `
		UPDATE comments SET comment = $1, updated_at = $2
		WHERE id = $3
		RETURNING comment, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		comment.Comment,
		time.Now().Local().UTC(),
		comment.ID,
	}

	err := c.DB.QueryRow(ctx, query, args...).Scan(
		&comment.Comment,
		&comment.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNoRecordFound
		default:
			return err
		}
	}

	return nil
}

func (c CommentModel) Delete(commentID, userID int64) error {
	query := `
		DELETE FROM comments
		WHERE id = $1 AND user_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.Exec(ctx, query, commentID, userID)
	if err != nil {
		return err
	}

	if rows.RowsAffected() == 0 {
		return ErrNoRecordFound
	}

	return nil
}
