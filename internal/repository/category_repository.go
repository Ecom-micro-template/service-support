package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/niaga-platform/service-support/internal/models"
	"gorm.io/gorm"
)

// CategoryRepository handles database operations for support categories
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// List retrieves all active categories
func (r *CategoryRepository) List(ctx context.Context, onlyActive bool) ([]models.Category, error) {
	var categories []models.Category
	query := r.db.WithContext(ctx).Order("priority ASC, name ASC")

	if onlyActive {
		query = query.Where("is_active = ?", true)
	}

	err := query.Find(&categories).Error
	return categories, err
}

// GetByID retrieves a category by ID
func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	var category models.Category
	err := r.db.WithContext(ctx).First(&category, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// Create creates a new category
func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

// Update updates a category
func (r *CategoryRepository) Update(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

// Delete deletes a category
func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Category{}, "id = ?", id).Error
}

// GetTicketCount returns the number of tickets in a category
func (r *CategoryRepository) GetTicketCount(ctx context.Context, categoryID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Ticket{}).
		Where("category_id = ?", categoryID).
		Count(&count).Error
	return count, err
}
