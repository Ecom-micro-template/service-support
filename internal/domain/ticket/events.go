package ticket

import (
	"time"

	"github.com/google/uuid"
)

// Event is the base interface for all ticket domain events.
type Event interface {
	EventType() string
	OccurredAt() time.Time
	AggregateID() uuid.UUID
}

// baseEvent contains common event fields.
type baseEvent struct {
	occurredAt  time.Time
	aggregateID uuid.UUID
}

func (e baseEvent) OccurredAt() time.Time  { return e.occurredAt }
func (e baseEvent) AggregateID() uuid.UUID { return e.aggregateID }

// TicketCreatedEvent is raised when a new ticket is created.
type TicketCreatedEvent struct {
	baseEvent
	TicketNumber string
	Subject      string
}

func (e TicketCreatedEvent) EventType() string { return "ticket.created" }

// NewTicketCreatedEvent creates a new TicketCreatedEvent.
func NewTicketCreatedEvent(ticketID uuid.UUID, ticketNumber, subject string) TicketCreatedEvent {
	return TicketCreatedEvent{
		baseEvent:    baseEvent{occurredAt: time.Now(), aggregateID: ticketID},
		TicketNumber: ticketNumber,
		Subject:      subject,
	}
}

// TicketAssignedEvent is raised when a ticket is assigned.
type TicketAssignedEvent struct {
	baseEvent
	AgentID uuid.UUID
}

func (e TicketAssignedEvent) EventType() string { return "ticket.assigned" }

// NewTicketAssignedEvent creates a new TicketAssignedEvent.
func NewTicketAssignedEvent(ticketID, agentID uuid.UUID) TicketAssignedEvent {
	return TicketAssignedEvent{
		baseEvent: baseEvent{occurredAt: time.Now(), aggregateID: ticketID},
		AgentID:   agentID,
	}
}

// TicketStatusChangedEvent is raised when ticket status changes.
type TicketStatusChangedEvent struct {
	baseEvent
	NewStatus string
}

func (e TicketStatusChangedEvent) EventType() string { return "ticket.status_changed" }

// NewTicketStatusChangedEvent creates a new TicketStatusChangedEvent.
func NewTicketStatusChangedEvent(ticketID uuid.UUID, newStatus string) TicketStatusChangedEvent {
	return TicketStatusChangedEvent{
		baseEvent: baseEvent{occurredAt: time.Now(), aggregateID: ticketID},
		NewStatus: newStatus,
	}
}

// TicketEscalatedEvent is raised when a ticket is escalated.
type TicketEscalatedEvent struct {
	baseEvent
	NewPriority string
	Reason      string
}

func (e TicketEscalatedEvent) EventType() string { return "ticket.escalated" }

// NewTicketEscalatedEvent creates a new TicketEscalatedEvent.
func NewTicketEscalatedEvent(ticketID uuid.UUID, newPriority, reason string) TicketEscalatedEvent {
	return TicketEscalatedEvent{
		baseEvent:   baseEvent{occurredAt: time.Now(), aggregateID: ticketID},
		NewPriority: newPriority,
		Reason:      reason,
	}
}

// TicketResolvedEvent is raised when a ticket is resolved.
type TicketResolvedEvent struct {
	baseEvent
	Resolution string
}

func (e TicketResolvedEvent) EventType() string { return "ticket.resolved" }

// NewTicketResolvedEvent creates a new TicketResolvedEvent.
func NewTicketResolvedEvent(ticketID uuid.UUID, resolution string) TicketResolvedEvent {
	return TicketResolvedEvent{
		baseEvent:  baseEvent{occurredAt: time.Now(), aggregateID: ticketID},
		Resolution: resolution,
	}
}

// TicketClosedEvent is raised when a ticket is closed.
type TicketClosedEvent struct {
	baseEvent
}

func (e TicketClosedEvent) EventType() string { return "ticket.closed" }

// NewTicketClosedEvent creates a new TicketClosedEvent.
func NewTicketClosedEvent(ticketID uuid.UUID) TicketClosedEvent {
	return TicketClosedEvent{
		baseEvent: baseEvent{occurredAt: time.Now(), aggregateID: ticketID},
	}
}

// TicketSLABreachedEvent is raised when SLA is breached.
type TicketSLABreachedEvent struct {
	baseEvent
	Deadline time.Time
}

func (e TicketSLABreachedEvent) EventType() string { return "ticket.sla_breached" }

// NewTicketSLABreachedEvent creates a new TicketSLABreachedEvent.
func NewTicketSLABreachedEvent(ticketID uuid.UUID, deadline time.Time) TicketSLABreachedEvent {
	return TicketSLABreachedEvent{
		baseEvent: baseEvent{occurredAt: time.Now(), aggregateID: ticketID},
		Deadline:  deadline,
	}
}
