package sso

import (
	"strings"
	"testing"
	"time"
)

func TestMintVerifyRoundTrip(t *testing.T) {
	m, err := NewMinter("test-shared-secret-at-least-32-chars!")
	if err != nil {
		t.Fatalf("minter: %v", err)
	}
	now := time.Now()
	raw, issued, err := m.Mint("user-1", "buyer", "buyer@test.ke", true, now)
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if !strings.HasPrefix(raw, "v1.") {
		t.Errorf("ticket should be versioned, got %q", raw[:10])
	}
	got, err := m.Verify(raw, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got.UserID != "user-1" || got.Role != "buyer" || got.Email != "buyer@test.ke" {
		t.Errorf("payload changed in transit: %+v", got)
	}
	if got.JTI != issued.JTI {
		t.Error("jti must survive the round trip")
	}
	if !got.Expiry.After(now.Add(119 * time.Second)) {
		t.Errorf("lifetime should be ~120s, expiry=%v", got.Expiry)
	}
}

func TestVerifyRejectsTampering(t *testing.T) {
	m, _ := NewMinter("test-shared-secret-at-least-32-chars!")
	raw, _, _ := m.Mint("user-1", "buyer", "buyer@test.ke", true, time.Now())
	// Flip the role inside the payload without re-signing.
	parts := strings.Split(raw, ".")
	parts[1] = parts[1][:len(parts[1])-4] + "AAAA"
	if _, err := m.Verify(strings.Join(parts, "."), time.Now()); err != ErrTicketSignature && err != ErrTicketMalformed {
		t.Fatalf("tampered ticket must be rejected, got %v", err)
	}
}

func TestVerifyRejectsWrongSecret(t *testing.T) {
	m1, _ := NewMinter("test-shared-secret-at-least-32-chars!")
	m2, _ := NewMinter("a-different-secret-also-long-enough!!")
	raw, _, _ := m1.Mint("user-1", "buyer", "buyer@test.ke", true, time.Now())
	if _, err := m2.Verify(raw, time.Now()); err != ErrTicketSignature {
		t.Fatalf("wrong secret must fail signature, got %v", err)
	}
}

func TestVerifyRejectsExpiry(t *testing.T) {
	m, _ := NewMinter("test-shared-secret-at-least-32-chars!")
	raw, _, _ := m.Mint("user-1", "buyer", "buyer@test.ke", true, time.Now().Add(-3*time.Minute))
	if _, err := m.Verify(raw, time.Now()); err != ErrTicketExpired {
		t.Fatalf("stale ticket must be rejected, got %v", err)
	}
}

func TestMinterRefusesEmptySecret(t *testing.T) {
	if _, err := NewMinter(""); err == nil {
		t.Fatal("empty secret must be refused")
	}
}

func TestTicketsAreUnique(t *testing.T) {
	m, _ := NewMinter("test-shared-secret-at-least-32-chars!")
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		raw, issued, err := m.Mint("user-1", "buyer", "buyer@test.ke", true, time.Now())
		if err != nil {
			t.Fatalf("mint: %v", err)
		}
		if seen[issued.JTI] || seen[raw] {
			t.Fatal("tickets must be unique (replay safety depends on it)")
		}
		seen[issued.JTI] = true
		seen[raw] = true
	}
}
