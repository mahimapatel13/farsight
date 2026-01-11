package emailtypes

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"regexp"
	"strings"
	"time"

	"budget-planner/internal/config"
	"budget-planner/pkg/logger"

	"github.com/google/uuid"
)

// SMTPProvider implements EmailProvider using SMTP
type SMTPProvider struct {
	config config.SMTPConfig
	logger *logger.Logger
}

// NewSMTPProvider creates a new SMTP provider instance
func NewSMTPProvider(config config.SMTPConfig, log *logger.Logger) *SMTPProvider {
	return &SMTPProvider{
		config: config,
		logger: log,
	}
}

// sendWithTLS sends an email using implicit TLS (Port 465)
func (p *SMTPProvider) sendWithTLS(addr string, auth smtp.Auth, email Email, message string) (string, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: p.config.Port != 465, // Only skip verification if not using standard TLS port
		ServerName:         p.config.Host,
		MinVersion:         tls.VersionTLS12, // Ensure minimum TLS 1.2 for security
	}

	// Log TLS configuration details
	p.logger.Info("TLS Configuration", "InsecureSkipVerify", tlsConfig.InsecureSkipVerify, "ServerName", tlsConfig.ServerName)

	// Set a timeout for the connection to avoid hanging
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	// Connect with timeout
	netConn, err := dialer.Dial("tcp", addr)
	if err != nil {
		p.logger.Error("TLS: Failed to establish TCP connection", "error", err)
		return "", fmt.Errorf("TCP connection failed: %w", err)
	}

	// Upgrade to TLS
	conn := tls.Client(netConn, tlsConfig)
	if err := conn.Handshake(); err != nil {
		netConn.Close()
		p.logger.Error("TLS: Handshake failed", "error", err)
		return "", fmt.Errorf("TLS handshake failed: %w", err)
	}

	client, err := smtp.NewClient(conn, p.config.Host)
	if err != nil {
		conn.Close()
		p.logger.Error("TLS: Failed to create SMTP client", "error", err)
		return "", fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate with SMTP server
	if err = client.Auth(auth); err != nil {
		return "", fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Set sender and recipients
	if err = client.Mail(p.config.FromEmail); err != nil {
		return "", fmt.Errorf("failed to set sender: %w", err)
	}
	for _, addr := range email.To {
		if err = client.Rcpt(addr); err != nil {
			return "", fmt.Errorf("failed to set recipient: %w", err)
		}
	}

	// Send email data
	wc, err := client.Data()
	if err != nil {
		return "", fmt.Errorf("failed to send data: %w", err)
	}
	defer wc.Close()

	_, err = wc.Write([]byte(message))
	if err != nil {
		return "", fmt.Errorf("failed to write message: %w", err)
	}

	p.logger.Info("SMTP email sent successfully via TLS", "to", email.To)
	return "smtp-tls-message-id", nil
}

// sendWithStartTLS sends an email using STARTTLS (Port 587)
func (p *SMTPProvider) sendWithStartTLS(addr string, auth smtp.Auth, email Email, message string) (string, error) {
	// Set a timeout for the connection to avoid hanging
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		p.logger.Error("STARTTLS: Failed to establish TCP connection", "error", err)
		return "", fmt.Errorf("failed to establish connection: %w", err)
	}

	client, err := smtp.NewClient(conn, p.config.Host)
	if err != nil {
		conn.Close()
		p.logger.Error("STARTTLS: Failed to create SMTP client", "error", err)
		return "", fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Check if server supports STARTTLS
	if ok, _ := client.Extension("STARTTLS"); !ok {
		p.logger.Warn("STARTTLS: Server does not support STARTTLS")
		return "", fmt.Errorf("server does not support STARTTLS")
	}

	// Start TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: p.config.Port != 587, // Only skip verification if not using standard STARTTLS port
		ServerName:         p.config.Host,
		MinVersion:         tls.VersionTLS12, // Ensure minimum TLS 1.2 for security
	}

	p.logger.Info("STARTTLS: Starting TLS negotiation", "host", p.config.Host)
	if err = client.StartTLS(tlsConfig); err != nil {
		p.logger.Error("STARTTLS: Negotiation failed", "error", err)
		return "", fmt.Errorf("STARTTLS negotiation failed: %w", err)
	}

	p.logger.Info("STARTTLS: TLS negotiation successful")

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return "", fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Set sender and recipients
	if err = client.Mail(p.config.FromEmail); err != nil {
		return "", fmt.Errorf("failed to set sender: %w", err)
	}
	for _, addr := range email.To {
		if err = client.Rcpt(addr); err != nil {
			return "", fmt.Errorf("failed to set recipient: %w", err)
		}
	}

	// Send email data
	wc, err := client.Data()
	if err != nil {
		return "", fmt.Errorf("failed to send data: %w", err)
	}
	defer wc.Close()

	_, err = wc.Write([]byte(message))
	if err != nil {
		return "", fmt.Errorf("failed to write message: %w", err)
	}

	p.logger.Info("SMTP email sent successfully via STARTTLS", "to", email.To)
	return "smtp-starttls-message-id", nil
}

