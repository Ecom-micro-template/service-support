// Package shared provides shared value objects for the support domain.
package shared

import (
	"errors"
	"fmt"
)

// TicketStatus represents the status of a support ticket.
// This is a value object with a state machine.
type TicketStatus string

// Ticket status constants
const (
	StatusOpen       TicketStatus = "open"
	StatusPending    TicketStatus = "pending"
	StatusInProgress TicketStatus = "in_progress"
	StatusResolved   TicketStatus = "resolved"
	StatusClosed     TicketStatus = "closed"
)

// validTransitions defines the allowed state transitions.
var validTransitions = map[TicketStatus][]TicketStatus{
	StatusOpen:       {StatusPending, StatusInProgress, StatusResolved},
	StatusPending:    {StatusOpen, StatusInProgress, StatusResolved},
	StatusInProgress: {StatusPending, StatusResolved},
	StatusResolved:   {StatusClosed, StatusOpen}, // Can reopen
	StatusClosed:     {},                         // Terminal state
}

// ErrInvalidTicketTransition is returned when an invalid status transition is attempted.
var ErrInvalidTicketTransition = errors.New("invalid ticket status transition")

// AllTicketStatuses returns all valid statuses.
func AllTicketStatuses() []TicketStatus {
	return []TicketStatus{StatusOpen, StatusPending, StatusInProgress, StatusResolved, StatusClosed}
}

// IsValid returns true if the status is valid.
func (s TicketStatus) IsValid() bool {
	switch s {
	case StatusOpen, StatusPending, StatusInProgress, StatusResolved, StatusClosed:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (s TicketStatus) String() string {
	return string(s)
}

// Label returns a human-readable label.
func (s TicketStatus) Label() string {
	switch s {
	case StatusOpen:
		return "Open"
	case StatusPending:
		return "Pending"
	case StatusInProgress:
		return "In Progress"
	case StatusResolved:
		return "Resolved"
	case StatusClosed:
		return "Closed"
	default:
		return "Unknown"
	}
}

// CanTransitionTo returns true if the ticket can transition to the target status.
func (s TicketStatus) CanTransitionTo(target TicketStatus) bool {
	allowed, exists := validTransitions[s]
	if !exists {
		return false
	}
	for _, status := range allowed {
		if status == target {
			return true
		}
	}
	return false
}

// TransitionTo attempts to transition to the target status.
func (s TicketStatus) TransitionTo(target TicketStatus) (TicketStatus, error) {
	if !s.CanTransitionTo(target) {
		return s, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidTicketTransition, s, target)
	}
	return target, nil
}

// ValidTransitions returns statuses this status can transition to.
func (s TicketStatus) ValidTransitions() []TicketStatus {
	transitions, exists := validTransitions[s]
	if !exists {
		return []TicketStatus{}
	}
	return transitions
}

// IsTerminal returns true if the status is terminal.
func (s TicketStatus) IsTerminal() bool {
	return len(s.ValidTransitions()) == 0
}

// IsOpen returns true if status is open.
func (s TicketStatus) IsOpen() bool {
	return s == StatusOpen
}

// IsPending returns true if status is pending.
func (s TicketStatus) IsPending() bool {
	return s == StatusPending
}

// IsInProgress returns true if status is in progress.
func (s TicketStatus) IsInProgress() bool {
	return s == StatusInProgress
}

// IsResolved returns true if status is resolved.
func (s TicketStatus) IsResolved() bool {
	return s == StatusResolved
}

// IsClosed returns true if status is closed.
func (s TicketStatus) IsClosed() bool {
	return s == StatusClosed
}

// IsActive returns true if the ticket is still active (not resolved or closed).
func (s TicketStatus) IsActive() bool {
	return s != StatusResolved && s != StatusClosed
}

// RequiresAction returns true if the ticket requires agent action.
func (s TicketStatus) RequiresAction() bool {
	return s == StatusOpen || s == StatusInProgress
}

// ParseTicketStatus parses a string into a TicketStatus.
func ParseTicketStatus(s string) (TicketStatus, error) {
	status := TicketStatus(s)
	if !status.IsValid() {
		return "", fmt.Errorf("invalid ticket status: %s", s)
	}
	return status, nil
}
