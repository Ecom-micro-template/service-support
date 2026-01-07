// Package models contains GORM persistence models for the support service.
//
// Deprecated: This package is being migrated to DDD architecture.
// For new development, use:
//   - Domain models: github.com/niaga-platform/service-support/internal/domain/ticket
//   - Persistence: github.com/niaga-platform/service-support/internal/infrastructure/persistence
//
// Existing code can continue using this package during the transition period.
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// TicketStatus represents the status of a support ticket
type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "open"
	TicketStatusPending    TicketStatus = "pending"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusResolved   TicketStatus = "resolved"
	TicketStatusClosed     TicketStatus = "closed"
)

// TicketPriority represents the priority of a support ticket
type TicketPriority string

const (
	TicketPriorityLow    TicketPriority = "low"
	TicketPriorityNormal TicketPriority = "normal"
	TicketPriorityHigh   TicketPriority = "high"
	TicketPriorityUrgent TicketPriority = "urgent"
)

// Ticket represents a support ticket
type Ticket struct {
	ID                  uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TicketNumber        string         `json:"ticket_number" gorm:"size:20;uniqueIndex;not null"`
	CustomerID          *uuid.UUID     `json:"customer_id" gorm:"type:uuid"`
	GuestEmail          string         `json:"guest_email" gorm:"size:255"`
	GuestName           string         `json:"guest_name" gorm:"size:255"`
	GuestPhone          string         `json:"guest_phone" gorm:"size:20"`
	CategoryID          *uuid.UUID     `json:"category_id" gorm:"type:uuid"`
	Category            *Category      `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Subject             string         `json:"subject" gorm:"size:255;not null"`
	Status              TicketStatus   `json:"status" gorm:"size:20;default:'open'"`
	Priority            TicketPriority `json:"priority" gorm:"size:20;default:'normal'"`
	AssignedTo          *uuid.UUID     `json:"assigned_to" gorm:"type:uuid"`
	AssignedToName      string         `json:"assigned_to_name" gorm:"-"`
	OrderID             *uuid.UUID     `json:"order_id" gorm:"type:uuid"`
	OrderNumber         string         `json:"order_number" gorm:"size:50"`
	SLADeadline         *time.Time     `json:"sla_deadline"`
	FirstResponseAt     *time.Time     `json:"first_response_at"`
	ResolvedAt          *time.Time     `json:"resolved_at"`
	ClosedAt            *time.Time     `json:"closed_at"`
	SatisfactionRating  *int           `json:"satisfaction_rating"`
	SatisfactionComment string         `json:"satisfaction_comment" gorm:"type:text"`
	Tags                pq.StringArray `json:"tags" gorm:"type:text[]"`
	Messages            []Message      `json:"messages,omitempty" gorm:"foreignKey:TicketID"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for Ticket
func (Ticket) TableName() string {
	return "support.tickets"
}

// IsOverdue checks if the ticket has exceeded its SLA deadline
func (t *Ticket) IsOverdue() bool {
	if t.SLADeadline == nil {
		return false
	}
	if t.Status == TicketStatusResolved || t.Status == TicketStatusClosed {
		return false
	}
	return time.Now().After(*t.SLADeadline)
}

// GetContactEmail returns the email of the ticket creator
func (t *Ticket) GetContactEmail() string {
	if t.GuestEmail != "" {
		return t.GuestEmail
	}
	return ""
}

// GetContactName returns the name of the ticket creator
func (t *Ticket) GetContactName() string {
	if t.GuestName != "" {
		return t.GuestName
	}
	return ""
}
