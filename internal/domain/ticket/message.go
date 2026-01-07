package ticket

import (
	"time"

	"github.com/google/uuid"
	"github.com/niaga-platform/service-support/internal/domain/shared"
)

// Attachment represents a file attachment.
type Attachment struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

// Message represents a message in a ticket conversation.
type Message struct {
	id          uuid.UUID
	ticketID    uuid.UUID
	senderType  shared.SenderType
	senderID    *uuid.UUID
	senderName  string
	senderEmail string
	content     string
	attachments []Attachment
	isInternal  bool
	readAt      *time.Time
	createdAt   time.Time
}

// MessageParams contains parameters for creating a Message.
type MessageParams struct {
	ID          uuid.UUID
	TicketID    uuid.UUID
	SenderType  string
	SenderID    *uuid.UUID
	SenderName  string
	SenderEmail string
	Content     string
	Attachments []Attachment
	IsInternal  bool
}

// NewMessage creates a new Message entity.
func NewMessage(params MessageParams) Message {
	id := params.ID
	if id == uuid.Nil {
		id = uuid.New()
	}

	senderType := shared.SenderCustomer
	if params.SenderType != "" {
		st, err := shared.ParseSenderType(params.SenderType)
		if err == nil {
			senderType = st
		}
	}

	return Message{
		id:          id,
		ticketID:    params.TicketID,
		senderType:  senderType,
		senderID:    params.SenderID,
		senderName:  params.SenderName,
		senderEmail: params.SenderEmail,
		content:     params.Content,
		attachments: params.Attachments,
		isInternal:  params.IsInternal,
		createdAt:   time.Now(),
	}
}

// CreateCustomerMessage creates a message from a customer.
func CreateCustomerMessage(ticketID uuid.UUID, customerID *uuid.UUID, name, email, content string) Message {
	return NewMessage(MessageParams{
		TicketID:    ticketID,
		SenderType:  string(shared.SenderCustomer),
		SenderID:    customerID,
		SenderName:  name,
		SenderEmail: email,
		Content:     content,
	})
}

// CreateAgentMessage creates a message from an agent.
func CreateAgentMessage(ticketID uuid.UUID, agentID uuid.UUID, name, email, content string, isInternal bool) Message {
	return NewMessage(MessageParams{
		TicketID:    ticketID,
		SenderType:  string(shared.SenderAgent),
		SenderID:    &agentID,
		SenderName:  name,
		SenderEmail: email,
		Content:     content,
		IsInternal:  isInternal,
	})
}

// CreateSystemMessage creates a system-generated message.
func CreateSystemMessage(ticketID uuid.UUID, content string) Message {
	return NewMessage(MessageParams{
		TicketID:   ticketID,
		SenderType: string(shared.SenderSystem),
		SenderName: "System",
		Content:    content,
	})
}

// Getters
func (m Message) ID() uuid.UUID                 { return m.id }
func (m Message) TicketID() uuid.UUID           { return m.ticketID }
func (m Message) SenderType() shared.SenderType { return m.senderType }
func (m Message) SenderID() *uuid.UUID          { return m.senderID }
func (m Message) SenderName() string            { return m.senderName }
func (m Message) SenderEmail() string           { return m.senderEmail }
func (m Message) Content() string               { return m.content }
func (m Message) Attachments() []Attachment     { return m.attachments }
func (m Message) IsInternal() bool              { return m.isInternal }
func (m Message) ReadAt() *time.Time            { return m.readAt }
func (m Message) CreatedAt() time.Time          { return m.createdAt }

// IsFromCustomer returns true if message is from customer.
func (m Message) IsFromCustomer() bool {
	return m.senderType.IsCustomer()
}

// IsFromAgent returns true if message is from agent.
func (m Message) IsFromAgent() bool {
	return m.senderType.IsAgent()
}

// IsFromSystem returns true if message is from system.
func (m Message) IsFromSystem() bool {
	return m.senderType.IsSystem()
}

// IsRead returns true if the message has been read.
func (m Message) IsRead() bool {
	return m.readAt != nil
}

// MarkAsRead marks the message as read.
func (m *Message) MarkAsRead() {
	if m.readAt == nil {
		now := time.Now()
		m.readAt = &now
	}
}

// HasAttachments returns true if the message has attachments.
func (m Message) HasAttachments() bool {
	return len(m.attachments) > 0
}
