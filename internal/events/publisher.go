package events

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/Ecom-micro-template/service-support/internal/models"
)

// Event types
const (
	EventTicketCreated  = "support.ticket.created"
	EventTicketUpdated  = "support.ticket.updated"
	EventTicketReplied  = "support.ticket.replied"
	EventTicketResolved = "support.ticket.resolved"
	EventTicketClosed   = "support.ticket.closed"
)

// Publisher handles NATS event publishing
type Publisher struct {
	nc *nats.Conn
}

// NewPublisher creates a new event publisher
func NewPublisher(nc *nats.Conn) *Publisher {
	return &Publisher{nc: nc}
}

// TicketCreatedEvent represents ticket creation event
type TicketCreatedEvent struct {
	TicketID     string `json:"ticket_id"`
	TicketNumber string `json:"ticket_number"`
	Subject      string `json:"subject"`
	CustomerID   string `json:"customer_id,omitempty"`
	GuestEmail   string `json:"guest_email,omitempty"`
	GuestName    string `json:"guest_name,omitempty"`
	CategoryID   string `json:"category_id,omitempty"`
	Priority     string `json:"priority"`
}

// TicketReplyEvent represents ticket reply event
type TicketReplyEvent struct {
	TicketID       string `json:"ticket_id"`
	TicketNumber   string `json:"ticket_number"`
	Subject        string `json:"subject"`
	MessageID      string `json:"message_id"`
	MessageContent string `json:"message_content"`
	SenderType     string `json:"sender_type"`
	CustomerID     string `json:"customer_id,omitempty"`
	GuestEmail     string `json:"guest_email,omitempty"`
	IsAgentReply   bool   `json:"is_agent_reply"`
}

// PublishTicketCreated publishes a ticket created event
func (p *Publisher) PublishTicketCreated(ticket *models.Ticket) error {
	if p.nc == nil {
		return nil
	}

	event := TicketCreatedEvent{
		TicketID:     ticket.ID.String(),
		TicketNumber: ticket.TicketNumber,
		Subject:      ticket.Subject,
		GuestEmail:   ticket.GuestEmail,
		GuestName:    ticket.GuestName,
		Priority:     string(ticket.Priority),
	}

	if ticket.CustomerID != nil {
		event.CustomerID = ticket.CustomerID.String()
	}
	if ticket.CategoryID != nil {
		event.CategoryID = ticket.CategoryID.String()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.nc.Publish(EventTicketCreated, data)
}

// PublishTicketReply publishes a ticket reply event
func (p *Publisher) PublishTicketReply(ticket *models.Ticket, message *models.Message, isAgentReply bool) error {
	if p.nc == nil {
		return nil
	}

	event := TicketReplyEvent{
		TicketID:       ticket.ID.String(),
		TicketNumber:   ticket.TicketNumber,
		Subject:        ticket.Subject,
		MessageID:      message.ID.String(),
		MessageContent: message.Content,
		SenderType:     string(message.SenderType),
		GuestEmail:     ticket.GuestEmail,
		IsAgentReply:   isAgentReply,
	}

	if ticket.CustomerID != nil {
		event.CustomerID = ticket.CustomerID.String()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.nc.Publish(EventTicketReplied, data)
}

// PublishTicketResolved publishes a ticket resolved event
func (p *Publisher) PublishTicketResolved(ticket *models.Ticket) error {
	if p.nc == nil {
		return nil
	}

	event := map[string]interface{}{
		"ticket_id":     ticket.ID.String(),
		"ticket_number": ticket.TicketNumber,
		"subject":       ticket.Subject,
		"customer_id":   "",
		"guest_email":   ticket.GuestEmail,
	}

	if ticket.CustomerID != nil {
		event["customer_id"] = ticket.CustomerID.String()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.nc.Publish(EventTicketResolved, data)
}
