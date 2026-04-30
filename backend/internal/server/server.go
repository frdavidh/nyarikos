package server

import (
	"context"
	"net/http"
	"time"

	"github.com/frdavidh/nyarikos/internal/config"
	"github.com/frdavidh/nyarikos/internal/models"
	"github.com/frdavidh/nyarikos/internal/redis"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	config         *config.Config
	logger         *zerolog.Logger
	db             *gorm.DB
	authService    services.AuthService
	userService    services.UserService
	kostService    services.KostService
	roomService    services.RoomService
	bookingService services.BookingService
	paymentService services.PaymentService
	uploadService  *services.UploadService
	redis          *redis.Client
}

func New(cfg *config.Config,
	logger *zerolog.Logger,
	db *gorm.DB,
	redisClient *redis.Client,
	authService services.AuthService,
	userService services.UserService,
	kostService services.KostService,
	roomService services.RoomService,
	bookingService services.BookingService,
	paymentService services.PaymentService,
	uploadService *services.UploadService,
) *Server {
	return &Server{
		config:         cfg,
		logger:         logger,
		db:             db,
		redis:          redisClient,
		authService:    authService,
		userService:    userService,
		kostService:    kostService,
		roomService:    roomService,
		bookingService: bookingService,
		paymentService: paymentService,
		uploadService:  uploadService,
	}
}

func (s *Server) SetupRoutes() *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(s.corsMiddleware())

	router.GET("/health", s.healthCheck)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Static("/uploads", "./uploads")

	authHandler := NewAuthHandler(s.authService)
	userHandler := NewUserHandler(s.userService)
	kostHandler := NewKostHandler(s.kostService, s.uploadService)
	roomHandler := NewRoomHandler(s.roomService)
	bookingHandler := NewBookingHandler(s.bookingService)
	paymentHandler := NewPaymentHandler(s.paymentService)

	api := router.Group("api/v1")
	api.Use(s.rateLimitMiddleware(100, time.Minute))
	authHandler.Routes(api)
	userHandler.Routes(api, s.authMiddleware())
	kostHandler.Routes(api, s.authMiddleware(), s.roleMiddleware(string(models.RolePemilik)))
	roomHandler.Routes(api, s.authMiddleware(), s.roleMiddleware(string(models.RolePemilik)))
	bookingHandler.Routes(api, s.authMiddleware())
	paymentHandler.Routes(api, s.authMiddleware())

	return router
}

func (s *Server) healthCheck(c *gin.Context) {
	response := gin.H{
		"status": "ok",
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err != nil {
			response["database"] = "unreachable"
		} else if err := sqlDB.PingContext(ctx); err != nil {
			response["database"] = "unreachable"
		} else {
			response["database"] = "connected"
		}
	} else {
		response["database"] = "disabled"
	}

	if s.redis != nil {
		if err := s.redis.RDB().Ping(ctx).Err(); err != nil {
			response["redis"] = "unreachable"
		} else {
			response["redis"] = "connected"
		}
	} else {
		response["redis"] = "disabled"
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
