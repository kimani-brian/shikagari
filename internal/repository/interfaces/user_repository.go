package interfaces

import (
	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
)

// UserRepository defines all database operations for the User entity.
type UserRepository interface {
	// Create persists a new user record.
	Create(user *domain.User) error

	// FindByID retrieves a user by their UUID.
	FindByID(id uuid.UUID) (*domain.User, error)

	// FindByEmail retrieves a user by email address (used for login).
	FindByEmail(email string) (*domain.User, error)

	// Update persists changes to an existing user record.
	Update(user *domain.User) error

	// UpdatePassword updates only the password hash for a user.
	UpdatePassword(id uuid.UUID, passwordHash string) error

	// SetVerified toggles the verified badge on a user.
	SetVerified(id uuid.UUID, verified bool) error

	// SetActive toggles the active status of a user (soft disable).
	SetActive(id uuid.UUID, active bool) error

	// SetRole updates the role of a user (admin action).
	SetRole(id uuid.UUID, role domain.UserRole) error

	// List returns a paginated list of all users (admin use).
	List(page, perPage int) ([]domain.User, int64, error)

	// Delete soft-deletes a user record.
	Delete(id uuid.UUID) error

	// ExistsByEmail checks if an email address is already registered.
	ExistsByEmail(email string) (bool, error)
}
