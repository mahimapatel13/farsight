package emailtypes

// import (
// 	"budget-planner/internal/domain/integration"

// 	"github.com/sendgrid/sendgrid-go"
// 	"github.com/sendgrid/sendgrid-go/helpers/mail"
// )

// // SendGridConfig holds SendGrid configuration
// type SendGridConfig struct {
// 	APIKey string
// 	From   string
// }

// // SendGridProvider implements EmailService using SendGrid
// type SendGridProvider struct {
// 	*Service
// 	config SendGridConfig
// 	client *sendgrid.Client
// }

// // NewSendGridProvider creates a new SendGrid email provider
// func NewSendGridProvider(config SendGridConfig, templateEngine *TemplateEngine) *SendGridProvider {
// 	return &SendGridProvider{
// 		Service: NewService(templateEngine),
// 		config:  config,
// 		client:  sendgrid.NewSendClient(config.APIKey),
// 	}
// }

// // Send sends an email via SendGrid
// func (p *SendGridProvider) Send(email integration.Email) (string, error) {
// 	// SendGrid implementation
// 	from := mail.NewEmail("", p.config.From)
// 	to := mail.NewEmail("", email.To[0]) // Simplification
// 	message := mail.NewSingleEmail(from, email.Subject, to, email.Body, email.Body)

// 	response, err := p.client.Send(message)
// 	if err != nil {
// 		return "", err
// 	}

// 	return response.Headers["X-Message-Id"], nil
// }

// // SendWithTemplate sends an email with a template via SendGrid
// func (p *SendGridProvider) SendWithTemplate(templateName string, data interface{}, to []string, subject string) (string, error) {
// 	// Render template
// 	body, err := p.templateEngine.Render(templateName, data)
// 	if err != nil {
// 		return "", err
// 	}

// 	// Create email
// 	email := integration.Email{
// 		To:      to,
// 		From:    p.config.From,
// 		Subject: subject,
// 		Body:    body,
// 		IsHTML:  true,
// 	}

// 	// Send email
// 	return p.Send(email)
// }

