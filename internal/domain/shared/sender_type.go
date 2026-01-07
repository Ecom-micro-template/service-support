package shared

import (
	"errors"
	"fmt"
)

// SenderType represents who sent a message.
type SenderType string

// Sender type constants
const (
	SenderCustomer SenderType = "customer"
	SenderAgent    SenderType = "agent"
	SenderSystem   SenderType = "system"
)

// ErrInvalidSenderType is returned for invalid sender types.
var ErrInvalidSenderType = errors.New("invalid sender type")

// AllSenderTypes returns all valid sender types.
func AllSenderTypes() []SenderType {
	return []SenderType{SenderCustomer, SenderAgent, SenderSystem}
}

// IsValid returns true if the sender type is valid.
func (t SenderType) IsValid() bool {
	switch t {
	case SenderCustomer, SenderAgent, SenderSystem:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (t SenderType) String() string {
	return string(t)
}

// Label returns a human-readable label.
func (t SenderType) Label() string {
	switch t {
	case SenderCustomer:
		return "Customer"
	case SenderAgent:
		return "Agent"
	case SenderSystem:
		return "System"
	default:
		return "Unknown"
	}
}

// IsCustomer returns true if sender is customer.
func (t SenderType) IsCustomer() bool {
	return t == SenderCustomer
}

// IsAgent returns true if sender is agent.
func (t SenderType) IsAgent() bool {
	return t == SenderAgent
}

// IsSystem returns true if sender is system.
func (t SenderType) IsSystem() bool {
	return t == SenderSystem
}

// IsHuman returns true if sender is human (customer or agent).
func (t SenderType) IsHuman() bool {
	return t == SenderCustomer || t == SenderAgent
}

// ParseSenderType parses a string into a SenderType.
func ParseSenderType(s string) (SenderType, error) {
	t := SenderType(s)
	if !t.IsValid() {
		return "", fmt.Errorf("%w: %s", ErrInvalidSenderType, s)
	}
	return t, nil
}
