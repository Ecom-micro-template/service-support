package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/Ecom-micro-template/service-support/internal/events"
	"github.com/Ecom-micro-template/service-support/internal/models"
	"github.com/Ecom-micro-template/service-support/internal/repository"
	"go.uber.org/zap"
)

// AdminHandler handles admin support management requests
type AdminHandler struct {
	ticketRepo        *repository.TicketRepository
	messageRepo       *repository.MessageRepository
	categoryRepo      *repository.CategoryRepository
	cannedResponseRepo *repository.CannedResponseRepository
	publisher         *events.Publisher
	logger            *zap.Logger
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	ticketRepo *repository.TicketRepository,
	messageRepo *repository.MessageRepository,
	categoryRepo *repository.CategoryRepository,
	cannedResponseRepo *repository.CannedResponseRepository,
	logger *zap.Logger,
) *AdminHandler {
	return &AdminHandler{
		ticketRepo:        ticketRepo,
		messageRepo:       messageRepo,
		categoryRepo:      categoryRepo,
		cannedResponseRepo: cannedResponseRepo,
		logger:            logger,
	}
}

// SetEventPublisher sets the event publisher
func (h *AdminHandler) SetEventPublisher(publisher *events.Publisher) {
	h.publisher = publisher
}

// ListTickets lists all tickets for admin
// GET /api/v1/admin/support/tickets
func (h *AdminHandler) ListTickets(c *gin.Context) {
	// Parse filters
	filter := repository.TicketFilter{
		Status:   c.Query("status"),
		Priority: c.Query("priority"),
		Search:   c.Query("search"),
	}

	if catID := c.Query("category_id"); catID != "" {
		id, err := uuid.Parse(catID)
		if err == nil {
			filter.CategoryID = &id
		}
	}

	if assignedTo := c.Query("assigned_to"); assignedTo != "" {
		id, err := uuid.Parse(assignedTo)
		if err == nil {
			filter.AssignedTo = &id
		}
	}

	if overdue := c.Query("overdue"); overdue == "true" {
		t := true
		filter.IsOverdue = &t
	}

	filter.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	filter.PerPage, _ = strconv.Atoi(c.DefaultQuery("per_page", "20"))

	tickets, total, err := h.ticketRepo.List(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list tickets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"message": "Failed to retrieve tickets"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tickets,
		"meta": gin.H{
			"page":     filter.Page,
			"per_page": filter.PerPage,
			"total":    total,
		},
	})
}

// GetTicket retrieves a specific ticket for admin
// GET /api/v1/admin/support/tickets/:id
func (h *AdminHandler) GetTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": "Invalid ticket ID"},
		})
		return
	}

	ticket, err := h.ticketRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   gin.H{"message": "Ticket not found"},
		})
		return
	}

	// Include internal notes for admin
	messages, _ := h.messageRepo.GetByTicketID(c.Request.Context(), ticket.ID, true)
	ticket.Messages = messages

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    ticket,
	})
}

// UpdateTicketRequest represents the request to update a ticket
type UpdateTicketRequest struct {
	Status     string     `json:"status"`
	Priority   string     `json:"priority"`
	CategoryID *uuid.UUID `json:"category_id"`
	AssignedTo *uuid.UUID `json:"assigned_to"`
	Tags       []string   `json:"tags"`
}

// UpdateTicket updates a ticket
// PUT /api/v1/admin/support/tickets/:id
func (h *AdminHandler) UpdateTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": "Invalid ticket ID"},
		})
		return
	}

	var req UpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": err.Error()},
		})
		return
	}

	ticket, err := h.ticketRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   gin.H{"message": "Ticket not found"},
		})
		return
	}

	// Get admin info
	adminIDStr, _ := c.Get("user_id")
	var adminID uuid.UUID
	switch v := adminIDStr.(type) {
	case string:
		adminID, _ = uuid.Parse(v)
	case uuid.UUID:
		adminID = v
	}
	adminName, _ := c.Get("email")

	// Update status if changed
	if req.Status != "" && req.Status != string(ticket.Status) {
		err := h.ticketRepo.UpdateStatus(
			c.Request.Context(),
			id,
			models.TicketStatus(req.Status),
			&adminID,
			adminName.(string),
			"",
		)
		if err != nil {
			h.logger.Error("Failed to update ticket status", zap.Error(err))
		}
		ticket.Status = models.TicketStatus(req.Status)
	}

	// Update other fields
	if req.Priority != "" {
		ticket.Priority = models.TicketPriority(req.Priority)
	}
	if req.CategoryID != nil {
		ticket.CategoryID = req.CategoryID
	}
	if req.AssignedTo != nil {
		ticket.AssignedTo = req.AssignedTo
	}
	if req.Tags != nil {
		ticket.Tags = req.Tags
	}

	if err := h.ticketRepo.Update(c.Request.Context(), ticket); err != nil {
		h.logger.Error("Failed to update ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"message": "Failed to update ticket"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    ticket,
		"message": "Ticket updated successfully",
	})
}

