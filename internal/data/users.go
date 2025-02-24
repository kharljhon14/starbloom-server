package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kharljhon14/starbloom-server/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNoRecordFound     = errors.New("record not found")
	ErrEditConflict      = errors.New("edit conflict")
	ErrDuplicateEmail    = errors.New("email already in use")
	ErrDuplicateUsername = errors.New("username already exist")
)

type UserModel struct {
	DB *pgx.Conn
}

var AnonymousUser = &User{}

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Activated bool      `json:"activated"`
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
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case "users_username_key":
				return ErrDuplicateUsername
			case "users_email_key":
				return ErrDuplicateEmail
			}
		}
		return err
	}

	return nil
}

func (m UserModel) GetUser(username string) (*User, error) {
	query := `
	SELECT id, username, email, hashed_password, first_name, last_name, created_at
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
		&user.Password.hash,
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

func (p *password) Matches(plainTextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plainTextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Username != "", "username", "username is required")
	v.Check(len(user.Username) >= 5, "username", "username must be atleast 5 characters")
	v.Check(len(user.Username) <= 60, "username", "username must not exceed 60 characters")

	v.Check(user.Email != "", "email", "email is required")
	v.Check(len(user.Email) <= 255, "email", "email must not exceed 255 characters")
	v.Check(validator.Matches(user.Email, validator.EmailRX), "email", "must be a valid email")

	v.Check(user.FirstName != "", "first_name", "first name is required")
	v.Check(len(user.FirstName) <= 255, "first_name", "first name must not execeed 255 characters")

	v.Check(user.LastName != "", "last_name", "last name is required")
	v.Check(len(user.LastName) <= 255, "last_name", "last name must not execeed 255 characters")

	if user.Password.plainText != nil {
		ValidatePlainTextPassword(v, *user.Password.plainText)
	}
}

func ValidatePlainTextPassword(v *validator.Validator, password string) {
	v.Check(password != "", "password", "password is required")
	v.Check(len(password) >= 8, "password", "password must be atleast 8 characters")
	v.Check(len(password) <= 60, "password", "password must not exceed 60 characters")
}
