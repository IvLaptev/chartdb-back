package emailsender

import (
	"fmt"
	"html/template"
	"io"

	"gopkg.in/gomail.v2"
)

const (
	createUserTemplate = "create_user_template.html"
)

type CustomSender struct {
	dialer          *gomail.Dialer
	sender          string
	serviceEndpoint string
	templates       *template.Template
}

func (s *CustomSender) sendMessage(to, subject string, patchMessage func(*gomail.Message) error) error {
	msg := gomail.NewMessage(gomail.SetCharset("UTF-8"))
	msg.SetHeader("From", fmt.Sprintf(`"ChartDB"<%s>`, s.sender))
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)

	err := patchMessage(msg)
	if err != nil {
		return fmt.Errorf("patch message: %w", err)
	}

	err = s.dialer.DialAndSend(msg)
	if err != nil {
		return fmt.Errorf("dial and send: %w", err)
	}

	return nil
}

func (s *CustomSender) SendCreateUserEmail(to string, token string) error {
	err := s.sendMessage(to, "Подтверждение регистрации", func(msg *gomail.Message) error {
		msg.AddAlternativeWriter("text/html", func(w io.Writer) error {
			return s.templates.ExecuteTemplate(w, createUserTemplate, struct {
				ServiceEndpoint string
				Token           string
			}{
				ServiceEndpoint: s.serviceEndpoint,
				Token:           token,
			})
		})
		return nil
	})
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func NewCustomSender(config *CustomEmailSenderConfig) (*CustomSender, error) {
	dialer := gomail.NewDialer(config.Host, config.Port, config.Username, config.Password)

	templates, err := template.ParseFiles(
		config.TemplatePath + "/" + createUserTemplate,
	)
	if err != nil {
		return nil, fmt.Errorf("parse files: %w", err)
	}

	return &CustomSender{
		dialer:          dialer,
		sender:          config.SenderEmail,
		serviceEndpoint: config.ServiceEndpoint,
		templates:       templates,
	}, nil
}
