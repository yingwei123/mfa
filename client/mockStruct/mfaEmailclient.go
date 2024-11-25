package mockStruct

import (
	"context"
)

type EmailClientMock struct {
	client           MockDialer
	MFATemplate      string
	ClientEmail      string
	SendMFAEmailFunc func(ctx context.Context, toEmail string, mfaCode string) error
}

func CreateEmailClientMock(client MockDialer, MFATemplate string, ClientEmail string, sendMFAEmailFunc func(ctx context.Context, toEmail string, mfaCode string) error) *EmailClientMock {
	return &EmailClientMock{
		client:      client,
		MFATemplate: MFATemplate,
		ClientEmail: ClientEmail,
	}
}

// SendMFAEmail mocks sending MFAEmail
func (g *EmailClientMock) SendMFAEmail(ctx context.Context, toEmail string, mfaCode string) error {
	if g.SendMFAEmailFunc != nil {
		return g.SendMFAEmailFunc(ctx, toEmail, mfaCode)
	}

	return nil
}
