package ticket

import (
	"time"

	"github.com/google/uuid"
	"github.com/niaga-platform/service-support/internal/domain/shared"
)

// StatusHistory represents a status change in a ticket.
type StatusHistory struct {
	id            uuid.UUID
	ticketID      uuid.UUID
	fromStatus    shared.TicketStatus
	toStatus      shared.TicketStatus
	changedBy     *uuid.UUID
	changedByName string
	notes         string
	createdAt     time.Time
}

// NewStatusHistory creates a new StatusHistory entry.
func NewStatusHistory(ticketID uuid.UUID, from, to shared.TicketStatus, changedBy *uuid.UUID, notes string) StatusHistory {
	return StatusHistory{
		id:         uuid.New(),
		ticketID:   ticketID,
		fromStatus: from,
		toStatus:   to,
		changedBy:  changedBy,
		notes:      notes,
		createdAt:  time.Now(),
	}
}

// Getters
func (h StatusHistory) ID() uuid.UUID                   { return h.id }
func (h StatusHistory) TicketID() uuid.UUID             { return h.ticketID }
func (h StatusHistory) FromStatus() shared.TicketStatus { return h.fromStatus }
func (h StatusHistory) ToStatus() shared.TicketStatus   { return h.toStatus }
func (h StatusHistory) ChangedBy() *uuid.UUID           { return h.changedBy }
func (h StatusHistory) ChangedByName() string           { return h.changedByName }
func (h StatusHistory) Notes() string                   { return h.notes }
func (h StatusHistory) CreatedAt() time.Time            { return h.createdAt }

// SetChangedByName sets the name of who made the change.
func (h *StatusHistory) SetChangedByName(name string) {
	h.changedByName = name
}

// IsEscalation returns true if this was an escalation.
func (h StatusHistory) IsEscalation() bool {
	return h.fromStatus.IsOpen() && h.toStatus.IsInProgress()
}

// IsResolution returns true if this was a resolution.
func (h StatusHistory) IsResolution() bool {
	return h.toStatus.IsResolved()
}

// IsClosure returns true if this was a closure.
func (h StatusHistory) IsClosure() bool {
	return h.toStatus.IsClosed()
}

// IsReopen returns true if this was a reopen.
func (h StatusHistory) IsReopen() bool {
	return h.fromStatus.IsResolved() && h.toStatus.IsOpen()
}
