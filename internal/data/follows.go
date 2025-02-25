package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrAlreadyFollowing = errors.New("already following")
)

type Follow struct {
	UserID     int64     `json:"user_id"`
	FollowerID int64     `json:"follower_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type FollowsModel struct {
	DB *pgx.Conn
}

func (f FollowsModel) Insert(userID, followerID int64) (*Follow, error) {

	query := `
	INSERT INTO follows (user_id, follower_id)
	VALUES ($1, $2)
	RETURNING created_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	follow := Follow{
		UserID:     userID,
		FollowerID: followerID,
	}

	err := f.DB.QueryRow(ctx, query, userID, followerID).Scan(&follow.CreatedAt)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			switch pgError.ConstraintName {
			case "unique_follow":
				return nil, ErrAlreadyFollowing
			default:
				return nil, err
			}
		}

		return nil, err

	}

	return &follow, nil
}

func (f FollowsModel) Delete(userID, followerID int64) error {
	query := `DELETE FROM follows WHERE user_id = $1 AND follower_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := f.DB.Exec(ctx, query, userID, followerID)
	if err != nil {
		return err
	}

	rows := res.RowsAffected()
	if rows == 0 {
		return ErrNoRecordFound
	}

	return nil
}

type FollowUser struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (f FollowsModel) GetFollowers(userID int64) ([]*FollowUser, error) {

	query := `
		SELECT u.id, u.username, u.first_name, u.last_name from users
		u INNER JOIN follows f ON u.id = f.follower_id
		WHERE user_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := f.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	users, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[FollowUser])
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (f FollowsModel) GetFollowingPosts(userID int64, limit, offset int) ([]*PostWithUser, error) {
	query := `
		SELECT p.id, p.user_id, p.content, p.created_at, p.updated_at,
		u.first_name, u.last_name
		FROM posts p INNER JOIN users u ON p.user_id = u.id
		WHERE user_id IN (SELECT user_id FROM follows where follower_id = $1)
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`

	args := []any{userID, limit, offset}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := f.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	posts, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[PostWithUser])
	if err != nil {
		return nil, err
	}

	return posts, nil

}
