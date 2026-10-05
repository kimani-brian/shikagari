package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"github.com/shikagari/api/internal/dto"
	"github.com/shikagari/api/internal/repository/interfaces"
	"github.com/shikagari/api/pkg/hash"
	"github.com/shikagari/api/pkg/jwt"
)

const passwordResetTTL = 15 * time.Minute
const passwordResetTokenBytes = 32

// AuthService handles user registration and authentication.
type AuthService struct {
	userRepo          interfaces.UserRepository
	resetRepo         interfaces.PasswordResetTokenRepository
	sessionRepo       interfaces.SessionRepository
	securityEventRepo interfaces.SecurityEventRepository
	hasher            *hash.Password
	jwtManager        *jwt.Manager
}

// NewAuthService constructs an AuthService with its dependencies.
func NewAuthService(
	userRepo interfaces.UserRepository,
	resetRepo interfaces.PasswordResetTokenRepository,
	sessionRepo interfaces.SessionRepository,
	securityEventRepo interfaces.SecurityEventRepository,
	hasher *hash.Password,
	jwtManager *jwt.Manager,
) *AuthService {
	return &AuthService{
		userRepo:          userRepo,
		resetRepo:         resetRepo,
		sessionRepo:       sessionRepo,
		securityEventRepo: securityEventRepo,
		hasher:            hasher,
		jwtManager:        jwtManager,
	}
}

// Register creates a new user account.
// Returns the auth response (token + user) on success.
func (s *AuthService) Register(req dto.RegisterRequest, meta dto.SessionMetadata) (*dto.AuthResponse, error) {
	// ── 1. Check for duplicate email ──────────────────────────────────────────
	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("an account with this email address already exists")
	}

	// ── 2. Hash the password ──────────────────────────────────────────────────
	passwordHash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, errors.New("failed to process password")
	}

	// ── 3. Determine role (defaults to buyer) ─────────────────────────────────
	role := domain.RoleBuyer
	if req.Role == string(domain.RoleDealer) {
		role = domain.RoleDealer
	}

	// ── 4. Persist the new user ───────────────────────────────────────────────
	user := &domain.User{
		FullName:     req.FullName,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: passwordHash,
		Role:         role,
		IsVerified:   false,
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create account")
	}

	// ── 5. Create session + JWT ──────────────────────────────────────────────
	session, err := s.startSession(user, meta)
	if err != nil {
		return nil, errors.New("account created but failed to start session")
	}

	token, err := s.jwtManager.Generate(user, session.ID)
	if err != nil {
		return nil, errors.New("account created but failed to generate token")
	}

	s.recordSecurityEvent(user.ID, "Account created", "info", "New account registered")

	return &dto.AuthResponse{
		Token: token,
		User:  dto.ToUserResponse(*user),
	}, nil
}

// Login authenticates a user with email and password.
// Returns the auth response (token + user) on success.
func (s *AuthService) Login(req dto.LoginRequest, meta dto.SessionMetadata) (*dto.AuthResponse, error) {
	// ── 1. Find user by email ─────────────────────────────────────────────────
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("failed to process login")
	}
	if user == nil {
		// Generic message — do not reveal whether the email exists
		return nil, errors.New("invalid email or password")
	}

	// ── 2. Check account is active ────────────────────────────────────────────
	if !user.IsActive {
		return nil, errors.New("this account has been suspended. Please contact support")
	}

	// ── 3. Verify password ────────────────────────────────────────────────────
	if err := s.hasher.Verify(req.Password, user.PasswordHash); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// ── 4. Persist session + issue JWT ───────────────────────────────────────
	session, err := s.startSession(user, meta)
	if err != nil {
		return nil, errors.New("failed to create session")
	}

	token, err := s.jwtManager.Generate(user, session.ID)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	s.recordSecurityEvent(user.ID, "Login", "info", fmt.Sprintf("Signed in on %s", session.Device))

	return &dto.AuthResponse{
		Token: token,
		User:  dto.ToUserResponse(*user),
	}, nil
}

// RequestPasswordReset generates a short-lived token and (eventually) emails it.
// To prevent user enumeration the method always returns nil errors for unknown emails.
func (s *AuthService) RequestPasswordReset(req dto.PasswordResetRequest) error {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return errors.New("failed to process password reset request")
	}
	if user == nil {
		return nil
	}

	if err := s.resetRepo.DeleteByUser(user.ID); err != nil {
		return errors.New("failed to process password reset request")
	}

	token, err := generatePasswordResetToken()
	if err != nil {
		return errors.New("failed to process password reset request")
	}

	reset := &domain.PasswordResetToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(passwordResetTTL),
	}

	if err := s.resetRepo.Create(reset); err != nil {
		return errors.New("failed to process password reset request")
	}

	return nil
}

// ResetPassword validates the token and updates the user's password.
func (s *AuthService) ResetPassword(req dto.PasswordResetConfirmRequest) error {
	resetToken, err := s.resetRepo.FindByToken(req.Token)
	if err != nil || resetToken == nil {
		return errors.New("invalid or expired token")
	}

	if resetToken.UsedAt != nil || time.Now().After(resetToken.ExpiresAt) {
		return errors.New("invalid or expired token")
	}

	user, err := s.userRepo.FindByID(resetToken.UserID)
	if err != nil || user == nil {
		return errors.New("invalid or expired token")
	}

	passwordHash, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		return errors.New("failed to process new password")
	}

	if err := s.userRepo.UpdatePassword(user.ID, passwordHash); err != nil {
		return errors.New("failed to reset password")
	}

	if err := s.resetRepo.MarkUsed(resetToken.ID, time.Now()); err != nil {
		return errors.New("failed to finalize password reset")
	}

	// Clean up any lingering tokens for this user, ignore errors to avoid masking success.
	_ = s.resetRepo.DeleteByUser(user.ID)

	s.recordSecurityEvent(user.ID, "Password reset", "info", "Password reset via email link")

	return nil
}

