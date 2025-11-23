package emailsender

type MockSender struct{}

func (*MockSender) SendCreateUserEmail(to string, token string) error {
	return nil
}

func NewMockSender() *MockSender {
	return &MockSender{}
}
