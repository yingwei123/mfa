package email

import (
	"context"
	"errors"

	"testing"

	"mfa/client/mockStruct"

	"github.com/stretchr/testify/assert"
	gomail "gopkg.in/gomail.v2"
)

const TestToEmail = "test@test.com"
const TestCode = "123456"
const TestClientEmail = "yingwei82599@gmail.com"
const EmailTemplate = "Your MFA code is {{.MfaCode}}."

func mockPassFunction(t *testing.T) func(msg ...*gomail.Message) error {
	return func(msg ...*gomail.Message) error {
		assert.Equal(t, msg[0].GetHeader("To")[0], TestToEmail)
		assert.Equal(t, msg[0].GetHeader("From")[0], TestClientEmail)
		assert.Equal(t, msg[0].GetHeader("Subject")[0], "MFA Code")

		return nil
	}
}

func mockFailFunction() func(msg ...*gomail.Message) error {
	return func(msg ...*gomail.Message) error {
		return errors.New("mock error")
	}
}

func testEmailClientSuite(t *testing.T) {
	t.Run("SendMFAEmailSuccess", testSendMFAEmailSuccess)
	t.Run("SendMFAEmailFail", testSendMFAEmailFail)
	t.Run("SendMFAWithEmptyTemplate", testSendMFAWithEmptyTemplate)
	t.Run("SendMFAWithInvalidTemplate", testSendMFAWithInvalidTemplate)
}

func testSendMFAEmailSuccess(t *testing.T) {
	dialer := mockStruct.CreateMockDialer(TestClientEmail, mockPassFunction(t))

	c, err := NewEmailClient(dialer, TestClientEmail, EmailTemplate)
	assert.Equal(t, err, nil)

	err = c.SendMFAEmail(context.Background(), TestToEmail, TestCode)
	assert.Equal(t, err, nil)
}

func testSendMFAEmailFail(t *testing.T) {
	dialer := mockStruct.CreateMockDialer(TestClientEmail, mockFailFunction())

	c, err := NewEmailClient(dialer, TestClientEmail, EmailTemplate)
	assert.Equal(t, err, nil)

	err = c.SendMFAEmail(context.Background(), TestToEmail, TestCode)
	assert.Equal(t, err.Error(), "failed to send MFA email: mock error")
}

func testSendMFAWithEmptyTemplate(t *testing.T) {
	dialer := mockStruct.CreateMockDialer(TestClientEmail, mockPassFunction(t))

	c, err := NewEmailClient(dialer, TestClientEmail, "")
	assert.Equal(t, err, nil)

	err = c.SendMFAEmail(context.Background(), TestToEmail, TestCode)
	assert.Equal(t, err, nil)
}

func testSendMFAWithInvalidTemplate(t *testing.T) {
	dialer := mockStruct.CreateMockDialer(TestClientEmail, mockPassFunction(t))

	_, err := NewEmailClient(dialer, TestClientEmail, "Invalid Template")
	assert.Equal(t, err.Error(), "MFA template must include '{{.MfaCode}}'")
}
