package service

import (
	"log"
)

type EmailService interface {
	SendNotificationEmail(to string, subject string, body string) error
}

type emailService struct{}

func NewEmailService() EmailService {
	return &emailService{}
}

func (s *emailService) SendNotificationEmail(to string, subject string, body string) error {
	log.Printf("[EMAIL API] Sending to: %s | Subject: %s | Body: %s", to, subject, body)
	return nil
}
