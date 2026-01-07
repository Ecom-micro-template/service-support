package shared

import (
	"errors"
	"fmt"
	"time"
)

// TicketPriority represents the priority of a support ticket.
type TicketPriority string

// Priority constants
const (
	PriorityLow    TicketPriority = "low"
	PriorityNormal TicketPriority = "normal"
	PriorityHigh   TicketPriority = "high"
	PriorityUrgent TicketPriority = "urgent"
)

// Priority SLA hours
const (
	SLAHoursLow    = 48
	SLAHoursNormal = 24
	SLAHoursHigh   = 8
	SLAHoursUrgent = 4
)

// ErrInvalidTicketPriority is returned for invalid priorities.
var ErrInvalidTicketPriority = errors.New("invalid ticket priority")

// AllTicketPriorities returns all valid priorities.
func AllTicketPriorities() []TicketPriority {
	return []TicketPriority{PriorityLow, PriorityNormal, PriorityHigh, PriorityUrgent}
}

// IsValid returns true if the priority is valid.
func (p TicketPriority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityNormal, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (p TicketPriority) String() string {
	return string(p)
}

// Label returns a human-readable label.
func (p TicketPriority) Label() string {
	switch p {
	case PriorityLow:
		return "Low"
	case PriorityNormal:
		return "Normal"
	case PriorityHigh:
		return "High"
	case PriorityUrgent:
		return "Urgent"
	default:
		return "Unknown"
	}
}

// SLAHours returns the SLA hours for this priority.
func (p TicketPriority) SLAHours() int {
	switch p {
	case PriorityLow:
		return SLAHoursLow
	case PriorityNormal:
		return SLAHoursNormal
	case PriorityHigh:
		return SLAHoursHigh
	case PriorityUrgent:
		return SLAHoursUrgent
	default:
		return SLAHoursNormal
	}
}

// SLADuration returns the SLA duration for this priority.
func (p TicketPriority) SLADuration() time.Duration {
	return time.Duration(p.SLAHours()) * time.Hour
}

// CalculateSLADeadline calculates the SLA deadline from the given start time.
func (p TicketPriority) CalculateSLADeadline(from time.Time) time.Time {
	return from.Add(p.SLADuration())
}

// Severity returns a numeric severity (higher = more urgent).
func (p TicketPriority) Severity() int {
	switch p {
	case PriorityUrgent:
		return 4
	case PriorityHigh:
		return 3
	case PriorityNormal:
		return 2
	case PriorityLow:
		return 1
	default:
		return 0
	}
}

// IsHigherThan returns true if this priority is higher than other.
func (p TicketPriority) IsHigherThan(other TicketPriority) bool {
	return p.Severity() > other.Severity()
}

// IsUrgent returns true if priority is urgent.
func (p TicketPriority) IsUrgent() bool {
	return p == PriorityUrgent
}

// IsHigh returns true if priority is high or urgent.
func (p TicketPriority) IsHigh() bool {
	return p == PriorityHigh || p == PriorityUrgent
}

// ParseTicketPriority parses a string into a TicketPriority.
func ParseTicketPriority(s string) (TicketPriority, error) {
	priority := TicketPriority(s)
	if !priority.IsValid() {
		return "", fmt.Errorf("%w: %s", ErrInvalidTicketPriority, s)
	}
	return priority, nil
}
