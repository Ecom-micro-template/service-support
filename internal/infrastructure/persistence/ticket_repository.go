package persistence

import (
	"context"

	"github.com/google/uuid"
	"github.com/Ecom-micro-template/service-support/internal/domain/shared"
	"github.com/Ecom-micro-template/service-support/internal/domain/ticket"
	"gorm.io/gorm"
)

// TicketRepository interface for ticket domain.
type TicketRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*ticket.Ticket, error)
	GetByNumber(ctx context.Context, number string) (*ticket.Ticket, error)
	List(ctx context.Context, page, pageSize int, status *string, assignedTo *uuid.UUID) ([]*ticket.Ticket, int64, error)
	GetOverdue(ctx context.Context) ([]*ticket.Ticket, error)
	Save(ctx context.Context, t *ticket.Ticket) error
	Update(ctx context.Context, t *ticket.Ticket) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// GormTicketRepository is the GORM implementation of TicketRepository.
type GormTicketRepository struct {
	db *gorm.DB
}

// NewTicketRepository creates a new GormTicketRepository.
func NewTicketRepository(db *gorm.DB) TicketRepository {
	return &GormTicketRepository{db: db}
}

// GetByID retrieves a ticket by ID.
func (r *GormTicketRepository) GetByID(ctx context.Context, id uuid.UUID) (*ticket.Ticket, error) {
	var model TicketModel
	err := r.db.WithContext(ctx).
		Preload("Messages").
		Preload("Category").
		Where("id = ?", id).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ticket.ErrTicketNotFound
		}
		return nil, err
	}
	return r.toDomain(&model)
}

// GetByNumber retrieves a ticket by ticket number.
func (r *GormTicketRepository) GetByNumber(ctx context.Context, number string) (*ticket.Ticket, error) {
	var model TicketModel
	err := r.db.WithContext(ctx).
		Preload("Messages").
		Where("ticket_number = ?", number).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ticket.ErrTicketNotFound
		}
		return nil, err
	}
	return r.toDomain(&model)
}

// List retrieves tickets with pagination and filters.
func (r *GormTicketRepository) List(ctx context.Context, page, pageSize int, status *string, assignedTo *uuid.UUID) ([]*ticket.Ticket, int64, error) {
	var models []TicketModel
	var total int64

	query := r.db.WithContext(ctx).Model(&TicketModel{})
	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
	}
	if assignedTo != nil {
		query = query.Where("assigned_to = ?", *assignedTo)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	tickets := make([]*ticket.Ticket, len(models))
	for i, m := range models {
		t, err := r.toDomain(&m)
		if err != nil {
			return nil, 0, err
		}
		tickets[i] = t
	}
	return tickets, total, nil
}

// GetOverdue retrieves overdue tickets.
func (r *GormTicketRepository) GetOverdue(ctx context.Context) ([]*ticket.Ticket, error) {
	var models []TicketModel

	err := r.db.WithContext(ctx).
		Where("sla_deadline < NOW()").
		Where("status NOT IN ?", []string{"resolved", "closed"}).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	tickets := make([]*ticket.Ticket, len(models))
	for i, m := range models {
		t, err := r.toDomain(&m)
		if err != nil {
			return nil, err
		}
		tickets[i] = t
	}
	return tickets, nil
}

// Save creates a new ticket.
func (r *GormTicketRepository) Save(ctx context.Context, t *ticket.Ticket) error {
	model := r.toModel(t)
	return r.db.WithContext(ctx).Create(model).Error
}

// Update updates an existing ticket.
func (r *GormTicketRepository) Update(ctx context.Context, t *ticket.Ticket) error {
	model := r.toModel(t)
	return r.db.WithContext(ctx).Save(model).Error
}

// Delete soft-deletes a ticket.
func (r *GormTicketRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&TicketModel{}, "id = ?", id).Error
}

// toDomain converts model to domain entity.
func (r *GormTicketRepository) toDomain(m *TicketModel) (*ticket.Ticket, error) {
	return ticket.NewTicket(ticket.TicketParams{
		ID:           m.ID,
		TicketNumber: m.TicketNumber,
		CustomerID:   m.CustomerID,
		GuestEmail:   m.GuestEmail,
		GuestName:    m.GuestName,
		GuestPhone:   m.GuestPhone,
		CategoryID:   m.CategoryID,
		Subject:      m.Subject,
		Priority:     m.Priority,
		OrderID:      m.OrderID,
		OrderNumber:  m.OrderNumber,
		Tags:         m.Tags,
	})
}

// toModel converts domain entity to model.
func (r *GormTicketRepository) toModel(t *ticket.Ticket) *TicketModel {
	status, _ := shared.ParseTicketStatus(string(t.Status()))
	priority, _ := shared.ParseTicketPriority(string(t.Priority()))

	return &TicketModel{
		ID:                  t.ID(),
		TicketNumber:        t.TicketNumber().Value(),
		CustomerID:          t.CustomerID(),
		GuestEmail:          t.GuestEmail(),
		GuestName:           t.GuestName(),
		GuestPhone:          t.GuestPhone(),
		CategoryID:          t.CategoryID(),
		Subject:             t.Subject(),
		Status:              status.String(),
		Priority:            priority.String(),
		AssignedTo:          t.AssignedTo(),
		OrderID:             t.OrderID(),
		OrderNumber:         t.OrderNumber(),
		SLADeadline:         t.SLADeadline(),
		FirstResponseAt:     t.FirstResponseAt(),
		ResolvedAt:          t.ResolvedAt(),
		ClosedAt:            t.ClosedAt(),
		SatisfactionRating:  t.SatisfactionRating(),
		SatisfactionComment: t.SatisfactionComment(),
		Tags:                t.Tags(),
		CreatedAt:           t.CreatedAt(),
		UpdatedAt:           t.UpdatedAt(),
	}
}
