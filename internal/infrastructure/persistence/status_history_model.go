package persistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StatusHistoryModel is the GORM persistence model for StatusHistory.
type StatusHistoryModel struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TicketID      uuid.UUID  `json:"ticket_id" gorm:"type:uuid;not null;index"`
	FromStatus    string     `json:"from_status" gorm:"size:20"`
	ToStatus      string     `json:"to_status" gorm:"size:20;not null"`
	ChangedBy     *uuid.UUID `json:"changed_by" gorm:"type:uuid"`
	ChangedByName string     `json:"changed_by_name" gorm:"size:255"`
	Notes         string     `json:"notes" gorm:"type:text"`
	CreatedAt     time.Time  `json:"created_at"`
}

// TableName specifies the table name.
func (StatusHistoryModel) TableName() string {
	return "support.status_history"
}

// BeforeCreate hook to generate UUID if not provided.
func (m *StatusHistoryModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
