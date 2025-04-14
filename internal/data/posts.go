package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Post struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostWithUser struct {
	ID           int64     `json:"id"`
	UserId       int64     `json:"user_id"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updted_at"`
	Username     string    `json:"username"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	LikeCount    int64     `json:"like_count"`
	LikedByUser  bool      `json:"liked_by_user"`
	CommentCount int64     `json:"comment_count"`
}

type PostModel struct {
	DB *pgxpool.Pool
}

func (p PostModel) Insert(post *Post) error {
	query := `
		INSERT INTO posts (user_id, content)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{post.UserID, post.Content}

	return p.DB.QueryRow(ctx, query, args...).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
}

func (p PostModel) Get(postID int64) (*Post, error) {
	query := `
		SELECT id, user_id, content, created_at, updated_at FROM posts
		WHERE id = $1
		LIMIT 1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var post Post
	err := p.DB.QueryRow(ctx, query, postID).Scan(
		&post.ID,
		&post.UserID,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecordFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (p PostModel) GetAll(userID int64, filter Filter) ([]*PostWithUser, Metadata, error) {
	query := `
		WITH total AS(
			SELECT COUNT(*) AS total_count FROM posts WHERE user_id = $1
		)
		SELECT total.total_count, p.*, u.username, u.first_name, u.last_name
		FROM posts p INNER JOIN users u 
		ON p.user_id = u.id
		CROSS JOIN total
		WHERE u.id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{userID, filter.limit(), filter.offset()}

	rows, err := p.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	// posts, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[PostWithUser])
	// if err != nil {
	// 	return nil, err
	// }

	totalRecords := 0
	postsWithUser := []*PostWithUser{}

	for rows.Next() {
		var post PostWithUser

		err := rows.Scan(
			&totalRecords,
			&post.ID,
			&post.UserId,
			&post.Content,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Username,
			&post.FirstName,
			&post.LastName,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		postsWithUser = append(postsWithUser, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filter.Page, filter.PageSize)

	return postsWithUser, metadata, nil
}

func (p PostModel) Update(post *Post) error {
	query := `
		UPDATE posts SET content = $1, updated_at = $2
		WHERE id = $3
		RETURNING content, updated_at
	`

	updatedAt := time.Now().Local().UTC()

	args := []any{
		post.Content,
		updatedAt,
		post.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.DB.QueryRow(ctx, query, args...).Scan(
		&post.Content,
		&post.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (p PostModel) Delete(postID int64) error {
	query := `
		DELETE FROM posts
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := p.DB.Exec(ctx, query, postID)
	if err != nil {
		return err
	}

	if rows.RowsAffected() == 0 {
		return ErrNoRecordFound
	}

	return nil
}
