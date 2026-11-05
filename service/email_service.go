package service

import (
	"log"
	"net/smtp"
	"os"
)

type EmailService interface {
	SendNotificationEmail(to string, subject string, body string) error
}

type emailService struct{}

func NewEmailService() EmailService {
	return &emailService{}
}

func (s *emailService) SendNotificationEmail(to string, subject string, body string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	auth := smtp.PlainAuth("", user, pass, host)

	sender := os.Getenv("SMTP_SENDER")
	if sender == "" {
		sender = "admin@wargapay.id"
	}

	// Format email standar MIME lengkap dengan header From, To, Subject
	msg := []byte("From: WargaPay RT 05 <" + sender + ">\r\nTo: " + to + "\r\nSubject: " + subject + "\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + body)

	addr := host + ":" + port
	err := smtp.SendMail(addr, auth, user, []string{to}, msg)
	if err != nil {
		log.Printf("[EMAIL ERROR] Gagal mengirim email ke %s: %v", to, err)
		return err
	}

	log.Printf("[EMAIL SMTP] Berhasil mengirim email asli ke: %s", to)
	return nil
}