func generateBoundary() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func chunkBase64(input string) string {
	var chunked strings.Builder
	for i := 0; i < len(input); i += 76 {
		end := i + 76
		if end > len(input) {
			end = len(input)
		}
		chunked.WriteString(input[i:end] + "\r\n")
	}
	return chunked.String()
}

// buildEmailMessage constructs the HTML email content with appropriate headers and attachments
func (p *SMTPProvider) buildEmailMessage(email Email) (string, error) {
	var builder strings.Builder

	// Generate a unique Message-ID with proper format
	hostname := p.config.Host
	if hostname == "" {
		hostname = "localhost.localdomain"
	}

	// Create a more standard Message-ID format
	messageID := fmt.Sprintf("<%s.%d.%d@%s>",
		strings.ReplaceAll(uuid.New().String(), "-", ""),
		time.Now().Unix(),
		time.Now().UnixNano()%100000,
		hostname)

	// Get current time for headers in RFC822 format
	currentTime := time.Now().Format(time.RFC1123Z)

	// ✉️ Enhanced headers to improve deliverability
	// Use a proper display name format
	builder.WriteString(fmt.Sprintf("From: \"Budget Planner\" <%s>\r\n", p.config.FromEmail))
	builder.WriteString(fmt.Sprintf("To: %s\r\n", email.JoinRecipients()))
	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.QEncoding.Encode("UTF-8", email.Subject)))

	// Add important headers to reduce spam probability
	builder.WriteString(fmt.Sprintf("Message-ID: %s\r\n", messageID))
	builder.WriteString(fmt.Sprintf("Date: %s\r\n", currentTime))
	builder.WriteString("MIME-Version: 1.0\r\n")

	// Add anti-spam headers
	builder.WriteString("X-Mailer: Budget Planner Email Service\r\n")
	builder.WriteString(fmt.Sprintf("Return-Path: <%s>\r\n", p.config.FromEmail))
	builder.WriteString(fmt.Sprintf("Reply-To: <%s>\r\n", p.config.FromEmail))

	// Add List-Unsubscribe header (important for deliverability)
	builder.WriteString(fmt.Sprintf("List-Unsubscribe: <mailto:%s?subject=unsubscribe>\r\n", p.config.FromEmail))

	// Always use multipart/alternative to provide both HTML and plain text versions
	// This significantly improves deliverability
	boundary := generateBoundary()
	builder.WriteString(fmt.Sprintf(
		"Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n",
		boundary,
	))

	// Add a brief comment to help email clients
	builder.WriteString("This is a multi-part message in MIME format.\r\n")

	// First add plain text version (important for spam prevention)
	builder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	builder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	builder.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")

	// Convert HTML to plain text (simple version)
	plainText := strings.ReplaceAll(email.Body, "<br>", "\r\n")
	plainText = strings.ReplaceAll(plainText, "<br/>", "\r\n")
	plainText = strings.ReplaceAll(plainText, "<br />", "\r\n")
	plainText = strings.ReplaceAll(plainText, "<p>", "\r\n")
	plainText = strings.ReplaceAll(plainText, "</p>", "\r\n")
	plainText = strings.ReplaceAll(plainText, "<div>", "\r\n")
	plainText = strings.ReplaceAll(plainText, "</div>", "\r\n")
	plainText = strings.ReplaceAll(plainText, "<li>", "- ")
	plainText = strings.ReplaceAll(plainText, "</li>", "\r\n")

	// Remove all other HTML tags
	re := regexp.MustCompile("<[^>]*>")
	plainText = re.ReplaceAllString(plainText, "")

	// Clean up multiple newlines
	re = regexp.MustCompile(`\r\n\s*\r\n`)
	plainText = re.ReplaceAllString(plainText, "\r\n\r\n")

	builder.WriteString(plainText + "\r\n\r\n")

	// Then add HTML version
	builder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	builder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	builder.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	builder.WriteString(email.Body + "\r\n")

	// If there are attachments, convert to multipart/mixed
	if len(email.Attachments) > 0 {
		// Close the alternative part
		builder.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

		// Create a new boundary for the mixed part
		mixedBoundary := generateBoundary()
		mixedContent := builder.String()

		// Reset the builder
		builder.Reset()

		// Create the mixed part headers
		builder.WriteString(fmt.Sprintf("From: \"Budget Planner\" <%s>\r\n", p.config.FromEmail))
		builder.WriteString(fmt.Sprintf("To: %s\r\n", email.JoinRecipients()))
		builder.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.QEncoding.Encode("UTF-8", email.Subject)))
		builder.WriteString(fmt.Sprintf("Message-ID: %s\r\n", messageID))
		builder.WriteString(fmt.Sprintf("Date: %s\r\n", currentTime))
		builder.WriteString("MIME-Version: 1.0\r\n")
		builder.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", mixedBoundary))

		// Add the alternative part as the first part of the mixed part
		builder.WriteString(fmt.Sprintf("--%s\r\n", mixedBoundary))
		builder.WriteString("Content-Type: multipart/alternative; boundary=\"" + boundary + "\"\r\n\r\n")
		builder.WriteString(mixedContent)

		// Add attachments
		for _, attachment := range email.Attachments {
			builder.WriteString(fmt.Sprintf("--%s\r\n", mixedBoundary))
			builder.WriteString(fmt.Sprintf(
				"Content-Type: %s\r\nContent-Disposition: attachment; filename=\"%s\"\r\n"+
					"Content-Transfer-Encoding: base64\r\n\r\n",
				attachment.ContentType,
				mime.QEncoding.Encode("UTF-8", attachment.Filename),
			))

			// Encode attachment content as base64
			encodedContent := base64.StdEncoding.EncodeToString(attachment.Content)
			builder.WriteString(chunkBase64(encodedContent) + "\r\n")
		}

		// Close the mixed part
		builder.WriteString(fmt.Sprintf("--%s--\r\n", mixedBoundary))
	} else {
		// Close the alternative part
		builder.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	}

	return builder.String(), nil
}

