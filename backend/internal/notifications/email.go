package notifications

import (
	"fmt"
	"net"
	"net/smtp"

	"github.com/rs/zerolog/log"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Email    string
	Password string
	From     string
}

type SimpleEmail struct {
	To      string
	Subject string
	Body    string
}

type EmailNotifier struct {
	config *SMTPConfig
}

func NewEmailNotifier(config *SMTPConfig) *EmailNotifier {
	return &EmailNotifier{
		config: config,
	}
}

func (e *EmailNotifier) SendSimpleEmail(email *SimpleEmail) error {
	addr := net.JoinHostPort(e.config.Host, fmt.Sprintf("%d", e.config.Port))

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close connection")
		}
	}()

	client, err := smtp.NewClient(conn, e.config.Host)
	if err != nil {
		return nil
	}
	defer func() {
		_ = client.Quit()
	}()

	if e.config.Email != "" || e.config.Password != "" {
		auth := smtp.PlainAuth("", e.config.Email, e.config.Password, e.config.Host)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(e.config.From); err != nil {
		return err
	}

	if err := client.Rcpt(email.To); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		e.config.From, email.To, email.Subject, email.Body)

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}

	return w.Close()
}

func (e *EmailNotifier) SendLoginNotification(userEmail string) error {
	email := &SimpleEmail{
		To:      userEmail,
		Subject: "Login Notification",
		Body: fmt.Sprintf(`Hello %s,
	
	Tes masuk`, userEmail),
	}

	return e.SendSimpleEmail(email)
}