// AdminReplyRequest represents admin reply to ticket
type AdminReplyRequest struct {
	Content    string `json:"content" binding:"required"`
	IsInternal bool   `json:"is_internal"`
	Attachments []struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Size     int64  `json:"size"`
		MimeType string `json:"mime_type"`
	} `json:"attachments"`
}

// ReplyToTicket sends admin reply to ticket
// POST /api/v1/admin/support/tickets/:id/reply
func (h *AdminHandler) ReplyToTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": "Invalid ticket ID"},
		})
		return
	}

	var req AdminReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": err.Error()},
		})
		return
	}

	ticket, err := h.ticketRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   gin.H{"message": "Ticket not found"},
		})
		return
	}

	// Get admin info
	adminIDStr, _ := c.Get("user_id")
	var adminID uuid.UUID
	switch v := adminIDStr.(type) {
	case string:
		adminID, _ = uuid.Parse(v)
	case uuid.UUID:
		adminID = v
	}
	adminEmail, _ := c.Get("email")

	// Convert attachments to JSON
	attachmentsJSON, _ := json.Marshal(req.Attachments)

	message := &models.Message{
		TicketID:    id,
		SenderType:  models.SenderTypeAgent,
		SenderID:    &adminID,
		SenderEmail: adminEmail.(string),
		Content:     req.Content,
		Attachments: attachmentsJSON,
		IsInternal:  req.IsInternal,
	}

	if err := h.messageRepo.Create(c.Request.Context(), message); err != nil {
		h.logger.Error("Failed to create reply", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"message": "Failed to send reply"},
		})
		return
	}

	// Publish event for notification (only for external replies)
	if h.publisher != nil && !req.IsInternal {
		h.publisher.PublishTicketReply(ticket, message, true)
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    message,
		"message": "Reply sent successfully",
	})
}

// AssignTicket assigns a ticket to an agent
// PUT /api/v1/admin/support/tickets/:id/assign
func (h *AdminHandler) AssignTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": "Invalid ticket ID"},
		})
		return
	}

	var req struct {
		AgentID uuid.UUID `json:"agent_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": err.Error()},
		})
		return
	}

	if err := h.ticketRepo.Assign(c.Request.Context(), id, req.AgentID); err != nil {
		h.logger.Error("Failed to assign ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"message": "Failed to assign ticket"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Ticket assigned successfully",
	})
}

// GetStats retrieves support statistics
// GET /api/v1/admin/support/stats
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.ticketRepo.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"message": "Failed to retrieve statistics"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// Categories management

// ListCategories lists all support categories
// GET /api/v1/admin/support/categories
func (h *AdminHandler) ListCategories(c *gin.Context) {
	categories, err := h.categoryRepo.List(c.Request.Context(), false)
	if err != nil {
		h.logger.Error("Failed to list categories", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"message": "Failed to retrieve categories"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    categories,
	})
}

// CreateCategoryRequest represents category creation request
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	NameMS      string `json:"name_ms"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SLAHours    int    `json:"sla_hours"`
	Priority    int    `json:"priority"`
	IsActive    *bool  `json:"is_active"`
}

// CreateCategory creates a new support category
// POST /api/v1/admin/support/categories
func (h *AdminHandler) CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": err.Error()},
		})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	slaHours := 24
	if req.SLAHours > 0 {
		slaHours = req.SLAHours
	}

	category := &models.Category{
		Name:        req.Name,
		NameMS:      req.NameMS,
		Description: req.Description,
		Icon:        req.Icon,
		SLAHours:    slaHours,
		Priority:    req.Priority,
		IsActive:    isActive,
	}

	if err := h.categoryRepo.Create(c.Request.Context(), category); err != nil {
		h.logger.Error("Failed to create category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"message": "Failed to create category"},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    category,
		"message": "Category created successfully",
	})
}

