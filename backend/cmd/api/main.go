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

	mainDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get database connection")
	}

	defer func() {
		if err := mainDB.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close database connection")
		}
	}()

	authService := services.NewAuthService(db, cfg)
	userService := services.NewUserService(db)
	kostService := services.NewKostService(db)
	roomService := services.NewRoomService(db)
	bookingService := services.NewBookingService(db)
	paymentService := services.NewPaymentService(db, &cfg.Midtrans)

	var uploadProvider interfaces.UploadProvider
	switch cfg.Upload.Provider {
	case "local":
		uploadProvider = providers.NewLocalUploadProvider(cfg.Upload.Path)
	case "s3":
		uploadProvider = providers.NewS3Provider(cfg)
	default:
		log.Fatal().Msg("invalid upload providerr")
	}

	uploadService := services.NewUploadService(uploadProvider)

	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%s", cfg.Server.Port)
	docs.SwaggerInfo.BasePath = "/api/v1"

	gin.SetMode(cfg.Server.GinMode)

	srv := server.New(
		cfg,
		&log,
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
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("failed to gracefully shutdown the server")
		return
	}

	log.Info().Msg("server stopped")
}
