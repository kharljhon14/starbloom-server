package data

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Models struct {
	Users    UserModel
	Tokens   TokenModel
	Follows  FollowsModel
	Posts    PostModel
	Likes    LikeModel
	Comments CommentModel
}

func NewModels(db *pgxpool.Pool) Models {
	return Models{
		Users:    UserModel{DB: db},
		Tokens:   TokenModel{DB: db},
		Follows:  FollowsModel{DB: db},
		Posts:    PostModel{DB: db},
		Likes:    LikeModel{DB: db},
		Comments: CommentModel{DB: db},
	}
}
