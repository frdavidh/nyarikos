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
	config      *config.Config
	db          *gorm.DB
	logger      *zerolog.Logger
	authHandler *AuthHandler
	userHandler *UserHandler
	kostHandler *KostHandler
}

func New(cfg *config.Config, db *gorm.DB, logger *zerolog.Logger) *Server {
	authService := services.NewAuthService(db, cfg)
	userService := services.NewUserService(db)
	kostService := services.NewKostService(db)

	return &Server{
		config: cfg,
		db:     db,
		logger: logger,

		authHandler: NewAuthHandler(authService),
		userHandler: NewUserHandler(userService),
		kostHandler: NewKostHandler(kostService),
	}
}

func (s *Server) SetupRoutes() *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(s.corsMiddleware())

	router.GET("/health", s.healthCheck)

	api := router.Group("api/v1")
	s.authHandler.Routes(api)
	s.userHandler.Routes(api, s.authMiddleware())
	s.kostHandler.Routes(api, s.authMiddleware(), s.roleMiddleware(string(models.RolePemilik)))

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
