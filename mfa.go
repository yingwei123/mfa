package mfa

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"time"

	gomail "github.com/yingwei123/mfa/client/email"
)

type MFACode struct {
	Code     string
	ExpireAt time.Time
}

type MFAClient struct {
	CodeDuration time.Duration
	Client       sync.Map // Replaces map[string]MFACode with sync.Map for concurrency
	EmailClient  gomail.EmailClient
	Logger       func(format string, v ...interface{})
}

type MFaClientInterface interface {
	SendMFAEmail(ctx context.Context, toEmail string) error
	VerifyMFA(toEmail string, mfaCode string) error
	GenerateRandomCode() (string, error)
}

func CreateMFAClient(codeDuration time.Duration, clientEmail, clientPassword, smtpServer string, smtpServerPort int, mfaEmail string) (MFaClientInterface, error) {
	dialer, err := gomail.CreateDialer(smtpServer, smtpServerPort, clientEmail, clientPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to create email dialer: %w", err)
	}

	gomailClient, err := gomail.NewEmailClient(dialer, clientEmail, mfaEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to create email client: %w", err)
	}

	return &MFAClient{
		CodeDuration: codeDuration,
		EmailClient:  *gomailClient,
		Client:       sync.Map{},
	}, nil
}

func (m *MFAClient) log(format string, v ...interface{}) {
	if m.Logger != nil {
		m.Logger(format, v...)
	}
}

// SendMFAEmail sends an MFA email to a given email address.
func (m *MFAClient) SendMFAEmail(ctx context.Context, toEmail string) error {
	mfaCode, err := m.GenerateRandomCode()
	if err != nil {
		return fmt.Errorf("failed to generate MFA code: %w", err)
	}

	// Clean up the code after it expires
	time.AfterFunc(m.CodeDuration, func() {
		m.Client.Delete(toEmail)
		m.log("MFA code expired and removed for email: %s", toEmail)
	})

	// Handle context cancellation
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled while sending MFA email")
	default:
	}

	// Send email
	err = m.EmailClient.SendMFAEmail(ctx, toEmail, mfaCode)
	if err != nil {
		return err
	}

	m.Client.Store(toEmail, MFACode{
		Code:     mfaCode,
		ExpireAt: time.Now().Add(m.CodeDuration),
	})

	m.log("MFA email sent to: %s", toEmail)
	return nil
}

// VerifyMFA verifies the MFA code for a given email address.
func (m *MFAClient) VerifyMFA(toEmail string, mfaCode string) error {
	val, ok := m.Client.Load(toEmail)
	if !ok {
		return errors.New("no MFA code found for this email")
	}

	codeData := val.(MFACode)
	if m.isExpired(codeData) {
		m.Client.Delete(toEmail)
		return errors.New("MFA code expired")
	}

	if mfaCode != codeData.Code {
		return errors.New("invalid MFA code")
	}

	m.Client.Delete(toEmail)
	m.log("MFA code verified and removed for email: %s", toEmail)
	return nil
}

// GenerateRandomCode generates a random 6-digit code.
func (m *MFAClient) GenerateRandomCode() (string, error) {
	var code int64
	buffer := make([]byte, 3) // 3 bytes generate 24 bits, sufficient for a 6-digit number

	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("failed to generate random code: %w", err)
	}

	code = int64(buffer[0])<<16 | int64(buffer[1])<<8 | int64(buffer[2])
	code = code % 1000000

	return fmt.Sprintf("%06d", code), nil
}

// isExpired checks if the MFA code has expired.
func (m *MFAClient) isExpired(codeData MFACode) bool {
	return time.Now().After(codeData.ExpireAt)
}
