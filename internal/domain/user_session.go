package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserSession represents a login instance for a user across devices.
type UserSession struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Device     string    `gorm:"type:varchar(120);not null" json:"device"`
	Browser    string    `gorm:"type:varchar(120);not null" json:"browser"`
	Location   string    `gorm:"type:varchar(120);not null" json:"location"`
	IPAddress  string    `gorm:"type:varchar(45);not null" json:"ip_address"`
	UserAgent  string    `gorm:"type:text" json:"user_agent"`
	LastActive time.Time `gorm:"not null" json:"last_active"`

	RevokedAt     *time.Time `json:"revoked_at"`
	RevokedReason string     `gorm:"type:varchar(160)" json:"revoked_reason"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
