package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"github.com/shikagari/api/internal/dto"
	"github.com/shikagari/api/internal/repository/interfaces"
)

// DealerService handles dealer profile creation, updates, and admin approval.
type DealerService struct {
	dealerRepo interfaces.DealerRepository
	userRepo   interfaces.UserRepository
}

// NewDealerService constructs a DealerService with its dependencies.
func NewDealerService(
	dealerRepo interfaces.DealerRepository,
	userRepo interfaces.UserRepository,
) *DealerService {
	return &DealerService{
		dealerRepo: dealerRepo,
		userRepo:   userRepo,
	}
}

// CreateProfile creates a dealer profile for the authenticated seller.
// A user may only have one dealer profile.
func (s *DealerService) CreateProfile(
	userID uuid.UUID,
	req dto.CreateDealerProfileRequest,
) (*dto.DealerProfileResponse, error) {
	// ── 1. Ensure user exists and has seller role ─────────────────────────────
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	if user.Role != domain.RoleSeller && user.Role != domain.RoleAdmin {
		return nil, errors.New("only users with the seller role can create a dealer profile")
	}

	// ── 2. Prevent duplicate profiles ────────────────────────────────────────
	exists, err := s.dealerRepo.ExistsByUserID(userID)
	if err != nil {
		return nil, errors.New("failed to check existing profile")
	}
	if exists {
		return nil, errors.New("you already have a dealer profile")
	}

	// ── 3. Check business registration number uniqueness ─────────────────────
	if req.BusinessRegNo != "" {
		regExists, err := s.dealerRepo.ExistsByBusinessRegNo(req.BusinessRegNo)
		if err != nil {
			return nil, errors.New("failed to validate business registration number")
		}
		if regExists {
			return nil, errors.New("a dealer profile with this business registration number already exists")
		}
	}

	// ── 4. Persist the profile ────────────────────────────────────────────────
	profile := &domain.DealerProfile{
		UserID:         userID,
		BusinessName:   req.BusinessName,
		BusinessRegNo:  req.BusinessRegNo,
		Location:       req.Location,
		Address:        req.Address,
		Description:    req.Description,
		KRAPIN:         req.KRAPIN,
		ApprovalStatus: domain.ApprovalPending,
	}

	if err := s.dealerRepo.Create(profile); err != nil {
		return nil, errors.New("failed to create dealer profile")
	}

	// ── 5. Reload with associations for response ──────────────────────────────
	profile, err = s.dealerRepo.FindByID(profile.ID)
	if err != nil || profile == nil {
		return nil, errors.New("profile created but failed to load")
	}

	res := dto.ToDealerProfileResponse(*profile)
	return &res, nil
}

// GetMyProfile retrieves the dealer profile for the authenticated user.
func (s *DealerService) GetMyProfile(userID uuid.UUID) (*dto.DealerProfileResponse, error) {
	profile, err := s.dealerRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("failed to retrieve dealer profile")
	}
	if profile == nil {
		return nil, errors.New("dealer profile not found")
	}

	res := dto.ToDealerProfileResponse(*profile)
	return &res, nil
}

// GetProfileByID retrieves any dealer profile by UUID (public + admin).
func (s *DealerService) GetProfileByID(id uuid.UUID) (*dto.DealerProfileResponse, error) {
	profile, err := s.dealerRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("failed to retrieve dealer profile")
	}
	if profile == nil {
		return nil, errors.New("dealer profile not found")
	}

	res := dto.ToDealerProfileResponse(*profile)
	return &res, nil
}

// UpdateProfile updates mutable fields on the authenticated dealer's profile.
// Profile must belong to the requesting user.
func (s *DealerService) UpdateProfile(
	userID uuid.UUID,
	req dto.UpdateDealerProfileRequest,
) (*dto.DealerProfileResponse, error) {
	profile, err := s.dealerRepo.FindByUserID(userID)
	if err != nil || profile == nil {
		return nil, errors.New("dealer profile not found")
	}

	// Apply only provided (non-zero) fields
	if req.BusinessName != "" {
		profile.BusinessName = req.BusinessName
	}
	if req.Location != "" {
		profile.Location = req.Location
	}
	if req.Address != "" {
		profile.Address = req.Address
	}
	if req.Description != "" {
		profile.Description = req.Description
	}
	if req.KRAPIN != "" {
		profile.KRAPIN = req.KRAPIN
	}

	if err := s.dealerRepo.Update(profile); err != nil {
		return nil, errors.New("failed to update dealer profile")
	}

	res := dto.ToDealerProfileResponse(*profile)
	return &res, nil
}

// UpdateLogo sets the dealer's logo URL after a successful image upload.
func (s *DealerService) UpdateLogo(userID uuid.UUID, logoURL string) error {
	profile, err := s.dealerRepo.FindByUserID(userID)
	if err != nil || profile == nil {
		return errors.New("dealer profile not found")
	}
	return s.dealerRepo.UpdateLogoURL(profile.ID, logoURL)
}

// ── Admin Operations ──────────────────────────────────────────────────────────

// AdminListProfiles returns a paginated list of dealer profiles.
// Optionally filtered by approval status.
func (s *DealerService) AdminListProfiles(
	status string,
	page, perPage int,
) ([]dto.DealerProfileResponse, int64, error) {
	approvalStatus := domain.ApprovalStatus(status)
	profiles, total, err := s.dealerRepo.List(approvalStatus, page, perPage)
	if err != nil {
		return nil, 0, errors.New("failed to retrieve dealer profiles")
	}

	var result []dto.DealerProfileResponse
	for _, p := range profiles {
		result = append(result, dto.ToDealerProfileResponse(p))
	}
	return result, total, nil
}

// AdminReviewProfile approves or rejects a dealer profile.
// Only an admin can call this.
func (s *DealerService) AdminReviewProfile(
	profileID uuid.UUID,
	adminID uuid.UUID,
	req dto.AdminReviewDealerRequest,
) (*dto.DealerProfileResponse, error) {
	profile, err := s.dealerRepo.FindByID(profileID)
	if err != nil || profile == nil {
		return nil, errors.New("dealer profile not found")
	}

	// Prevent re-review of already-decided profiles
	if profile.ApprovalStatus != domain.ApprovalPending {
		return nil, errors.New("this profile has already been reviewed")
	}

	newStatus := domain.ApprovalStatus(req.ApprovalStatus)
	if err := s.dealerRepo.UpdateApprovalStatus(profileID, newStatus, adminID); err != nil {
		return nil, errors.New("failed to update approval status")
	}

	// Reload for updated response
	profile, err = s.dealerRepo.FindByID(profileID)
	if err != nil || profile == nil {
		return nil, errors.New("profile updated but failed to reload")
	}

	res := dto.ToDealerProfileResponse(*profile)
	return &res, nil
}
