package persistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CannedResponseModel is the GORM persistence model for CannedResponse.
type CannedResponseModel struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Title      string     `json:"title" gorm:"size:255;not null"`
	Content    string     `json:"content" gorm:"type:text;not null"`
	CategoryID *uuid.UUID `json:"category_id" gorm:"type:uuid"`
	Shortcut   string     `json:"shortcut" gorm:"size:50;index"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`
	UsageCount int        `json:"usage_count" gorm:"default:0"`
	CreatedBy  *uuid.UUID `json:"created_by" gorm:"type:uuid"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// TableName specifies the table name.
func (CannedResponseModel) TableName() string {
	return "support.canned_responses"
}

// BeforeCreate hook to generate UUID if not provided.
func (m *CannedResponseModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
