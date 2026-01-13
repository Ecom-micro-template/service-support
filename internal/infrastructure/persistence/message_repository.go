package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/Ecom-micro-template/service-support/internal/domain"
	"gorm.io/gorm"
)

// MessageRepository handles database operations for messages
type MessageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create creates a new message and updates ticket
func (r *MessageRepository) Create(ctx context.Context, message *domain.Message) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the message
		if err := tx.Create(message).Error; err != nil {
			return err
		}

		// Update ticket's updated_at
		updates := map[string]interface{}{
			"updated_at": time.Now(),
		}

		// If this is the first agent response, record first_response_at
		if message.SenderType == domain.SenderTypeAgent {
			var ticket domain.Ticket
			if err := tx.First(&ticket, "id = ?", message.TicketID).Error; err != nil {
				return err
			}

			if ticket.FirstResponseAt == nil {
				updates["first_response_at"] = time.Now()
			}

			// Update status to in_progress if currently open
			if ticket.Status == domain.TicketStatusOpen {
				updates["status"] = domain.TicketStatusInProgress
			}
		}

		return tx.Model(&domain.Ticket{}).
			Where("id = ?", message.TicketID).
			Updates(updates).Error
	})
}

// GetByTicketID retrieves all messages for a ticket
func (r *MessageRepository) GetByTicketID(ctx context.Context, ticketID uuid.UUID, includeInternal bool) ([]domain.Message, error) {
	var messages []domain.Message
	query := r.db.WithContext(ctx).Where("ticket_id = ?", ticketID)

	if !includeInternal {
		query = query.Where("is_internal = ?", false)
	}

	err := query.Order("created_at ASC").Find(&messages).Error
	return messages, err
}

// GetByID retrieves a message by ID
func (r *MessageRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	var message domain.Message
	err := r.db.WithContext(ctx).First(&message, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// MarkAsRead marks a message as read
func (r *MessageRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.Message{}).
		Where("id = ? AND read_at IS NULL", id).
		Update("read_at", now).Error
}

// MarkAllAsRead marks all messages in a ticket as read
func (r *MessageRepository) MarkAllAsRead(ctx context.Context, ticketID uuid.UUID, senderType domain.SenderType) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.Message{}).
		Where("ticket_id = ? AND sender_type != ? AND read_at IS NULL", ticketID, senderType).
		Update("read_at", now).Error
}

// GetUnreadCount returns count of unread messages for a ticket
func (r *MessageRepository) GetUnreadCount(ctx context.Context, ticketID uuid.UUID, forSenderType domain.SenderType) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Message{}).
		Where("ticket_id = ? AND sender_type != ? AND read_at IS NULL", ticketID, forSenderType).
		Count(&count).Error
	return count, err
}
