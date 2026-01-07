package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/nats-io/nats.go"
	liblogger "github.com/niaga-platform/lib-common/logger"
	libmiddleware "github.com/niaga-platform/lib-common/middleware"
	"github.com/niaga-platform/service-support/internal/config"
	"github.com/niaga-platform/service-support/internal/events"
	"github.com/niaga-platform/service-support/internal/handlers"
	"github.com/niaga-platform/service-support/internal/repository"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db             *gorm.DB
	cfg            *config.Config
	zapLogger      *zap.Logger
	natsClient     *nats.Conn
	eventPublisher *events.Publisher
)

func main() {
	// Initialize logger
	var err error
	zapLogger, err = liblogger.NewLogger(os.Getenv("APP_ENV"))
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()
	zapLogger.Info("Starting Support Service...")

	// Load configuration
	cfg = config.Load()
	zapLogger.Info("Configuration loaded",
		zap.Int("port", cfg.ServicePort),
		zap.String("environment", cfg.Environment))

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db, err = gorm.Open(postgres.Open(cfg.Database.GetDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		zapLogger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		zapLogger.Fatal("Failed to get underlying sql.DB", zap.Error(err))
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)
	zapLogger.Info("Database connected")

	// Initialize NATS
	natsClient, err = nats.Connect(cfg.NatsURL)
	if err != nil {
		zapLogger.Warn("NATS connection failed (events will be disabled)", zap.Error(err))
	} else {
		zapLogger.Info("NATS connected")
		eventPublisher = events.NewPublisher(natsClient)
	}

	// Initialize repositories
	ticketRepo := repository.NewTicketRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	cannedResponseRepo := repository.NewCannedResponseRepository(db)

	// Initialize handlers
	ticketHandler := handlers.NewTicketHandler(ticketRepo, messageRepo, zapLogger)
	adminHandler := handlers.NewAdminHandler(ticketRepo, messageRepo, categoryRepo, cannedResponseRepo, zapLogger)

	// Wire event publisher
	if eventPublisher != nil {
		ticketHandler.SetEventPublisher(eventPublisher)
		adminHandler.SetEventPublisher(eventPublisher)
		zapLogger.Info("Event publisher wired to handlers")
	}

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// CORS
	allowedOrigins := getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001,http://localhost:3002")
	router.Use(libmiddleware.CORSWithOrigins(allowedOrigins))

	// Request ID
	router.Use(libmiddleware.RequestID())

	// Security headers
	router.Use(libmiddleware.SecurityHeaders())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "support",
			"time":    time.Now().UTC(),
		})
	})

	router.GET("/health/ready", func(c *gin.Context) {
		sqlDB, err := db.DB()
		dbHealthy := err == nil && sqlDB.Ping() == nil

		status := "healthy"
		if !dbHealthy {
			status = "unhealthy"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  status,
			"service": "support",
			"dependencies": gin.H{
				"postgres": gin.H{"status": func() string {
					if dbHealthy {
						return "healthy"
					}
					return "unhealthy"
				}()},
				"nats": gin.H{"status": func() string {
					if natsClient != nil && natsClient.IsConnected() {
						return "healthy"
					}
					return "disconnected"
				}()},
			},
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public support routes
		support := v1.Group("/support")
		{
			// Contact form (no auth required)
			support.POST("/contact", ticketHandler.SubmitContactForm)

			// Categories (public - for contact form dropdown)
			support.GET("/categories", func(c *gin.Context) {
				categories, err := categoryRepo.List(c.Request.Context(), true)
				if err != nil {
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
			})

			// Authenticated customer routes
			authed := support.Group("")
			authed.Use(AuthMiddleware(cfg.JWTSecret))
			{
				authed.POST("/tickets", ticketHandler.Create)
				authed.GET("/tickets", ticketHandler.List)
				authed.GET("/tickets/:id", ticketHandler.GetByID)
				authed.POST("/tickets/:id/messages", ticketHandler.AddMessage)
				authed.POST("/tickets/:id/rate", ticketHandler.RateTicket)
			}
		}

		// Admin support routes
		admin := v1.Group("/admin/support")
		admin.Use(AuthMiddleware(cfg.JWTSecret))
		admin.Use(AdminRoleMiddleware())
		{
			// Dashboard stats
			admin.GET("/stats", adminHandler.GetStats)

			// Ticket management
			admin.GET("/tickets", adminHandler.ListTickets)
			admin.GET("/tickets/:id", adminHandler.GetTicket)
			admin.PUT("/tickets/:id", adminHandler.UpdateTicket)
			admin.POST("/tickets/:id/reply", adminHandler.ReplyToTicket)
			admin.PUT("/tickets/:id/assign", adminHandler.AssignTicket)

			// Category management
			admin.GET("/categories", adminHandler.ListCategories)
			admin.POST("/categories", adminHandler.CreateCategory)
			admin.PUT("/categories/:id", adminHandler.UpdateCategory)
			admin.DELETE("/categories/:id", adminHandler.DeleteCategory)

			// Canned responses
			admin.GET("/canned-responses", adminHandler.ListCannedResponses)
			admin.POST("/canned-responses", adminHandler.CreateCannedResponse)
			admin.PUT("/canned-responses/:id", adminHandler.UpdateCannedResponse)
			admin.DELETE("/canned-responses/:id", adminHandler.DeleteCannedResponse)
		}
	}

	// Start server
	port := fmt.Sprintf(":%d", cfg.ServicePort)
	srv := &http.Server{
		Addr:         port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		zapLogger.Info("Support service starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Shutting down server...")

	if natsClient != nil {
		natsClient.Close()
		zapLogger.Info("NATS connection closed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zapLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("Support service stopped")
}

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userID, exists := claims["user_id"]; exists {
				c.Set("user_id", userID)
			}
			if email, exists := claims["email"]; exists {
				c.Set("email", email)
			}
			if role, exists := claims["role"]; exists {
				c.Set("role", role)
			}
		}

		c.Next()
	}
}

// AdminRoleMiddleware ensures user has admin/support role
func AdminRoleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		allowedRoles := []string{"admin", "super_admin", "support", "manager"}
		isAllowed := false
		for _, r := range allowedRoles {
			if roleStr == r {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
