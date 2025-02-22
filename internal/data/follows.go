package data

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
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
