package data

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNoRecordFound = errors.New("record not found")
)

type UserModel struct {
	DB *pgx.Conn
}

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Activaed  bool      `json:"activated"`
	Password  password  `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type password struct {
	plainText *string
	hash      []byte
}

func (m UserModel) Insert(user *User) error {
	query := `
	INSERT INTO users (username, email, first_name, last_name, hashed_password)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at
	`

	args := []interface{}{
		user.Username,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Password.hash,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, args...).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (m UserModel) GetUser(username string) (*User, error) {
	query := `
	SELECT id, username, email, first_name, last_name, created_at
	FROM users
	WHERE username = $1
	`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecordFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (p *password) Set(plainTextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 12)
	if err != nil {
		return err
	}

	p.plainText = &plainTextPassword
	p.hash = hash

	return nil
}
