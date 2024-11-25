package mockStruct

import (
	g "gopkg.in/gomail.v2"
)

type MockDialer struct {
	Username string
	SendFunc func(msg ...*g.Message) error
}

// CreateMockDialer is a helper to instantiate a MockDialer with default behavior.
func CreateMockDialer(username string, sendFunc func(msg ...*g.Message) error) *MockDialer {
	return &MockDialer{
		Username: username,
		SendFunc: sendFunc,
	}
}

func (m *MockDialer) DialAndSend(msg ...*g.Message) error {
	if m.SendFunc != nil {
		return m.SendFunc(msg[0])
	}

	return nil
}
