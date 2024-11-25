package email

import (
	"bytes"
	"context"
	"fmt"
	"net/mail"
	"strings"
	"text/template"

	d "github.com/yingwei123/mfa/client/email/default"

	gomail "gopkg.in/gomail.v2"
)

type Dialer interface {
	DialAndSend(msg ...*gomail.Message) error
}

type EmailClient struct {
	Client      Dialer
	MFATemplate string
	ClientEmail string
}

type EmailClientInterface interface {
	SendMFAEmail(ctx context.Context, toEmail string, mfaCode string) error
}

// isValidEmail validates an email address format.
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// NewEmailClient creates a new EmailClient instance.
func NewEmailClient(dialer Dialer, clientEmail string, mfaTemplate string) (*EmailClient, error) {
	if mfaTemplate == "" {
		mfaTemplate = d.MFATemplate
	}

	if !strings.Contains(mfaTemplate, "{{.MfaCode}}") {
		return nil, fmt.Errorf("MFA template must include '{{.MfaCode}}'")
	}

	return &EmailClient{
		Client:      dialer,
		MFATemplate: mfaTemplate,
		ClientEmail: clientEmail,
	}, nil
}

// SendMFAEmail sends an MFA email with the given data to the specified email address.
func (g *EmailClient) SendMFAEmail(ctx context.Context, toEmail string, mfaCode string) error {
	if !isValidEmail(toEmail) {
		return fmt.Errorf("invalid recipient email address: %s", toEmail)
	}

	body, err := executeTemplate(g.MFATemplate, mfaCode)
	if err != nil {
		return fmt.Errorf("failed to execute MFA template: %w", err)
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("email sending canceled or timed out")
	default:
	}

	return g.sendEmail(toEmail, "MFA Code", body)
}

// sendEmail handles the low-level logic of constructing and sending an email.
func (g *EmailClient) sendEmail(toEmail, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", g.ClientEmail)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	if err := g.Client.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send MFA email: %w", err)
	}

	return nil
}

// executeTemplate parses and executes a template string with the given data.
func executeTemplate(tmpl string, mfaCode string) (string, error) {
	var buffer bytes.Buffer
	parsedTemplate, err := template.New("mfaEmailTemplate").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}

	data := struct {
		MfaCode string
	}{
		MfaCode: mfaCode,
	}

	if err := parsedTemplate.Execute(&buffer, data); err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	return buffer.String(), nil
}

// CreateDialer is a utility function to create a gomail.Dialer with validation.
func CreateDialer(smtpServer string, smtpPort int, email, password string) (*gomail.Dialer, error) {
	if smtpServer == "" {
		return nil, fmt.Errorf("SMTP server address cannot be empty")
	}

	if smtpPort <= 0 || smtpPort > 65535 {
		return nil, fmt.Errorf("invalid SMTP port: %d", smtpPort)
	}

	if !isValidEmail(email) {
		return nil, fmt.Errorf("invalid email address: %s", email)
	}

	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	dialer := gomail.NewDialer(smtpServer, smtpPort, email, password)
	return dialer, nil
}
