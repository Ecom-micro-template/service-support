package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/Ecom-micro-template/service-support/internal/domain"
	"gorm.io/gorm"
)

// TicketRepository handles database operations for tickets
type TicketRepository struct {
	db *gorm.DB
}

// NewTicketRepository creates a new ticket repository
func NewTicketRepository(db *gorm.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

// TicketFilter represents filters for listing tickets
type TicketFilter struct {
	Status      string
	Priority    string
	CategoryID  *uuid.UUID
	CustomerID  *uuid.UUID
	AssignedTo  *uuid.UUID
	OrderID     *uuid.UUID
	Search      string
	IsOverdue   *bool
	Page        int
	PerPage     int
}

// TicketStats represents ticket statistics
type TicketStats struct {
	TotalOpen       int64   `json:"total_open"`
	TotalPending    int64   `json:"total_pending"`
	TotalInProgress int64   `json:"total_in_progress"`
	TotalResolved   int64   `json:"total_resolved"`
	TotalClosed     int64   `json:"total_closed"`
	TotalOverdue    int64   `json:"total_overdue"`
	AvgResponseTime float64 `json:"avg_response_time_hours"`
	AvgResolutionTime float64 `json:"avg_resolution_time_hours"`
	SatisfactionRate float64 `json:"satisfaction_rate"`
}

// Create creates a new ticket
func (r *TicketRepository) Create(ctx context.Context, ticket *models.Ticket) error {
	return r.db.WithContext(ctx).Create(ticket).Error
}

// GetByID retrieves a ticket by ID
func (r *TicketRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Ticket, error) {
	var ticket models.Ticket
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&ticket, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

// GetByTicketNumber retrieves a ticket by ticket number
func (r *TicketRepository) GetByTicketNumber(ctx context.Context, ticketNumber string) (*models.Ticket, error) {
	var ticket models.Ticket
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&ticket, "ticket_number = ?", ticketNumber).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

// List retrieves tickets with filters
func (r *TicketRepository) List(ctx context.Context, filter TicketFilter) ([]models.Ticket, int64, error) {
	var tickets []models.Ticket
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Ticket{})

	// Apply filters
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Priority != "" {
		query = query.Where("priority = ?", filter.Priority)
	}
	if filter.CategoryID != nil {
		query = query.Where("category_id = ?", filter.CategoryID)
	}
	if filter.CustomerID != nil {
		query = query.Where("customer_id = ?", filter.CustomerID)
	}
	if filter.AssignedTo != nil {
		query = query.Where("assigned_to = ?", filter.AssignedTo)
	}
	if filter.OrderID != nil {
		query = query.Where("order_id = ?", filter.OrderID)
	}
	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		query = query.Where("subject ILIKE ? OR ticket_number ILIKE ? OR guest_email ILIKE ? OR guest_name ILIKE ?",
			search, search, search, search)
	}
	if filter.IsOverdue != nil && *filter.IsOverdue {
		query = query.Where("sla_deadline < ? AND status NOT IN ('resolved', 'closed')", time.Now())
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	if filter.PerPage <= 0 {
		filter.PerPage = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.PerPage

	// Fetch with preloads
	err := query.
		Preload("Category").
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.PerPage).
		Find(&tickets).Error
	if err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

// ListByCustomer retrieves tickets for a specific customer
func (r *TicketRepository) ListByCustomer(ctx context.Context, customerID uuid.UUID, page, perPage int) ([]models.Ticket, int64, error) {
	return r.List(ctx, TicketFilter{
		CustomerID: &customerID,
		Page:       page,
		PerPage:    perPage,
	})
}

// ListByEmail retrieves tickets for a specific email (for guests)
func (r *TicketRepository) ListByEmail(ctx context.Context, email string, page, perPage int) ([]models.Ticket, int64, error) {
	var tickets []models.Ticket
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Ticket{}).Where("guest_email = ?", email)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if perPage <= 0 {
		perPage = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * perPage

	err := query.
		Preload("Category").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&tickets).Error
	if err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

// Update updates a ticket
func (r *TicketRepository) Update(ctx context.Context, ticket *models.Ticket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

// UpdateStatus updates ticket status and records history
func (r *TicketRepository) UpdateStatus(ctx context.Context, ticketID uuid.UUID, newStatus models.TicketStatus, changedBy *uuid.UUID, changedByName, notes string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get current ticket
		var ticket models.Ticket
		if err := tx.First(&ticket, "id = ?", ticketID).Error; err != nil {
			return err
		}

		oldStatus := string(ticket.Status)

		// Update ticket status
		updates := map[string]interface{}{
			"status":     newStatus,
			"updated_at": time.Now(),
		}

		if newStatus == models.TicketStatusResolved && ticket.ResolvedAt == nil {
			now := time.Now()
			updates["resolved_at"] = now
		}
		if newStatus == models.TicketStatusClosed && ticket.ClosedAt == nil {
			now := time.Now()
			updates["closed_at"] = now
		}

		if err := tx.Model(&ticket).Updates(updates).Error; err != nil {
			return err
		}

		// Create status history
		history := &models.StatusHistory{
			TicketID:      ticketID,
			FromStatus:    oldStatus,
			ToStatus:      string(newStatus),
			ChangedBy:     changedBy,
			ChangedByName: changedByName,
			Notes:         notes,
		}

		return tx.Create(history).Error
	})
}

// Assign assigns a ticket to an agent
func (r *TicketRepository) Assign(ctx context.Context, ticketID, agentID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Ticket{}).
		Where("id = ?", ticketID).
		Update("assigned_to", agentID).Error
}

