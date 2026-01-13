package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/Ecom-micro-template/service-support/internal/domain"
	"gorm.io/gorm"
)

// CannedResponseRepository handles database operations for canned responses
type CannedResponseRepository struct {
	db *gorm.DB
}

// NewCannedResponseRepository creates a new canned response repository
func NewCannedResponseRepository(db *gorm.DB) *CannedResponseRepository {
	return &CannedResponseRepository{db: db}
}

// List retrieves all canned responses
func (r *CannedResponseRepository) List(ctx context.Context, categoryID *uuid.UUID, onlyActive bool) ([]models.CannedResponse, error) {
	var responses []models.CannedResponse
	query := r.db.WithContext(ctx).Order("usage_count DESC, title ASC")

	if categoryID != nil {
		query = query.Where("category_id = ? OR category_id IS NULL", categoryID)
	}

	if onlyActive {
		query = query.Where("is_active = ?", true)
	}

	err := query.Find(&responses).Error
	return responses, err
}

// GetByID retrieves a canned response by ID
func (r *CannedResponseRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CannedResponse, error) {
	var response models.CannedResponse
	err := r.db.WithContext(ctx).First(&response, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// GetByShortcut retrieves a canned response by shortcut
func (r *CannedResponseRepository) GetByShortcut(ctx context.Context, shortcut string) (*models.CannedResponse, error) {
	var response models.CannedResponse
	err := r.db.WithContext(ctx).
		Where("shortcut = ? AND is_active = ?", shortcut, true).
		First(&response).Error
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Create creates a new canned response
func (r *CannedResponseRepository) Create(ctx context.Context, response *models.CannedResponse) error {
	return r.db.WithContext(ctx).Create(response).Error
}

// Update updates a canned response
func (r *CannedResponseRepository) Update(ctx context.Context, response *models.CannedResponse) error {
	return r.db.WithContext(ctx).Save(response).Error
}

// Delete deletes a canned response
func (r *CannedResponseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.CannedResponse{}, "id = ?", id).Error
}

// IncrementUsage increments the usage count of a canned response
func (r *CannedResponseRepository) IncrementUsage(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.CannedResponse{}).
		Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}
