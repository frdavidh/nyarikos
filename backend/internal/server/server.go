package server

import (
	"net/http"

	"github.com/frdavidh/nyarikos/internal/config"
	"github.com/frdavidh/nyarikos/internal/models"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Server struct {
	config        *config.Config
	db            *gorm.DB
	logger        *zerolog.Logger
	authService   services.AuthService
	userService   services.UserService
	kostService   services.KostService
	roomService   services.RoomService
	uploadService *services.UploadService
}

func New(cfg *config.Config,
	logger *zerolog.Logger,
	authService services.AuthService,
	userService services.UserService,
	kostService services.KostService,
	roomService services.RoomService,
	uploadService *services.UploadService,
) *Server {
	return &Server{
		config:        cfg,
		logger:        logger,
		authService:   authService,
		userService:   userService,
		kostService:   kostService,
		roomService:   roomService,
		uploadService: uploadService,
	}
}

func (s *Server) SetupRoutes() *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(s.corsMiddleware())

	router.GET("/health", s.healthCheck)
	router.Static("/uploads", "./uploads")

	authHandler := NewAuthHandler(s.authService)
	userHandler := NewUserHandler(s.userService)
	kostHandler := NewKostHandler(s.kostService, s.uploadService)
	roomHandler := NewRoomHandler(s.roomService)

	api := router.Group("api/v1")
	authHandler.Routes(api)
	userHandler.Routes(api, s.authMiddleware())
	kostHandler.Routes(api, s.authMiddleware(), s.roleMiddleware(string(models.RolePemilik)))
	roomHandler.Routes(api, s.authMiddleware(), s.roleMiddleware(string(models.RolePemilik)))

	return router
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
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