// UpdateCategory updates a support category
// PUT /api/v1/admin/support/categories/:id
func (h *AdminHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": "Invalid category ID"},
		})
		return
	}

	category, err := h.categoryRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   gin.H{"message": "Category not found"},
		})
		return
	}

	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": err.Error()},
		})
		return
	}

	if req.Name != "" {
		category.Name = req.Name
	}
	if req.NameMS != "" {
		category.NameMS = req.NameMS
	}
	if req.Description != "" {
		category.Description = req.Description
	}
	if req.Icon != "" {
		category.Icon = req.Icon
	}
	if req.SLAHours > 0 {
		category.SLAHours = req.SLAHours
	}
	if req.Priority > 0 {
		category.Priority = req.Priority
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := h.categoryRepo.Update(c.Request.Context(), category); err != nil {
		h.logger.Error("Failed to update category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"message": "Failed to update category"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    category,
		"message": "Category updated successfully",
	})
}

// DeleteCategory deletes a support category
// DELETE /api/v1/admin/support/categories/:id
func (h *AdminHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": "Invalid category ID"},
		})
		return
	}

	// Check if category has tickets
	count, _ := h.categoryRepo.GetTicketCount(c.Request.Context(), id)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": "Cannot delete category with existing tickets"},
		})
		return
	}

	if err := h.categoryRepo.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"message": "Failed to delete category"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Category deleted successfully",
	})
}

// Canned Responses management

// ListCannedResponses lists all canned responses
// GET /api/v1/admin/support/canned-responses
func (h *AdminHandler) ListCannedResponses(c *gin.Context) {
	var categoryID *uuid.UUID
	if catIDStr := c.Query("category_id"); catIDStr != "" {
		id, err := uuid.Parse(catIDStr)
		if err == nil {
			categoryID = &id
		}
	}

	responses, err := h.cannedResponseRepo.List(c.Request.Context(), categoryID, false)
	if err != nil {
		h.logger.Error("Failed to list canned responses", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"message": "Failed to retrieve canned responses"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
	})
}

// CreateCannedResponse creates a new canned response
// POST /api/v1/admin/support/canned-responses
func (h *AdminHandler) CreateCannedResponse(c *gin.Context) {
	var req struct {
		Title      string     `json:"title" binding:"required"`
		Content    string     `json:"content" binding:"required"`
		CategoryID *uuid.UUID `json:"category_id"`
		Shortcut   string     `json:"shortcut"`
		IsActive   *bool      `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": err.Error()},
		})
		return
	}

	// Get creator ID
	creatorIDStr, _ := c.Get("user_id")
	var creatorID uuid.UUID
	switch v := creatorIDStr.(type) {
	case string:
		creatorID, _ = uuid.Parse(v)
	case uuid.UUID:
		creatorID = v
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	response := &models.CannedResponse{
		Title:      req.Title,
		Content:    req.Content,
		CategoryID: req.CategoryID,
		Shortcut:   req.Shortcut,
		IsActive:   isActive,
		CreatedBy:  &creatorID,
	}

	if err := h.cannedResponseRepo.Create(c.Request.Context(), response); err != nil {
		h.logger.Error("Failed to create canned response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"message": "Failed to create canned response"},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
		"message": "Canned response created successfully",
	})
}

// UpdateCannedResponse updates a canned response
// PUT /api/v1/admin/support/canned-responses/:id
func (h *AdminHandler) UpdateCannedResponse(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": "Invalid canned response ID"},
		})
		return
	}

	response, err := h.cannedResponseRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   gin.H{"message": "Canned response not found"},
		})
		return
	}

	var req struct {
		Title      string     `json:"title"`
		Content    string     `json:"content"`
		CategoryID *uuid.UUID `json:"category_id"`
		Shortcut   string     `json:"shortcut"`
		IsActive   *bool      `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": err.Error()},
		})
		return
	}

	if req.Title != "" {
		response.Title = req.Title
	}
	if req.Content != "" {
		response.Content = req.Content
	}
	if req.CategoryID != nil {
		response.CategoryID = req.CategoryID
	}
	if req.Shortcut != "" {
		response.Shortcut = req.Shortcut
	}
	if req.IsActive != nil {
		response.IsActive = *req.IsActive
	}

	if err := h.cannedResponseRepo.Update(c.Request.Context(), response); err != nil {
		h.logger.Error("Failed to update canned response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"message": "Failed to update canned response"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Canned response updated successfully",
	})
}

// DeleteCannedResponse deletes a canned response
// DELETE /api/v1/admin/support/canned-responses/:id
func (h *AdminHandler) DeleteCannedResponse(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": "Invalid canned response ID"},
		})
		return
	}

	if err := h.cannedResponseRepo.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete canned response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"message": "Failed to delete canned response"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Canned response deleted successfully",
	})
}
