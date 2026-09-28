package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"github.com/shikagari/api/internal/dto"
	"github.com/shikagari/api/internal/repository/interfaces"
)

// PrivateSellerService handles private seller profile lifecycle and admin approval.
type PrivateSellerService struct {
	sellerRepo interfaces.PrivateSellerRepository
	userRepo   interfaces.UserRepository
}

// NewPrivateSellerService constructs a PrivateSellerService with its dependencies.
func NewPrivateSellerService(
	sellerRepo interfaces.PrivateSellerRepository,
	userRepo interfaces.UserRepository,
) *PrivateSellerService {
	return &PrivateSellerService{
		sellerRepo: sellerRepo,
		userRepo:   userRepo,
	}
}

// CreateProfile creates a private seller profile for the authenticated user.
func (s *PrivateSellerService) CreateProfile(
	userID uuid.UUID,
	req dto.CreatePrivateSellerRequest,
) (*dto.PrivateSellerProfileResponse, error) {
	// ── 1. Validate user role ─────────────────────────────────────────────────
	// Buyers may also apply: this is the self-serve upgrade path (the buyer
	// keeps the buyer role until an admin reviews the profile and flips the
	// role to seller). Dealers use the dealer profile flow instead.
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	hasAccess := user.Role == domain.RoleBuyer || user.Role == domain.RoleSeller || user.Role == domain.RoleAdmin
	if !hasAccess {
		return nil, errors.New("your account is not permitted to request private seller approval")
	}

	// ── 2. Prevent duplicate profiles ────────────────────────────────────────
	exists, err := s.sellerRepo.ExistsByUserID(userID)
	if err != nil {
		return nil, errors.New("failed to check existing profile")
	}
	if exists {
		return nil, errors.New("you already have a private seller profile")
	}

	// ── 3. Validate National ID uniqueness ────────────────────────────────────
	idExists, err := s.sellerRepo.ExistsByNationalID(req.NationalIDNo)
	if err != nil {
		return nil, errors.New("failed to validate national ID")
	}
	if idExists {
		return nil, errors.New("a profile with this national ID number already exists")
	}

	// ── 4. Persist the profile ────────────────────────────────────────────────
	profile := &domain.PrivateSellerProfile{
		UserID:          userID,
		NationalIDNo:    req.NationalIDNo,
		Location:        req.Location,
		Address:         req.Address,
		ProfilePhotoURL: req.ProfilePhotoURL,
		Bio:             req.Bio,
		ApprovalStatus:  domain.ApprovalPending,
	}

	if err := s.sellerRepo.Create(profile); err != nil {
		return nil, errors.New("failed to create private seller profile")
	}

	// ── 5. Reload with associations ───────────────────────────────────────────
	profile, err = s.sellerRepo.FindByID(profile.ID)
	if err != nil || profile == nil {
		return nil, errors.New("profile created but failed to load")
	}

	res := dto.ToPrivateSellerProfileResponse(*profile)
	return &res, nil
}

// GetMyProfile retrieves the private seller profile for the authenticated user.
func (s *PrivateSellerService) GetMyProfile(userID uuid.UUID) (*dto.PrivateSellerProfileResponse, error) {
	profile, err := s.sellerRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("failed to retrieve private seller profile")
	}
	if profile == nil {
		return nil, errors.New("private seller profile not found")
	}

	res := dto.ToPrivateSellerProfileResponse(*profile)
	return &res, nil
}

// GetProfileByID retrieves any private seller profile by UUID.
func (s *PrivateSellerService) GetProfileByID(id uuid.UUID) (*dto.PrivateSellerProfileResponse, error) {
	profile, err := s.sellerRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("failed to retrieve private seller profile")
	}
	if profile == nil {
		return nil, errors.New("private seller profile not found")
	}

	res := dto.ToPrivateSellerProfileResponse(*profile)
	return &res, nil
}

// UpdateProfile updates mutable fields on the authenticated private seller's profile.
func (s *PrivateSellerService) UpdateProfile(
	userID uuid.UUID,
	req dto.UpdatePrivateSellerRequest,
) (*dto.PrivateSellerProfileResponse, error) {
	profile, err := s.sellerRepo.FindByUserID(userID)
	if err != nil || profile == nil {
		return nil, errors.New("private seller profile not found")
	}

	if req.Location != "" {
		profile.Location = req.Location
	}
	if req.Address != "" {
		profile.Address = req.Address
	}
	if req.ProfilePhotoURL != "" {
		profile.ProfilePhotoURL = req.ProfilePhotoURL
	}
	if req.Bio != "" {
		profile.Bio = req.Bio
	}

	if err := s.sellerRepo.Update(profile); err != nil {
		return nil, errors.New("failed to update private seller profile")
	}

	res := dto.ToPrivateSellerProfileResponse(*profile)
	return &res, nil
}

// UpdateProfilePhoto sets the profile photo URL after a successful upload.
func (s *PrivateSellerService) UpdateProfilePhoto(userID uuid.UUID, photoURL string) error {
	profile, err := s.sellerRepo.FindByUserID(userID)
	if err != nil || profile == nil {
		return errors.New("private seller profile not found")
	}
	return s.sellerRepo.UpdateProfilePhotoURL(profile.ID, photoURL)
}

// ── Admin Operations ──────────────────────────────────────────────────────────

// AdminListProfiles returns a paginated list of private seller profiles.
func (s *PrivateSellerService) AdminListProfiles(
	status string,
	page, perPage int,
) ([]dto.PrivateSellerProfileResponse, int64, error) {
	approvalStatus := domain.ApprovalStatus(status)
	profiles, total, err := s.sellerRepo.List(approvalStatus, page, perPage)
	if err != nil {
		return nil, 0, errors.New("failed to retrieve private seller profiles")
	}

	var result []dto.PrivateSellerProfileResponse
	for _, p := range profiles {
		result = append(result, dto.ToPrivateSellerProfileResponse(p))
	}
	return result, total, nil
}

// AdminReviewProfile approves or rejects a private seller profile.
func (s *PrivateSellerService) AdminReviewProfile(
	profileID uuid.UUID,
	adminID uuid.UUID,
	req dto.AdminReviewSellerRequest,
) (*dto.PrivateSellerProfileResponse, error) {
	profile, err := s.sellerRepo.FindByID(profileID)
	if err != nil || profile == nil {
		return nil, errors.New("private seller profile not found")
	}

	if profile.ApprovalStatus != domain.ApprovalPending {
		return nil, errors.New("this profile has already been reviewed")
	}

	newStatus := domain.ApprovalStatus(req.ApprovalStatus)
	if err := s.sellerRepo.UpdateApprovalStatus(profileID, newStatus, adminID); err != nil {
		return nil, errors.New("failed to update approval status")
	}

	profile, err = s.sellerRepo.FindByID(profileID)
	if err != nil || profile == nil {
		return nil, errors.New("profile updated but failed to reload")
	}

	res := dto.ToPrivateSellerProfileResponse(*profile)
	return &res, nil
}
