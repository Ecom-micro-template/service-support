package ticket

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/niaga-platform/service-support/internal/domain/shared"
)

// Domain errors for Ticket aggregate
var (
	ErrTicketNotFound  = errors.New("ticket not found")
	ErrInvalidTicket   = errors.New("invalid ticket data")
	ErrCannotModify    = errors.New("ticket cannot be modified in current state")
	ErrAlreadyAssigned = errors.New("ticket is already assigned")
	ErrNotAssigned     = errors.New("ticket is not assigned")
	ErrSLABreached     = errors.New("SLA deadline has been breached")
)

// Ticket is the aggregate root for support tickets.
type Ticket struct {
	id                  uuid.UUID
	ticketNumber        shared.TicketNumber
	customerID          *uuid.UUID
	guestEmail          string
	guestName           string
	guestPhone          string
	categoryID          *uuid.UUID
	subject             string
	status              shared.TicketStatus
	priority            shared.TicketPriority
	assignedTo          *uuid.UUID
	orderID             *uuid.UUID
	orderNumber         string
	slaDeadline         *time.Time
	firstResponseAt     *time.Time
	resolvedAt          *time.Time
	closedAt            *time.Time
	satisfactionRating  *int
	satisfactionComment string
	tags                []string
	messages            []Message
	statusHistory       []StatusHistory
	createdAt           time.Time
	updatedAt           time.Time

	// Domain events
	events []Event
}

// TicketParams contains parameters for creating a Ticket.
type TicketParams struct {
	ID           uuid.UUID
	TicketNumber string
	CustomerID   *uuid.UUID
	GuestEmail   string
	GuestName    string
	GuestPhone   string
	CategoryID   *uuid.UUID
	Subject      string
	Priority     string
	OrderID      *uuid.UUID
	OrderNumber  string
	Tags         []string
	SLAHours     int
}

// NewTicket creates a new Ticket aggregate.
func NewTicket(params TicketParams) (*Ticket, error) {
	if params.Subject == "" {
		return nil, errors.New("subject is required")
	}
	if params.CustomerID == nil && params.GuestEmail == "" {
		return nil, errors.New("customer or guest email is required")
	}

	id := params.ID
	if id == uuid.Nil {
		id = uuid.New()
	}

	ticketNumber := shared.GenerateTicketNumber()
	if params.TicketNumber != "" {
		tn, err := shared.NewTicketNumber(params.TicketNumber)
		if err == nil {
			ticketNumber = tn
		}
	}

	priority := shared.PriorityNormal
	if params.Priority != "" {
		p, err := shared.ParseTicketPriority(params.Priority)
		if err == nil {
			priority = p
		}
	}

	now := time.Now()
	slaDeadline := priority.CalculateSLADeadline(now)
	if params.SLAHours > 0 {
		slaDeadline = now.Add(time.Duration(params.SLAHours) * time.Hour)
	}

	ticket := &Ticket{
		id:            id,
		ticketNumber:  ticketNumber,
		customerID:    params.CustomerID,
		guestEmail:    params.GuestEmail,
		guestName:     params.GuestName,
		guestPhone:    params.GuestPhone,
		categoryID:    params.CategoryID,
		subject:       params.Subject,
		status:        shared.StatusOpen,
		priority:      priority,
		orderID:       params.OrderID,
		orderNumber:   params.OrderNumber,
		slaDeadline:   &slaDeadline,
		tags:          params.Tags,
		messages:      make([]Message, 0),
		statusHistory: make([]StatusHistory, 0),
		createdAt:     now,
		updatedAt:     now,
		events:        make([]Event, 0),
	}

	ticket.addEvent(NewTicketCreatedEvent(id, ticketNumber.Value(), params.Subject))

	return ticket, nil
}

// Getters
func (t *Ticket) ID() uuid.UUID                     { return t.id }
func (t *Ticket) TicketNumber() shared.TicketNumber { return t.ticketNumber }
func (t *Ticket) CustomerID() *uuid.UUID            { return t.customerID }
func (t *Ticket) GuestEmail() string                { return t.guestEmail }
func (t *Ticket) GuestName() string                 { return t.guestName }
func (t *Ticket) GuestPhone() string                { return t.guestPhone }
func (t *Ticket) CategoryID() *uuid.UUID            { return t.categoryID }
func (t *Ticket) Subject() string                   { return t.subject }
func (t *Ticket) Status() shared.TicketStatus       { return t.status }
func (t *Ticket) Priority() shared.TicketPriority   { return t.priority }
func (t *Ticket) AssignedTo() *uuid.UUID            { return t.assignedTo }
func (t *Ticket) OrderID() *uuid.UUID               { return t.orderID }
func (t *Ticket) OrderNumber() string               { return t.orderNumber }
func (t *Ticket) SLADeadline() *time.Time           { return t.slaDeadline }
func (t *Ticket) FirstResponseAt() *time.Time       { return t.firstResponseAt }
func (t *Ticket) ResolvedAt() *time.Time            { return t.resolvedAt }
func (t *Ticket) ClosedAt() *time.Time              { return t.closedAt }
func (t *Ticket) SatisfactionRating() *int          { return t.satisfactionRating }
func (t *Ticket) SatisfactionComment() string       { return t.satisfactionComment }
func (t *Ticket) Tags() []string                    { return t.tags }
func (t *Ticket) Messages() []Message               { return t.messages }
func (t *Ticket) StatusHistory() []StatusHistory    { return t.statusHistory }
func (t *Ticket) CreatedAt() time.Time              { return t.createdAt }
func (t *Ticket) UpdatedAt() time.Time              { return t.updatedAt }

// ContactEmail returns the email of the ticket creator.
func (t *Ticket) ContactEmail() string {
	return t.guestEmail
}

