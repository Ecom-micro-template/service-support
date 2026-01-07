// Package persistence contains GORM models and repository implementations.
package persistence

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// TicketModel is the GORM persistence model for Ticket.
type TicketModel struct {
	ID                  uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TicketNumber        string         `json:"ticket_number" gorm:"size:20;uniqueIndex;not null"`
	CustomerID          *uuid.UUID     `json:"customer_id" gorm:"type:uuid"`
	GuestEmail          string         `json:"guest_email" gorm:"size:255"`
	GuestName           string         `json:"guest_name" gorm:"size:255"`
	GuestPhone          string         `json:"guest_phone" gorm:"size:20"`
	CategoryID          *uuid.UUID     `json:"category_id" gorm:"type:uuid"`
	Category            *CategoryModel `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Subject             string         `json:"subject" gorm:"size:255;not null"`
	Status              string         `json:"status" gorm:"size:20;default:'open'"`
	Priority            string         `json:"priority" gorm:"size:20;default:'normal'"`
	AssignedTo          *uuid.UUID     `json:"assigned_to" gorm:"type:uuid"`
	OrderID             *uuid.UUID     `json:"order_id" gorm:"type:uuid"`
	OrderNumber         string         `json:"order_number" gorm:"size:50"`
	SLADeadline         *time.Time     `json:"sla_deadline"`
	FirstResponseAt     *time.Time     `json:"first_response_at"`
	ResolvedAt          *time.Time     `json:"resolved_at"`
	ClosedAt            *time.Time     `json:"closed_at"`
	SatisfactionRating  *int           `json:"satisfaction_rating"`
	SatisfactionComment string         `json:"satisfaction_comment" gorm:"type:text"`
	Tags                pq.StringArray `json:"tags" gorm:"type:text[]"`
	Messages            []MessageModel `json:"messages,omitempty" gorm:"foreignKey:TicketID"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name.
func (TicketModel) TableName() string {
	return "support.tickets"
}

// BeforeCreate hook to generate UUID if not provided.
func (m *TicketModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// IsOverdue checks if the ticket has exceeded its SLA deadline.
func (m *TicketModel) IsOverdue() bool {
	if m.SLADeadline == nil {
		return false
	}
	if m.Status == "resolved" || m.Status == "closed" {
		return false
	}
	return time.Now().After(*m.SLADeadline)
}

// MessageModel is the GORM persistence model for Message.
type MessageModel struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TicketID    uuid.UUID  `json:"ticket_id" gorm:"type:uuid;not null;index"`
	SenderType  string     `json:"sender_type" gorm:"size:20;not null"`
	SenderID    *uuid.UUID `json:"sender_id" gorm:"type:uuid"`
	SenderName  string     `json:"sender_name" gorm:"size:255"`
	SenderEmail string     `json:"sender_email" gorm:"size:255"`
	Content     string     `json:"content" gorm:"type:text;not null"`
	Attachments string     `json:"attachments" gorm:"type:jsonb;default:'[]'"` // JSON array
	IsInternal  bool       `json:"is_internal" gorm:"default:false"`
	ReadAt      *time.Time `json:"read_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// TableName specifies the table name.
func (MessageModel) TableName() string {
	return "support.messages"
}

// BeforeCreate hook.
func (m *MessageModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
