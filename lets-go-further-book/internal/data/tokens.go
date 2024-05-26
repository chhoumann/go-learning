package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"greenlight.bagerbach.com/internal/validator"
)

const (
	ScopeActivation = "activation"
)

type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scope     string
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// 16 doesn't mean the plaintext tokens are 16 characters long, but that they have an underlying entropy of 16 bytes of randomness
	// The length of the plaintext token depends on the 16 random bytes are encoded to create a string.
	// Since we'll encode them to a base-32 string, it'll be 26 characters long
	randomBytes := make([]byte, 16)
	// Fill randomBytes with random data using the OS' CSPRNG
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}

	// We encode the random bytes to a base32-encoded string
	// By default, base-32 strings may be padded with '=' characters, so we use WithPadding to remove them, we don't need them
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// Now we generate the hash of the plaintext token.
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:] // It returned an array, so we convert it to a slice to make it easier to work with

	return token, nil
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 characters long")
}

type TokenModel struct {
	DB *sql.DB
}

func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(token *Token) error {
	query := `
		INSERT INTO tokens (hash, user_id, expiry, scope)
		VALUES ($1, $2, $3, $4)
	`
	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

func (m TokenModel) DeleteAllForUser(userID int64, scope string) error {
	query := `
		DELETE FROM tokens
		WHERE user_id = $1 AND scope = $2
	`
	args := []interface{}{userID, scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}
