package service

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"github.com/shikagari/api/internal/dto"
	"github.com/shikagari/api/internal/repository/interfaces"
	"github.com/shikagari/api/pkg/hash"
	"github.com/shikagari/api/pkg/response"
)

// UserService handles user profile management and admin user operations.
type UserService struct {
	userRepo interfaces.UserRepository
	hasher   *hash.Password
}

// NewUserService constructs a UserService with its dependencies.
func NewUserService(
	userRepo interfaces.UserRepository,
	hasher *hash.Password,
) *UserService {
	return &UserService{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

// GetProfile retrieves the authenticated user's own profile.
func (s *UserService) GetProfile(userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("failed to retrieve profile")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	res := dto.ToUserResponse(*user)
	return &res, nil
}

// UpdateProfile updates the authenticated user's full name, phone, and email.
// Email changes are checked for uniqueness; anything else is left untouched.
func (s *UserService) UpdateProfile(userID uuid.UUID, req dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Email != "" && !strings.EqualFold(req.Email, user.Email) {
		taken, err := s.userRepo.ExistsByEmail(req.Email)
		if err != nil {
			return nil, errors.New("failed to validate email address")
		}
		if taken {
			return nil, errors.New("this email address is already in use")
		}
		user.Email = req.Email
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, errors.New("failed to update profile")
	}

	res := dto.ToUserResponse(*user)
	return &res, nil
}

// ChangePassword verifies the current password and sets a new one.
func (s *UserService) ChangePassword(userID uuid.UUID, req dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// ── Verify current password ───────────────────────────────────────────────
	if err := s.hasher.Verify(req.CurrentPassword, user.PasswordHash); err != nil {
		return errors.New("current password is incorrect")
	}

	// ── Hash and persist new password ─────────────────────────────────────────
	newHash, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		return errors.New("failed to process new password")
	}

	return s.userRepo.UpdatePassword(userID, newHash)
}

// ── Admin Operations ──────────────────────────────────────────────────────────

// AdminListUsers returns a paginated list of all platform users.
func (s *UserService) AdminListUsers(page, perPage int) ([]dto.UserResponse, int64, error) {
	users, total, err := s.userRepo.List(page, perPage)
	if err != nil {
		return nil, 0, errors.New("failed to retrieve users")
	}

	var result []dto.UserResponse
	for _, u := range users {
		result = append(result, dto.ToUserResponse(u))
	}
	return result, total, nil
}

// AdminGetUser retrieves any user by ID (admin only).
func (s *UserService) AdminGetUser(userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	res := dto.ToUserResponse(*user)
	return &res, nil
}

// AdminUpdateUser allows an admin to change a user's role, verified badge,
// or active status.
func (s *UserService) AdminUpdateUser(
	targetID uuid.UUID,
	req dto.AdminUpdateUserRequest,
) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(targetID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	if req.Role != nil {
		if err := s.userRepo.SetRole(targetID, domain.UserRole(*req.Role)); err != nil {
			return nil, errors.New("failed to update role")
		}
		user.Role = domain.UserRole(*req.Role)
	}

	if req.IsVerified != nil {
		if err := s.userRepo.SetVerified(targetID, *req.IsVerified); err != nil {
			return nil, errors.New("failed to update verified status")
		}
		user.IsVerified = *req.IsVerified
	}

	if req.IsActive != nil {
		if err := s.userRepo.SetActive(targetID, *req.IsActive); err != nil {
			return nil, errors.New("failed to update active status")
		}
		user.IsActive = *req.IsActive
	}

	res := dto.ToUserResponse(*user)
	return &res, nil
}

// AdminDeleteUser soft-deletes a user account.
func (s *UserService) AdminDeleteUser(targetID uuid.UUID) error {
	user, err := s.userRepo.FindByID(targetID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// Prevent deletion of admin accounts via this endpoint
	if user.Role == domain.RoleAdmin {
		return errors.New("admin accounts cannot be deleted through this endpoint")
	}

	return s.userRepo.Delete(targetID)
}

// calculateTotalPages is a shared pagination helper.
func calculateTotalPages(total int64, perPage int) int {
	pages := int(total) / perPage
	if int(total)%perPage != 0 {
		pages++
	}
	return pages
}

// buildMeta constructs a response.Meta for paginated endpoints.
func buildMeta(page, perPage int, total int64) *response.Meta {
	return &response.Meta{
		Page:       page,
		PerPage:    perPage,
		TotalItems: total,
		TotalPages: calculateTotalPages(total, perPage),
	}
}