// Delete soft deletes a ticket
func (r *TicketRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Ticket{}, "id = ?", id).Error
}

// GetStats returns ticket statistics
func (r *TicketRepository) GetStats(ctx context.Context) (*TicketStats, error) {
	stats := &TicketStats{}

	// Count by status
	r.db.WithContext(ctx).Model(&models.Ticket{}).
		Where("status = ?", models.TicketStatusOpen).
		Count(&stats.TotalOpen)
	r.db.WithContext(ctx).Model(&models.Ticket{}).
		Where("status = ?", models.TicketStatusPending).
		Count(&stats.TotalPending)
	r.db.WithContext(ctx).Model(&models.Ticket{}).
		Where("status = ?", models.TicketStatusInProgress).
		Count(&stats.TotalInProgress)
	r.db.WithContext(ctx).Model(&models.Ticket{}).
		Where("status = ?", models.TicketStatusResolved).
		Count(&stats.TotalResolved)
	r.db.WithContext(ctx).Model(&models.Ticket{}).
		Where("status = ?", models.TicketStatusClosed).
		Count(&stats.TotalClosed)

	// Count overdue
	r.db.WithContext(ctx).Model(&models.Ticket{}).
		Where("sla_deadline < ? AND status NOT IN ('resolved', 'closed')", time.Now()).
		Count(&stats.TotalOverdue)

	// Calculate average response time (hours)
	var avgResponse struct {
		Avg float64
	}
	r.db.WithContext(ctx).Model(&models.Ticket{}).
		Select("AVG(EXTRACT(EPOCH FROM (first_response_at - created_at)) / 3600) as avg").
		Where("first_response_at IS NOT NULL").
		Scan(&avgResponse)
	stats.AvgResponseTime = avgResponse.Avg

	// Calculate average resolution time (hours)
	var avgResolution struct {
		Avg float64
	}
	r.db.WithContext(ctx).Model(&models.Ticket{}).
		Select("AVG(EXTRACT(EPOCH FROM (resolved_at - created_at)) / 3600) as avg").
		Where("resolved_at IS NOT NULL").
		Scan(&avgResolution)
	stats.AvgResolutionTime = avgResolution.Avg

	// Calculate satisfaction rate (percentage of 4-5 ratings)
	var satisfactionData struct {
		Total   int64
		Satisfied int64
	}
	r.db.WithContext(ctx).Model(&models.Ticket{}).
		Select("COUNT(*) as total, COUNT(CASE WHEN satisfaction_rating >= 4 THEN 1 END) as satisfied").
		Where("satisfaction_rating IS NOT NULL").
		Scan(&satisfactionData)
	if satisfactionData.Total > 0 {
		stats.SatisfactionRate = float64(satisfactionData.Satisfied) / float64(satisfactionData.Total) * 100
	}

	return stats, nil
}
