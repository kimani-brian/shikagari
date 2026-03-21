package domain

import (
	"time"

	"github.com/google/uuid"
)

// SecurityEvent captures notable actions related to account safety.
type SecurityEvent struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID   uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Label    string    `gorm:"type:varchar(120);not null" json:"label"`
	Severity string    `gorm:"type:varchar(20);not null" json:"severity"`
	Details  string    `gorm:"type:text" json:"details"`

	CreatedAt time.Time `json:"created_at"`
}
