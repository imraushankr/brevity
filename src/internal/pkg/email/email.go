package email

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/imraushankr/brevity/server/src/configs"
	"github.com/imraushankr/brevity/server/src/internal/pkg/logger"
)

type EmailService struct {
	cfg *configs.EmailConfig
}

func NewEmailService(cfg *configs.EmailConfig) *EmailService {
	return &EmailService{cfg: cfg}
}

func (e *EmailService) SendVerificationEmail(to, verificationLink string) error {
	subject := "Verify Your Email Address"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h1>Welcome to Brevity!</h1>
			<p>Please click the link below to verify your email address:</p>
			<a href="%s">Verify Email</a>
			<p>If you didn't request this, please ignore this email.</p>
		</body>
		</html>
	`, verificationLink)

	return e.sendEmail(to, subject, body)
}

func (e *EmailService) SendPasswordResetEmail(to, resetLink string) error {
	subject := "Password Reset Request"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h1>Password Reset</h1>
			<p>You requested a password reset. Click the link below to reset your password:</p>
			<a href="%s">Reset Password</a>
			<p>This link will expire in 15 minutes.</p>
			<p>If you didn't request this, please ignore this email.</p>
		</body>
		</html>
	`, resetLink)

	return e.sendEmail(to, subject, body)
}

func (e *EmailService) sendEmail(to, subject, body string) error {
	from := e.cfg.SMTP.FromEmail
	msg := []byte(fmt.Sprintf("From: %s\r\n", from) +
		fmt.Sprintf("To: %s\r\n", to) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n" +
		"\r\n" +
		body)

	auth := smtp.PlainAuth("", e.cfg.SMTP.Username, e.cfg.SMTP.Password, e.cfg.SMTP.Host)
	addr := fmt.Sprintf("%s:%d", e.cfg.SMTP.Host, e.cfg.SMTP.Port)

	// Split to address in case there are multiple recipients
	toAddresses := strings.Split(to, ";")

	if err := smtp.SendMail(addr, auth, from, toAddresses, msg); err != nil {
		logger.Error("Failed to send email",
			logger.ErrorField(err),
			logger.String("to", to),
			logger.String("subject", subject))
		return fmt.Errorf("failed to send email: %w", err)
	}

	logger.Info("Email sent successfully",
		logger.String("to", to),
		logger.String("subject", subject))

	return nil
}