// ContactName returns the name of the ticket creator.
func (t *Ticket) ContactName() string {
	return t.guestName
}

// IsOverdue checks if the ticket has exceeded its SLA deadline.
func (t *Ticket) IsOverdue() bool {
	if t.slaDeadline == nil {
		return false
	}
	if !t.status.IsActive() {
		return false
	}
	return time.Now().After(*t.slaDeadline)
}

// --- Behavior Methods ---

// Assign assigns the ticket to an agent.
func (t *Ticket) Assign(agentID uuid.UUID, changedBy *uuid.UUID) error {
	if t.status.IsClosed() {
		return ErrCannotModify
	}

	t.assignedTo = &agentID
	t.updatedAt = time.Now()

	if t.status.IsOpen() {
		t.transitionStatus(shared.StatusInProgress, changedBy, "Assigned to agent")
	}

	t.addEvent(NewTicketAssignedEvent(t.id, agentID))
	return nil
}

// Unassign removes the agent assignment.
func (t *Ticket) Unassign(changedBy *uuid.UUID) error {
	if t.assignedTo == nil {
		return ErrNotAssigned
	}

	t.assignedTo = nil
	t.updatedAt = time.Now()

	if t.status.IsInProgress() {
		t.transitionStatus(shared.StatusOpen, changedBy, "Unassigned")
	}

	return nil
}

// Escalate escalates the ticket priority.
func (t *Ticket) Escalate(reason string, changedBy *uuid.UUID) error {
	if t.status.IsClosed() {
		return ErrCannotModify
	}

	currentSeverity := t.priority.Severity()
	if currentSeverity >= 4 {
		return errors.New("ticket is already at highest priority")
	}

	newPriority := shared.PriorityUrgent
	if currentSeverity < 2 {
		newPriority = shared.PriorityNormal
	} else if currentSeverity < 3 {
		newPriority = shared.PriorityHigh
	}

	t.priority = newPriority
	t.slaDeadline = new(time.Time)
	*t.slaDeadline = newPriority.CalculateSLADeadline(time.Now())
	t.updatedAt = time.Now()

	t.addEvent(NewTicketEscalatedEvent(t.id, string(newPriority), reason))
	return nil
}

// SetPending sets the ticket to pending status.
func (t *Ticket) SetPending(reason string, changedBy *uuid.UUID) error {
	return t.transitionStatus(shared.StatusPending, changedBy, reason)
}

// Resolve resolves the ticket.
func (t *Ticket) Resolve(resolution string, changedBy *uuid.UUID) error {
	if err := t.transitionStatus(shared.StatusResolved, changedBy, resolution); err != nil {
		return err
	}

	now := time.Now()
	t.resolvedAt = &now
	t.addEvent(NewTicketResolvedEvent(t.id, resolution))
	return nil
}

// Close closes the ticket.
func (t *Ticket) Close(changedBy *uuid.UUID) error {
	if err := t.transitionStatus(shared.StatusClosed, changedBy, "Ticket closed"); err != nil {
		return err
	}

	now := time.Now()
	t.closedAt = &now
	t.addEvent(NewTicketClosedEvent(t.id))
	return nil
}

// Reopen reopens a resolved ticket.
func (t *Ticket) Reopen(reason string, changedBy *uuid.UUID) error {
	if !t.status.IsResolved() {
		return ErrCannotModify
	}

	t.resolvedAt = nil
	return t.transitionStatus(shared.StatusOpen, changedBy, reason)
}

// AddMessage adds a message to the ticket.
func (t *Ticket) AddMessage(msg Message) {
	t.messages = append(t.messages, msg)
	t.updatedAt = time.Now()

	// Track first response
	if t.firstResponseAt == nil && msg.SenderType().IsAgent() {
		now := time.Now()
		t.firstResponseAt = &now
	}
}

// RateSatisfaction records customer satisfaction.
func (t *Ticket) RateSatisfaction(rating int, comment string) error {
	if !t.status.IsResolved() && !t.status.IsClosed() {
		return errors.New("can only rate resolved or closed tickets")
	}
	if rating < 1 || rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}

	t.satisfactionRating = &rating
	t.satisfactionComment = comment
	t.updatedAt = time.Now()
	return nil
}

// SetCategory sets the ticket category.
func (t *Ticket) SetCategory(categoryID uuid.UUID) {
	t.categoryID = &categoryID
	t.updatedAt = time.Now()
}

// AddTag adds a tag to the ticket.
func (t *Ticket) AddTag(tag string) {
	for _, existingTag := range t.tags {
		if existingTag == tag {
			return
		}
	}
	t.tags = append(t.tags, tag)
	t.updatedAt = time.Now()
}

// RemoveTag removes a tag from the ticket.
func (t *Ticket) RemoveTag(tag string) {
	for i, existingTag := range t.tags {
		if existingTag == tag {
			t.tags = append(t.tags[:i], t.tags[i+1:]...)
			t.updatedAt = time.Now()
			return
		}
	}
}

// transitionStatus transitions the ticket to a new status.
func (t *Ticket) transitionStatus(target shared.TicketStatus, changedBy *uuid.UUID, notes string) error {
	if !t.status.CanTransitionTo(target) {
		return ErrCannotModify
	}

	history := NewStatusHistory(t.id, t.status, target, changedBy, notes)
	t.statusHistory = append(t.statusHistory, history)
	t.status = target
	t.updatedAt = time.Now()

	t.addEvent(NewTicketStatusChangedEvent(t.id, string(target)))
	return nil
}

// Events returns and clears the collected domain events.
func (t *Ticket) Events() []Event {
	events := t.events
	t.events = make([]Event, 0)
	return events
}

func (t *Ticket) addEvent(event Event) {
	t.events = append(t.events, event)
}
