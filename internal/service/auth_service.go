package service

import (
	"errors"

	"github.com/shikagari/api/internal/domain"
	"github.com/shikagari/api/internal/dto"
	"github.com/shikagari/api/internal/repository/interfaces"
	"github.com/shikagari/api/pkg/hash"
	"github.com/shikagari/api/pkg/jwt"
)

// AuthService handles user registration and authentication.
type AuthService struct {
	userRepo   interfaces.UserRepository
	hasher     *hash.Password
	jwtManager *jwt.Manager
}

// NewAuthService constructs an AuthService with its dependencies.
func NewAuthService(
	userRepo interfaces.UserRepository,
	hasher *hash.Password,
	jwtManager *jwt.Manager,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		hasher:     hasher,
		jwtManager: jwtManager,
	}
}

// Register creates a new user account.
// Returns the auth response (token + user) on success.
func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
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
	if req.Role == string(domain.RoleSeller) {
		role = domain.RoleSeller
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

	// ── 5. Generate JWT ───────────────────────────────────────────────────────
	token, err := s.jwtManager.Generate(user)
	if err != nil {
		return nil, errors.New("account created but failed to generate token")
	}

	return &dto.AuthResponse{
		Token: token,
		User:  dto.ToUserResponse(*user),
	}, nil
}

// Login authenticates a user with email and password.
// Returns the auth response (token + user) on success.
func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
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

	// ── 4. Generate JWT ───────────────────────────────────────────────────────
	token, err := s.jwtManager.Generate(user)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &dto.AuthResponse{
		Token: token,
		User:  dto.ToUserResponse(*user),
	}, nil
}
