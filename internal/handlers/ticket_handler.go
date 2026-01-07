package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/niaga-platform/service-support/internal/events"
	"github.com/niaga-platform/service-support/internal/models"
	"github.com/niaga-platform/service-support/internal/repository"
	"go.uber.org/zap"
)

// TicketHandler handles ticket-related requests
type TicketHandler struct {
	ticketRepo  *repository.TicketRepository
	messageRepo *repository.MessageRepository
	publisher   *events.Publisher
	logger      *zap.Logger
}

// NewTicketHandler creates a new ticket handler
func NewTicketHandler(
	ticketRepo *repository.TicketRepository,
	messageRepo *repository.MessageRepository,
	logger *zap.Logger,
) *TicketHandler {
	return &TicketHandler{
		ticketRepo:  ticketRepo,
		messageRepo: messageRepo,
		logger:      logger,
	}
}

// SetEventPublisher sets the event publisher for notifications
func (h *TicketHandler) SetEventPublisher(publisher *events.Publisher) {
	h.publisher = publisher
}

// CreateTicketRequest represents the request to create a ticket
type CreateTicketRequest struct {
	Subject     string     `json:"subject" binding:"required"`
	Message     string     `json:"message" binding:"required"`
	CategoryID  *uuid.UUID `json:"category_id"`
	Priority    string     `json:"priority"`
	OrderID     *uuid.UUID `json:"order_id"`
	OrderNumber string     `json:"order_number"`
	// For guest contact form
	GuestEmail  string `json:"guest_email"`
	GuestName   string `json:"guest_name"`
	GuestPhone  string `json:"guest_phone"`
}

// Create creates a new support ticket (authenticated user)
// POST /api/v1/support/tickets
func (h *TicketHandler) Create(c *gin.Context) {
	var req CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": err.Error()},
		})
		return
	}

	// Get customer ID from auth context
	customerIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   gin.H{"message": "Not authenticated"},
		})
		return
	}

	var customerID uuid.UUID
	switch v := customerIDStr.(type) {
	case string:
		var err error
		customerID, err = uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   gin.H{"message": "Invalid user ID"},
			})
			return
		}
	case uuid.UUID:
		customerID = v
	}

	// Set priority
	priority := models.TicketPriorityNormal
	if req.Priority != "" {
		priority = models.TicketPriority(req.Priority)
	}

	// Create ticket
	ticket := &models.Ticket{
		CustomerID:  &customerID,
		Subject:     req.Subject,
		CategoryID:  req.CategoryID,
		Priority:    priority,
		Status:      models.TicketStatusOpen,
		OrderID:     req.OrderID,
		OrderNumber: req.OrderNumber,
	}

	if err := h.ticketRepo.Create(c.Request.Context(), ticket); err != nil {
		h.logger.Error("Failed to create ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"message": "Failed to create ticket"},
		})
		return
	}

	// Create initial message
	message := &models.Message{
		TicketID:   ticket.ID,
		SenderType: models.SenderTypeCustomer,
		SenderID:   &customerID,
		Content:    req.Message,
	}

	if err := h.messageRepo.Create(c.Request.Context(), message); err != nil {
		h.logger.Error("Failed to create initial message", zap.Error(err))
	}

	// Publish event for notification
	if h.publisher != nil {
		h.publisher.PublishTicketCreated(ticket)
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    ticket,
		"message": "Ticket created successfully",
	})
}

// SubmitContactForm handles guest contact form submission
// POST /api/v1/support/contact
func (h *TicketHandler) SubmitContactForm(c *gin.Context) {
	var req CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": err.Error()},
		})
		return
	}

	// Validate guest info
	if req.GuestEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": "Email is required"},
		})
		return
	}

	// Create ticket for guest
	ticket := &models.Ticket{
		GuestEmail: req.GuestEmail,
		GuestName:  req.GuestName,
		GuestPhone: req.GuestPhone,
		Subject:    req.Subject,
		CategoryID: req.CategoryID,
		Priority:   models.TicketPriorityNormal,
		Status:     models.TicketStatusOpen,
	}

	if err := h.ticketRepo.Create(c.Request.Context(), ticket); err != nil {
		h.logger.Error("Failed to create contact form ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"message": "Failed to submit contact form"},
		})
		return
	}

	// Create initial message
	message := &models.Message{
		TicketID:    ticket.ID,
		SenderType:  models.SenderTypeCustomer,
		SenderName:  req.GuestName,
		SenderEmail: req.GuestEmail,
		Content:     req.Message,
	}

	if err := h.messageRepo.Create(c.Request.Context(), message); err != nil {
		h.logger.Error("Failed to create initial message", zap.Error(err))
	}

	// Publish event for notification
	if h.publisher != nil {
		h.publisher.PublishTicketCreated(ticket)
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"ticket_number": ticket.TicketNumber,
		},
		"message": "Thank you for contacting us. We will respond to your inquiry soon.",
	})
}