// ListSecurityEvents returns recent persisted security events for the authenticated user.
func (s *AuthService) ListSecurityEvents(userID uuid.UUID) ([]dto.SecurityEventResponse, error) {
	records, err := s.securityEventRepo.ListRecentByUser(userID, 20)
	if err != nil {
		return nil, errors.New("failed to load security events")
	}

	resp := make([]dto.SecurityEventResponse, 0, len(records))
	for _, event := range records {
		resp = append(resp, dto.SecurityEventResponse{
			ID:        event.ID.String(),
			Label:     event.Label,
			Timestamp: event.CreatedAt.Format(time.RFC3339),
			Severity:  event.Severity,
			Details:   event.Details,
		})
	}

	return resp, nil
}

// ListSessions returns the active sessions for the authenticated user.
func (s *AuthService) ListSessions(userID uuid.UUID, currentSessionID uuid.UUID) ([]dto.SessionResponse, error) {
	records, err := s.sessionRepo.ListByUser(userID)
	if err != nil {
		return nil, errors.New("failed to load sessions")
	}

	resp := make([]dto.SessionResponse, 0, len(records))
	for _, session := range records {
		if session.RevokedAt != nil {
			continue
		}
		resp = append(resp, dto.SessionResponse{
			ID:         session.ID.String(),
			Device:     session.Device,
			Browser:    session.Browser,
			Location:   session.Location,
			LastActive: session.LastActive.Format(time.RFC3339),
			Current:    session.ID == currentSessionID,
		})
	}

	return resp, nil
}

// RevokeSession removes an active session for a user.
func (s *AuthService) RevokeSession(userID uuid.UUID, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return errors.New("failed to locate session")
	}
	if session == nil || session.UserID != userID {
		return errors.New("session not found")
	}
	if session.RevokedAt != nil {
		return nil
	}

	if err := s.sessionRepo.MarkRevoked(userID, sessionID, "revoked_by_user"); err != nil {
		return errors.New("failed to revoke session")
	}

	s.recordSecurityEvent(userID, "Session revoked", "warning", fmt.Sprintf("Signed out %s", session.Device))
	return nil
}

// Logout invalidates the current user's active session.
func (s *AuthService) Logout(userID uuid.UUID, sessionID uuid.UUID) error {
	if sessionID == uuid.Nil {
		return errors.New("session not found")
	}

	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return errors.New("failed to locate session")
	}
	if session == nil || session.UserID != userID {
		return errors.New("session not found")
	}
	if session.RevokedAt != nil {
		return nil
	}

	if err := s.sessionRepo.MarkRevoked(userID, sessionID, "logout"); err != nil {
		return errors.New("failed to sign out")
	}

	s.recordSecurityEvent(userID, "Logout", "info", fmt.Sprintf("Signed out %s", session.Device))
	return nil
}

func (s *AuthService) startSession(user *domain.User, meta dto.SessionMetadata) (*domain.UserSession, error) {
	if user == nil {
		return nil, errors.New("user required")
	}

	device := coalesce(meta.Device, deriveDevice(meta.UserAgent))
	browser := coalesce(meta.Browser, deriveBrowser(meta.UserAgent))
	location := coalesce(meta.Location, coalesce(meta.IPAddress, "Unknown location"))
	ip := coalesce(meta.IPAddress, "unknown")

	session := &domain.UserSession{
		UserID:     user.ID,
		Device:     device,
		Browser:    browser,
		Location:   location,
		IPAddress:  ip,
		UserAgent:  meta.UserAgent,
		LastActive: time.Now(),
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *AuthService) recordSecurityEvent(userID uuid.UUID, label, severity, details string) {
	if s.securityEventRepo == nil || userID == uuid.Nil {
		return
	}

	event := &domain.SecurityEvent{
		UserID:   userID,
		Label:    label,
		Severity: severity,
		Details:  details,
	}

	_ = s.securityEventRepo.Create(event)
}

func deriveDevice(userAgent string) string {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "iphone"):
		return "iPhone"
	case strings.Contains(ua, "ipad"):
		return "iPad"
	case strings.Contains(ua, "android"):
		return "Android"
	case strings.Contains(ua, "windows"):
		return "Windows PC"
	case strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os"):
		return "Mac"
	default:
		if userAgent == "" {
			return "Unknown device"
		}
		return userAgent
	}
}

func deriveBrowser(userAgent string) string {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "chrome") && !strings.Contains(ua, "edge"):
		return "Chrome"
	case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
		return "Safari"
	case strings.Contains(ua, "firefox"):
		return "Firefox"
	case strings.Contains(ua, "edg"):
		return "Edge"
	default:
		if userAgent == "" {
			return "Unknown browser"
		}
		return userAgent
	}
}

func coalesce(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func generatePasswordResetToken() (string, error) {
	buf := make([]byte, passwordResetTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
