// Package domain contains GORM persistence models for the support service.
//
// Deprecated: This package is being migrated to DDD architecture.
// For new development, use:
//   - Domain models: github.com/Ecom-micro-template/service-support/internal/domain/category
//   - Persistence: github.com/Ecom-micro-template/service-support/internal/infrastructure/persistence
//
// Existing code can continue using this package during the transition period.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Category represents a support ticket category
type Category struct {
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

// TableName specifies the table name for Category
func (Category) TableName() string {
	return "support.categories"
}
