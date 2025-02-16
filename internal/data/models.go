package data

import (
	"github.com/jackc/pgx/v5"
)

type Models struct {
	Users UserModel
}

func NewModels(db *pgx.Conn) Models {
	return Models{
		Users: UserModel{DB: db},
	}
}
