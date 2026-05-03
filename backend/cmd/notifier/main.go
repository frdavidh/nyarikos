package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/frdavidh/nyarikos/internal/config"
	"github.com/frdavidh/nyarikos/internal/notifications"
	"github.com/hibiken/asynq"
)

func main() {
	log.Println("Starting notification service...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	emailConfig := &notifications.SMTPConfig{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
	}
	emailNotifier := notifications.NewEmailNotifier(emailConfig)

	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 10,
	})

	mux := asynq.NewServeMux()
	mux.HandleFunc(notifications.TypeEmailLoginNotification, func(ctx context.Context, task *asynq.Task) error {
		var payload notifications.LoginNotificationPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("unmarshal payload: %w", err)
		}

		log.Printf("Sending login notification to %s", payload.Email)
		return emailNotifier.SendLoginNotification(payload.Email)
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down notification service...")
		srv.Shutdown()
	}()

	log.Println("Notification service started. Waiting for tasks...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
