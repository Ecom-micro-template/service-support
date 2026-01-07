package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// SenderType represents who sent the message
type SenderType string

const (
	SenderTypeCustomer SenderType = "customer"
	SenderTypeAgent    SenderType = "agent"
	SenderTypeSystem   SenderType = "system"
)

// Message represents a message in a ticket conversation
type Message struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TicketID    uuid.UUID      `json:"ticket_id" gorm:"type:uuid;not null;index"`
	SenderType  SenderType     `json:"sender_type" gorm:"size:20;not null"`
	SenderID    *uuid.UUID     `json:"sender_id" gorm:"type:uuid"`
	SenderName  string         `json:"sender_name" gorm:"size:255"`
	SenderEmail string         `json:"sender_email" gorm:"size:255"`
	Content     string         `json:"content" gorm:"type:text;not null"`
	Attachments datatypes.JSON `json:"attachments" gorm:"type:jsonb;default:'[]'"`
	IsInternal  bool           `json:"is_internal" gorm:"default:false"`
	ReadAt      *time.Time     `json:"read_at"`
	CreatedAt   time.Time      `json:"created_at"`
}

// TableName specifies the table name for Message
func (Message) TableName() string {
	return "support.messages"
}

// Attachment represents a file attachment in a message
type Attachment struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

// StatusHistory represents a status change in a ticket
type StatusHistory struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TicketID      uuid.UUID  `json:"ticket_id" gorm:"type:uuid;not null;index"`
	FromStatus    string     `json:"from_status" gorm:"size:20"`
	ToStatus      string     `json:"to_status" gorm:"size:20;not null"`
	ChangedBy     *uuid.UUID `json:"changed_by" gorm:"type:uuid"`
	ChangedByName string     `json:"changed_by_name" gorm:"size:255"`
	Notes         string     `json:"notes" gorm:"type:text"`
	CreatedAt     time.Time  `json:"created_at"`
}

// TableName specifies the table name for StatusHistory
func (StatusHistory) TableName() string {
	return "support.status_history"
}

// CannedResponse represents a pre-written response template
type CannedResponse struct {
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

// TableName specifies the table name for CannedResponse
func (CannedResponse) TableName() string {
	return "support.canned_responses"
}
