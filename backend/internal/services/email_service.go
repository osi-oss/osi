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

// SendAuthCode отправляет код подтверждения для входа/регистрации
func (e *EmailService) SendAuthCode(toEmail, code string) error {
	subject := "Код подтверждения OSI"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #333;">Ваш код подтверждения</h2>
			<p style="font-size: 16px; color: #666;">Используйте этот код для входа в систему OSI:</p>
			<div style="background: #f5f5f5; padding: 20px; text-align: center; margin: 20px 0; border-radius: 8px;">
				<span style="font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #333;">%s</span>
			</div>
			<p style="font-size: 14px; color: #999;">Код действителен 20 минут.</p>
			<p style="font-size: 14px; color: #999;">Если вы не запрашивали этот код, проигнорируйте это письмо.</p>
		</body>
		</html>
	`, code)

	return e.sendEmail(toEmail, subject, body)
}

// SendWelcomeEmail отправляет приветственное письмо
func (e *EmailService) SendWelcomeEmail(toEmail, userName string) error {
	subject := "Добро пожаловать в OSI!"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #333;">Добро пожаловать!</h2>
			<p style="font-size: 16px; color: #666;">Привет %s,</p>
			<p style="font-size: 16px; color: #666;">Спасибо за регистрацию в системе OSI.</p>
			<p style="font-size: 16px; color: #666;">Теперь вы можете пользоваться всеми возможностями платформы.</p>
		</body>
		</html>
	`, userName)

	return e.sendEmail(toEmail, subject, body)
}

// SendInviteEmail отправляет приглашение в организацию
func (e *EmailService) SendInviteEmail(toEmail, orgName, positionName, inviteLink string) error {
	subject := fmt.Sprintf("Приглашение в %s", orgName)

	linkHTML := ""
	if inviteLink != "" {
		linkHTML = fmt.Sprintf(`<p style="text-align: center; margin-top: 30px;">
			<a href="%s" style="background-color: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; font-weight: bold;">
				Перейти к приглашению
			</a>
		</p>`, inviteLink)
	}

	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; color: #333;">
			<h2 style="color: #333; border-bottom: 2px solid #007bff; padding-bottom: 10px;">Вас пригласили в организацию</h2>
			
			<p style="font-size: 16px; margin: 20px 0;">Здравствуйте!</p>
			
			<p style="font-size: 16px; margin: 20px 0;">
				Вас пригласили на должность <strong style="color: #007bff;">%s</strong> 
				в организацию <strong style="color: #007bff;">%s</strong>
			</p>
			
			<h3 style="color: #333; margin-top: 30px;">Как принять приглашение:</h3>
			<ol style="font-size: 15px; line-height: 1.8;">
				<li>Убедитесь, что вы зарегистрированы и вошли в систему</li>
				<li>Если у вас нет аккаунта, зарегистрируйтесь с этим email адресом</li>
				<li>После входа в систему найдите приглашение в разделе "Мои приглашения"</li>
				<li>Нажмите кнопку "Принять" чтобы присоединиться к организации</li>
			</ol>
			
			%s
			
			<hr style="border: none; border-top: 1px solid #e0e0e0; margin: 30px 0;">
			
			<p style="font-size: 12px; color: #999;">
				Это письмо отправлено автоматически. 
				Если вы не запрашивали это приглашение, просто проигнорируйте письмо.
			</p>
			
			<p style="font-size: 12px; color: #999;">
				Если у вас есть вопросы, свяжитесь с администратором организации.
			</p>
		</body>
		</html>
	`, positionName, orgName, linkHTML)

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
