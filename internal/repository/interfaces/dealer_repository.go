package interfaces

import (
	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
)

// DealerRepository defines all database operations for DealerProfile.
type DealerRepository interface {
	// Create persists a new dealer profile.
	Create(profile *domain.DealerProfile) error

	// FindByID retrieves a dealer profile by its UUID.
	FindByID(id uuid.UUID) (*domain.DealerProfile, error)

	// FindByUserID retrieves a dealer profile by the owning user's UUID.
	FindByUserID(userID uuid.UUID) (*domain.DealerProfile, error)

	// Update persists changes to a dealer profile.
	Update(profile *domain.DealerProfile) error

	// UpdateApprovalStatus sets the approval status and records who approved it.
	UpdateApprovalStatus(id uuid.UUID, status domain.ApprovalStatus, approvedByID uuid.UUID) error

	// UpdateLogoURL sets the logo image URL after upload.
	UpdateLogoURL(id uuid.UUID, logoURL string) error

	// List returns a paginated list of dealer profiles filtered by approval status.
	// Pass an empty string to return all statuses.
	List(status domain.ApprovalStatus, page, perPage int) ([]domain.DealerProfile, int64, error)

	// Delete soft-deletes a dealer profile.
	Delete(id uuid.UUID) error

	// ExistsByUserID checks if a dealer profile already exists for a user.
	ExistsByUserID(userID uuid.UUID) (bool, error)

	// ExistsByBusinessRegNo checks for duplicate business registration numbers.
	ExistsByBusinessRegNo(regNo string) (bool, error)
}
