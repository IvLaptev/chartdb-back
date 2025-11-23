package emailsender

type EmailSender interface {
	SendCreateUserEmail(to string, token string) error
}
