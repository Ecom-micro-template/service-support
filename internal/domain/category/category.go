package category

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Domain errors for Category entity
var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrInvalidCategory  = errors.New("invalid category data")
)

// Category represents a support ticket category.
type Category struct {
	id          uuid.UUID
	name        string
	nameMS      string // Malay translation
	description string
	icon        string
	slaHours    int
	priority    int
	isActive    bool
	createdAt   time.Time
	updatedAt   time.Time
}

// CategoryParams contains parameters for creating a Category.
type CategoryParams struct {
	ID          uuid.UUID
	Name        string
	NameMS      string
	Description string
	Icon        string
	SLAHours    int
	Priority    int
	IsActive    bool
}

// NewCategory creates a new Category entity.
func NewCategory(params CategoryParams) (*Category, error) {
	if params.Name == "" {
		return nil, errors.New("name is required")
	}

	id := params.ID
	if id == uuid.Nil {
		id = uuid.New()
	}

	slaHours := params.SLAHours
	if slaHours <= 0 {
		slaHours = 24 // Default 24 hours
	}

	now := time.Now()
	return &Category{
		id:          id,
		name:        params.Name,
		nameMS:      params.NameMS,
		description: params.Description,
		icon:        params.Icon,
		slaHours:    slaHours,
		priority:    params.Priority,
		isActive:    params.IsActive,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// Getters
func (c *Category) ID() uuid.UUID        { return c.id }
func (c *Category) Name() string         { return c.name }
func (c *Category) NameMS() string       { return c.nameMS }
func (c *Category) Description() string  { return c.description }
func (c *Category) Icon() string         { return c.icon }
func (c *Category) SLAHours() int        { return c.slaHours }
func (c *Category) Priority() int        { return c.priority }
func (c *Category) IsActive() bool       { return c.isActive }
func (c *Category) CreatedAt() time.Time { return c.createdAt }
func (c *Category) UpdatedAt() time.Time { return c.updatedAt }

// LocalizedName returns the localized name.
func (c *Category) LocalizedName(lang string) string {
	if lang == "ms" && c.nameMS != "" {
		return c.nameMS
	}
	return c.name
}

// --- Behavior Methods ---

// Update updates the category details.
func (c *Category) Update(name, nameMS, description, icon string) {
	if name != "" {
		c.name = name
	}
	c.nameMS = nameMS
	c.description = description
	c.icon = icon
	c.updatedAt = time.Now()
}

// SetSLAHours sets the SLA hours.
func (c *Category) SetSLAHours(hours int) {
	if hours > 0 {
		c.slaHours = hours
		c.updatedAt = time.Now()
	}
}

// SetPriority sets the display priority.
func (c *Category) SetPriority(priority int) {
	c.priority = priority
	c.updatedAt = time.Now()
}

// Activate activates the category.
func (c *Category) Activate() {
	c.isActive = true
	c.updatedAt = time.Now()
}

// Deactivate deactivates the category.
func (c *Category) Deactivate() {
	c.isActive = false
	c.updatedAt = time.Now()
}