// List lists tickets for authenticated user
// GET /api/v1/support/tickets
func (h *TicketHandler) List(c *gin.Context) {
	// Get customer ID from auth context
	customerIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   gin.H{"message": "Not authenticated"},
		})
		return
	}

	var customerID uuid.UUID
	switch v := customerIDStr.(type) {
	case string:
		var err error
		customerID, err = uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   gin.H{"message": "Invalid user ID"},
			})
			return
		}
	case uuid.UUID:
		customerID = v
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	tickets, total, err := h.ticketRepo.ListByCustomer(c.Request.Context(), customerID, page, perPage)
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
			"page":     page,
			"per_page": perPage,
			"total":    total,
		},
	})
}

// GetByID retrieves a specific ticket
// GET /api/v1/support/tickets/:id
func (h *TicketHandler) GetByID(c *gin.Context) {
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

	// Verify ownership for authenticated users
	customerIDStr, exists := c.Get("user_id")
	if exists {
		var customerID uuid.UUID
		switch v := customerIDStr.(type) {
		case string:
			customerID, _ = uuid.Parse(v)
		case uuid.UUID:
			customerID = v
		}

		if ticket.CustomerID != nil && *ticket.CustomerID != customerID {
			// Check if user is admin
			role, _ := c.Get("role")
			if role != "admin" && role != "super_admin" && role != "support" {
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"error":   gin.H{"message": "Access denied"},
				})
				return
			}
		}
	}

	// Get messages (exclude internal notes for customers)
	includeInternal := false
	role, _ := c.Get("role")
	if role == "admin" || role == "super_admin" || role == "support" {
		includeInternal = true
	}

	messages, _ := h.messageRepo.GetByTicketID(c.Request.Context(), ticket.ID, includeInternal)
	ticket.Messages = messages

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    ticket,
	})
}

// AddMessageRequest represents the request to add a message
type AddMessageRequest struct {
	Content     string `json:"content" binding:"required"`
	Attachments []struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Size     int64  `json:"size"`
		MimeType string `json:"mime_type"`
	} `json:"attachments"`
}

// AddMessage adds a message to a ticket
// POST /api/v1/support/tickets/:id/messages
func (h *TicketHandler) AddMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": "Invalid ticket ID"},
		})
		return
	}

	var req AddMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": err.Error()},
		})
		return
	}

	// Verify ticket exists and user has access
	ticket, err := h.ticketRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   gin.H{"message": "Ticket not found"},
		})
		return
	}

	// Get user info
	customerIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   gin.H{"message": "Not authenticated"},
		})
		return
	}

	var senderID uuid.UUID
	switch v := customerIDStr.(type) {
	case string:
		senderID, _ = uuid.Parse(v)
	case uuid.UUID:
		senderID = v
	}

	// Determine sender type
	senderType := models.SenderTypeCustomer
	if ticket.CustomerID != nil && *ticket.CustomerID != senderID {
		role, _ := c.Get("role")
		if role == "admin" || role == "super_admin" || role == "support" {
			senderType = models.SenderTypeAgent
		} else {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   gin.H{"message": "Access denied"},
			})
			return
		}
	}

	// Convert attachments to JSON
	attachmentsJSON, _ := json.Marshal(req.Attachments)

	message := &models.Message{
		TicketID:    id,
		SenderType:  senderType,
		SenderID:    &senderID,
		Content:     req.Content,
		Attachments: attachmentsJSON,
	}

	if err := h.messageRepo.Create(c.Request.Context(), message); err != nil {
		h.logger.Error("Failed to create message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"message": "Failed to send message"},
		})
		return
	}

	// Publish event for notification
	if h.publisher != nil {
		h.publisher.PublishTicketReply(ticket, message, senderType == models.SenderTypeAgent)
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    message,
		"message": "Message sent successfully",
	})
}

// RateTicket allows customer to rate resolved ticket
// POST /api/v1/support/tickets/:id/rate
func (h *TicketHandler) RateTicket(c *gin.Context) {
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
		Rating  int    `json:"rating" binding:"required,min=1,max=5"`
		Comment string `json:"comment"`
	}
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

	// Only allow rating for resolved/closed tickets
	if ticket.Status != models.TicketStatusResolved && ticket.Status != models.TicketStatusClosed {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"message": "Can only rate resolved or closed tickets"},
		})
		return
	}

	ticket.SatisfactionRating = &req.Rating
	ticket.SatisfactionComment = req.Comment

	if err := h.ticketRepo.Update(c.Request.Context(), ticket); err != nil {
		h.logger.Error("Failed to rate ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"message": "Failed to submit rating"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thank you for your feedback!",
	})
}
