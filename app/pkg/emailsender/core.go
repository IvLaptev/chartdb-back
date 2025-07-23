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

type EmailSenderConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	TemplatePath    string `yaml:"template_path"`
	ServiceEndpoint string `yaml:"service_endpoint"`
}

type EmailSender interface {
	SendCreateUserEmail(to string, token string) error
}

type GomailSender struct {
	dialer          *gomail.Dialer
	username        string
	serviceEndpoint string
	templates       *template.Template
}

func (s *GomailSender) sendMessage(to, subject string, patchMessage func(*gomail.Message) error) error {
	msg := gomail.NewMessage(gomail.SetCharset("UTF-8"))
	msg.SetHeader("From", fmt.Sprintf(`"ChartDB"<%s>`, s.username))
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

func (s *GomailSender) SendCreateUserEmail(to string, token string) error {
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

func NewGomailSender(config EmailSenderConfig) (*GomailSender, error) {
	dialer := gomail.NewDialer(config.Host, config.Port, config.Username, config.Password)

	templates, err := template.ParseFiles(
		config.TemplatePath + "/" + createUserTemplate,
	)
	if err != nil {
		return nil, fmt.Errorf("parse files: %w", err)
	}

	return &GomailSender{
		dialer:          dialer,
		username:        config.Username,
		serviceEndpoint: config.ServiceEndpoint,
		templates:       templates,
	}, nil
}
