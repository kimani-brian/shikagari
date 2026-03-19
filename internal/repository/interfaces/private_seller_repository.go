package interfaces

import (
	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
)

// PrivateSellerRepository defines all database operations for PrivateSellerProfile.
type PrivateSellerRepository interface {
	// Create persists a new private seller profile.
	Create(profile *domain.PrivateSellerProfile) error

	// FindByID retrieves a private seller profile by its UUID.
	FindByID(id uuid.UUID) (*domain.PrivateSellerProfile, error)

	// FindByUserID retrieves a private seller profile by the owning user's UUID.
	FindByUserID(userID uuid.UUID) (*domain.PrivateSellerProfile, error)

	// Update persists changes to a private seller profile.
	Update(profile *domain.PrivateSellerProfile) error

	// UpdateApprovalStatus sets the approval status and records who approved it.
	UpdateApprovalStatus(id uuid.UUID, status domain.ApprovalStatus, approvedByID uuid.UUID) error

	// UpdateProfilePhotoURL sets the profile photo URL after upload.
	UpdateProfilePhotoURL(id uuid.UUID, photoURL string) error

	// List returns a paginated list of private seller profiles by approval status.
	// Pass an empty string to return all statuses.
	List(status domain.ApprovalStatus, page, perPage int) ([]domain.PrivateSellerProfile, int64, error)

	// Delete soft-deletes a private seller profile.
	Delete(id uuid.UUID) error

	// ExistsByUserID checks if a private seller profile already exists for a user.
	ExistsByUserID(userID uuid.UUID) (bool, error)

	// ExistsByNationalID checks for duplicate national ID numbers.
	ExistsByNationalID(nationalID string) (bool, error)
}