// Send sends an HTML email using SMTP with TLS/STARTTLS and returns an EmailResponse
func (p *SMTPProvider) Send(ctx context.Context, email *Email) (*EmailResponse, error) {
	// ✅ Validate email before sending
	if err := email.Validate(); err != nil {
		p.logger.Error("SMTP: Invalid email", "error", err)
		return nil, fmt.Errorf("email validation failed: %w", err)
	}

	// Prepare the email content
	message, err := p.buildEmailMessage(*email)
	if err != nil {
		p.logger.Error("SMTP: Failed to build email message", "error", err)
		return nil, fmt.Errorf("failed to build email content: %w", err)
	}

	p.logger.Info("SMTP: Preparing to send email", "to", email.To, "subject", email.Subject)

	// Initialize placeholders
	var messageID string
	var sendErr error

	// Try all connection methods in sequence until one succeeds
	messageID, sendErr = p.tryAllConnectionMethods(ctx, email, message)

	// ❌ Handle email send error
	if sendErr != nil {
		p.logger.Error("SMTP: Failed to send email", "error", sendErr)
		return nil, fmt.Errorf("failed to send email: %w", sendErr)
	}

	// ✅ Success: Return EmailResponse with "sent" status
	p.logger.Info("SMTP: Email sent successfully", "to", email.To, "message_id", messageID)

	return &EmailResponse{
		MessageID: messageID,
		Status:    EmailStatusSent, // Use EmailStatusSent for consistency
		SentAt:    time.Now(),
	}, nil
}

