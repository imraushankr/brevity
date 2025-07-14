package email

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/imraushankr/brevity/server/src/configs"
	"github.com/imraushankr/brevity/server/src/internal/pkg/logger"
)

const (
	emailTimeout = 15 * time.Second
)

type EmailService struct {
	cfg    *configs.EmailConfig
	logger logger.Logger
}

func NewEmailService(cfg *configs.EmailConfig, log logger.Logger) *EmailService {
	return &EmailService{
		cfg:    cfg,
		logger: log,
	}
}

func (e *EmailService) SendVerificationEmail(to, verificationLink string) error {
	const subject = "Verify Your Email Address"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Welcome to Brevity!</h2>
			<p>Please click the link below to verify your email address:</p>
			<p><a href="%s" style="color: #2563eb; text-decoration: underline;">Verify Email</a></p>
			<p>If you didn't request this, please ignore this email.</p>
			<hr>
			<small>This link will expire in 24 hours.</small>
		</body>
		</html>
	`, verificationLink)

	return e.sendEmail(to, subject, body)
}

func (e *EmailService) SendPasswordResetEmail(to, resetLink string) error {
	const subject = "Password Reset Request"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Password Reset</h2>
			<p>You requested a password reset. Click the link below to reset your password:</p>
			<p><a href="%s" style="color: #2563eb; text-decoration: underline;">Reset Password</a></p>
			<p>This link will expire in 15 minutes.</p>
			<p>If you didn't request this, please secure your account.</p>
			<hr>
			<small>Brevity Security Team</small>
		</body>
		</html>
	`, resetLink)

	return e.sendEmail(to, subject, body)
}

func (e *EmailService) sendEmail(to, subject, body string) error {
	from := e.cfg.SMTP.FromEmail
	if from == "" {
		return fmt.Errorf("from email address not configured")
	}

	// Validate recipient email
	if to == "" || !strings.Contains(to, "@") {
		return fmt.Errorf("invalid recipient email address")
	}

	// Construct MIME email
	headers := map[string]string{
		"From":         from,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=UTF-8",
	}

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n" + body)

	auth := smtp.PlainAuth("", e.cfg.SMTP.Username, e.cfg.SMTP.Password, e.cfg.SMTP.Host)
	addr := fmt.Sprintf("%s:%d", e.cfg.SMTP.Host, e.cfg.SMTP.Port)

	// Create a channel to handle SMTP send with timeout
	done := make(chan error, 1)
	go func() {
		err := smtp.SendMail(
			addr,
			auth,
			from,
			strings.Split(to, ";"),
			[]byte(msg.String()),
		)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			e.logger.Error("Failed to send email",
				logger.ErrorField(err),
				logger.String("to", to),
				logger.String("subject", subject))
			return fmt.Errorf("email delivery failed: %w", err)
		}
		e.logger.Info("Email sent successfully",
			logger.String("to", to),
			logger.String("subject", subject))
		return nil
	case <-time.After(emailTimeout):
		e.logger.Error("Email sending timed out",
			logger.String("to", to),
			logger.String("subject", subject))
		return fmt.Errorf("email sending timed out after %s", emailTimeout)
	}
}