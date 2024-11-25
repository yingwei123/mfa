package client

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/yingwei123/mfa/client/email"
	"github.com/yingwei123/mfa/client/mockStruct"

	"github.com/stretchr/testify/assert"
	"gopkg.in/gomail.v2"
)

const TestCodeDuration = 100 * time.Millisecond
const OpperationPadding = 5 * time.Millisecond
const TestClientEmail = "test@test.com"

func mockPassFunction() func(msg ...*gomail.Message) error {
	return func(msg ...*gomail.Message) error {
		return nil
	}
}

func mockFailFunction() func(msg ...*gomail.Message) error {
	return func(msg ...*gomail.Message) error {
		return errors.New("mock error")
	}
}

func createTestMFAClient(t *testing.T, sendEmailFunc func(msg ...*gomail.Message) error) MFAClient {
	dialer := mockStruct.CreateMockDialer(TestClientEmail, sendEmailFunc)
	mfaEmailClient, err := email.NewEmailClient(dialer, TestClientEmail, "Your MFA code is {{.MfaCode}}.")

	assert.Equal(t, err, nil)
	assert.Equal(t, mfaEmailClient.ClientEmail, TestClientEmail)
	assert.Equal(t, mfaEmailClient.MFATemplate, "Your MFA code is {{.MfaCode}}.")
	assert.Equal(t, mfaEmailClient.Client, dialer)

	mfaClient := MFAClient{
		CodeDuration: TestCodeDuration,
		Client:       sync.Map{},
		EmailClient:  *mfaEmailClient,
		Logger:       func(format string, v ...interface{}) {},
	}

	assert.Equal(t, mfaClient.EmailClient, *mfaEmailClient)
	assert.Equal(t, mfaClient.CodeDuration, TestCodeDuration)
	assert.Equal(t, mfaClient.Client, sync.Map{})

	return mfaClient
}

func testMFAClientSuite(t *testing.T) {
	t.Run("GenerateRandomCode", testGenerateRandomCode)
	t.Run("SendMFAEmailSuccess", testSendMFAEmailSuccess)
	t.Run("SendMFAEmailFail", testSendMFAEmailFail)
	t.Run("VerifyMFAEmailSuccess", testVerifyMFAEmailSuccess)
	t.Run("VerifyMFAEmailFail", testVerifyMFAEmailFail)
	t.Run("VerifyMFAEmailExpired", testVerifyMFAEmailExpired)
	t.Run("VerifyMFAEmailNoCode", testVerifyMFAEmailNoCode)
}

func testGenerateRandomCode(t *testing.T) {
	mfaClient := createTestMFAClient(t, mockPassFunction())
	code, err := mfaClient.GenerateRandomCode()

	assert.Equal(t, err, nil)
	assert.Equal(t, len(code), 6)
}

func testSendMFAEmailSuccess(t *testing.T) {
	mfaClient := createTestMFAClient(t, mockPassFunction())
	err := mfaClient.SendMFAEmail(context.Background(), "test@test.com")

	client, ok := mfaClient.Client.Load("test@test.com")
	codeData := client.(MFACode)

	assert.Equal(t, true, ok)
	assert.Equal(t, len(codeData.Code), 6)
	assert.Equal(t, nil, err)

	time.Sleep(TestCodeDuration + OpperationPadding)

	clientAfterExpire, ok := mfaClient.Client.Load("test@test.com")

	assert.Equal(t, false, ok)
	assert.Equal(t, nil, clientAfterExpire)
}

func testSendMFAEmailFail(t *testing.T) {
	mfaClient := createTestMFAClient(t, mockFailFunction())

	err := mfaClient.SendMFAEmail(context.Background(), "test@test.com")

	assert.Equal(t, "failed to send MFA email: mock error", err.Error())

	_, ok := mfaClient.Client.Load("test@test.com")
	assert.Equal(t, false, ok)
}

func testSendMFAEmailContextCancel(t *testing.T) {
	mfaClient := createTestMFAClient(t, mockPassFunction())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := mfaClient.SendMFAEmail(ctx, "test@test.com")

	assert.Equal(t, "context canceled while sending MFA email", err.Error())

	_, ok := mfaClient.Client.Load("test@test.com")
	assert.Equal(t, false, ok)
}

func testVerifyMFAEmailSuccess(t *testing.T) {
	mfaClient := createTestMFAClient(t, mockPassFunction())
	mfaClient.Client.Store(TestClientEmail, MFACode{
		Code:     "123456",
		ExpireAt: time.Now().Add(TestCodeDuration),
	})

	err := mfaClient.VerifyMFA(TestClientEmail, "123456")

	assert.Equal(t, nil, err)

	_, ok := mfaClient.Client.Load(TestClientEmail) //test if the code is deleted after successful verification
	assert.Equal(t, false, ok)
}

func testVerifyMFAEmailFail(t *testing.T) {
	mfaClient := createTestMFAClient(t, mockPassFunction())
	mfaClient.Client.Store(TestClientEmail, MFACode{
		Code:     "234566",
		ExpireAt: time.Now().Add(TestCodeDuration),
	})

	err := mfaClient.VerifyMFA(TestClientEmail, "123456")

	assert.Equal(t, "invalid MFA code", err.Error())

	_, ok := mfaClient.Client.Load(TestClientEmail) //test code still exist after failed verification
	assert.Equal(t, true, ok)
}

func testVerifyMFAEmailExpired(t *testing.T) {
	mfaClient := createTestMFAClient(t, mockPassFunction())
	mfaClient.Client.Store(TestClientEmail, MFACode{
		Code:     "123456",
		ExpireAt: time.Now().Add(-TestCodeDuration),
	})

	err := mfaClient.VerifyMFA(TestClientEmail, "123456")

	assert.Equal(t, "MFA code expired", err.Error())

	_, ok := mfaClient.Client.Load(TestClientEmail) //test code is deleted after expired
	assert.Equal(t, false, ok)
}

func testVerifyMFAEmailNoCode(t *testing.T) {
	mfaClient := createTestMFAClient(t, mockPassFunction())

	err := mfaClient.VerifyMFA(TestClientEmail, "123456")

	assert.Equal(t, "no MFA code found for this email", err.Error())
}
