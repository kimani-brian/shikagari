package hash

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const (
	// bcryptCost defines the work factor for bcrypt hashing.
	// Cost 12 is a good balance between security and performance.
	bcryptCost = 12
)

// Password wraps bcrypt operations for hashing and verification.
type Password struct{}

// New returns a new Password utility instance.
func New() *Password {
	return &Password{}
}

// Hash takes a plain-text password and returns a bcrypt hash.
// The hash is safe to store directly in the database.
func (p *Password) Hash(plain string) (string, error) {
	if plain == "" {
		return "", errors.New("password cannot be empty")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

// Verify compares a plain-text password against a stored bcrypt hash.
// Returns nil if the password matches, or an error if it does not.
func (p *Password) Verify(plain, hashed string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}
