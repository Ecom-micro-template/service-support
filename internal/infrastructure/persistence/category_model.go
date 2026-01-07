package persistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoryModel is the GORM persistence model for Category.
type CategoryModel struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	NameMS      string    `json:"name_ms" gorm:"column:name_ms;size:100"`
	Description string    `json:"description" gorm:"type:text"`
	Icon        string    `json:"icon" gorm:"size:50"`
	SLAHours    int       `json:"sla_hours" gorm:"default:24"`
	Priority    int       `json:"priority" gorm:"default:0"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name.
func (CategoryModel) TableName() string {
	return "support.categories"
}

// BeforeCreate hook to generate UUID if not provided.
func (m *CategoryModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
