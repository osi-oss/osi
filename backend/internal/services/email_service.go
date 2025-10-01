package services

import (
	"fmt"
	"net/smtp"
	"strings"
)

type EmailService struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

func NewEmailService(host, port, user, password, fromEmail, fromName string) *EmailService {
	return &EmailService{
		SMTPHost:     host,
		SMTPPort:     port,
		SMTPUser:     user,
		SMTPPassword: password,
		FromEmail:    fromEmail,
		FromName:     fromName,
	}
}

// SendPasswordResetEmail отправляет email с токеном для восстановления пароля
func (e *EmailService) SendPasswordResetEmail(toEmail, resetToken, baseURL string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", baseURL, resetToken)

	subject := "Восстановление пароля"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Восстановление пароля</h2>
			<p>Вы запросили восстановление пароля для вашего аккаунта.</p>
			<p>Перейдите по ссылке ниже для сброса пароля:</p>
			<p><a href="%s">Сбросить пароль</a></p>
			<p>Ссылка действительна в течение 1 часа.</p>
			<p>Если вы не запрашивали восстановление пароля, проигнорируйте это письмо.</p>
		</body>
		</html>
	`, resetURL)

	return e.sendEmail(toEmail, subject, body)
}

// SendWelcomeEmail отправляет приветственное письмо
func (e *EmailService) SendWelcomeEmail(toEmail, userName string) error {
	subject := "Добро пожаловать в OSI!"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Добро пожаловать!</h2>
			<p>Привет %s,</p>
			<p>Спасибо за регистрацию в нашей системе OSI.</p>
			<p>Теперь вы можете пользоваться всеми возможностями платформы.</p>
		</body>
		</html>
	`, userName)

	return e.sendEmail(toEmail, subject, body)
}

// sendEmail внутренний метод для отправки email
func (e *EmailService) sendEmail(to, subject, body string) error {
	// Настройка аутентификации
	auth := smtp.PlainAuth("", e.SMTPUser, e.SMTPPassword, e.SMTPHost)

	// Формирование сообщения
	msg := fmt.Sprintf("To: %s\r\n"+
		"From: %s <%s>\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", to, e.FromName, e.FromEmail, subject, body)

	// Отправка email
	addr := fmt.Sprintf("%s:%s", e.SMTPHost, e.SMTPPort)
	return smtp.SendMail(addr, auth, e.FromEmail, []string{to}, []byte(msg))
}

// ValidateEmail проверяет валидность email адреса
func (e *EmailService) ValidateEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
