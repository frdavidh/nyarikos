package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/frdavidh/nyarikos/docs"
	"github.com/frdavidh/nyarikos/internal/config"
	"github.com/frdavidh/nyarikos/internal/database"
	"github.com/frdavidh/nyarikos/internal/interfaces"
	"github.com/frdavidh/nyarikos/internal/logger"
	"github.com/frdavidh/nyarikos/internal/providers"
	"github.com/frdavidh/nyarikos/internal/redis"
	"github.com/frdavidh/nyarikos/internal/server"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/gin-gonic/gin"
)

//	@title						Nyarikos API
//	@version					1.0
//	@description				API for Nyarikos - Kost Management System
//	@host						localhost:8080
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization

func main() {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	redisClient, err := redis.New(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}

	authService := services.NewAuthService(db, cfg, redisClient)
	userService := services.NewUserService(db)
	kostService := services.NewKostService(db)
	roomService := services.NewRoomService(db)
	bookingService := services.NewBookingService(db, redisClient)
	paymentService := services.NewPaymentService(db, &cfg.Midtrans)

	var uploadProvider interfaces.UploadProvider
	switch cfg.Upload.Provider {
	case "local":
		uploadProvider = providers.NewLocalUploadProvider(cfg.Upload.Path)
	case "s3":
		uploadProvider, err = providers.NewS3Provider(context.Background(), cfg)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to initialize S3 provider")
		}
	default:
		log.Fatal().Msg("invalid upload provider")
	}

	uploadService := services.NewUploadService(uploadProvider)

	swaggerHost := cfg.Server.SwaggerHost
	if swaggerHost == "" {
		swaggerHost = fmt.Sprintf("localhost:%s", cfg.Server.Port)
	}
	docs.SwaggerInfo.Host = swaggerHost
	docs.SwaggerInfo.BasePath = "/api/v1"

	gin.SetMode(cfg.Server.GinMode)

	srv := server.New(
		cfg,
		&log,
		db,
		redisClient,
		authService,
		userService,
		kostService,
		roomService,
		bookingService,
		paymentService,
		uploadService,
	)

	router := srv.SetupRoutes()
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.Server.Port).Msg("server is running")
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server shutdown")
	}
	cancel()

	if err := redisClient.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close redis conn")
	}

	mainDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get database connection")
	} else {
		if err := mainDB.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close db conn")
		}
	}

	log.Info().Msg("server stopped")
}
