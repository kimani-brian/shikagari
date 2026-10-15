// Package sso mints and verifies single-use handoff tickets that carry a
// marketplace identity into the auction engine.
//
// A ticket is v1.<base64url(payload)>.<base64url(hmac-sha256)>, where the
// payload names the user, their role, an expiry 120 seconds out, and a
// unique id. The engine verifies the signature with the shared secret, checks
// the expiry, and consumes the id exactly once — so a stolen ticket is
// near-useless and a replayed one is detectable.
package sso

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	// TicketVersion prefixes every ticket so the format can evolve.
	TicketVersion = "v1"
	// TicketLifetime bounds how long a minted ticket is redeemable.
	TicketLifetime = 120 * time.Second
)

// Ticket is the verified payload of a handoff ticket.
type Ticket struct {
	UserID     string    `json:"user_id"`
	Role       string    `json:"role"`
	Email      string    `json:"email"`
	IsVerified bool      `json:"is_verified"`
	Expiry     time.Time `json:"exp"`
	JTI        string    `json:"jti"`
}

var (
	// ErrTicketMalformed is returned when the ticket does not parse.
	ErrTicketMalformed = errors.New("malformed ticket")
	// ErrTicketSignature is returned when the signature does not verify.
	ErrTicketSignature = errors.New("invalid ticket signature")
	// ErrTicketExpired is returned when the ticket is past its lifetime.
	ErrTicketExpired = errors.New("ticket expired")
)

// Minter signs tickets with the shared secret.
type Minter struct {
	secret []byte
}

// NewMinter builds a minter. An empty secret is refused: an unsigned ticket
// is worse than no tickets.
func NewMinter(secret string) (*Minter, error) {
	if secret == "" {
		return nil, errors.New("sso shared secret is required")
	}
	return &Minter{secret: []byte(secret)}, nil
}

// Mint issues a ticket for the given identity, valid for TicketLifetime.
func (m *Minter) Mint(userID, role, email string, verified bool, now time.Time) (string, *Ticket, error) {
	jtiBytes := make([]byte, 16)
	if _, err := rand.Read(jtiBytes); err != nil {
		return "", nil, fmt.Errorf("jti: %w", err)
	}
	t := &Ticket{
		UserID:     userID,
		Role:       role,
		Email:      email,
		IsVerified: verified,
		Expiry:     now.Add(TicketLifetime).UTC(),
		JTI:        uuid.NewString(),
	}
	payload, err := json.Marshal(t)
	if err != nil {
		return "", nil, err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(TicketVersion + "." + encoded))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return TicketVersion + "." + encoded + "." + sig, t, nil
}

// Verify checks the structure, signature and expiry of a ticket. One-time use
// is enforced by the consumer (the engine records the JTI in Redis).
func (m *Minter) Verify(raw string, now time.Time) (*Ticket, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 || parts[0] != TicketVersion {
		return nil, ErrTicketMalformed
	}
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	want := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	got, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrTicketMalformed
	}
	wantBytes, _ := base64.RawURLEncoding.DecodeString(want)
	if !hmac.Equal(got, wantBytes) {
		return nil, ErrTicketSignature
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrTicketMalformed
	}
	var t Ticket
	if err := json.Unmarshal(payload, &t); err != nil {
		return nil, ErrTicketMalformed
	}
	if t.UserID == "" || t.JTI == "" {
		return nil, ErrTicketMalformed
	}
	if now.After(t.Expiry) {
		return nil, ErrTicketExpired
	}
	return &t, nil
}
