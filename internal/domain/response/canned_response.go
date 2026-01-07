package response

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Domain errors for CannedResponse entity
var (
	ErrResponseNotFound = errors.New("canned response not found")
	ErrInvalidResponse  = errors.New("invalid canned response data")
)

// CannedResponse represents a pre-written response template.
type CannedResponse struct {
	id         uuid.UUID
	title      string
	content    string
	categoryID *uuid.UUID
	shortcut   string
	isActive   bool
	usageCount int
	createdBy  *uuid.UUID
	createdAt  time.Time
	updatedAt  time.Time
}

// CannedResponseParams contains parameters for creating a CannedResponse.
type CannedResponseParams struct {
	ID         uuid.UUID
	Title      string
	Content    string
	CategoryID *uuid.UUID
	Shortcut   string
	IsActive   bool
	CreatedBy  *uuid.UUID
}

// NewCannedResponse creates a new CannedResponse entity.
func NewCannedResponse(params CannedResponseParams) (*CannedResponse, error) {
	if params.Title == "" {
		return nil, errors.New("title is required")
	}
	if params.Content == "" {
		return nil, errors.New("content is required")
	}

	id := params.ID
	if id == uuid.Nil {
		id = uuid.New()
	}

	now := time.Now()
	return &CannedResponse{
		id:         id,
		title:      params.Title,
		content:    params.Content,
		categoryID: params.CategoryID,
		shortcut:   params.Shortcut,
		isActive:   params.IsActive,
		usageCount: 0,
		createdBy:  params.CreatedBy,
		createdAt:  now,
		updatedAt:  now,
	}, nil
}

// Getters
func (r *CannedResponse) ID() uuid.UUID          { return r.id }
func (r *CannedResponse) Title() string          { return r.title }
func (r *CannedResponse) Content() string        { return r.content }
func (r *CannedResponse) CategoryID() *uuid.UUID { return r.categoryID }
func (r *CannedResponse) Shortcut() string       { return r.shortcut }
func (r *CannedResponse) IsActive() bool         { return r.isActive }
func (r *CannedResponse) UsageCount() int        { return r.usageCount }
func (r *CannedResponse) CreatedBy() *uuid.UUID  { return r.createdBy }
func (r *CannedResponse) CreatedAt() time.Time   { return r.createdAt }
func (r *CannedResponse) UpdatedAt() time.Time   { return r.updatedAt }

// --- Behavior Methods ---

// Update updates the response details.
func (r *CannedResponse) Update(title, content, shortcut string) {
	if title != "" {
		r.title = title
	}
	if content != "" {
		r.content = content
	}
	r.shortcut = shortcut
	r.updatedAt = time.Now()
}

// SetCategory sets the category.
func (r *CannedResponse) SetCategory(categoryID *uuid.UUID) {
	r.categoryID = categoryID
	r.updatedAt = time.Now()
}

// IncrementUsage increments the usage count.
func (r *CannedResponse) IncrementUsage() {
	r.usageCount++
	r.updatedAt = time.Now()
}

// Activate activates the response.
func (r *CannedResponse) Activate() {
	r.isActive = true
	r.updatedAt = time.Now()
}

// Deactivate deactivates the response.
func (r *CannedResponse) Deactivate() {
	r.isActive = false
	r.updatedAt = time.Now()
}

// HasShortcut returns true if response has a shortcut.
func (r *CannedResponse) HasShortcut() bool {
	return r.shortcut != ""
}
