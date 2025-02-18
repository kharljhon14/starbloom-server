package data

import (
	"github.com/jackc/pgx/v5"
)

type Models struct {
	Users  UserModel
	Tokens TokenModel
}

func NewModels(db *pgx.Conn) Models {
	return Models{
		Users:  UserModel{DB: db},
		Tokens: TokenModel{DB: db},
	}
}