// GetSenderEmail returns the configured sender email
func (p *SMTPProvider) GetSenderEmail() string {
	return p.config.FromEmail
}

// tryAllConnectionMethods attempts to connect using all available methods
func (p *SMTPProvider) tryAllConnectionMethods(ctx context.Context, email *Email, message string) (string, error) {
	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)
	auth := smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.Host)

	// Try all methods in order of security preference
	methods := []struct {
		name string
		fn   func(string, smtp.Auth, Email, string) (string, error)
	}{
		{"TLS", p.sendWithTLS},
		{"STARTTLS", p.sendWithStartTLS},
		{"Plain", func(addr string, auth smtp.Auth, email Email, message string) (string, error) {
			err := smtp.SendMail(addr, auth, p.config.FromEmail, email.To, []byte(message))
			if err != nil {
				return "", err
			}
			return "smtp-plain-message-id", nil
		}},
	}

	var lastErr error
	for _, method := range methods {
		p.logger.Info("SMTP: Attempting to send email using method", "method", method.name)
		if messageID, err := method.fn(addr, auth, *email, message); err == nil {
			p.logger.Info("SMTP: Email sent successfully", "method", method.name)
			return messageID, nil
		} else {
			p.logger.Warn("SMTP: Method failed", "method", method.name, "error", err)
			lastErr = err
		}
	}

	return "", fmt.Errorf("all SMTP connection methods failed, last error: %w", lastErr)
}

// HealthCheck verifies if the SMTP service is reachable.
func (p *SMTPProvider) HealthCheck(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		p.logger.Error("SMTP health check failed", "error", err)
		return fmt.Errorf("SMTP service not reachable: %w", err)
	}
	defer conn.Close()

	// Try to create an SMTP client
	client, err := smtp.NewClient(conn, p.config.Host)
	if err != nil {
		p.logger.Error("SMTP health check failed", "error", err)
		return fmt.Errorf("SMTP service not responding correctly: %w", err)
	}
	defer client.Close()

	p.logger.Info("SMTP health check passed")
	return nil
}

// BatchSend sends multiple emails using the SMTP provider
func (p *SMTPProvider) BatchSend(ctx context.Context, emails []*Email) ([]*EmailResponse, error) {
	var responses []*EmailResponse

	for _, email := range emails {
		messageResponse, err := p.Send(ctx, email)
		if err != nil {
			p.logger.Error("Failed to send batch email", "error", err, "to", email.To, "subject", email.Subject)
			responses = append(responses, &EmailResponse{
				MessageID: "",
				Status:    EmailStatusFailed,
				SentAt:    time.Now(),
			})
			continue
		}
		response := &EmailResponse{
			MessageID: messageResponse.MessageID,
			Status:    EmailStatusSent,
			SentAt:    time.Now(),
		}
		responses = append(responses, response)
		p.logger.Info("Batch email sent successfully", "to", email.To, "subject", email.Subject, "message_id", messageResponse.MessageID)
	}
	return responses, nil
}

// Name returns the name of the provider
func (p *SMTPProvider) Name() string {
	return "smtp"
}
