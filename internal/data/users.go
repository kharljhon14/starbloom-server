package data

import (
	"time"

	"github.com/jackc/pgx/v5"
)

type UserModel struct {
	DB *pgx.Conn
}

type User struct {
	ID             int64     `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Activaed       bool      `json:"activated"`
	HashedPassword string    `json:"_"`
	CreatedAt      time.Time `json:"created_at"`
}
