package shared

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TicketNumber represents a unique ticket identifier.
type TicketNumber struct {
	value string
}

// TicketNumber format: TKT-YYYYMMDD-XXXX
var ticketNumberRegex = regexp.MustCompile(`^TKT-\d{8}-\d{4}$`)

// ErrInvalidTicketNumber is returned for invalid ticket numbers.
var ErrInvalidTicketNumber = errors.New("invalid ticket number format")

// NewTicketNumber creates a new TicketNumber with validation.
func NewTicketNumber(number string) (TicketNumber, error) {
	number = strings.TrimSpace(strings.ToUpper(number))
	if !ticketNumberRegex.MatchString(number) {
		return TicketNumber{}, ErrInvalidTicketNumber
	}
	return TicketNumber{value: number}, nil
}

// GenerateTicketNumber generates a new unique ticket number.
// Format: TKT-YYYYMMDD-XXXX
func GenerateTicketNumber() TicketNumber {
	now := time.Now()
	seq := now.UnixNano() % 10000
	value := fmt.Sprintf("TKT-%s-%04d", now.Format("20060102"), seq)
	return TicketNumber{value: value}
}

// GenerateTicketNumberFromSequence generates a ticket number from a sequence.
func GenerateTicketNumberFromSequence(seq int64) TicketNumber {
	now := time.Now()
	value := fmt.Sprintf("TKT-%s-%04d", now.Format("20060102"), seq%10000)
	return TicketNumber{value: value}
}

// Value returns the ticket number string.
func (n TicketNumber) Value() string {
	return n.value
}

// String returns the string representation.
func (n TicketNumber) String() string {
	return n.value
}

// IsEmpty returns true if the ticket number is empty.
func (n TicketNumber) IsEmpty() bool {
	return n.value == ""
}

// Date returns the date portion of the ticket number.
func (n TicketNumber) Date() (time.Time, error) {
	if n.IsEmpty() {
		return time.Time{}, ErrInvalidTicketNumber
	}
	parts := strings.Split(n.value, "-")
	if len(parts) != 3 {
		return time.Time{}, ErrInvalidTicketNumber
	}
	return time.Parse("20060102", parts[1])
}

// Sequence returns the sequence portion of the ticket number.
func (n TicketNumber) Sequence() int {
	if n.IsEmpty() {
		return 0
	}
	parts := strings.Split(n.value, "-")
	if len(parts) != 3 {
		return 0
	}
	seq, _ := strconv.Atoi(parts[2])
	return seq
}

// Equals compares two ticket numbers.
func (n TicketNumber) Equals(other TicketNumber) bool {
	return n.value == other.value
}
