package email

import (
	"fmt"
	"fowergram/internal/core/domain"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Service interface {
	SendVerificationEmail(to, code string) error
	SendLoginNotification(to string, device *domain.DeviceSession) error
	SendPasswordResetEmail(to, code string) error
}

type emailService struct {
	client      *sendgrid.Client
	senderEmail string
	senderName  string
	templateIDs map[string]string
}

func NewEmailService(apiKey, senderEmail, senderName string) Service {
	return &emailService{
		client:      sendgrid.NewSendClient(apiKey),
		senderEmail: senderEmail,
		senderName:  senderName,
		templateIDs: map[string]string{
			"verification": "d-xxx",
			"login":        "d-yyy",
			"reset":        "d-zzz",
		},
	}
}

func (s *emailService) SendVerificationEmail(to, code string) error {
	from := mail.NewEmail(s.senderName, s.senderEmail)
	subject := "Verify your email"
	toEmail := mail.NewEmail("", to)
	plainTextContent := "Your verification code is: " + code
	htmlContent := "<p>Your verification code is: <strong>" + code + "</strong></p>"

	message := mail.NewSingleEmail(from, subject, toEmail, plainTextContent, htmlContent)
	_, err := s.client.Send(message)
	return err
}

func (s *emailService) SendLoginNotification(to string, device *domain.DeviceSession) error {
	from := mail.NewEmail(s.senderName, s.senderEmail)
	subject := "New Login Detected"
	toEmail := mail.NewEmail("", to)
	plainTextContent := fmt.Sprintf("New login detected from %s using %s at %s", device.Location, device.DeviceType, device.LastActive.Format(time.RFC1123))
	htmlContent := fmt.Sprintf("<p>New login detected from <strong>%s</strong> using <strong>%s</strong> at <strong>%s</strong></p>", device.Location, device.DeviceType, device.LastActive.Format(time.RFC1123))

	message := mail.NewSingleEmail(from, subject, toEmail, plainTextContent, htmlContent)
	_, err := s.client.Send(message)
	return err
}

func (s *emailService) SendPasswordResetEmail(to, code string) error {
	from := mail.NewEmail(s.senderName, s.senderEmail)
	subject := "Reset Your Password"
	toEmail := mail.NewEmail("", to)
	plainTextContent := fmt.Sprintf("Your password reset code is: %s\nThis code will expire in 1 hour.", code)
	htmlContent := fmt.Sprintf("<p>Your password reset code is: <strong>%s</strong></p><p>This code will expire in 1 hour.</p>", code)

	message := mail.NewSingleEmail(from, subject, toEmail, plainTextContent, htmlContent)
	_, err := s.client.Send(message)
	return err
}
