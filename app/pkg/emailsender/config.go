package emailsender

type EmailSenderType string

const (
	MockEmailSenderType   EmailSenderType = "mock"
	CustomEmailSenderType EmailSenderType = "custom"
)

type CustomEmailSenderConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	SenderEmail     string `yaml:"sender_email"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password" env:"EMAIL_SENDER_PASSWORD"`
	TemplatePath    string `yaml:"template_path"`
	ServiceEndpoint string `yaml:"service_endpoint"`
}

type EmailSenderConfig struct {
	Type              EmailSenderType          `yaml:"type"`
	CustomEmailSender *CustomEmailSenderConfig `yaml:"custom_email_sender"`
}
