package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kharljhon14/starbloom-server/internal/validator"
)

const (
	ScopeAuthentication = "authentication"
	ScopeAuthorization  = "authorization"
)

type Token struct {
	PlainText string    `json:"plain_text"`
	Hash      []byte    `json:"-"`
	ExpiredAt time.Time `json:"expired_at"`
	UserID    int64     `json:"-"`
	Scope     string    `json:"-"`
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID:    userID,
		ExpiredAt: time.Now().Add(ttl),
		Scope:     scope,
	}

	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.
		WithPadding(base32.NoPadding).
		EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(token.PlainText))

	token.Hash = hash[:]

	return token, nil
}

func ValidateTokenPlainText(v *validator.Validator, tokenPlainText string) {
	v.Check(tokenPlainText != "", "token", "token must be provided")
	v.Check(len(tokenPlainText) == 26, "token", "must be 26 bytes long")
}

type TokenModel struct {
	DB *pgxpool.Pool
}

func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.insert(token)
	if err != nil {
		return nil, err
	}

	return token, err
}

func (m TokenModel) insert(token *Token) error {
	query := `
	INSERT INTO tokens (hash, user_id, expired_at, scope)
	VALUES ($1, $2, $3, $4)
	`

	args := []interface{}{
		token.Hash,
		token.UserID,
		token.ExpiredAt,
		token.Scope,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, query, args...)

	return err
}

func (m TokenModel) GetForToken(tokenScope, tokenPlainText string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlainText))

	query := `
		SELECT 
		u.id, 
		u.username,
		u.email,
		u.first_name,
		u.last_name,
		u.hashed_password,
		u.created_at,
		u.activated
		FROM users u
		INNER JOIN tokens t
		ON u.id = t.user_id
		WHERE t.hash = $1
		AND t.scope = $2
		AND t.expired_at > $3
	`

	args := []any{
		tokenHash[:],
		tokenScope,
		time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User

	err := m.DB.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password.hash,
		&user.CreatedAt,
		&user.Activated,
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

func (m TokenModel) DeleteAllForUser(scope string, userId int64) error {
	query := `
	DELETE FROM tokens
	WHERE scope = $1 AND user_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, query, scope, userId)
	return err
}
