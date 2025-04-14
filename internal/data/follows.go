package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
	DB *pgxpool.Pool
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

func (f FollowsModel) GetFollowers(userID int64, filters Filter) ([]*FollowUser, Metadata, error) {
	query := `
		WITH total AS(
			SELECT COUNT(*) AS total_count FROM follows WHERE user_id = $1
		)
		SELECT total.total_count, u.id, u.username, u.first_name, u.last_name from users
		u INNER JOIN follows f ON u.id = f.follower_id
		CROSS JOIN total
		WHERE user_id = $1 LIMIT $2 OFFSET $3
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{userID, filters.limit(), filters.offset()}

	rows, err := f.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	users := []*FollowUser{}

	// users, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[FollowUser])
	// if err != nil {
	// 	return nil, Metadata{}, err
	// }

	for rows.Next() {
		var user FollowUser

		err := rows.Scan(
			&totalRecords,
			&user.UserID,
			&user.Username,
			&user.FirstName,
			&user.LastName,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return users, metadata, nil
}

func (f FollowsModel) GetFollowingPosts(userID int64, filters Filter) ([]*PostWithUser, Metadata, error) {
	query := `
		WITH total AS(
			SELECT COUNT(*) AS total_count FROM posts WHERE user_id IN (SELECT user_id FROM follows WHERE follower_id = $1)
		),
		like_counts AS(
			SELECT post_id, COUNT(*) AS like_count
			FROM likes
			GROUP BY post_id
		),
		user_likes AS (
   			SELECT post_id 
    		FROM likes 
   			WHERE user_id = $1
		),
		comment_counts AS(
			SELECT post_id, COUNT(*) AS comment_count
			FROM comments
			GROUP BY post_id
		)
		SELECT total.total_count, p.id, p.user_id, p.content, p.created_at, p.updated_at,
		u.username, u.first_name, u.last_name, COALESCE(l.like_count, 0) AS like_count,
		CASE WHEN ul.post_id IS NOT NULL THEN true ELSE false END AS liked_by_user,
		COALESCE(c.comment_count, 0) AS comment_count
		FROM posts p INNER JOIN users u ON p.user_id = u.id
		LEFT JOIN like_counts l ON p.id = l.post_id
		LEFT JOIN user_likes ul ON p.id = ul.post_id
		LEFT JOIN comment_counts c ON p.id = c.post_id
		CROSS JOIN total
		WHERE user_id IN (SELECT user_id FROM follows where follower_id = $1) OR p.user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`

	args := []any{userID, filters.limit(), filters.offset()}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := f.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	// posts, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[PostWithUser])
	// if err != nil {
	// 	return nil, err
	// }

	totalRecords := 0
	posts := []*PostWithUser{}

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
			&post.LikeCount,
			&post.LikedByUser,
			&post.CommentCount,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return posts, metadata, nil

}
